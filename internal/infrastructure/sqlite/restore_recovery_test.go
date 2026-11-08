package sqlite

import (
	"errors"
	"path/filepath"
	"testing"
)

func recoveryTestJournal(phase RestorePhase, hadPrevious bool) RestoreJournal {
	root := filepath.Clean("state")
	journal := RestoreJournal{
		Version: restoreJournalVersion,
		TransactionID: "recovery-test",
		Phase: phase,
		TargetPath: filepath.Join(root, "state.db"),
		StagingPath: filepath.Join(root, ".stage.db"),
		PreviousDatabaseExisted: boolPointer(hadPrevious),
		ArchiveSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ArchiveSize: 100,
	}
	if hadPrevious {
		journal.RollbackPath = filepath.Join(root, ".rollback.db")
	}
	return journal
}

func TestPlanRestoreRecoveryConservativeDecisions(t *testing.T) {
	tests := []struct {
		name string
		phase RestorePhase
		hadPrevious bool
		files RestoreRecoveryFiles
		wantAction RestoreRecoveryAction
		wantRemoveStaging bool
		wantPreserveTarget bool
		wantRestoreRollback bool
		wantVerifyTarget bool
	}{
		{
			name: "prepared preserves existing database and cleans stage",
			phase: RestorePhasePrepared, hadPrevious: true,
			files: RestoreRecoveryFiles{TargetExists: true, StagingExists: true},
			wantAction: RestoreRecoveryCleanupStaging, wantRemoveStaging: true,
		},
		{
			name: "staging without prior database keeps target absent",
			phase: RestorePhaseStaging, hadPrevious: false,
			files: RestoreRecoveryFiles{StagingExists: true},
			wantAction: RestoreRecoveryCleanupStaging, wantRemoveStaging: true,
		},
		{
			name: "previous preserved restores rollback",
			phase: RestorePhasePreviousPreserved, hadPrevious: true,
			files: RestoreRecoveryFiles{StagingExists: true, RollbackExists: true},
			wantAction: RestoreRecoveryRestorePrevious, wantRemoveStaging: true, wantRestoreRollback: true,
		},
		{
			name: "published with prior target preserves candidate then rolls back",
			phase: RestorePhasePublished, hadPrevious: true,
			files: RestoreRecoveryFiles{TargetExists: true, RollbackExists: true},
			wantAction: RestoreRecoveryRestorePrevious, wantPreserveTarget: true, wantRestoreRollback: true,
		},
		{
			name: "reopening with missing candidate still restores rollback",
			phase: RestorePhaseReopening, hadPrevious: true,
			files: RestoreRecoveryFiles{RollbackExists: true},
			wantAction: RestoreRecoveryRestorePrevious, wantRestoreRollback: true,
		},
		{
			name: "published without prior database preserves candidate for manual review",
			phase: RestorePhasePublished, hadPrevious: false,
			files: RestoreRecoveryFiles{TargetExists: true},
			wantAction: RestoreRecoveryPreserveForManualReview,
		},
		{
			name: "committed with prior database requires verification and retains rollback",
			phase: RestorePhaseCommitted, hadPrevious: true,
			files: RestoreRecoveryFiles{TargetExists: true, RollbackExists: true},
			wantAction: RestoreRecoveryVerifyCommitted, wantVerifyTarget: true,
		},
		{
			name: "committed first database requires verification",
			phase: RestorePhaseCommitted, hadPrevious: false,
			files: RestoreRecoveryFiles{TargetExists: true},
			wantAction: RestoreRecoveryVerifyCommitted, wantVerifyTarget: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			journal := recoveryTestJournal(test.phase, test.hadPrevious)
			plan, err := PlanRestoreRecovery(journal, test.files)
			if err != nil {
				t.Fatalf("PlanRestoreRecovery() error = %v", err)
			}
			if plan.Action != test.wantAction ||
				plan.RemoveStaging != test.wantRemoveStaging ||
				plan.PreserveTarget != test.wantPreserveTarget ||
				plan.RestoreRollback != test.wantRestoreRollback ||
				plan.VerifyTarget != test.wantVerifyTarget {
				t.Fatalf("PlanRestoreRecovery() = %#v", plan)
			}
		})
	}
}

func TestPlanRestoreRecoveryFailsClosedOnAmbiguousStates(t *testing.T) {
	tests := []struct {
		name string
		journal RestoreJournal
		files RestoreRecoveryFiles
	}{
		{
			name: "missing previous database during pre-publication",
			journal: recoveryTestJournal(RestorePhaseStaging, true),
			files: RestoreRecoveryFiles{StagingExists: true, RollbackExists: true},
		},
		{
			name: "rollback missing after publication",
			journal: recoveryTestJournal(RestorePhasePublished, true),
			files: RestoreRecoveryFiles{TargetExists: true},
		},
		{
			name: "target unexpectedly present after previous-preserved",
			journal: recoveryTestJournal(RestorePhasePreviousPreserved, true),
			files: RestoreRecoveryFiles{TargetExists: true, RollbackExists: true},
		},
		{
			name: "first publication missing target",
			journal: recoveryTestJournal(RestorePhasePublished, false),
			files: RestoreRecoveryFiles{},
		},
		{
			name: "committed database missing",
			journal: recoveryTestJournal(RestorePhaseCommitted, false),
			files: RestoreRecoveryFiles{StagingExists: true},
		},
		{
			name: "unknown phase",
			journal: recoveryTestJournal(RestorePhase("unknown"), false),
			files: RestoreRecoveryFiles{},
		},
		{
			name: "unsupported journal version",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.Version++
				return j
			}(),
			files: RestoreRecoveryFiles{},
		},
		{
			name: "empty target path",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.TargetPath = ""
				return j
			}(),
			files: RestoreRecoveryFiles{},
		},
		{
			name: "missing prior database metadata",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.PreviousDatabaseExisted = nil
				return j
			}(),
			files: RestoreRecoveryFiles{},
		},
		{
			name: "unexpected rollback path for first database",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.RollbackPath = filepath.Join("state", ".unexpected-rollback.db")
				return j
			}(),
			files: RestoreRecoveryFiles{TargetExists: false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := PlanRestoreRecovery(test.journal, test.files)
			if !errors.Is(err, ErrRestoreRecoveryAmbiguous) {
				t.Fatalf("PlanRestoreRecovery() = %#v, %v; want ErrRestoreRecoveryAmbiguous", plan, err)
			}
			if plan != (RestoreRecoveryPlan{}) {
				t.Fatalf("ambiguous state returned an actionable plan: %#v", plan)
			}
		})
	}
}


func TestPlanRestoreRecoveryRejectsCollidingPaths(t *testing.T) {
	tests := []struct {
		name string
		journal RestoreJournal
	}{
		{
			name: "target equals staging",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.StagingPath = j.TargetPath
				return j
			}(),
		},
		{
			name: "rollback equals target",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePublished, true)
				j.RollbackPath = j.TargetPath
				return j
			}(),
		},
		{
			name: "rollback equals staging",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePublished, true)
				j.RollbackPath = j.StagingPath
				return j
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := PlanRestoreRecovery(test.journal, RestoreRecoveryFiles{TargetExists: true, RollbackExists: true})
			if !errors.Is(err, ErrRestoreRecoveryAmbiguous) {
				t.Fatalf("PlanRestoreRecovery() = %#v, %v; want ErrRestoreRecoveryAmbiguous", plan, err)
			}
			if plan != (RestoreRecoveryPlan{}) {
				t.Fatalf("colliding paths returned actionable plan: %#v", plan)
			}
		})
	}
}

func TestPlanRestoreRecoveryRequiresVerifiedFileIdentities(t *testing.T) {
	journal := recoveryTestJournal(RestorePhasePublished, true)
	journal.CandidateSHA256 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	journal.CandidateSize = 200
	journal.PreviousSHA256 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	journal.PreviousSize = 300

	tests := []struct {
		name  string
		files RestoreRecoveryFiles
		want  RestoreRecoveryAction
		wantErr bool
	}{
		{
			name: "verified candidate and previous permit rollback plan",
			files: RestoreRecoveryFiles{
				Target: RestoreFileCandidate,
				Rollback: RestoreFilePrevious,
				Staging: RestoreFileMissing,
				Quarantine: RestoreFileMissing,
			},
			want: RestoreRecoveryRestorePrevious,
		},
		{
			name: "candidate digest mismatch is ambiguous",
			files: RestoreRecoveryFiles{
				Target: RestoreFileMismatch,
				Rollback: RestoreFilePrevious,
				Staging: RestoreFileMissing,
				Quarantine: RestoreFileMissing,
			},
			wantErr: true,
		},
		{
			name: "previous digest mismatch is ambiguous",
			files: RestoreRecoveryFiles{
				Target: RestoreFileCandidate,
				Rollback: RestoreFileMismatch,
				Staging: RestoreFileMissing,
				Quarantine: RestoreFileMissing,
			},
			wantErr: true,
		},
		{
			name: "unreadable candidate is ambiguous",
			files: RestoreRecoveryFiles{
				Target: RestoreFileUnreadable,
				Rollback: RestoreFilePrevious,
				Staging: RestoreFileMissing,
				Quarantine: RestoreFileMissing,
			},
			wantErr: true,
		},
		{
			name: "unverified staging artifact cannot be deleted",
			files: RestoreRecoveryFiles{
				Target: RestoreFilePrevious,
				Staging: RestoreFileUnverified,
				Rollback: RestoreFileMissing,
				Quarantine: RestoreFileMissing,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			phaseJournal := journal
			if test.name == "unverified staging artifact cannot be deleted" {
				phaseJournal = recoveryTestJournal(RestorePhaseStaging, true)
			}
			plan, err := PlanRestoreRecovery(phaseJournal, test.files)
			if test.wantErr {
				if !errors.Is(err, ErrRestoreRecoveryAmbiguous) {
					t.Fatalf("PlanRestoreRecovery() = %#v, %v; want ambiguity", plan, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("PlanRestoreRecovery() error = %v", err)
			}
			if plan.Action != test.want {
				t.Fatalf("PlanRestoreRecovery() action = %q; want %q", plan.Action, test.want)
			}
		})
	}
}
