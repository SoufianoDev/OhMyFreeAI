package sqlite

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func validV3TestJournal(root string) RestoreJournal {
	return RestoreJournal{
		Version: restoreJournalVersion,
		TransactionID: "v3-transaction",
		Phase: RestorePhasePrepared,
		TargetPath: filepath.Join(root, "state.db"),
		StagingPath: filepath.Join(root, ".stage.db"),
		RollbackPath: filepath.Join(root, ".rollback.db"),
		CandidateQuarantinePath: filepath.Join(root, ".candidate-quarantine.db"),
		PreviousDatabaseExisted: boolPointer(true),
		ArchiveSHA256: strings.Repeat("a", 64),
		ArchiveSize: 4096,
		CandidateSHA256: strings.Repeat("b", 64),
		CandidateSize: 8192,
		PreviousSHA256: strings.Repeat("c", 64),
		PreviousSize: 16384,
		UpdatedAt: time.Now().UTC(),
	}
}

func TestRestoreJournalV3RoundTripPersistsIntentAndContentIdentities(t *testing.T) {
	root := restoreTestDir(t)
	journal := validV3TestJournal(root)
	if err := WriteRestoreJournal(root, journal); err != nil {
		t.Fatalf("write v3 prepared journal: %v", err)
	}
	journal.Phase = RestorePhaseStaging
	journal.UpdatedAt = journal.UpdatedAt.Add(time.Second)
	if err := WriteRestoreJournal(root, journal); err != nil {
		t.Fatalf("write v3 staging journal: %v", err)
	}
	journal.Phase = RestorePhasePreservingPrevious
	journal.UpdatedAt = journal.UpdatedAt.Add(time.Second)
	if err := WriteRestoreJournal(root, journal); err != nil {
		t.Fatalf("write v3 intent journal: %v", err)
	}
	got, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read v3 intent journal: %v", err)
	}
	if got.Version != 3 || got.Phase != RestorePhasePreservingPrevious {
		t.Fatalf("journal version/phase = %d/%q, want 3/preserving-previous", got.Version, got.Phase)
	}
	if got.CandidateSHA256 != journal.CandidateSHA256 || got.CandidateSize != journal.CandidateSize ||
		got.PreviousSHA256 != journal.PreviousSHA256 || got.PreviousSize != journal.PreviousSize {
		t.Fatalf("journal content identities were not preserved: %#v", got)
	}
	wantQuarantinePath := canonicalTestRestorePath(t, filepath.Join(root, ".candidate-quarantine.db"))
	if got.CandidateQuarantinePath != wantQuarantinePath {
		t.Fatalf("quarantine path = %q", got.CandidateQuarantinePath)
	}
}

func TestRestoreJournalV3RejectsVersionTwo(t *testing.T) {
	root := restoreTestDir(t)
	journal := validV3TestJournal(root)
	journal.Version = 2
	data, err := json.Marshal(journal)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, restoreJournalFileName), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("read v2 journal error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestRestoreJournalV3RejectsIdentityPairsAndInvalidDigests(t *testing.T) {
	tests := []struct {
		name string
		mutate func(*RestoreJournal)
	}{
		{"candidate digest without size", func(j *RestoreJournal) { j.CandidateSize = 0 }},
		{"candidate size without digest", func(j *RestoreJournal) { j.CandidateSHA256 = "" }},
		{"candidate malformed digest", func(j *RestoreJournal) { j.CandidateSHA256 = "xyz" }},
		{"candidate uppercase digest", func(j *RestoreJournal) { j.CandidateSHA256 = strings.Repeat("B", 64) }},
		{"previous digest without size", func(j *RestoreJournal) { j.PreviousSize = 0 }},
		{"previous size without digest", func(j *RestoreJournal) { j.PreviousSHA256 = "" }},
		{"previous malformed digest", func(j *RestoreJournal) { j.PreviousSHA256 = "xyz" }},
		{"previous identity with no previous database", func(j *RestoreJournal) { j.PreviousDatabaseExisted = boolPointer(false); j.RollbackPath = "" }},
		{"previous database without identity", func(j *RestoreJournal) { j.PreviousSHA256 = ""; j.PreviousSize = 0 }},
		{"candidate identity missing before publishing", func(j *RestoreJournal) { j.Phase = RestorePhasePublishing; j.CandidateSHA256 = ""; j.CandidateSize = 0 }},
		{"candidate identity missing after publication", func(j *RestoreJournal) { j.Phase = RestorePhaseCommitted; j.CandidateSHA256 = ""; j.CandidateSize = 0 }},
		{"previous identity missing before preservation intent", func(j *RestoreJournal) { j.Phase = RestorePhasePreservingPrevious; j.PreviousSHA256 = ""; j.PreviousSize = 0 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := restoreTestDir(t)
			journal := validV3TestJournal(root)
			test.mutate(&journal)
			if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
				t.Fatalf("WriteRestoreJournal() error = %v, want ErrRestoreJournalInvalid", err)
			}
		})
	}
}

func TestRestoreJournalV3RejectsQuarantinePathCollisions(t *testing.T) {
	root := restoreTestDir(t)
	journal := validV3TestJournal(root)
	journal.CandidateQuarantinePath = journal.TargetPath
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("quarantine/target collision error = %v, want ErrRestoreJournalInvalid", err)
	}
	journal = validV3TestJournal(root)
	journal.CandidateQuarantinePath = journal.StagingPath
	if err := WriteRestoreJournal(root, journal); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("quarantine/staging collision error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestRestoreJournalV3FirstDatabaseOmitsPreviousIdentity(t *testing.T) {
	root := restoreTestDir(t)
	journal := validV3TestJournal(root)
	journal.PreviousDatabaseExisted = boolPointer(false)
	journal.RollbackPath = ""
	journal.PreviousSHA256 = ""
	journal.PreviousSize = 0
	if err := WriteRestoreJournal(root, journal); err != nil {
		t.Fatalf("write first-database journal: %v", err)
	}
	if _, err := ReadRestoreJournal(root); err != nil {
		t.Fatalf("read first-database journal: %v", err)
	}
}
