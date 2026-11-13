//go:build linux || windows

package sqlite

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func restoreJournalReadFixture(t *testing.T, root string) []byte {
	t.Helper()
	journal := validTestRestoreJournal(root)
	data, err := json.Marshal(restoreJournalLedger{
		Version: restoreJournalLedgerVersion,
		Transactions: []RestoreJournal{journal},
	})
	if err != nil {
		t.Fatalf("encode journal read fixture: %v", err)
	}
	return data
}

func TestReadRestoreJournalRejectsUnsafeLedgerEntries(t *testing.T) {
	t.Run("directory", func(t *testing.T) {
		root := restoreTestDir(t)
		if err := os.Mkdir(filepath.Join(root, restoreJournalFileName), 0o700); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
			t.Fatalf("read directory ledger error = %v, want ErrRestoreJournalInvalid", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		root := restoreTestDir(t)
		path := filepath.Join(root, restoreJournalFileName)
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
			t.Fatalf("read empty ledger error = %v, want ErrRestoreJournalInvalid", err)
		}
	})

	t.Run("oversized file", func(t *testing.T) {
		root := restoreTestDir(t)
		path := filepath.Join(root, restoreJournalFileName)
		if err := os.WriteFile(path, make([]byte, maxRestoreJournalBytes+1), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
			t.Fatalf("read oversized ledger error = %v, want ErrRestoreJournalInvalid", err)
		}
	})

	t.Run("hard link", func(t *testing.T) {
		root := restoreTestDir(t)
		source := filepath.Join(root, "ledger-source")
		if err := os.WriteFile(source, restoreJournalReadFixture(t, root), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Link(source, filepath.Join(root, restoreJournalFileName)); err != nil {
			t.Fatalf("create ledger hard link: %v", err)
		}
		if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
			t.Fatalf("read hard-linked ledger error = %v, want ErrRestoreJournalInvalid", err)
		}
	})

	t.Run("broad permissions", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("POSIX mode-bit policy is Linux-specific")
		}
		root := restoreTestDir(t)
		path := filepath.Join(root, restoreJournalFileName)
		if err := os.WriteFile(path, restoreJournalReadFixture(t, root), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0o640); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
			t.Fatalf("read broadly accessible ledger error = %v, want ErrRestoreJournalInvalid", err)
		}
	})
}
