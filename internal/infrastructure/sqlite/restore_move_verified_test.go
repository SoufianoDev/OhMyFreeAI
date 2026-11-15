//go:build linux || windows

package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func restoreIdentityForTest(content []byte) RestoreContentIdentity {
	sum := sha256.Sum256(content)
	return RestoreContentIdentity{Size: uint64(len(content)), SHA256: hex.EncodeToString(sum[:])}
}

func TestMoveRestoreFileNoReplaceVerified(t *testing.T) {
	directory := restoreTestDir(t)
	source := filepath.Join(directory, "candidate.staging")
	destination := filepath.Join(directory, "candidate.db")
	content := []byte("verified candidate bytes")

	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := moveRestoreFileNoReplaceVerified(source, destination, restoreIdentityForTest(content)); err != nil {
		t.Fatalf("moveRestoreFileNoReplaceVerified() error = %v", err)
	}
	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists after verified move: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("destination content = %q, want %q", got, content)
	}
}

func TestMoveRestoreFileNoReplaceVerifiedPreservesIdentityMismatch(t *testing.T) {
	directory := restoreTestDir(t)
	source := filepath.Join(directory, "candidate.staging")
	destination := filepath.Join(directory, "candidate.db")
	actual := []byte("different candidate bytes")
	expected := restoreIdentityForTest([]byte("journaled candidate bytes"))

	if err := os.WriteFile(source, actual, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := moveRestoreFileNoReplaceVerified(source, destination, expected); err == nil {
		t.Fatal("accepted moved content that did not match the journaled identity")
	}
	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("source should have been moved before post-move verification: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("mismatched destination was not preserved for recovery: %v", err)
	}
	if string(got) != string(actual) {
		t.Fatalf("preserved destination content = %q, want original moved bytes %q", got, actual)
	}
}

func TestMoveRestoreFileNoReplaceVerifiedRejectsInvalidExpectedIdentityBeforeMove(t *testing.T) {
	directory := restoreTestDir(t)
	source := filepath.Join(directory, "candidate.staging")
	destination := filepath.Join(directory, "candidate.db")
	content := []byte("candidate bytes")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatal(err)
	}

	err := moveRestoreFileNoReplaceVerified(source, destination, RestoreContentIdentity{Size: uint64(len(content)), SHA256: "invalid"})
	if err == nil {
		t.Fatal("accepted an invalid expected content identity")
	}
	if _, err := os.Lstat(source); err != nil {
		t.Fatalf("source was changed despite invalid expected identity: %v", err)
	}
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		t.Fatalf("destination was created despite invalid expected identity: %v", err)
	}
}

func TestMoveRestoreFileNoReplaceVerifiedRejectsSymlinkSubstitutionAfterPreflight(t *testing.T) {
	directory := restoreTestDir(t)
	source := filepath.Join(directory, "candidate.staging")
	destination := filepath.Join(directory, "candidate.db")
	external := filepath.Join(directory, "external.db")
	expectedContent := []byte("journaled candidate bytes")
	externalContent := []byte("external file must never be followed")

	if err := os.WriteFile(source, expectedContent, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(external, externalContent, 0o600); err != nil {
		t.Fatal(err)
	}

	err := moveRestoreFileNoReplaceVerifiedWithHooks(
		source,
		destination,
		restoreIdentityForTest(expectedContent),
		func() error {
			if err := os.Remove(source); err != nil {
				return err
			}
			if err := os.Symlink(external, source); err != nil {
				if runtime.GOOS == "windows" {
					t.Skipf("Windows runner does not permit creating a file symlink: %v", err)
				}
				return err
			}
			return nil
		},
		nil,
	)
	if err == nil {
		t.Fatal("accepted a source symlink substituted after the regular-file preflight")
	}

	linkPath := source
	if runtime.GOOS == "linux" {
		if _, err := os.Lstat(source); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Linux source leaf after rename = %v, want not-exist", err)
		}
		linkPath = destination
		if _, err := readRestoreContentIdentityPlatform(destination); err == nil {
			t.Fatal("followed the symlink moved into the destination")
		}
	} else {
		if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Windows destination after rejecting reparse source = %v, want not-exist", err)
		}
	}

	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("inspect preserved source symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("preserved artifact mode = %v, want symlink", info.Mode())
	}
	gotTarget, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("read preserved symlink target: %v", err)
	}
	if gotTarget != external {
		t.Fatalf("preserved symlink target = %q, want %q", gotTarget, external)
	}
	got, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(externalContent) {
		t.Fatalf("external target content changed: got %q, want %q", got, externalContent)
	}
}
