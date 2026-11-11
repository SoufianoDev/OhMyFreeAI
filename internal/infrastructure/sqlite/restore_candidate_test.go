package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"testing/fstest"

	domainconfig "github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
	"github.com/SoufianoDev/OhMyFreeAI/migrations"
)

func createRestoreCandidateSnapshot(t *testing.T) (string, *Store, domainconfig.Config) {
	t.Helper()
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "source.db"))
	if err != nil {
		t.Fatalf("open restore candidate source: %v", err)
	}
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		_ = store.Close()
		t.Fatalf("migrate restore candidate source: %v", err)
	}
	expected := domainconfig.Defaults()
	expected.Gateway.Port = 18111
	expected.UI.Language = "en"
	if err := store.Save(ctx, expected); err != nil {
		_ = store.Close()
		t.Fatalf("save restore candidate source: %v", err)
	}
	snapshotPath := filepath.Join(t.TempDir(), "source-snapshot.db")
	if err := store.BackupTo(ctx, snapshotPath); err != nil {
		_ = store.Close()
		t.Fatalf("create restore candidate source snapshot: %v", err)
	}
	return snapshotPath, store, expected
}

func TestBuildRestoreCandidateReturnsVerifiedPrivateArtifact(t *testing.T) {
	ctx := context.Background()
	sourcePath, sourceStore, expected := createRestoreCandidateSnapshot(t)
	defer sourceStore.Close()

	root := restoreTestDir(t)
	candidate, err := BuildRestoreCandidate(ctx, sourcePath, root, migrations.Files, ".")
	if err != nil {
		t.Fatalf("build restore candidate: %v", err)
	}
	if candidate.Path == "" || candidate.Identity.Size == 0 || candidate.Identity.SHA256 == "" {
		t.Fatalf("candidate identity is incomplete: %#v", candidate)
	}
	if !samePath(filepath.Dir(candidate.Path), root) {
		t.Fatalf("candidate path %q is outside private state directory %q", candidate.Path, root)
	}
	actualIdentity, err := readRestoreContentIdentityPlatform(candidate.Path)
	if err != nil {
		t.Fatalf("read candidate identity: %v", err)
	}
	if !actualIdentity.matches(candidate.Identity.SHA256, candidate.Identity.Size) {
		t.Fatalf("candidate identity changed: got %#v, want %#v", actualIdentity, candidate.Identity)
	}
	// Check the builder's closed-artifact contract before opening the file again.
	// A read-only SQLite verification connection may itself create WAL/SHM
	// sidecars for a WAL-mode database, even though the builder left none.
	if err := ensureNoSQLiteSidecars(candidate.Path); err != nil {
		t.Fatalf("candidate retained SQLite sidecars: %v", err)
	}
	if err := verifyBackupDatabase(ctx, candidate.Path); err != nil {
		t.Fatalf("verify built candidate database: %v", err)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(candidate.Path)
		if err != nil {
			t.Fatalf("stat candidate: %v", err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("candidate permissions = %o, want 600", info.Mode().Perm())
		}
	}
	actualSourceConfig, err := sourceStore.Load(ctx)
	if err != nil {
		t.Fatalf("read source configuration after candidate build: %v", err)
	}
	if !reflect.DeepEqual(actualSourceConfig, expected) {
		t.Fatalf("source database changed during candidate build:\n got: %#v\nwant: %#v", actualSourceConfig, expected)
	}
}


func TestBuildRestoreCandidateAtPathUsesExclusiveJournalSelectedPath(t *testing.T) {
	ctx := context.Background()
	sourcePath, sourceStore, _ := createRestoreCandidateSnapshot(t)
	defer sourceStore.Close()

	root := restoreTestDir(t)
	stagingPath := filepath.Join(root, ".restore-stage-transaction.sqlite")
	candidate, err := BuildRestoreCandidateAtPath(ctx, sourcePath, root, stagingPath, migrations.Files, ".")
	if err != nil {
		t.Fatalf("build candidate at journal-selected path: %v", err)
	}
	if candidate.Path != stagingPath {
		t.Fatalf("candidate path = %q, want exact journal-selected path %q", candidate.Path, stagingPath)
	}
	actual, err := readRestoreContentIdentityPlatform(stagingPath)
	if err != nil {
		t.Fatalf("read staged candidate identity: %v", err)
	}
	if !actual.matches(candidate.Identity.SHA256, candidate.Identity.Size) {
		t.Fatalf("staged identity = %#v, returned identity = %#v", actual, candidate.Identity)
	}
	if err := ensureNoSQLiteSidecars(stagingPath); err != nil {
		t.Fatalf("staged candidate retained SQLite sidecars: %v", err)
	}
}

func TestBuildRestoreCandidateAtPathNeverReplacesExistingArtifact(t *testing.T) {
	ctx := context.Background()
	sourcePath, sourceStore, _ := createRestoreCandidateSnapshot(t)
	defer sourceStore.Close()

	root := restoreTestDir(t)
	stagingPath := filepath.Join(root, ".restore-stage-collision.sqlite")
	sentinel := []byte("preserve pre-existing evidence")
	if err := os.WriteFile(stagingPath, sentinel, 0o600); err != nil {
		t.Fatalf("create existing staging evidence: %v", err)
	}
	candidate, err := BuildRestoreCandidateAtPath(ctx, sourcePath, root, stagingPath, migrations.Files, ".")
	if err == nil {
		t.Fatal("candidate builder unexpectedly replaced an existing artifact")
	}
	if candidate.Path != "" {
		t.Fatalf("candidate path = %q after O_EXCL collision, want empty", candidate.Path)
	}
	actual, readErr := os.ReadFile(stagingPath)
	if readErr != nil {
		t.Fatalf("read preserved staging evidence: %v", readErr)
	}
	if string(actual) != string(sentinel) {
		t.Fatalf("existing staging evidence changed: %q", actual)
	}
}

func TestBuildRestoreCandidateAtPathPreservesPartialArtifactOnFailure(t *testing.T) {
	ctx := context.Background()
	sourcePath, sourceStore, _ := createRestoreCandidateSnapshot(t)
	defer sourceStore.Close()

	root := restoreTestDir(t)
	stagingPath := filepath.Join(root, ".restore-stage-failed.sqlite")
	candidate, err := BuildRestoreCandidateAtPath(ctx, sourcePath, root, stagingPath, fstest.MapFS{}, ".")
	if !errors.Is(err, ErrMigrationVersionTooNew) {
		t.Fatalf("candidate build error = %v, want ErrMigrationVersionTooNew", err)
	}
	if candidate.Path != stagingPath {
		t.Fatalf("preserved candidate path = %q, want %q", candidate.Path, stagingPath)
	}
	info, statErr := os.Stat(stagingPath)
	if statErr != nil {
		t.Fatalf("partial candidate artifact was removed: %v", statErr)
	}
	if info.Size() == 0 {
		t.Fatal("partial candidate artifact is unexpectedly empty")
	}
}

func TestBuildRestoreCandidateAtPathRejectsPathsOutsideStateDirectory(t *testing.T) {
	ctx := context.Background()
	sourcePath, sourceStore, _ := createRestoreCandidateSnapshot(t)
	defer sourceStore.Close()

	root := restoreTestDir(t)
	outside := restoreTestDir(t)
	candidate, err := BuildRestoreCandidateAtPath(ctx, sourcePath, root, filepath.Join(outside, "candidate.sqlite"), migrations.Files, ".")
	if !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("outside candidate path error = %v, want ErrRestoreJournalInvalid", err)
	}
	if candidate.Path != "" {
		t.Fatalf("outside candidate path was accepted: %#v", candidate)
	}
}

func TestBuildRestoreCandidatePreservesArtifactOnMigrationFailure(t *testing.T) {
	ctx := context.Background()
	sourcePath, sourceStore, _ := createRestoreCandidateSnapshot(t)
	defer sourceStore.Close()

	root := restoreTestDir(t)
	candidate, err := BuildRestoreCandidate(ctx, sourcePath, root, fstest.MapFS{}, ".")
	if !errors.Is(err, ErrMigrationVersionTooNew) {
		t.Fatalf("migration error = %v, want ErrMigrationVersionTooNew", err)
	}
	if candidate.Path == "" {
		t.Fatal("failed build did not return the preserved candidate path")
	}
	info, statErr := os.Stat(candidate.Path)
	if statErr != nil {
		t.Fatalf("failed candidate artifact was removed: %v", statErr)
	}
	if info.Size() == 0 {
		t.Fatal("failed candidate artifact is unexpectedly empty")
	}
}

func TestBuildRestoreCandidateHonorsCanceledContextBeforeCreatingArtifact(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	root := restoreTestDir(t)

	candidate, err := BuildRestoreCandidate(ctx, filepath.Join(t.TempDir(), "unused.db"), root, migrations.Files, ".")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("build error = %v, want context.Canceled", err)
	}
	if candidate.Path != "" {
		t.Fatalf("canceled build created candidate artifact %q", candidate.Path)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read state directory: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("canceled build left artifacts: %#v", entries)
	}
}

func TestSamePathDetectsHardLinkedFiles(t *testing.T) {
	directory := t.TempDir()
	original := filepath.Join(directory, "original.sqlite")
	alias := filepath.Join(directory, "alias.sqlite")
	if err := os.WriteFile(original, []byte("same file identity"), 0o600); err != nil {
		t.Fatalf("write original file: %v", err)
	}
	if err := os.Link(original, alias); err != nil {
		t.Skipf("hard links are unavailable on this filesystem: %v", err)
	}
	if !samePath(original, alias) {
		t.Fatal("samePath did not recognize two names for the same file")
	}
}
