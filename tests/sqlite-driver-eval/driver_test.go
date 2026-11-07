package eval

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "modernc.org/sqlite"
)

func TestSQLiteDriverEvaluation(t *testing.T) {
	driver := os.Getenv("OMFAI_SQLITE_DRIVER")
	switch driver {
	case "sqlite", "sqlite3":
	default:
		t.Fatalf("OMFAI_SQLITE_DRIVER must be sqlite or sqlite3; got %q", driver)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	dbPath := filepath.Join(t.TempDir(), "evaluation.db")
	db, err := sql.Open(driver, dbPath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	db.SetConnMaxLifetime(0)

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	var sqliteVersion string
	if err := db.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&sqliteVersion); err != nil {
		t.Fatalf("read SQLite version: %v", err)
	}
	t.Logf("driver=%s sqlite=%s", driver, sqliteVersion)

	var journalMode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode=WAL").Scan(&journalMode); err != nil {
		t.Fatalf("enable WAL: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("WAL mode not enabled; got %q", journalMode)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("get connection: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
		conn.Close()
		t.Fatalf("enable foreign keys: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "PRAGMA busy_timeout=10000"); err != nil {
		conn.Close()
		t.Fatalf("set busy timeout: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `
		CREATE TABLE parent (
			id INTEGER PRIMARY KEY
		);
		CREATE TABLE child (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER NOT NULL REFERENCES parent(id)
		);
		CREATE TABLE writes (
			id INTEGER PRIMARY KEY,
			worker INTEGER NOT NULL,
			sequence INTEGER NOT NULL,
			UNIQUE(worker, sequence)
		);
		CREATE TABLE schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		);
		INSERT INTO parent(id) VALUES (1);
	`); err != nil {
		conn.Close()
		t.Fatalf("create migration fixture schema: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "INSERT INTO child(id, parent_id) VALUES (1, 999)"); err == nil {
		conn.Close()
		t.Fatal("foreign-key violation was accepted")
	}
	if _, err := conn.ExecContext(ctx, "INSERT INTO child(id, parent_id) VALUES (2, 1)"); err != nil {
		conn.Close()
		t.Fatalf("valid foreign-key insert failed: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close initial connection: %v", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO parent(id) VALUES (2)"); err != nil {
		tx.Rollback()
		t.Fatalf("transaction insert: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	var rolledBack int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM parent WHERE id=2").Scan(&rolledBack); err != nil {
		t.Fatalf("verify rollback: %v", err)
	}
	if rolledBack != 0 {
		t.Fatalf("rollback did not remove row; count=%d", rolledBack)
	}

	const workers, writesPerWorker = 4, 25
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			c, err := db.Conn(ctx)
			if err != nil {
				errs <- fmt.Errorf("worker %d acquire connection: %w", worker, err)
				return
			}
			defer c.Close()
			if _, err := c.ExecContext(ctx, "PRAGMA busy_timeout=10000"); err != nil {
				errs <- fmt.Errorf("worker %d busy timeout: %w", worker, err)
				return
			}
			if _, err := c.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
				errs <- fmt.Errorf("worker %d foreign keys: %w", worker, err)
				return
			}
			for sequence := 0; sequence < writesPerWorker; sequence++ {
				if _, err := c.ExecContext(ctx,
					"INSERT INTO writes(worker, sequence) VALUES (?, ?)", worker, sequence); err != nil {
					errs <- fmt.Errorf("worker %d write %d: %w", worker, sequence, err)
					return
				}
			}
		}(worker)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if t.Failed() {
		return
	}

	var writeCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM writes").Scan(&writeCount); err != nil {
		t.Fatalf("count concurrent writes: %v", err)
	}
	if writeCount != workers*writesPerWorker {
		t.Fatalf("lost concurrent writes: got %d, want %d", writeCount, workers*writesPerWorker)
	}

	var integrity string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		t.Fatalf("integrity_check: %v", err)
	}
	if integrity != "ok" {
		t.Fatalf("integrity_check returned %q", integrity)
	}

	var busy, logFrames, checkpointed int
	if err := db.QueryRowContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)").Scan(&busy, &logFrames, &checkpointed); err != nil {
		t.Fatalf("checkpoint WAL: %v", err)
	}
	if busy != 0 {
		t.Fatalf("WAL checkpoint remained busy: busy=%d log=%d checkpointed=%d", busy, logFrames, checkpointed)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	backupPath := filepath.Join(t.TempDir(), "evaluation-backup.db")
	data, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("read checkpointed database for backup: %v", err)
	}
	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		t.Fatalf("write database backup: %v", err)
	}

	backup, err := sql.Open(driver, backupPath)
	if err != nil {
		t.Fatalf("open copied backup: %v", err)
	}
	defer backup.Close()
	if err := backup.PingContext(ctx); err != nil {
		t.Fatalf("reopen copied backup: %v", err)
	}
	if err := backup.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		t.Fatalf("backup integrity_check: %v", err)
	}
	if integrity != "ok" {
		t.Fatalf("backup integrity_check returned %q", integrity)
	}
	if err := backup.QueryRowContext(ctx, "SELECT COUNT(*) FROM writes").Scan(&writeCount); err != nil {
		t.Fatalf("read copied backup: %v", err)
	}
	if writeCount != workers*writesPerWorker {
		t.Fatalf("backup lost writes: got %d, want %d", writeCount, workers*writesPerWorker)
	}
}
