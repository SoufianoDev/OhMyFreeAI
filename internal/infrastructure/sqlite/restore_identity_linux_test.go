//go:build linux

package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestRestoreArtifactMetadataDetectsInodeStatusChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate.sqlite")
	if err := os.WriteFile(path, []byte("stable-sized content"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	before, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	// chmod changes inode ctime without changing file content, size, or mtime.
	if err := file.Chmod(0o400); err != nil {
		t.Fatal(err)
	}
	after, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("test setup unexpectedly changed file size or mtime")
	}

	stable, err := restoreArtifactMetadataUnchanged(before, after)
	if err != nil {
		t.Fatalf("restoreArtifactMetadataUnchanged() error = %v", err)
	}
	if stable {
		t.Fatal("restoreArtifactMetadataUnchanged() accepted an inode status change")
	}
}

func TestRestoreArtifactMetadataDetectsContentChangeWhenMtimeIsRestored(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate.sqlite")
	if err := os.WriteFile(path, []byte("original content"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	before, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	original := []byte("original content")
	replacement := []byte("modified content")
	if len(original) != len(replacement) {
		t.Fatal("test fixture must preserve file size")
	}
	if _, err := file.WriteAt(replacement, 0); err != nil {
		t.Fatal(err)
	}
	// Simulate a writer restoring mtime after an in-place content update.
	if err := os.Chtimes(path, before.ModTime(), before.ModTime()); err != nil {
		t.Fatal(err)
	}
	after, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("test setup did not preserve file size and mtime")
	}

	beforeStat, ok := before.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("before stat metadata type = %T, want *syscall.Stat_t", before.Sys())
	}
	afterStat, ok := after.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("after stat metadata type = %T, want *syscall.Stat_t", after.Sys())
	}
	if beforeStat.Ctim == afterStat.Ctim {
		t.Skip("filesystem timestamp resolution did not expose the content update through ctime")
	}

	stable, err := restoreArtifactMetadataUnchanged(before, after)
	if err != nil {
		t.Fatalf("restoreArtifactMetadataUnchanged() error = %v", err)
	}
	if stable {
		t.Fatal("restoreArtifactMetadataUnchanged() accepted an in-place content update with restored mtime")
	}
}

func TestRestoreArtifactMetadataAcceptsUnchangedInode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate.sqlite")
	if err := os.WriteFile(path, []byte("stable content"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	before, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	after, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	stable, err := restoreArtifactMetadataUnchanged(before, after)
	if err != nil {
		t.Fatalf("restoreArtifactMetadataUnchanged() error = %v", err)
	}
	if !stable {
		t.Fatal("restoreArtifactMetadataUnchanged() rejected unchanged inode metadata")
	}
}

func TestOpenVerifiedRestoreArtifactRetainsHandleAfterPathReplacement(t *testing.T) {
	stateDirectory := t.TempDir()
	path := filepath.Join(stateDirectory, "candidate.sqlite")
	movedPath := filepath.Join(stateDirectory, "original.sqlite")
	original := []byte("original verified SQLite candidate")
	replacement := []byte("replacement candidate with same length")
	if len(original) != len(replacement) {
		t.Fatal("test fixtures must have equal lengths")
	}
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}

	artifact, err := openVerifiedRestoreArtifact(stateDirectory, path)
	if err != nil {
		t.Fatalf("openVerifiedRestoreArtifact() error = %v", err)
	}
	defer artifact.close()

	if artifact.identity.Size != uint64(len(original)) {
		t.Fatalf("verified size = %d, want %d", artifact.identity.Size, len(original))
	}
	digest := sha256.Sum256(original)
	if artifact.identity.SHA256 != hex.EncodeToString(digest[:]) {
		t.Fatalf("verified SHA-256 = %q, want original content digest", artifact.identity.SHA256)
	}

	// Replace the directory entry after verification. The retained descriptor
	// must still refer to the exact inode whose content was hashed.
	if err := os.Rename(path, movedPath); err != nil {
		t.Fatalf("move original artifact away from its inspected path: %v", err)
	}
	if err := os.WriteFile(path, replacement, 0o600); err != nil {
		t.Fatalf("create replacement at original path: %v", err)
	}

	got, err := io.ReadAll(artifact.file)
	if err != nil {
		t.Fatalf("read retained verified handle: %v", err)
	}
	if string(got) != string(original) {
		t.Fatalf("retained handle read %q, want original verified bytes %q", got, original)
	}
}
