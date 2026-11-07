package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"os"
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

	backupPath := filepath.Join(directory, "placeholder.sqlite")
	if err := os.WriteFile(backupPath, []byte("unused because target is busy"), 0o600); err != nil {
		t.Fatalf("create placeholder source: %v", err)
	}
	if _, err := RestoreDatabase(ctx, backupPath, targetPath, migrations.Files, "."); !errors.Is(err, ErrRestoreTargetBusy) {
		t.Fatalf("expected active-target error, got %v", err)
	}
}
