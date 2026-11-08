//go:build linux

package sqlite

import (
	"os"
	"path/filepath"
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
