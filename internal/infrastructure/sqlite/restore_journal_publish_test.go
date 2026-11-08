package sqlite

import (
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
