package sqlite

import (
	"errors"
	"os"
	"path/filepath"
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

func TestPublishRestoreJournalNoReplacePreservesExistingFile(t *testing.T) {
	root := restoreTestDir(t)
	source := filepath.Join(root, ".restore-journal-next.tmp")
	destination := filepath.Join(root, restoreJournalFileName)
	if err := os.WriteFile(source, []byte("next journal"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("existing journal"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := publishRestoreJournalNoReplace(source, destination); !errors.Is(err, os.ErrExist) {
		t.Fatalf("publishRestoreJournalNoReplace() error = %v, want os.ErrExist", err)
	}
	for path, want := range map[string]string{
		source:      "next journal",
		destination: "existing journal",
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
