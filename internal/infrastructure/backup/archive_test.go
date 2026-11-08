package backup

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	appconfig "github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
	"github.com/SoufianoDev/OhMyFreeAI/internal/infrastructure/sqlite"
	"github.com/SoufianoDev/OhMyFreeAI/migrations"
)

func TestManagerCreateWritesStrictOmfaibackupArchive(t *testing.T) {
	ctx := context.Background()
	store, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate SQLite store: %v", err)
	}
	expected := appconfig.Defaults()
	expected.Gateway.Port = 18085
	if err := store.Save(ctx, expected); err != nil {
		t.Fatalf("save configuration: %v", err)
	}

	manager := Manager{ApplicationVersion: "0.1.0", Migrations: migrations.Files, MigrationDir: "."}
	archivePath := filepath.Join(t.TempDir(), "settings.omfaibackup")
	if err := manager.Create(ctx, store, archivePath); err != nil {
		t.Fatalf("create OhMyFreeAI backup: %v", err)
	}
	info, err := os.Stat(archivePath)
	if err != nil {
		t.Fatalf("stat backup archive: %v", err)
	}
	if info.Size() <= 0 {
		t.Fatal("backup archive is empty")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("backup archive permissions = %o, want 600", info.Mode().Perm())
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("open backup archive: %v", err)
	}
	defer reader.Close()
	if len(reader.File) != 2 {
		t.Fatalf("archive contains %d entries, want exactly 2", len(reader.File))
	}
	names := map[string]bool{}
	for _, entry := range reader.File {
		names[entry.Name] = true
		if entry.Mode().Perm() != 0o600 {
			t.Fatalf("archive entry %s permissions = %o, want 600", entry.Name, entry.Mode().Perm())
		}
	}
	if !names[manifestName] || !names[databaseName] {
		t.Fatalf("archive entries = %v, want only manifest and SQLite database", names)
	}

	manifestBytes, err := readArchiveEntry(findZipEntry(t, reader.File, manifestName), maxManifestBytes)
	if err != nil {
		t.Fatalf("read archive manifest: %v", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode archive manifest: %v", err)
	}
	if manifest.Format != archiveFormat || manifest.FormatVersion != archiveVersion || manifest.SchemaVersion != 1 {
		t.Fatalf("unexpected archive manifest: %#v", manifest)
	}
	if len(manifest.Files) != 1 || manifest.Files[0].Path != databaseName {
		t.Fatalf("archive contains unexpected file manifest: %#v", manifest.Files)
	}
}

func TestManagerRestoreFailsClosedWithoutTouchingTarget(t *testing.T) {
	directory := t.TempDir()
	targetPath := filepath.Join(directory, "state.db")
	original := []byte("existing application state must remain untouched")
	if err := os.WriteFile(targetPath, original, 0o600); err != nil {
		t.Fatalf("create existing target: %v", err)
	}

	manager := Manager{}
	archivePath := filepath.Join(directory, "missing.omfaibackup")
	if _, err := manager.Restore(context.Background(), archivePath, targetPath); !errors.Is(err, sqlite.ErrRestoreExecutorUnavailable) {
		t.Fatalf("Manager.Restore() error = %v, want ErrRestoreExecutorUnavailable", err)
	}

	actual, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read existing target after disabled restore: %v", err)
	}
	if !reflect.DeepEqual(actual, original) {
		t.Fatalf("target changed after disabled restore: got %q, want %q", actual, original)
	}
	if _, err := os.Stat(archivePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("disabled restore unexpectedly accessed or created archive path: %v", err)
	}
}

func TestManagerCreateDoesNotOverwriteExistingArchive(t *testing.T) {
	ctx := context.Background()
	store, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate SQLite store: %v", err)
	}
	manager := Manager{ApplicationVersion: "0.1.0", Migrations: migrations.Files, MigrationDir: "."}
	archivePath := filepath.Join(t.TempDir(), "existing.omfaibackup")
	original := []byte("keep existing archive")
	if err := os.WriteFile(archivePath, original, 0o600); err != nil {
		t.Fatalf("create existing archive: %v", err)
	}
	if err := manager.Create(ctx, store, archivePath); !errors.Is(err, ErrBackupArchiveExists) {
		t.Fatalf("expected existing-archive error, got %v", err)
	}
	actual, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read existing archive: %v", err)
	}
	if !reflect.DeepEqual(actual, original) {
		t.Fatalf("existing archive changed: %q", actual)
	}
}

func findZipEntry(t *testing.T, entries []*zip.File, name string) *zip.File {
	t.Helper()
	for _, entry := range entries {
		if entry.Name == name {
			return entry
		}
	}
	t.Fatalf("archive entry %q not found", name)
	return nil
}


func TestExtractArchiveRejectsPathTraversalEntry(t *testing.T) {
	ctx := context.Background()
	store, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate SQLite store: %v", err)
	}
	manager := Manager{ApplicationVersion: "0.1.0", Migrations: migrations.Files, MigrationDir: "."}
	archivePath := filepath.Join(t.TempDir(), "safe.omfaibackup")
	if err := manager.Create(ctx, store, archivePath); err != nil {
		t.Fatalf("create valid backup archive: %v", err)
	}

	traversalPath := filepath.Join(t.TempDir(), "traversal.omfaibackup")
	rewriteArchiveEntryName(t, archivePath, traversalPath, databaseName, "../state.sqlite")
	extractedPath := filepath.Join(t.TempDir(), "should-not-exist.sqlite")
	if _, err := extractArchive(ctx, traversalPath, extractedPath); !errors.Is(err, ErrInvalidBackupArchive) {
		t.Fatalf("expected path-traversal rejection, got %v", err)
	}
	if _, err := os.Stat(extractedPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("path-traversal archive unexpectedly extracted a file, stat error: %v", err)
	}
}

func rewriteArchiveEntryName(t *testing.T, sourcePath, destinationPath, oldName, newName string) {
	t.Helper()
	source, err := zip.OpenReader(sourcePath)
	if err != nil {
		t.Fatalf("open source archive: %v", err)
	}
	defer source.Close()
	output, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("create rewritten archive: %v", err)
	}
	writer := zip.NewWriter(output)
	for _, entry := range source.File {
		reader, err := entry.Open()
		if err != nil {
			t.Fatalf("open archive entry: %v", err)
		}
		data, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			t.Fatalf("read archive entry: %v", err)
		}
		name := entry.Name
		if name == oldName {
			name = newName
		}
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(0o600)
		entryWriter, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatalf("create rewritten archive entry: %v", err)
		}
		if _, err := entryWriter.Write(data); err != nil {
			t.Fatalf("write rewritten archive entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close rewritten archive: %v", err)
	}
	if err := output.Close(); err != nil {
		t.Fatalf("close rewritten archive file: %v", err)
	}
}


func TestExtractArchiveHonorsCancellation(t *testing.T) {
	store, err := sqlite.Open(context.Background(), filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(context.Background(), migrations.Files, "."); err != nil {
		t.Fatalf("migrate SQLite store: %v", err)
	}

	manager := Manager{ApplicationVersion: "0.1.0", Migrations: migrations.Files, MigrationDir: "."}
	archivePath := filepath.Join(t.TempDir(), "cancel.omfaibackup")
	if err := manager.Create(context.Background(), store, archivePath); err != nil {
		t.Fatalf("create valid backup archive: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	extractedPath := filepath.Join(t.TempDir(), "cancelled.sqlite")
	if _, err := extractArchive(ctx, archivePath, extractedPath); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
