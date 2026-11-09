package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
	"github.com/SoufianoDev/OhMyFreeAI/migrations"
)

func TestOpenMigratesAndPersistsConfig(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "state", "ohmyfreeai.db")
	store, err := Open(ctx, databasePath)
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("apply production migrations: %v", err)
	}
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("reapply migrations should be a no-op: %v", err)
	}

	var version int
	if err := store.DB().QueryRowContext(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		t.Fatalf("read applied migration version: %v", err)
	}
	if version != 1 {
		t.Fatalf("unexpected migration version: got %d, want 1", version)
	}

	defaults, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("load default configuration: %v", err)
	}
	if !reflect.DeepEqual(defaults, config.Defaults()) {
		t.Fatalf("unexpected default configuration: %#v", defaults)
	}

	expected := config.Defaults()
	expected.UI.Language = "en"
	expected.UI.Theme = "dark"
	expected.Gateway.Port = 18082
	if err := store.Save(ctx, expected); err != nil {
		t.Fatalf("save configuration: %v", err)
	}
	actual, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("reload configuration: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("configuration round-trip mismatch:\n got: %#v\nwant: %#v", actual, expected)
	}

	// Hold four connections at once so the test verifies connection-scoped
	// PRAGMAs on every newly opened pooled connection.
	connections := make([]*sql.Conn, 0, maxOpenConns)
	for i := 0; i < maxOpenConns; i++ {
		conn, err := store.DB().Conn(ctx)
		if err != nil {
			t.Fatalf("acquire connection %d: %v", i, err)
		}
		connections = append(connections, conn)

		var foreignKeys, busyTimeout, synchronous int
		if err := conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
			t.Fatalf("read foreign_keys on connection %d: %v", i, err)
		}
		if foreignKeys != 1 {
			t.Fatalf("foreign_keys disabled on connection %d", i)
		}
		if err := conn.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
			t.Fatalf("read busy_timeout on connection %d: %v", i, err)
		}
		if busyTimeout != busyTimeoutMS {
			t.Fatalf("unexpected busy_timeout on connection %d: got %d, want %d", i, busyTimeout, busyTimeoutMS)
		}
		if err := conn.QueryRowContext(ctx, "PRAGMA synchronous").Scan(&synchronous); err != nil {
			t.Fatalf("read synchronous on connection %d: %v", i, err)
		}
		if synchronous != 2 {
			t.Fatalf("synchronous is %d on connection %d, want FULL (2)", synchronous, i)
		}
	}
	for _, conn := range connections {
		if err := conn.Close(); err != nil {
			t.Errorf("release SQLite connection: %v", err)
		}
	}

	var journalMode string
	if err := store.DB().QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("read journal mode: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("unexpected journal mode: %q", journalMode)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(databasePath)
		if err != nil {
			t.Fatalf("stat SQLite database: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("database file permissions are %04o, want 0600", got)
		}
	}
}

func TestMigrationFailureRollsBackWholeMigration(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "rollback.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()

	broken := fstest.MapFS{
		"001_broken.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE transient (id INTEGER PRIMARY KEY); INSERT INTO missing_table(id) VALUES (1);",
		)},
	}
	if err := store.Migrate(ctx, broken, "."); err == nil {
		t.Fatal("expected broken migration to fail")
	}

	var transientCount, historyCount int
	if err := store.DB().QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='transient'").Scan(&transientCount); err != nil {
		t.Fatalf("check rolled-back table: %v", err)
	}
	if transientCount != 0 {
		t.Fatal("failed migration left a partially created table")
	}
	if err := store.DB().QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&historyCount); err != nil {
		t.Fatalf("check migration history rollback: %v", err)
	}
	if historyCount != 0 {
		t.Fatal("failed migration left a migration history table")
	}
}

func TestMigrationGapIsRejected(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "gap.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()

	gapped := fstest.MapFS{
		"002_later.sql": &fstest.MapFile{Data: []byte("CREATE TABLE later (id INTEGER PRIMARY KEY);")},
	}
	if err := store.Migrate(ctx, gapped, "."); !errors.Is(err, ErrMigrationGap) {
		t.Fatalf("expected migration gap error, got %v", err)
	}
}

func TestSaveRejectsInvalidConfiguration(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "invalid-config.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("apply production migrations: %v", err)
	}

	cfg := config.Defaults()
	cfg.Gateway.Host = "0.0.0.0"
	if err := store.Save(ctx, cfg); !errors.Is(err, config.ErrInvalidConfig) {
		t.Fatalf("expected invalid config error, got %v", err)
	}
}

func TestOpenRejectsInMemoryPath(t *testing.T) {
	if _, err := Open(context.Background(), ":memory:"); !errors.Is(err, ErrInvalidDatabasePath) {
		t.Fatalf("expected invalid database path error, got %v", err)
	}
}


func TestMigrationHistoryGapIsRejected(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "history-gap.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()

	_, err = store.DB().ExecContext(ctx,
		"CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, description TEXT NOT NULL, applied_at TEXT NOT NULL)")
	if err != nil {
		t.Fatalf("create migration history fixture: %v", err)
	}
	_, err = store.DB().ExecContext(ctx,
		"INSERT INTO schema_migrations(version, description, applied_at) VALUES (1, 'first', '2026-11-07T00:00:00Z'), (3, 'third', '2026-11-07T00:00:00Z')")
	if err != nil {
		t.Fatalf("insert corrupt migration history fixture: %v", err)
	}

	if err := store.Migrate(ctx, migrations.Files, "."); !errors.Is(err, ErrMigrationHistoryCorrupt) {
		t.Fatalf("expected corrupt migration history error, got %v", err)
	}
}


func TestBackupToProducesVerifiedSnapshot(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open source SQLite store: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate source SQLite store: %v", err)
	}

	expected := config.Defaults()
	expected.Gateway.Port = 18082
	expected.UI.Language = "en"
	if err := store.Save(ctx, expected); err != nil {
		t.Fatalf("save source configuration: %v", err)
	}

	backupPath := filepath.Join(t.TempDir(), "state-snapshot.sqlite")
	if err := store.BackupTo(ctx, backupPath); err != nil {
		t.Fatalf("create online SQLite backup: %v", err)
	}
	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("stat SQLite backup: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("SQLite backup is empty")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("SQLite backup permissions = %o, want 600", info.Mode().Perm())
	}

	backupStore, err := Open(ctx, backupPath)
	if err != nil {
		t.Fatalf("open SQLite backup: %v", err)
	}
	defer backupStore.Close()
	if err := backupStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("verify backup migration history: %v", err)
	}
	actual, err := backupStore.Load(ctx)
	if err != nil {
		t.Fatalf("load configuration from backup: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("backup configuration mismatch:\n got: %#v\nwant: %#v", actual, expected)
	}
}

func TestBackupToDoesNotOverwriteExistingDestination(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate SQLite store: %v", err)
	}

	backupPath := filepath.Join(t.TempDir(), "existing.sqlite")
	original := []byte("preserve this file")
	if err := os.WriteFile(backupPath, original, 0o600); err != nil {
		t.Fatalf("create existing destination: %v", err)
	}
	if err := store.BackupTo(ctx, backupPath); !errors.Is(err, ErrBackupDestinationExists) {
		t.Fatalf("expected existing-destination error, got %v", err)
	}
	actual, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read existing destination: %v", err)
	}
	if !reflect.DeepEqual(actual, original) {
		t.Fatalf("existing destination changed: %q", actual)
	}
}

func TestBackupToHonorsCanceledContext(t *testing.T) {
	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	backupPath := filepath.Join(t.TempDir(), "canceled.sqlite")
	if err := store.BackupTo(ctx, backupPath); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if _, err := os.Stat(backupPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("canceled backup should not publish a destination, stat error: %v", err)
	}
}


func TestOpenRestrictsExistingDatabasePermissions(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "application-data")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatalf("create permissive database directory: %v", err)
	}
	databasePath := filepath.Join(directory, "state.db")
	if err := os.WriteFile(databasePath, []byte{}, 0o644); err != nil {
		t.Fatalf("create permissive database file: %v", err)
	}

	store, err := Open(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("open existing SQLite database: %v", err)
	}
	defer store.Close()

	if runtime.GOOS != "windows" {
		directoryInfo, err := os.Stat(directory)
		if err != nil {
			t.Fatalf("stat application data directory: %v", err)
		}
		if got := directoryInfo.Mode().Perm(); got != 0o700 {
			t.Fatalf("application data directory permissions are %04o, want 0700", got)
		}
		fileInfo, err := os.Stat(databasePath)
		if err != nil {
			t.Fatalf("stat SQLite database file: %v", err)
		}
		if got := fileInfo.Mode().Perm(); got != 0o600 {
			t.Fatalf("SQLite database file permissions are %04o, want 0600", got)
		}
	}
}


func TestMigrationRejectsDatabaseNewerThanAvailableMigrations(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "newer-schema.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()

	_, err = store.DB().ExecContext(ctx,
		"CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, description TEXT NOT NULL, applied_at TEXT NOT NULL)")
	if err != nil {
		t.Fatalf("create migration history fixture: %v", err)
	}
	_, err = store.DB().ExecContext(ctx,
		"INSERT INTO schema_migrations(version, description, applied_at) VALUES (1, 'first', '2026-11-07T00:00:00Z'), (2, 'second', '2026-11-07T00:00:00Z')")
	if err != nil {
		t.Fatalf("insert newer migration history fixture: %v", err)
	}

	available := fstest.MapFS{
		"001_initial.sql": &fstest.MapFile{Data: []byte("CREATE TABLE should_not_run (id INTEGER PRIMARY KEY);")},
	}
	if err := store.Migrate(ctx, available, "."); !errors.Is(err, ErrMigrationVersionTooNew) {
		t.Fatalf("expected newer-schema error, got %v", err)
	}
}


func TestRestoreDatabaseRestoresSnapshotAndPreservesRollback(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "source.db")
	sourceStore, err := Open(ctx, sourcePath)
	if err != nil {
		t.Fatalf("open source database: %v", err)
	}
	if err := sourceStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate source database: %v", err)
	}
	sourceConfig := config.Defaults()
	sourceConfig.Gateway.Port = 18082
	if err := sourceStore.Save(ctx, sourceConfig); err != nil {
		t.Fatalf("save source configuration: %v", err)
	}

	backupPath := filepath.Join(directory, "source-snapshot.sqlite")
	if err := sourceStore.BackupTo(ctx, backupPath); err != nil {
		t.Fatalf("create source snapshot: %v", err)
	}
	if err := sourceStore.Close(); err != nil {
		t.Fatalf("close source database: %v", err)
	}

	targetPath := filepath.Join(directory, "target.db")
	targetStore, err := Open(ctx, targetPath)
	if err != nil {
		t.Fatalf("open target database: %v", err)
	}
	if err := targetStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate target database: %v", err)
	}
	targetConfig := config.Defaults()
	targetConfig.Gateway.Port = 18083
	if err := targetStore.Save(ctx, targetConfig); err != nil {
		t.Fatalf("save target configuration: %v", err)
	}
	if err := targetStore.Close(); err != nil {
		t.Fatalf("close target database: %v", err)
	}

	rollbackPath, err := RestoreDatabase(ctx, backupPath, targetPath, migrations.Files, ".")
	if err != nil {
		t.Fatalf("restore SQLite snapshot: %v", err)
	}
	if rollbackPath == "" {
		t.Fatal("restore did not preserve the previous database")
	}
	if _, err := os.Stat(rollbackPath); err != nil {
		t.Fatalf("stat preserved rollback database: %v", err)
	}

	restoredStore, err := Open(ctx, targetPath)
	if err != nil {
		t.Fatalf("open restored database: %v", err)
	}
	if err := restoredStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("verify restored database migrations: %v", err)
	}
	restoredConfig, err := restoredStore.Load(ctx)
	if err != nil {
		t.Fatalf("load restored configuration: %v", err)
	}
	if !reflect.DeepEqual(restoredConfig, sourceConfig) {
		t.Fatalf("restored configuration mismatch:\n got: %#v\nwant: %#v", restoredConfig, sourceConfig)
	}
	if err := restoredStore.Close(); err != nil {
		t.Fatalf("close restored database: %v", err)
	}

	previousStore, err := Open(ctx, rollbackPath)
	if err != nil {
		t.Fatalf("open preserved rollback database: %v", err)
	}
	defer previousStore.Close()
	if err := previousStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("verify rollback database migrations: %v", err)
	}
	previousConfig, err := previousStore.Load(ctx)
	if err != nil {
		t.Fatalf("load preserved configuration: %v", err)
	}
	if !reflect.DeepEqual(previousConfig, targetConfig) {
		t.Fatalf("rollback configuration mismatch:\n got: %#v\nwant: %#v", previousConfig, targetConfig)
	}
}

func TestRestoreDatabaseRejectsActiveTarget(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	sourceStore, err := Open(ctx, filepath.Join(directory, "source.db"))
	if err != nil {
		t.Fatalf("open source database: %v", err)
	}
	if err := sourceStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate source database: %v", err)
	}
	backupPath := filepath.Join(directory, "source-snapshot.sqlite")
	if err := sourceStore.BackupTo(ctx, backupPath); err != nil {
		t.Fatalf("create source snapshot: %v", err)
	}
	if err := sourceStore.Close(); err != nil {
		t.Fatalf("close source database: %v", err)
	}

	targetPath := filepath.Join(directory, "active.db")
	store, err := Open(ctx, targetPath)
	if err != nil {
		t.Fatalf("open active target database: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate active target database: %v", err)
	}
	if _, err := store.DB().ExecContext(ctx, "CREATE TABLE active_restore_fixture (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("write active target database: %v", err)
	}

	if _, err := RestoreDatabase(ctx, backupPath, targetPath, migrations.Files, "."); !errors.Is(err, ErrRestoreTargetBusy) {
		t.Fatalf("expected active-target error, got %v", err)
	}
}


func TestRestoreDatabaseRefusesNonSQLiteTarget(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	sourceStore, err := Open(ctx, filepath.Join(directory, "source.db"))
	if err != nil {
		t.Fatalf("open source database: %v", err)
	}
	if err := sourceStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate source database: %v", err)
	}
	backupPath := filepath.Join(directory, "source-snapshot.sqlite")
	if err := sourceStore.BackupTo(ctx, backupPath); err != nil {
		t.Fatalf("create source snapshot: %v", err)
	}
	if err := sourceStore.Close(); err != nil {
		t.Fatalf("close source database: %v", err)
	}

	targetPath := filepath.Join(directory, "not-a-database")
	original := []byte("preserve unrelated user file")
	if err := os.WriteFile(targetPath, original, 0o600); err != nil {
		t.Fatalf("create unrelated target file: %v", err)
	}
	if _, err := RestoreDatabase(ctx, backupPath, targetPath, migrations.Files, "."); !errors.Is(err, ErrRestoreTargetInvalid) {
		t.Fatalf("expected invalid-target error, got %v", err)
	}
	actual, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read unrelated target file: %v", err)
	}
	if !reflect.DeepEqual(actual, original) {
		t.Fatalf("unrelated target file changed: %q", actual)
	}
}


func TestRestoreDatabaseRejectsCorruptSourceAndPreservesTarget(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	targetPath := filepath.Join(directory, "target.db")
	targetStore, err := Open(ctx, targetPath)
	if err != nil {
		t.Fatalf("open target database: %v", err)
	}
	if err := targetStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate target database: %v", err)
	}
	expected := config.Defaults()
	expected.Gateway.Port = 18084
	if err := targetStore.Save(ctx, expected); err != nil {
		t.Fatalf("save target configuration: %v", err)
	}
	if err := targetStore.Close(); err != nil {
		t.Fatalf("close target database: %v", err)
	}

	corruptBackup := filepath.Join(directory, "corrupt.sqlite")
	if err := os.WriteFile(corruptBackup, []byte("not a SQLite backup"), 0o600); err != nil {
		t.Fatalf("write corrupt backup fixture: %v", err)
	}
	if _, err := RestoreDatabase(ctx, corruptBackup, targetPath, migrations.Files, "."); !errors.Is(err, ErrRestoreSourceInvalid) {
		t.Fatalf("expected invalid-source error, got %v", err)
	}

	verifiedStore, err := Open(ctx, targetPath)
	if err != nil {
		t.Fatalf("reopen unchanged target database: %v", err)
	}
	defer verifiedStore.Close()
	actual, err := verifiedStore.Load(ctx)
	if err != nil {
		t.Fatalf("load unchanged target configuration: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("target changed after rejected restore:\n got: %#v\nwant: %#v", actual, expected)
	}
}


func TestOpenExistingRejectsMissingDatabaseWithoutCreatingIt(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "missing", "state.db")
	if _, err := OpenExisting(context.Background(), databasePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("OpenExisting error = %v, want os.ErrNotExist", err)
	}
	if _, err := os.Stat(databasePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("OpenExisting created missing database, stat error = %v", err)
	}
	if _, err := os.Stat(filepath.Dir(databasePath)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("OpenExisting created missing directory, stat error = %v", err)
	}
}

func TestOpenExistingOpensRegularDatabaseWithoutChangingFilePermissions(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "existing.db")
	if err := os.WriteFile(databasePath, nil, 0o640); err != nil {
		t.Fatalf("create database fixture: %v", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(databasePath, 0o640); err != nil {
			t.Fatalf("set database fixture permissions: %v", err)
		}
	}
	store, err := OpenExisting(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}
	defer store.Close()
	if store.DatabasePath() != databasePath {
		t.Fatalf("database path = %q, want %q", store.DatabasePath(), databasePath)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(databasePath)
		if err != nil {
			t.Fatalf("stat database fixture: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o640 {
			t.Fatalf("OpenExisting changed database permissions to %04o, want 0640", got)
		}
	}
}

func TestRestoreDatabasePreservesExistingTargetDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX directory permission bits are not portable to Windows")
	}

	ctx := context.Background()
	root := t.TempDir()
	sourceStore, err := Open(ctx, filepath.Join(root, "source.db"))
	if err != nil {
		t.Fatalf("open source database: %v", err)
	}
	if err := sourceStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate source database: %v", err)
	}
	backupPath := filepath.Join(root, "source-snapshot.sqlite")
	if err := sourceStore.BackupTo(ctx, backupPath); err != nil {
		t.Fatalf("create source snapshot: %v", err)
	}
	if err := sourceStore.Close(); err != nil {
		t.Fatalf("close source database: %v", err)
	}

	targetDirectory := filepath.Join(root, "existing-state-directory")
	if err := os.Mkdir(targetDirectory, 0o750); err != nil {
		t.Fatalf("create target directory: %v", err)
	}
	if err := os.Chmod(targetDirectory, 0o750); err != nil {
		t.Fatalf("set target directory permissions: %v", err)
	}

	if _, err := RestoreDatabase(ctx, backupPath, filepath.Join(targetDirectory, "state.db"), migrations.Files, "."); err != nil {
		t.Fatalf("restore SQLite snapshot: %v", err)
	}
	info, err := os.Stat(targetDirectory)
	if err != nil {
		t.Fatalf("stat target directory: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o750 {
		t.Fatalf("restore changed existing directory permissions to %04o, want 0750", got)
	}
}


// These tests exercise SQLite's in-place online-backup restore primitive as
// an architectural alternative to replacing the live database filename. They
// do not enable the application Restore lifecycle or its recovery executor.
func TestRestoreFromPathUpdatesExistingDatabaseInPlace(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()

	sourceStore, err := Open(ctx, filepath.Join(directory, "source.db"))
	if err != nil {
		t.Fatalf("open source database: %v", err)
	}
	if err := sourceStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate source database: %v", err)
	}
	sourceConfig := config.Defaults()
	sourceConfig.Gateway.Port = 18082
	if err := sourceStore.Save(ctx, sourceConfig); err != nil {
		t.Fatalf("save source configuration: %v", err)
	}
	backupPath := filepath.Join(directory, "source-snapshot.sqlite")
	if err := sourceStore.BackupTo(ctx, backupPath); err != nil {
		t.Fatalf("create source snapshot: %v", err)
	}
	if err := sourceStore.Close(); err != nil {
		t.Fatalf("close source database: %v", err)
	}

	targetPath := filepath.Join(directory, "target.db")
	targetStore, err := Open(ctx, targetPath)
	if err != nil {
		t.Fatalf("open target database: %v", err)
	}
	defer targetStore.Close()
	if err := targetStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate target database: %v", err)
	}
	targetConfig := config.Defaults()
	targetConfig.Gateway.Port = 18083
	if err := targetStore.Save(ctx, targetConfig); err != nil {
		t.Fatalf("save target configuration: %v", err)
	}

	before, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("stat target database before restore: %v", err)
	}
	if err := targetStore.restoreFromPath(ctx, backupPath); err != nil {
		t.Fatalf("restore snapshot into existing database: %v", err)
	}
	after, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("stat target database after restore: %v", err)
	}
	if !os.SameFile(before, after) {
		t.Fatal("in-place SQLite restore unexpectedly replaced the target file identity")
	}
	actual, err := targetStore.Load(ctx)
	if err != nil {
		t.Fatalf("load restored target configuration: %v", err)
	}
	if !reflect.DeepEqual(actual, sourceConfig) {
		t.Fatalf("in-place restore mismatch:\\n got: %#v\\nwant: %#v", actual, sourceConfig)
	}
}

func TestRestoreFromPathFailedCopyPreservesExistingDatabase(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	targetPath := filepath.Join(directory, "target.db")
	targetStore, err := Open(ctx, targetPath)
	if err != nil {
		t.Fatalf("open target database: %v", err)
	}
	defer targetStore.Close()
	if err := targetStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate target database: %v", err)
	}
	expected := config.Defaults()
	expected.Gateway.Port = 18083
	if err := targetStore.Save(ctx, expected); err != nil {
		t.Fatalf("save target configuration: %v", err)
	}

	invalidSource := filepath.Join(directory, "not-a-database.sqlite")
	if err := os.WriteFile(invalidSource, []byte("not a SQLite database"), 0o600); err != nil {
		t.Fatalf("write invalid source fixture: %v", err)
	}
	if err := targetStore.restoreFromPath(ctx, invalidSource); err == nil {
		t.Fatal("expected invalid source restore to fail")
	}
	actual, err := targetStore.Load(ctx)
	if err != nil {
		t.Fatalf("load target after failed restore: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("failed in-place restore changed target:\\n got: %#v\\nwant: %#v", actual, expected)
	}
}


const restoreCrashTestModeEnv = "OHMYFREEAI_RESTORE_CRASH_TEST_MODE"

// TestRestoreCrashSubprocessHelper is invoked only by the parent crash test.
// os.Exit intentionally bypasses SQLite and Go cleanup to simulate process death.
func TestRestoreCrashSubprocessHelper(t *testing.T) {
	mode := os.Getenv(restoreCrashTestModeEnv)
	if mode == "" {
		return
	}
	targetPath := os.Getenv("OHMYFREEAI_RESTORE_CRASH_TARGET")
	sourcePath := os.Getenv("OHMYFREEAI_RESTORE_CRASH_SOURCE")
	if targetPath == "" || sourcePath == "" {
		t.Fatal("crash helper requires target and source paths")
	}
	store, err := Open(context.Background(), targetPath)
	if err != nil {
		t.Fatalf("open target in crash helper: %v", err)
	}
	if err := preparePersistentRestoreJournalForTest(context.Background(), store); err != nil {
		t.Fatalf("prepare persistent rollback journal in crash helper: %v", err)
	}

	switch mode {
	case "partial":
		connection, err := store.db.Conn(context.Background())
		if err != nil {
			t.Fatalf("reserve target connection in crash helper: %v", err)
		}
		err = connection.Raw(func(raw any) error {
			driverConnection, ok := raw.(onlineRestoreConnection)
			if !ok {
				return ErrBackupUnsupported
			}
			backup, err := driverConnection.NewRestore(readOnlyDataSourceName(sourcePath))
			if err != nil {
				return err
			}
			more, stepErr := backup.Step(1)
			if stepErr != nil {
				_ = backup.Finish()
				return stepErr
			}
			if !more {
				_ = backup.Finish()
				return errors.New("crash fixture did not require more than one page")
			}
			os.Exit(0)
			return nil
		})
		t.Fatalf("partial restore crash helper unexpectedly returned: %v", err)
	case "committed":
		if err := store.restoreFromPath(context.Background(), sourcePath); err != nil {
			t.Fatalf("complete restore in crash helper: %v", err)
		}
		os.Exit(0)
	default:
		t.Fatalf("unknown restore crash helper mode %q", mode)
	}
}

func TestRestoreFromPathRecoversAfterProcessDeath(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	sourceStore, err := Open(ctx, filepath.Join(directory, "source.db"))
	if err != nil {
		t.Fatalf("open source database: %v", err)
	}
	if err := sourceStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate source database: %v", err)
	}
	sourceConfig := config.Defaults()
	sourceConfig.Gateway.Port = 18082
	if err := sourceStore.Save(ctx, sourceConfig); err != nil {
		t.Fatalf("save source configuration: %v", err)
	}
	if _, err := sourceStore.DB().ExecContext(ctx, "CREATE TABLE restore_crash_payload (data BLOB NOT NULL)"); err != nil {
		t.Fatalf("create large source payload table: %v", err)
	}
	if _, err := sourceStore.DB().ExecContext(ctx, "INSERT INTO restore_crash_payload(data) VALUES (zeroblob(1048576))"); err != nil {
		t.Fatalf("insert large source payload: %v", err)
	}
	sourcePath := filepath.Join(directory, "source-snapshot.sqlite")
	if err := sourceStore.BackupTo(ctx, sourcePath); err != nil {
		t.Fatalf("create source snapshot: %v", err)
	}
	if err := sourceStore.Close(); err != nil {
		t.Fatalf("close source database: %v", err)
	}

	targetPath := filepath.Join(directory, "target.db")
	targetStore, err := Open(ctx, targetPath)
	if err != nil {
		t.Fatalf("open target database: %v", err)
	}
	if err := targetStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate target database: %v", err)
	}
	previousConfig := config.Defaults()
	previousConfig.Gateway.Port = 18083
	if err := targetStore.Save(ctx, previousConfig); err != nil {
		t.Fatalf("save previous target configuration: %v", err)
	}
	if err := targetStore.Close(); err != nil {
		t.Fatalf("close target before crash test: %v", err)
	}

	runCrashChild := func(mode string) {
		t.Helper()
		command := exec.Command(os.Args[0], "-test.run=^TestRestoreCrashSubprocessHelper$")
		command.Env = append(os.Environ(),
			restoreCrashTestModeEnv+"="+mode,
			"OHMYFREEAI_RESTORE_CRASH_TARGET="+targetPath,
			"OHMYFREEAI_RESTORE_CRASH_SOURCE="+sourcePath,
		)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("crash helper mode %q failed: %v\\n%s", mode, err, output)
		}
	}

	// Kill the child after it has copied one page but before SQLITE_DONE and
	// backup_finish. SQLite must recover the original target on next open.
	runCrashChild("partial")
	recoveredPrevious, err := OpenExisting(ctx, targetPath)
	if err != nil {
		t.Fatalf("reopen target after partial-restore process death: %v", err)
	}
	actualPrevious, err := recoveredPrevious.Load(ctx)
	if err != nil {
		t.Fatalf("load target after partial-restore process death: %v", err)
	}
	if !reflect.DeepEqual(actualPrevious, previousConfig) {
		t.Fatalf("partial restore changed target after process death:\\n got: %#v\\nwant: %#v", actualPrevious, previousConfig)
	}
	if err := recoveredPrevious.Close(); err != nil {
		t.Fatalf("close target after partial-restore recovery: %v", err)
	}

	// Kill the child after the backup API has completed and returned, but
	// before the owning Store is cleanly closed. The committed snapshot must
	// be visible after SQLite performs WAL recovery on the next open.
	runCrashChild("committed")
	recoveredCandidate, err := OpenExisting(ctx, targetPath)
	if err != nil {
		t.Fatalf("reopen target after committed-restore process death: %v", err)
	}
	actualCandidate, err := recoveredCandidate.Load(ctx)
	if err != nil {
		t.Fatalf("load target after committed-restore process death: %v", err)
	}
	if !reflect.DeepEqual(actualCandidate, sourceConfig) {
		t.Fatalf("committed restore missing after process death:\\n got: %#v\\nwant: %#v", actualCandidate, sourceConfig)
	}
	if err := recoveredCandidate.Close(); err != nil {
		t.Fatalf("close target after committed-restore recovery: %v", err)
	}
}


func preparePersistentRestoreJournalForTest(ctx context.Context, store *Store) error {
	store.db.SetMaxOpenConns(1)
	store.db.SetMaxIdleConns(1)
	var journalMode string
	if err := store.db.QueryRowContext(ctx, "PRAGMA journal_mode=PERSIST").Scan(&journalMode); err != nil {
		return fmt.Errorf("set SQLite PERSIST journal mode: %w", err)
	}
	if journalMode != "persist" {
		return fmt.Errorf("SQLite refused PERSIST journal mode; active mode is %q", journalMode)
	}
	current, err := store.Load(ctx)
	if err != nil {
		return fmt.Errorf("load configuration before persistent-journal preparation: %w", err)
	}
	if err := store.Save(ctx, current); err != nil {
		return fmt.Errorf("prepare persistent rollback journal: %w", err)
	}
	if _, err := os.Stat(store.path + "-journal"); err != nil {
		return fmt.Errorf("persistent rollback journal was not retained: %w", err)
	}
	return nil
}

func TestRestoreFromPathPERSISTModeSupportsDifferentSourcePageSize(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "source-1024.db")
	rawSource, err := sql.Open(driverName, backupDataSourceName(sourcePath))
	if err != nil {
		t.Fatalf("open raw source database: %v", err)
	}
	rawSource.SetMaxOpenConns(1)
	if _, err := rawSource.ExecContext(ctx, "PRAGMA page_size=1024"); err != nil {
		t.Fatalf("set source page size: %v", err)
	}
	if _, err := rawSource.ExecContext(ctx, "CREATE TABLE page_size_seed (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("persist source page size: %v", err)
	}
	if err := rawSource.Close(); err != nil {
		t.Fatalf("close raw source database: %v", err)
	}

	sourceStore, err := OpenExisting(ctx, sourcePath)
	if err != nil {
		t.Fatalf("open source database: %v", err)
	}
	if err := sourceStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate source database: %v", err)
	}
	sourceConfig := config.Defaults()
	sourceConfig.Gateway.Port = 18082
	if err := sourceStore.Save(ctx, sourceConfig); err != nil {
		t.Fatalf("save source configuration: %v", err)
	}
	snapshotPath := filepath.Join(directory, "source-1024-snapshot.sqlite")
	if err := sourceStore.BackupTo(ctx, snapshotPath); err != nil {
		t.Fatalf("create source snapshot: %v", err)
	}
	if err := sourceStore.Close(); err != nil {
		t.Fatalf("close source database: %v", err)
	}

	snapshotDB, err := sql.Open(driverName, readOnlyDataSourceName(snapshotPath))
	if err != nil {
		t.Fatalf("open source snapshot for page-size check: %v", err)
	}
	var snapshotPageSize int
	if err := snapshotDB.QueryRowContext(ctx, "PRAGMA page_size").Scan(&snapshotPageSize); err != nil {
		_ = snapshotDB.Close()
		t.Fatalf("read source snapshot page size: %v", err)
	}
	if err := snapshotDB.Close(); err != nil {
		t.Fatalf("close source snapshot page-size check: %v", err)
	}
	if snapshotPageSize != 1024 {
		t.Fatalf("source snapshot page size = %d, want 1024", snapshotPageSize)
	}

	targetPath := filepath.Join(directory, "target-4096.db")
	targetStore, err := Open(ctx, targetPath)
	if err != nil {
		t.Fatalf("open target database: %v", err)
	}
	targetStoreClosed := false
	defer func() {
		if !targetStoreClosed {
			_ = targetStore.Close()
		}
	}()
	if err := targetStore.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate target database: %v", err)
	}
	previousConfig := config.Defaults()
	previousConfig.Gateway.Port = 18083
	if err := targetStore.Save(ctx, previousConfig); err != nil {
		t.Fatalf("save previous target configuration: %v", err)
	}

	// SQLite documents SQLITE_READONLY when a WAL destination and source
	// have different page sizes. The failed attempt must leave the target intact.
	if err := targetStore.restoreFromPath(ctx, snapshotPath); err == nil {
		t.Fatal("WAL-mode restore unexpectedly accepted a different source page size")
	}
	actualPrevious, err := targetStore.Load(ctx)
	if err != nil {
		t.Fatalf("load target after rejected WAL restore: %v", err)
	}
	if !reflect.DeepEqual(actualPrevious, previousConfig) {
		t.Fatalf("rejected WAL restore changed target:\\n got: %#v\\nwant: %#v", actualPrevious, previousConfig)
	}

	// PERSIST uses a rollback journal whose file is retained and invalidated
	// by zeroing its header at commit, avoiding journal-name deletion.
	if err := preparePersistentRestoreJournalForTest(ctx, targetStore); err != nil {
		t.Fatalf("prepare persistent target journal: %v", err)
	}
	before, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("stat target before in-place restore: %v", err)
	}
	if err := targetStore.restoreFromPath(ctx, snapshotPath); err != nil {
		t.Fatalf("restore different-page-size snapshot in PERSIST mode: %v", err)
	}
	after, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("stat target after in-place restore: %v", err)
	}
	if !os.SameFile(before, after) {
		t.Fatal("PERSIST-mode restore unexpectedly replaced the target file identity")
	}
	if _, err := os.Stat(targetPath + "-journal"); err != nil {
		t.Fatalf("PERSIST rollback journal was not retained after restore: %v", err)
	}

	// Reopen before inspecting page_size: the connection that performed the
	// backup may retain its pre-restore B-tree page-size cache.
	if err := targetStore.Close(); err != nil {
		t.Fatalf("close target after in-place restore: %v", err)
	}
	targetStoreClosed = true
	if err := verifyBackupDatabase(ctx, targetPath); err != nil {
		t.Fatalf("verify in-place restored target: %v", err)
	}
	restoredStore, err := OpenExisting(ctx, targetPath)
	if err != nil {
		t.Fatalf("reopen in-place restored target: %v", err)
	}
	defer restoredStore.Close()
	var restoredPageSize int
	if err := restoredStore.DB().QueryRowContext(ctx, "PRAGMA page_size").Scan(&restoredPageSize); err != nil {
		t.Fatalf("read restored target page size after reopen: %v", err)
	}
	if restoredPageSize != 1024 {
		t.Fatalf("restored target page size after reopen = %d, want 1024", restoredPageSize)
	}
	actual, err := restoredStore.Load(ctx)
	if err != nil {
		t.Fatalf("load restored target configuration: %v", err)
	}
	if !reflect.DeepEqual(actual, sourceConfig) {
		t.Fatalf("PERSIST-mode restore mismatch:\\n got: %#v\\nwant: %#v", actual, sourceConfig)
	}
}
