package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPublishRestoreJournalReplacesExistingFile(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, ".restore-journal-next.tmp")
	destination := filepath.Join(root, restoreJournalFileName)
	if err := os.WriteFile(source, []byte("next journal"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("previous journal"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := publishRestoreJournal(source, destination); err != nil {
		t.Fatalf("publishRestoreJournal() error = %v", err)
	}
	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("temporary journal still exists or cannot be inspected: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "next journal" {
		t.Fatalf("published journal = %q, want %q", got, "next journal")
	}
}

func TestPublishRestoreJournalNoReplaceCreatesMissingFile(t *testing.T) {
	skipUnlessRestoreNoReplaceSupported(t)
	root := restoreTestDir(t)
	source := filepath.Join(root, ".restore-journal-next.tmp")
	destination := filepath.Join(root, restoreJournalFileName)
	data := []byte("initial journal")
	if err := os.WriteFile(source, data, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := publishRestoreJournalNoReplace(source, destination, restoreContentIdentityForTest(data)); err != nil {
		t.Fatalf("publishRestoreJournalNoReplace() error = %v", err)
	}
	if _, err := os.Lstat(source); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary journal still exists or cannot be inspected: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("published journal = %q, want %q", got, data)
	}
}

func TestPublishRestoreJournalNoReplacePreservesExistingFile(t *testing.T) {
	skipUnlessRestoreNoReplaceSupported(t)
	root := restoreTestDir(t)
	source := filepath.Join(root, ".restore-journal-next.tmp")
	destination := filepath.Join(root, restoreJournalFileName)
	sourceData := []byte("next journal")
	destinationData := []byte("existing journal")
	if err := os.WriteFile(source, sourceData, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, destinationData, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := publishRestoreJournalNoReplace(source, destination, restoreContentIdentityForTest(sourceData)); !errors.Is(err, os.ErrExist) {
		t.Fatalf("publishRestoreJournalNoReplace() error = %v, want os.ErrExist", err)
	}
	for path, want := range map[string]string{
		source:      string(sourceData),
		destination: string(destinationData),
	} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read preserved journal %s: %v", path, err)
		}
		if string(got) != want {
			t.Fatalf("journal %s = %q, want %q", path, got, want)
		}
	}
}

func TestPublishRestoreJournalNoReplacePreservesMismatchedArtifact(t *testing.T) {
	skipUnlessRestoreNoReplaceSupported(t)
	root := restoreTestDir(t)
	source := filepath.Join(root, ".restore-journal-next.tmp")
	destination := filepath.Join(root, restoreJournalFileName)
	data := []byte("published journal")
	if err := os.WriteFile(source, data, 0o600); err != nil {
		t.Fatal(err)
	}

	expected := restoreContentIdentityForTest([]byte("different journal"))
	if err := publishRestoreJournalNoReplace(source, destination, expected); err == nil {
		t.Fatal("publishRestoreJournalNoReplace() accepted a mismatched artifact")
	}
	if _, err := os.Lstat(source); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source remains or cannot be inspected after the move: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read preserved destination: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("destination = %q, want preserved artifact %q", got, data)
	}
}

func restoreContentIdentityForTest(data []byte) RestoreContentIdentity {
	digest := sha256.Sum256(data)
	return RestoreContentIdentity{Size: uint64(len(data)), SHA256: hex.EncodeToString(digest[:])}
}

func skipUnlessRestoreNoReplaceSupported(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		t.Skip("atomic no-replace Restore moves are qualified only on Linux and Windows")
	}
}
