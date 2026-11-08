package sqlite

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"reflect"
	"strings"
	"testing"
	"time"
)

func boolPointer(value bool) *bool { return &value }

func validTestRestoreJournal(root string) RestoreJournal {
	return RestoreJournal{
		Version:       restoreJournalVersion,
		TransactionID: "test-transaction-001",
		Phase:         RestorePhasePrepared,
		TargetPath:    filepath.Join(root, "state.db"),
		StagingPath:   filepath.Join(root, ".restore-stage.db"),
		ArchiveSHA256: strings.Repeat("a", 64),
		ArchiveSize:   4096,
		UpdatedAt:     time.Now().UTC(),
	}
}

func canonicalTestRestorePath(t *testing.T, path string) string {
	t.Helper()
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		t.Fatalf("resolve test path parent: %v", err)
	}
	return filepath.Join(parent, filepath.Base(path))
}

func TestRestoreJournalRoundTripAndAtomicPhaseUpdate(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.RollbackPath = filepath.Join(root, ".restore-rollback.db")
	if err := WriteRestoreJournal(root, journal); err != nil {
		t.Fatalf("write initial restore journal: %v", err)
	}

	got, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read initial restore journal: %v", err)
	}
	want := journal
	want.TargetPath = canonicalTestRestorePath(t, journal.TargetPath)
	want.StagingPath = canonicalTestRestorePath(t, journal.StagingPath)
	want.RollbackPath = canonicalTestRestorePath(t, journal.RollbackPath)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("read journal = %#v, want %#v", got, want)
	}
	info, err := os.Stat(filepath.Join(root, restoreJournalFileName))
	if err != nil {
		t.Fatalf("stat restore journal: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("journal permissions = %o, want 600", info.Mode().Perm())
	}

	journal.Phase = RestorePhasePreviousPreserved
	journal.UpdatedAt = time.Now().UTC().Add(time.Second)
	if err := WriteRestoreJournal(root, journal); err != nil {
		t.Fatalf("atomically update restore journal: %v", err)
	}
	got, err = ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read updated restore journal: %v", err)
	}
	if got.Phase != RestorePhasePreviousPreserved || !got.UpdatedAt.Equal(journal.UpdatedAt) {
		t.Fatalf("updated journal = %#v, want phase %q and timestamp %s", got, journal.Phase, journal.UpdatedAt)
	}
	if err := RemoveRestoreJournal(root); err != nil {
		t.Fatalf("remove restore journal: %v", err)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read removed journal error = %v, want os.ErrNotExist", err)
	}
}

func TestWriteRestoreJournalFillsMissingTimestamp(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.UpdatedAt = time.Time{}
	before := time.Now().Add(-time.Second)
	if err := WriteRestoreJournal(root, journal); err != nil {
		t.Fatalf("write journal without timestamp: %v", err)
	}
	got, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read journal: %v", err)
	}
	if got.UpdatedAt.IsZero() || got.UpdatedAt.Before(before) {
		t.Fatalf("journal timestamp was not initialized: %s", got.UpdatedAt)
	}
}

func TestWriteRestoreJournalRejectsPathsOutsideStateDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.StagingPath = filepath.Join(outside, "stage.db")
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write with external staging path error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestWriteRestoreJournalRejectsInvalidPhaseAndArchiveIdentity(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.Phase = RestorePhase("unknown")
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write with unknown phase error = %v, want ErrRestoreJournalInvalid", err)
	}

	journal = validTestRestoreJournal(root)
	journal.PreviousDatabaseExisted = nil
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write without previous-database state error = %v, want ErrRestoreJournalInvalid", err)
	}

	journal = validTestRestoreJournal(root)
	journal.ArchiveSHA256 = "not-a-sha256"
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write with invalid archive identity error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestReadRestoreJournalRejectsUnknownFieldsAndTrailingData(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, restoreJournalFileName)
	for _, content := range []string{
		`{"version":1,"unknown":true}`,
		`{} { "second": true }`,
	} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write malformed journal: %v", err)
		}
		if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
			t.Fatalf("read malformed journal error = %v, want ErrRestoreJournalInvalid", err)
		}
	}
}

func TestReadRestoreJournalRejectsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require elevated privileges on Windows")
	}
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	if err := os.Symlink(target, filepath.Join(root, restoreJournalFileName)); err != nil {
		t.Fatalf("create journal symlink: %v", err)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("read symlink journal error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestWriteRestoreJournalRejectsSymlinkedTransactionPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require elevated privileges on Windows")
	}
	root := t.TempDir()
	outside := filepath.Join(root, "outside.db")
	if err := os.WriteFile(outside, []byte("not a database"), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	link := filepath.Join(root, "stage-link.db")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("create staging symlink: %v", err)
	}
	journal := validTestRestoreJournal(root)
	journal.StagingPath = link
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write with symlinked staging path error = %v, want ErrRestoreJournalInvalid", err)
	}
}
