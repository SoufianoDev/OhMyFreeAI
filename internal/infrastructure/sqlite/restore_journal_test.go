package sqlite

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
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
		RollbackPath: filepath.Join(root, ".restore-rollback.db"),
		CandidateQuarantinePath: filepath.Join(root, ".restore-candidate-quarantine.db"),
		PreviousDatabaseExisted: boolPointer(true),
		CandidateSHA256: strings.Repeat("b", 64),
		CandidateSize: 8192,
		PreviousSHA256: strings.Repeat("c", 64),
		PreviousSize: 4096,
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
	want.CandidateQuarantinePath = canonicalTestRestorePath(t, journal.CandidateQuarantinePath)
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
	if runtime.GOOS == "windows" {
		if err := RemoveRestoreJournal(root); !errors.Is(err, errRestoreJournalRemovalUnsupported) {
			t.Fatalf("remove restore journal error = %v, want unsupported", err)
		}
		if _, err := ReadRestoreJournal(root); err != nil {
			t.Fatalf("journal disappeared after unsupported removal: %v", err)
		}
		return
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


func TestRestoreJournalPreservesExplicitNoPreviousDatabase(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.PreviousDatabaseExisted = boolPointer(false)
	journal.RollbackPath = ""
	journal.PreviousSHA256 = ""
	journal.PreviousSize = 0
	if err := WriteRestoreJournal(root, journal); err != nil {
		t.Fatalf("write journal with no previous database: %v", err)
	}
	got, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read journal with no previous database: %v", err)
	}
	if got.PreviousDatabaseExisted == nil || *got.PreviousDatabaseExisted {
		t.Fatalf("previous-database metadata = %#v, want explicit false", got.PreviousDatabaseExisted)
	}
}

func TestReadRestoreJournalRejectsMissingPreviousDatabaseMetadata(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.PreviousDatabaseExisted = nil
	data, err := json.Marshal(journal)
	if err != nil {
		t.Fatalf("marshal journal without previous-database metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, restoreJournalFileName), data, 0o600); err != nil {
		t.Fatalf("write journal without previous-database metadata: %v", err)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("read journal without previous-database metadata error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestReadRestoreJournalRejectsLegacyVersionOne(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.Version = 1
	data, err := json.Marshal(journal)
	if err != nil {
		t.Fatalf("marshal legacy journal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, restoreJournalFileName), data, 0o600); err != nil {
		t.Fatalf("write legacy journal: %v", err)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("read version-one journal error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestReadRestoreJournalRejectsLegacyVersionTwo(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.Version = 2
	data, err := json.Marshal(journal)
	if err != nil {
		t.Fatalf("marshal legacy journal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, restoreJournalFileName), data, 0o600); err != nil {
		t.Fatalf("write legacy journal: %v", err)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("read version-two journal error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestRestoreJournalV3AcceptsIntentAndTerminalPhases(t *testing.T) {
	root := t.TempDir()
	for _, phase := range []RestorePhase{
		RestorePhasePreservingPrevious,
		RestorePhasePublishing,
		RestorePhaseRestoringPrevious,
		RestorePhaseRolledBack,
	} {
		t.Run(string(phase), func(t *testing.T) {
			journal := validTestRestoreJournal(root)
			journal.Phase = phase
			if err := WriteRestoreJournal(root, journal); err != nil {
				t.Fatalf("write v3 phase %q: %v", phase, err)
			}
			got, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatalf("read v3 phase %q: %v", phase, err)
			}
			if got.Phase != phase {
				t.Fatalf("phase = %q, want %q", got.Phase, phase)
			}
		})
	}
}

func TestWriteRestoreJournalRequiresCandidateIdentityBeforePublication(t *testing.T) {
	root := t.TempDir()
	for _, phase := range []RestorePhase{
		RestorePhasePublishing,
		RestorePhasePublished,
		RestorePhaseReopening,
		RestorePhaseRestoringPrevious,
		RestorePhaseCommitted,
		RestorePhaseRolledBack,
	} {
		t.Run(string(phase), func(t *testing.T) {
			journal := validTestRestoreJournal(root)
			journal.Phase = phase
			journal.CandidateSHA256 = ""
			journal.CandidateSize = 0
			if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
				t.Fatalf("write without candidate identity error = %v, want ErrRestoreJournalInvalid", err)
			}
		})
	}
}

func TestWriteRestoreJournalRejectsIncompleteOrInvalidContentIdentities(t *testing.T) {
	root := t.TempDir()
	tests := []struct {
		name string
		mutate func(*RestoreJournal)
	}{
		{name: "candidate digest without size", mutate: func(j *RestoreJournal) { j.CandidateSize = 0 }},
		{name: "candidate size without digest", mutate: func(j *RestoreJournal) { j.CandidateSHA256 = "" }},
		{name: "candidate malformed digest", mutate: func(j *RestoreJournal) { j.CandidateSHA256 = "ABC" }},
		{name: "candidate zero size", mutate: func(j *RestoreJournal) { j.CandidateSize = 0 }},
		{name: "previous digest without size", mutate: func(j *RestoreJournal) { j.PreviousSize = 0 }},
		{name: "previous size without digest", mutate: func(j *RestoreJournal) { j.PreviousSHA256 = "" }},
		{name: "previous malformed digest", mutate: func(j *RestoreJournal) { j.PreviousSHA256 = "not-hex" }},
		{name: "previous zero size", mutate: func(j *RestoreJournal) { j.PreviousSize = 0 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			journal := validTestRestoreJournal(root)
			test.mutate(&journal)
			if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
				t.Fatalf("write invalid content identity error = %v, want ErrRestoreJournalInvalid", err)
			}
		})
	}
}

func TestWriteRestoreJournalRequiresPreviousIdentityIffPreviousDatabaseExists(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.PreviousSHA256 = ""
	journal.PreviousSize = 0
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write without previous identity error = %v, want ErrRestoreJournalInvalid", err)
	}

	journal = validTestRestoreJournal(root)
	journal.PreviousDatabaseExisted = boolPointer(false)
	journal.RollbackPath = ""
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write unexpected previous identity error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestWriteRestoreJournalRejectsCandidateQuarantinePathCollisions(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		filepath.Join(root, "state.db"),
		filepath.Join(root, ".restore-stage.db"),
		filepath.Join(root, ".restore-rollback.db"),
	} {
		journal := validTestRestoreJournal(root)
		journal.CandidateQuarantinePath = path
		if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
			t.Fatalf("write with quarantine collision %q error = %v, want ErrRestoreJournalInvalid", path, err)
		}
	}
}

func TestWriteRestoreJournalRejectsCandidateQuarantineOutsideStateDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.CandidateQuarantinePath = filepath.Join(outside, "candidate.db")
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write with external quarantine path error = %v, want ErrRestoreJournalInvalid", err)
	}
}



func TestWriteRestoreJournalRejectsInconsistentRollbackMetadata(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.RollbackPath = ""
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write with missing rollback path error = %v, want ErrRestoreJournalInvalid", err)
	}

	journal = validTestRestoreJournal(root)
	journal.PreviousDatabaseExisted = boolPointer(false)
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write unexpected rollback path error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestReadRestoreJournalRejectsInconsistentRollbackMetadata(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.RollbackPath = ""
	data, err := json.Marshal(journal)
	if err != nil {
		t.Fatalf("marshal journal without rollback path: %v", err)
	}
	path := filepath.Join(root, restoreJournalFileName)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write journal without rollback path: %v", err)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("read journal without rollback path error = %v, want ErrRestoreJournalInvalid", err)
	}

	journal = validTestRestoreJournal(root)
	journal.PreviousDatabaseExisted = boolPointer(false)
	data, err = json.Marshal(journal)
	if err != nil {
		t.Fatalf("marshal journal with unexpected rollback path: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write journal with unexpected rollback path: %v", err)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("read journal with unexpected rollback path error = %v, want ErrRestoreJournalInvalid", err)
	}
}


func TestWriteRestoreJournalRejectsNonRegularTransactionPaths(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "stage-directory")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatalf("create staging directory: %v", err)
	}
	journal := validTestRestoreJournal(root)
	journal.StagingPath = directory
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("write with directory staging path error = %v, want ErrRestoreJournalInvalid", err)
	}
}
