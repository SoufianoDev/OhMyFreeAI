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

	_, err = store.DB().ExecContext(ctx, `
		CREATE TABLE schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TEXT NOT NULL
		);
		INSERT INTO schema_migrations(version, description, applied_at)
		VALUES (1, 'first', '2026-11-07T00:00:00Z'), (3, 'third', '2026-11-07T00:00:00Z');
	`)
	if err != nil {
		t.Fatalf("create corrupt migration history fixture: %v", err)
	}

	if err := store.Migrate(ctx, migrations.Files, "."); !errors.Is(err, ErrMigrationHistoryCorrupt) {
		t.Fatalf("expected corrupt migration history error, got %v", err)
	}
}
