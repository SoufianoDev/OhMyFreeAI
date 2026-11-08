//go:build windows

package sqlite

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveRestoreJournalFailsClosed(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, restoreJournalFileName)
	if err := os.WriteFile(path, []byte("terminal journal"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := removeRestoreJournal(root, path)
	if !errors.Is(err, errRestoreJournalRemovalUnsupported) {
		t.Fatalf("removeRestoreJournal() error = %v, want unsupported", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("journal was removed despite unsupported durable deletion: %v", err)
	}
}
