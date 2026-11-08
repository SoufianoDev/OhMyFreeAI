//go:build linux

package sqlite

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMoveRestoreFileNoReplaceRejectsSymlinkedContainingDirectory(t *testing.T) {
	root := t.TempDir()
	realDirectory := filepath.Join(root, "real")
	if err := os.Mkdir(realDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(realDirectory, "source.db")
	if err := os.WriteFile(source, []byte("candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(realDirectory, alias); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skip("creating symlinks is not permitted")
		}
		t.Fatal(err)
	}

	err := moveRestoreFileNoReplace(filepath.Join(alias, "source.db"), filepath.Join(alias, "target.db"))
	if err == nil {
		t.Fatal("moveRestoreFileNoReplace() accepted a symlinked containing directory")
	}
	if _, err := os.Lstat(source); err != nil {
		t.Fatalf("real source was changed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(realDirectory, "target.db")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination unexpectedly exists: %v", err)
	}
}
