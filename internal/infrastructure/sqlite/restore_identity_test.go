//go:build linux || windows

package sqlite

import (
	"errors"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestReadRestoreContentIdentityFromNoFollowHandle(t *testing.T) {
	directory := restoreTestDir(t)
	path := filepath.Join(directory, "candidate.db")
	content := []byte("verified SQLite restore candidate")
	if err := os.WriteFile(path, content, 0o600); err != nil { t.Fatal(err) }
	identity, err := readRestoreContentIdentityPlatform(path)
	if err != nil { t.Fatalf("readRestoreContentIdentityPlatform() error = %v", err) }
	sum := sha256.Sum256(content)
	if identity.Size != uint64(len(content)) || identity.SHA256 != hex.EncodeToString(sum[:]) {
		t.Fatalf("identity = %#v, want size %d and SHA-256 %x", identity, len(content), sum)
	}
	if !identity.matches(hex.EncodeToString(sum[:]), uint64(len(content))) { t.Fatal("matches() rejected exact identity") }
	if identity.matches(hex.EncodeToString(sum[:]), uint64(len(content)+1)) ||
		identity.matches("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", uint64(len(content))) {
		t.Fatal("matches() accepted mismatched identity")
	}
}

func TestReadRestoreContentIdentityRejectsUnsafeEntries(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		if _, err := readRestoreContentIdentityPlatform(filepath.Join(restoreTestDir(t), "missing.db")); err == nil {
			t.Fatal("accepted a missing file")
		}
	})
	t.Run("directory", func(t *testing.T) {
		path := filepath.Join(restoreTestDir(t), "candidate.db")
		if err := os.Mkdir(path, 0o700); err != nil { t.Fatal(err) }
		if _, err := readRestoreContentIdentityPlatform(path); err == nil { t.Fatal("accepted a directory") }
	})
	t.Run("symbolic link", func(t *testing.T) {
		directory := restoreTestDir(t)
		target, link := filepath.Join(directory, "real.db"), filepath.Join(directory, "candidate.db")
		if err := os.WriteFile(target, []byte("must not follow"), 0o600); err != nil { t.Fatal(err) }
		if err := os.Symlink(target, link); err != nil {
			if errors.Is(err, os.ErrPermission) { t.Skip("creating symlinks is not permitted") }
			t.Fatal(err)
		}
		if _, err := readRestoreContentIdentityPlatform(link); err == nil { t.Fatal("followed a symbolic link") }
	})
	t.Run("invalid leaf", func(t *testing.T) {
		for _, path := range []string{"", ".", ".."} {
			if _, err := readRestoreContentIdentityPlatform(path); err == nil {
				t.Fatalf("accepted invalid leaf %q", path)
			}
		}
	})
	t.Run("empty", func(t *testing.T) {
		path := filepath.Join(restoreTestDir(t), "empty.db")
		if err := os.WriteFile(path, nil, 0o600); err != nil { t.Fatal(err) }
		if _, err := readRestoreContentIdentityPlatform(path); err == nil { t.Fatal("accepted an empty file") }
	})
}
