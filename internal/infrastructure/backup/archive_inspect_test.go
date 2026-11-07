package backup

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/SoufianoDev/OhMyFreeAI/internal/infrastructure/sqlite"
	"github.com/SoufianoDev/OhMyFreeAI/migrations"
)

func TestManagerInspectReturnsValidatedManifestWithoutChangingLiveDatabase(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	inspectionTempRoot := filepath.Join(root, "inspection-tmp")
	if err := os.Mkdir(inspectionTempRoot, 0o700); err != nil {
		t.Fatalf("create isolated inspection temp root: %v", err)
	}
	t.Setenv("TMPDIR", inspectionTempRoot)
	t.Setenv("TMP", inspectionTempRoot)
	t.Setenv("TEMP", inspectionTempRoot)
	databasePath := filepath.Join(root, "state.db")
	store, err := sqlite.Open(ctx, databasePath)
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate SQLite store: %v", err)
	}
	manager := Manager{ApplicationVersion: "0.1.0", Migrations: migrations.Files, MigrationDir: "."}
	archivePath := filepath.Join(root, "valid.omfaibackup")
	if err := manager.Create(ctx, store, archivePath); err != nil {
		t.Fatalf("create backup: %v", err)
	}
	before, err := os.ReadFile(databasePath)
	if err != nil {
		t.Fatalf("read live database before inspect: %v", err)
	}

	manifest, err := manager.Inspect(ctx, archivePath)
	if err != nil {
		t.Fatalf("inspect valid backup: %v", err)
	}
	if manifest.Format != archiveFormat || manifest.FormatVersion != archiveVersion {
		t.Fatalf("unexpected manifest format: %#v", manifest)
	}
	if manifest.ApplicationVersion != "0.1.0" || manifest.SchemaVersion != 1 {
		t.Fatalf("unexpected manifest metadata: %#v", manifest)
	}
	if len(manifest.Files) != 1 || manifest.Files[0].Path != databaseName {
		t.Fatalf("unexpected manifest file records: %#v", manifest.Files)
	}

	after, err := os.ReadFile(databasePath)
	if err != nil {
		t.Fatalf("read live database after inspect: %v", err)
	}
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("inspect modified the live database")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close SQLite store: %v", err)
	}
	remaining, err := os.ReadDir(inspectionTempRoot)
	if err != nil {
		t.Fatalf("inspect temporary directory after inspection: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("inspection left temporary files behind: %v", remaining)
	}
}

func TestManagerInspectRejectsInvalidArchive(t *testing.T) {
	manager := Manager{}
	_, err := manager.Inspect(context.Background(), filepath.Join(t.TempDir(), "missing.omfaibackup"))
	if !errors.Is(err, ErrInvalidBackupArchive) {
		t.Fatalf("expected invalid archive error, got %v", err)
	}
}

func TestManagerInspectHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	manager := Manager{}
	_, err := manager.Inspect(ctx, filepath.Join(t.TempDir(), "cancel.omfaibackup"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestManagerInspectRejectsWrongExtension(t *testing.T) {
	manager := Manager{}
	_, err := manager.Inspect(context.Background(), filepath.Join(t.TempDir(), "backup.zip"))
	if !errors.Is(err, ErrInvalidBackupPath) {
		t.Fatalf("expected invalid backup path, got %v", err)
	}
}
