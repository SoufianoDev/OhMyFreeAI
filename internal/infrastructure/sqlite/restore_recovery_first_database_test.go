package sqlite

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRestoreRecoveryExecutorPreservesFirstDatabaseForManualReview(t *testing.T) {
	tests := []struct {
		name              string
		phase             RestorePhase
		candidateLocation string
	}{
		{name: "prepared before candidate exists", phase: RestorePhasePrepared},
		{name: "staging candidate", phase: RestorePhaseStaging, candidateLocation: "staging"},
		{name: "publishing candidate before move", phase: RestorePhasePublishing, candidateLocation: "staging"},
		{name: "published candidate", phase: RestorePhasePublished, candidateLocation: "target"},
		{name: "published candidate already quarantined", phase: RestorePhasePublished, candidateLocation: "quarantine"},
		{name: "reopening candidate", phase: RestorePhaseReopening, candidateLocation: "target"},
		{name: "committed candidate", phase: RestorePhaseCommitted, candidateLocation: "target"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := restoreTestDir(t)
			targetPath := filepath.Join(root, "state.db")
			stagingPath := filepath.Join(root, ".restore-stage.db")
			quarantinePath := filepath.Join(root, ".restore-candidate-quarantine.db")
			candidate := []byte("first database candidate evidence")
			digest := sha256.Sum256(candidate)

			journal := RestoreJournal{
				Version:                 restoreJournalVersion,
				TransactionID:           "first-database-manual-review",
				Phase:                   RestorePhasePrepared,
				TargetPath:              targetPath,
				StagingPath:             stagingPath,
				CandidateQuarantinePath: quarantinePath,
				PreviousDatabaseExisted: boolPointer(false),
				ArchiveSHA256:           strings.Repeat("a", 64),
				ArchiveSize:             1,
				UpdatedAt:               time.Now().UTC(),
			}
			if err := WriteRestoreJournal(root, journal); err != nil {
				t.Fatalf("write prepared first-database journal: %v", err)
			}

			if test.phase != RestorePhasePrepared {
				journal.Phase = RestorePhaseStaging
				journal.CandidateSHA256 = hex.EncodeToString(digest[:])
				journal.CandidateSize = uint64(len(candidate))
				journal.UpdatedAt = journal.UpdatedAt.Add(time.Second)
				if err := WriteRestoreJournal(root, journal); err != nil {
					t.Fatalf("write staging first-database journal: %v", err)
				}

				if test.phase != RestorePhaseStaging {
					transitions := []RestorePhase{
						RestorePhasePublishing,
						RestorePhasePublished,
						RestorePhaseReopening,
						RestorePhaseCommitted,
					}
					for _, phase := range transitions {
						journal.Phase = phase
						journal.UpdatedAt = journal.UpdatedAt.Add(time.Second)
						if err := WriteRestoreJournal(root, journal); err != nil {
							t.Fatalf("advance first-database journal to %q: %v", phase, err)
						}
						if phase == test.phase {
							break
						}
					}
				}
			}

			candidatePath := ""
			switch test.candidateLocation {
			case "staging":
				candidatePath = stagingPath
			case "target":
				candidatePath = targetPath
			case "quarantine":
				candidatePath = quarantinePath
			}
			if candidatePath != "" {
				if err := os.WriteFile(candidatePath, candidate, 0o600); err != nil {
					t.Fatalf("write candidate evidence at %s: %v", candidatePath, err)
				}
			}

			before, err := inspectRestoreRecoveryFiles(root, journal)
			if err != nil {
				t.Fatalf("inspect first-database artifacts before recovery: %v", err)
			}
			executor, err := NewRestoreRecoveryExecutor(root)
			if err != nil {
				t.Fatalf("create recovery executor: %v", err)
			}

			openCalls, moveCalls, journalWrites := 0, 0, 0
			realWrite := executor.operations.writeJournal
			executor.operations.openStore = func(context.Context, string) (*Store, error) {
				openCalls++
				return nil, errors.New("first-database recovery must not open SQLite")
			}
			executor.operations.move = func(string, string, RestoreContentIdentity) error {
				moveCalls++
				return errors.New("first-database recovery must not move candidate evidence")
			}
			executor.operations.writeJournal = func(lease *restoreJournalLease, next RestoreJournal, boundary string) error {
				journalWrites++
				return realWrite(lease, next, boundary)
			}

			store, err := executor.Recover(context.Background())
			if store != nil {
				_ = store.Close()
				t.Fatal("first-database recovery returned a Store before manual review")
			}
			if !errors.Is(err, ErrRestoreRecoveryManualReview) {
				t.Fatalf("Recover() error = %v, want ErrRestoreRecoveryManualReview", err)
			}
			if openCalls != 0 || moveCalls != 0 || journalWrites != 0 {
				t.Fatalf("manual-review recovery performed operations: open=%d move=%d journal-write=%d", openCalls, moveCalls, journalWrites)
			}

			current, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatalf("read retained first-database journal: %v", err)
			}
			if current.Phase != test.phase {
				t.Fatalf("first-database journal phase changed to %q, want %q", current.Phase, test.phase)
			}
			after, err := inspectRestoreRecoveryFiles(root, current)
			if err != nil {
				t.Fatalf("inspect first-database artifacts after manual review: %v", err)
			}
			if after != before {
				t.Fatalf("first-database artifact identities changed during manual review: before=%#v after=%#v", before, after)
			}
			if candidatePath != "" {
				actual, err := os.ReadFile(candidatePath)
				if err != nil {
					t.Fatalf("read preserved candidate evidence: %v", err)
				}
				if string(actual) != string(candidate) {
					t.Fatalf("candidate evidence changed: got %q, want %q", actual, candidate)
				}
			}
		})
	}
}
