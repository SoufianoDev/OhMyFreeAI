//go:build linux || windows

package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func restoreIdentityForTest(content []byte) RestoreContentIdentity {
	sum := sha256.Sum256(content)
	return RestoreContentIdentity{Size: uint64(len(content)), SHA256: hex.EncodeToString(sum[:])}
}

func TestMoveRestoreFileNoReplaceVerified(t *testing.T) {
	directory := t.TempDir()
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
	directory := t.TempDir()
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
	directory := t.TempDir()
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
