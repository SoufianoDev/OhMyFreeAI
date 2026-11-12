package eval

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	persistCrashChildEnv = "OMFAI_PERSIST_CRASH_CHILD"
	persistCrashPathEnv  = "OMFAI_PERSIST_CRASH_DB_PATH"
)

func TestPersistentJournalRecoversAfterAbruptProcessDeath(t *testing.T) {
	driver := selectedDriver(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	dbPath := filepath.Join(t.TempDir(), "restore-control-crash.db")
	db, err := sql.Open(driver, dbPath)
	if err != nil {
		t.Fatalf("open crash-test database: %v", err)
	}
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		t.Fatalf("get crash-test connection: %v", err)
	}
	requirePersistentJournalMode(t, ctx, conn)
	requireSynchronousFull(t, ctx, conn)
	if _, err := conn.ExecContext(ctx, "CREATE TABLE journal_events (id INTEGER PRIMARY KEY, phase TEXT NOT NULL, payload BLOB NOT NULL)"); err != nil {
		_ = conn.Close()
		_ = db.Close()
		t.Fatalf("create crash-test journal table: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "INSERT INTO journal_events(phase, payload) VALUES ('committed-before-crash', X'01')"); err != nil {
		_ = conn.Close()
		_ = db.Close()
		t.Fatalf("write committed baseline event: %v", err)
	}
	if err := conn.Close(); err != nil {
		_ = db.Close()
		t.Fatalf("close baseline connection: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close baseline database: %v", err)
	}

	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPersistentJournalCrashHelper$")
	command.Env = append(os.Environ(),
		"OMFAI_SQLITE_DRIVER="+driver,
		persistCrashChildEnv+"=1",
		persistCrashPathEnv+"="+dbPath,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("crash child failed before forced process exit: %v\n%s", err, output)
	}

	reopened, err := sql.Open(driver, dbPath)
	if err != nil {
		t.Fatalf("reopen database after process death: %v", err)
	}
	reopened.SetMaxOpenConns(1)
	reopenedConn, err := reopened.Conn(ctx)
	if err != nil {
		_ = reopened.Close()
		t.Fatalf("get recovery connection: %v", err)
	}
	defer reopenedConn.Close()
	defer reopened.Close()

	var committedCount, interruptedCount int
	if err := reopenedConn.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM journal_events WHERE phase='committed-before-crash'").Scan(&committedCount); err != nil {
		t.Fatalf("read committed baseline after recovery: %v", err)
	}
	if committedCount != 1 {
		t.Fatalf("committed baseline count after recovery = %d, want 1", committedCount)
	}
	if err := reopenedConn.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM journal_events WHERE phase='uncommitted-before-crash'").Scan(&interruptedCount); err != nil {
		t.Fatalf("read interrupted rows after recovery: %v", err)
	}
	if interruptedCount != 0 {
		t.Fatalf("SQLite retained %d uncommitted rows after abrupt process death", interruptedCount)
	}
	var integrity string
	if err := reopenedConn.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		t.Fatalf("run integrity_check after crash recovery: %v", err)
	}
	if integrity != "ok" {
		t.Fatalf("integrity_check after crash recovery returned %q", integrity)
	}
}

func TestPersistentJournalCrashHelper(t *testing.T) {
	if os.Getenv(persistCrashChildEnv) != "1" {
		t.Skip("subprocess-only crash helper")
	}
	driver := selectedDriver(t)
	dbPath := os.Getenv(persistCrashPathEnv)
	if dbPath == "" {
		t.Fatal("crash child database path is missing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	db, err := sql.Open(driver, dbPath)
	if err != nil {
		t.Fatalf("open child database: %v", err)
	}
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("get child database connection: %v", err)
	}
	requirePersistentJournalMode(t, ctx, conn)
	requireSynchronousFull(t, ctx, conn)
	if _, err := conn.ExecContext(ctx, "PRAGMA cache_size=1"); err != nil {
		t.Fatalf("set child cache_size: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "PRAGMA cache_spill=ON"); err != nil {
		t.Fatalf("enable child cache spill: %v", err)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin interrupted transaction: %v", err)
	}
	statement, err := tx.PrepareContext(ctx, "INSERT INTO journal_events(phase, payload) VALUES (?, ?)")
	if err != nil {
		t.Fatalf("prepare interrupted writes: %v", err)
	}
	payload := strings.Repeat("x", 8192)
	for index := 0; index < 256; index++ {
		if _, err := statement.ExecContext(ctx, "uncommitted-before-crash", payload); err != nil {
			t.Fatalf("write interrupted row %d: %v", index, err)
		}
	}
	if err := statement.Close(); err != nil {
		t.Fatalf("close interrupted write statement: %v", err)
	}

	journalPath := dbPath + "-journal"
	info, err := os.Stat(journalPath)
	if err != nil {
		t.Fatalf("inspect hot rollback journal before forced process death: %v", err)
	}
	if !info.Mode().IsRegular() || info.Size() <= 8 {
		t.Fatalf("rollback journal is not a non-empty regular file before process death: mode=%s size=%d", info.Mode(), info.Size())
	}
	file, err := os.Open(journalPath)
	if err != nil {
		t.Fatalf("open hot rollback journal header: %v", err)
	}
	header := make([]byte, 8)
	_, readErr := file.Read(header)
	closeErr := file.Close()
	if readErr != nil {
		t.Fatalf("read hot rollback journal header: %v", readErr)
	}
	if closeErr != nil {
		t.Fatalf("close hot rollback journal header: %v", closeErr)
	}
	allZero := true
	for _, value := range header {
		if value != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Fatal("rollback journal header is zero before the transaction commits; crash fixture is not active")
	}

	// Deliberately bypass database/sql and SQLite cleanup. The parent process
	// reopens the same file and verifies SQLite rolls the interrupted write back.
	os.Exit(0)
}
