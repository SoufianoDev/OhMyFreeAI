// Test-first checkpoint: safely inspect restore artifact identity from an opened file handle.
package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectRestoreArtifactReturnsContentIdentity(t *testing.T) {
	stateDirectory := t.TempDir()
	path := filepath.Join(stateDirectory, "candidate.sqlite")
	content := []byte("verified SQLite restore candidate")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := inspectRestoreArtifact(stateDirectory, path)
	if err != nil {
		t.Fatalf("inspectRestoreArtifact() error = %v", err)
	}
	sum := sha256.Sum256(content)
	want := RestoreContentIdentity{
		SHA256: hex.EncodeToString(sum[:]),
		Size:   uint64(len(content)),
	}
	if got != want {
		t.Fatalf("inspectRestoreArtifact() = %#v, want %#v", got, want)
	}
}

func TestInspectRestoreArtifactRejectsUnsafePaths(t *testing.T) {
	stateDirectory := t.TempDir()
	outsideDirectory := t.TempDir()
	outside := filepath.Join(outsideDirectory, "outside.sqlite")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
	}{
		{name: "outside state directory", path: outside},
		{name: "state directory itself", path: stateDirectory},
		{name: "missing file", path: filepath.Join(stateDirectory, "missing.sqlite")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := inspectRestoreArtifact(stateDirectory, test.path); err == nil {
				t.Fatal("inspectRestoreArtifact() unexpectedly accepted an unsafe or missing path")
			}
		})
	}
}

func TestInspectRestoreArtifactRejectsSymlinksAndNonRegularFiles(t *testing.T) {
	stateDirectory := t.TempDir()
	realPath := filepath.Join(stateDirectory, "real.sqlite")
	if err := os.WriteFile(realPath, []byte("real database"), 0o600); err != nil {
		t.Fatal(err)
	}

	symlinkPath := filepath.Join(stateDirectory, "linked.sqlite")
	if err := os.Symlink(realPath, symlinkPath); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skip("creating symlinks is not permitted")
		}
		t.Fatal(err)
	}
	if _, err := inspectRestoreArtifact(stateDirectory, symlinkPath); err == nil {
		t.Fatal("inspectRestoreArtifact() accepted a symbolic-link artifact")
	}

	directoryPath := filepath.Join(stateDirectory, "directory.sqlite")
	if err := os.Mkdir(directoryPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := inspectRestoreArtifact(stateDirectory, directoryPath); err == nil {
		t.Fatal("inspectRestoreArtifact() accepted a directory as a database artifact")
	}
}

func TestInspectRestoreArtifactRejectsSymlinkedParent(t *testing.T) {
	root := t.TempDir()
	stateDirectory := filepath.Join(root, "state")
	realDirectory := filepath.Join(stateDirectory, "artifacts")
	if err := os.MkdirAll(realDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(realDirectory, "candidate.sqlite")
	if err := os.WriteFile(path, []byte("candidate"), 0o600); err != nil {
		t.Fatal(err)
	}

	alias := filepath.Join(root, "alias")
	if err := os.Symlink(realDirectory, alias); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skip("creating symlinks is not permitted")
		}
		t.Fatal(err)
	}
	if _, err := inspectRestoreArtifact(stateDirectory, filepath.Join(alias, "candidate.sqlite")); err == nil {
		t.Fatal("inspectRestoreArtifact() accepted a symlinked parent directory")
	}
}
