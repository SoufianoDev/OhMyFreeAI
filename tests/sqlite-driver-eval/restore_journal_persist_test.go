package eval

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	ncrucessqlite "github.com/ncruces/go-sqlite3"
)

func TestPersistentRollbackJournalLifecycle(t *testing.T) {
	driver := selectedDriver(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbPath := filepath.Join(t.TempDir(), "restore-control.db")
	db, err := sql.Open(driver, dbPath)
	if err != nil {
		t.Fatalf("open restore-control database: %v", err)
	}
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		t.Fatalf("get restore-control connection: %v", err)
	}

	requirePersistentJournalMode(t, ctx, conn)
	requireSynchronousFull(t, ctx, conn)
	commitRestoreControlWrite(t, ctx, conn, true)

	journalPath := dbPath + "-journal"
	firstInfo := requirePersistentJournalFile(t, journalPath)
	requireZeroJournalHeader(t, journalPath)

	if err := conn.Close(); err != nil {
		_ = db.Close()
		t.Fatalf("close first restore-control connection: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close first restore-control database: %v", err)
	}
	afterCloseInfo := requirePersistentJournalFile(t, journalPath)
	if !os.SameFile(firstInfo, afterCloseInfo) {
		t.Fatal("PERSIST sidecar was replaced or removed when the database connection closed")
	}

	// PERSIST is connection-scoped. A fresh connection must explicitly set and
	// verify it again before any journaled transaction is started.
	reopened, err := sql.Open(driver, dbPath)
	if err != nil {
		t.Fatalf("reopen restore-control database: %v", err)
	}
	reopened.SetMaxOpenConns(1)
	reopenedConn, err := reopened.Conn(ctx)
	if err != nil {
		_ = reopened.Close()
		t.Fatalf("get reopened restore-control connection: %v", err)
	}
	var mode string
	if err := reopenedConn.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("read journal mode after reopen: %v", err)
	}
	if strings.EqualFold(mode, "persist") {
		t.Fatal("PERSIST unexpectedly survived close/reopen; keep the explicit per-connection mode check")
	}

	requirePersistentJournalMode(t, ctx, reopenedConn)
	requireSynchronousFull(t, ctx, reopenedConn)
	commitRestoreControlWrite(t, ctx, reopenedConn, false)

	secondInfo := requirePersistentJournalFile(t, journalPath)
	if !os.SameFile(firstInfo, secondInfo) {
		t.Fatal("reopened PERSIST connection did not reuse the pre-provisioned rollback-journal sidecar")
	}
	requireZeroJournalHeader(t, journalPath)

	if err := reopenedConn.Close(); err != nil {
		_ = reopened.Close()
		t.Fatalf("close reopened restore-control connection: %v", err)
	}
	if err := reopened.Close(); err != nil {
		t.Fatalf("close reopened restore-control database: %v", err)
	}
}

func TestSQLiteSilentlyRecreatesMissingRestoreControlArtifacts(t *testing.T) {
	driver := selectedDriver(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// database/sql drivers create a missing file-backed database on first use.
	// A restore-control adapter must therefore preflight its provisioned DB
	// without following links and must not rely on sql.Open/Ping to fail closed.
	testDirectory, err := os.MkdirTemp(".", ".restore-control-mode-rw-*")
	if err != nil {
		t.Fatalf("create same-volume restore-control test directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(testDirectory); err != nil {
			t.Errorf("remove restore-control test directory: %v", err)
		}
	})

	missingDBPath := filepath.Join(testDirectory, "missing-control.db")
	missingDB, err := sql.Open(driver, missingDBPath)
	if err != nil {
		t.Fatalf("open missing database handle: %v", err)
	}
	if err := missingDB.PingContext(ctx); err != nil {
		_ = missingDB.Close()
		t.Fatalf("probe missing database creation behavior: %v", err)
	}
	if err := missingDB.Close(); err != nil {
		t.Fatalf("close newly created database: %v", err)
	}
	if info, err := os.Lstat(missingDBPath); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("driver did not create a regular database file; info=%v err=%v", info, err)
	}

	// Likewise, deleting a PERSIST sidecar does not make SQLite reject the
	// next transaction. The candidate design must detect the missing sidecar
	// before opening the control DB or writing a journal phase.
	dbPath := filepath.Join(t.TempDir(), "provisioned-control.db")
	db, err := sql.Open(driver, dbPath)
	if err != nil {
		t.Fatalf("open provisioned database: %v", err)
	}
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		t.Fatalf("get provisioned connection: %v", err)
	}
	requirePersistentJournalMode(t, ctx, conn)
	requireSynchronousFull(t, ctx, conn)
	commitRestoreControlWrite(t, ctx, conn, true)
	journalPath := dbPath + "-journal"
	_ = requirePersistentJournalFile(t, journalPath)
	if err := conn.Close(); err != nil {
		_ = db.Close()
		t.Fatalf("close provisioned connection: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close provisioned database: %v", err)
	}
	if err := os.Remove(journalPath); err != nil {
		t.Fatalf("remove sidecar to simulate incomplete provisioning: %v", err)
	}

	reopened, err := sql.Open(driver, dbPath)
	if err != nil {
		t.Fatalf("reopen database with missing sidecar: %v", err)
	}
	reopened.SetMaxOpenConns(1)
	reopenedConn, err := reopened.Conn(ctx)
	if err != nil {
		_ = reopened.Close()
		t.Fatalf("get reopened connection with missing sidecar: %v", err)
	}
	requirePersistentJournalMode(t, ctx, reopenedConn)
	requireSynchronousFull(t, ctx, reopenedConn)
	commitRestoreControlWrite(t, ctx, reopenedConn, false)
	_ = requirePersistentJournalFile(t, journalPath)
	if err := reopenedConn.Close(); err != nil {
		_ = reopened.Close()
		t.Fatalf("close recreated-sidecar connection: %v", err)
	}
	if err := reopened.Close(); err != nil {
		t.Fatalf("close recreated-sidecar database: %v", err)
	}
}

func requirePersistentJournalMode(t *testing.T, ctx context.Context, conn *sql.Conn) {
	t.Helper()
	var mode string
	if err := conn.QueryRowContext(ctx, "PRAGMA journal_mode=PERSIST").Scan(&mode); err != nil {
		t.Fatalf("enable PERSIST journal mode: %v", err)
	}
	if !strings.EqualFold(mode, "persist") {
		t.Fatalf("SQLite refused PERSIST journal mode; got %q", mode)
	}
}

func requireSynchronousFull(t *testing.T, ctx context.Context, conn *sql.Conn) {
	t.Helper()
	if _, err := conn.ExecContext(ctx, "PRAGMA synchronous=FULL"); err != nil {
		t.Fatalf("set synchronous=FULL: %v", err)
	}
	var synchronous int
	if err := conn.QueryRowContext(ctx, "PRAGMA synchronous").Scan(&synchronous); err != nil {
		t.Fatalf("read synchronous mode: %v", err)
	}
	if synchronous != 2 {
		t.Fatalf("synchronous mode = %d, want FULL (2)", synchronous)
	}
}

func commitRestoreControlWrite(t *testing.T, ctx context.Context, conn *sql.Conn, createTable bool) {
	t.Helper()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin restore-control transaction: %v", err)
	}
	if createTable {
		if _, err := tx.ExecContext(ctx, "CREATE TABLE journal_events (id INTEGER PRIMARY KEY, phase TEXT NOT NULL)"); err != nil {
			_ = tx.Rollback()
			t.Fatalf("create journal event table: %v", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO journal_events(phase) VALUES (?)", "prepared"); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert journal event: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit restore-control transaction: %v", err)
	}
}

func requirePersistentJournalFile(t *testing.T, path string) os.FileInfo {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("inspect persistent rollback-journal sidecar %s: %v", path, err)
	}
	if !info.Mode().IsRegular() || info.Size() < 8 {
		t.Fatalf("persistent rollback-journal sidecar is not a regular file with a header: mode=%s size=%d", info.Mode(), info.Size())
	}
	return info
}

func requireZeroJournalHeader(t *testing.T, path string) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open persistent rollback-journal header: %v", err)
	}
	defer file.Close()
	header := make([]byte, 8)
	if _, err := io.ReadFull(file, header); err != nil {
		t.Fatalf("read persistent rollback-journal header: %v", err)
	}
	for index, value := range header {
		if value != 0 {
			t.Fatalf("PERSIST journal header byte %d = %#02x, want zero after commit", index, value)
		}
	}
}

func TestSQLiteExistingOnlyOpenRequiresExistingDatabase(t *testing.T) {
	driver := selectedDriver(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	missingPath := filepath.Join(t.TempDir(), "missing-control.db")
	if err := openExistingOnlySQLiteDatabase(ctx, t, driver, missingPath); err == nil {
		t.Fatal("existing-only open accepted a missing restore-control database")
	}
	if _, err := os.Lstat(missingPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("existing-only open created a missing database: lstat error = %v", err)
	}

	existingPath := filepath.Join(t.TempDir(), "existing-control.db")
	db, err := sql.Open(driver, existingPath)
	if err != nil {
		t.Fatalf("open database for existing-only probe: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("provision database for existing-only probe: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close provisioned database: %v", err)
	}
	if err := openExistingOnlySQLiteDatabase(ctx, t, driver, existingPath); err != nil {
		t.Fatalf("existing-only open rejected an existing database: %v", err)
	}
}

func openExistingOnlySQLiteDatabase(ctx context.Context, t *testing.T, driver, path string) error {
	t.Helper()
	if driver == "sqlite3" && runtime.GOOS == "windows" {
		// The pinned ncruces Windows VFS mishandles absolute file:///C:/ URIs.
		// Its direct OpenFlags API supports an existing-only native-path open.
		conn, err := ncrucessqlite.OpenFlags(path, ncrucessqlite.OPEN_READWRITE)
		if err != nil {
			return err
		}
		return conn.Close()
	}

	db, err := sql.Open(driver, restoreControlSQLiteURI(t, path))
	if err != nil {
		return err
	}
	defer db.Close()
	return db.PingContext(ctx)
}

func restoreControlSQLiteURI(t *testing.T, path string) string {
	t.Helper()
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("resolve SQLite URI path: %v", err)
	}
	slashPath := filepath.ToSlash(absolute)
	if runtime.GOOS == "windows" && !strings.HasPrefix(slashPath, "/") {
		slashPath = "/" + slashPath
	}
	uri := url.URL{Scheme: "file", Path: slashPath}
	query := url.Values{}
	query.Set("mode", "rw")
	uri.RawQuery = query.Encode()
	return uri.String()
}


func TestPersistentSidecarSymlinkProbe(t *testing.T) {
	driver := selectedDriver(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbPath := filepath.Join(t.TempDir(), "restore-control.db")
	db, err := sql.Open(driver, dbPath)
	if err != nil {
		t.Fatalf("open restore-control database: %v", err)
	}
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		t.Fatalf("get restore-control connection: %v", err)
	}
	requirePersistentJournalMode(t, ctx, conn)
	requireSynchronousFull(t, ctx, conn)
	commitRestoreControlWrite(t, ctx, conn, true)
	if err := conn.Close(); err != nil {
		_ = db.Close()
		t.Fatalf("close restore-control connection: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close restore-control database: %v", err)
	}

	journalPath := dbPath + "-journal"
	if err := os.Remove(journalPath); err != nil {
		t.Fatalf("remove provisioned PERSIST sidecar: %v", err)
	}
	targetPath := filepath.Join(t.TempDir(), "outside-journal-target")
	sentinel := []byte("unrelated file that must not become a SQLite rollback journal")
	if err := os.WriteFile(targetPath, sentinel, 0o600); err != nil {
		t.Fatalf("write outside journal target: %v", err)
	}
	if err := os.Symlink(targetPath, journalPath); err != nil {
		if runtime.GOOS == "windows" && errors.Is(err, os.ErrPermission) {
			t.Skip("creating a Windows symbolic link requires an unavailable privilege")
		}
		t.Fatalf("create PERSIST sidecar symlink: %v", err)
	}

	reopened, openErr := sql.Open(driver, dbPath)
	var connectionErr, modeErr, syncErr, writeErr, connectionCloseErr, databaseCloseErr error
	var mode string
	if openErr == nil {
		reopened.SetMaxOpenConns(1)
		var reopenedConn *sql.Conn
		reopenedConn, connectionErr = reopened.Conn(ctx)
		if connectionErr == nil {
			modeErr = reopenedConn.QueryRowContext(ctx, "PRAGMA journal_mode=PERSIST").Scan(&mode)
			if modeErr == nil && strings.EqualFold(mode, "persist") {
				_, syncErr = reopenedConn.ExecContext(ctx, "PRAGMA synchronous=FULL")
				if syncErr == nil {
					_, writeErr = reopenedConn.ExecContext(ctx, "INSERT INTO journal_events(phase) VALUES (?)", "symlink-probe")
				}
			}
			connectionCloseErr = reopenedConn.Close()
		}
		databaseCloseErr = reopened.Close()
	}

	after, readErr := os.ReadFile(targetPath)
	targetChanged := readErr != nil || !bytes.Equal(after, sentinel)
	linkInfo, linkErr := os.Lstat(journalPath)
	symlinkRetained := linkErr == nil && linkInfo.Mode()&os.ModeSymlink != 0
	t.Logf("PERSIST sidecar symlink probe: driver=%s open_error=%v connection_error=%v mode=%q mode_error=%v sync_error=%v write_error=%v close_errors=%v/%v target_changed=%t symlink_retained=%t link_error=%v target_read_error=%v",
		driver, openErr, connectionErr, mode, modeErr, syncErr, writeErr, connectionCloseErr, databaseCloseErr, targetChanged, symlinkRetained, linkErr, readErr)
	switch {
	case targetChanged:
		t.Log("PERSIST sidecar path followed the symbolic link; this driver cannot satisfy the no-follow sidecar requirement")
	case symlinkRetained && (connectionErr != nil || modeErr != nil || syncErr != nil || writeErr != nil):
		t.Log("PERSIST sidecar symbolic link was retained and the operation was rejected")
	case !symlinkRetained && !targetChanged:
		t.Log("PERSIST sidecar symbolic link was replaced or removed without modifying its target")
	default:
		t.Log("PERSIST sidecar symbolic-link behavior is inconclusive and requires further inspection")
	}
}
