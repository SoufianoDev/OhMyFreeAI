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
		CandidateQuarantinePath: filepath.Join(root, ".candidate-quarantine.db"),
		PreviousDatabaseExisted: boolPointer(hadPrevious),
		CandidateSHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		CandidateSize: 200,
		ArchiveSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ArchiveSize: 100,
	}
	if hadPrevious {
		journal.RollbackPath = filepath.Join(root, ".rollback.db")
		journal.PreviousSHA256 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
		journal.PreviousSize = 300
	}
	return journal
}

func recoveryFiles(target, staging, rollback, quarantine RestoreFileIdentity) RestoreRecoveryFiles {
	return RestoreRecoveryFiles{Target: target, Staging: staging, Rollback: rollback, Quarantine: quarantine}
}

func TestPlanRestoreRecoveryConservativeDecisions(t *testing.T) {
	tests := []struct {
		name string
		phase RestorePhase
		hadPrevious bool
		files RestoreRecoveryFiles
		wantAction RestoreRecoveryAction
		wantRemoveStaging bool
		wantPreserveStaging bool
		wantPreserveTarget bool
		wantRestoreRollback bool
		wantVerifyTarget bool
	}{
		{
			name: "prepared preserves previous database and removes verified stage",
			phase: RestorePhasePrepared, hadPrevious: true,
			files: recoveryFiles(RestoreFilePrevious, RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing),
			wantAction: RestoreRecoveryCleanupStaging, wantRemoveStaging: true,
		},
		{
			name: "staging first database keeps target absent",
			phase: RestorePhaseStaging, hadPrevious: false,
			files: recoveryFiles(RestoreFileMissing, RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing),
			wantAction: RestoreRecoveryCleanupStaging, wantRemoveStaging: true,
		},
		{
			name: "previous preserved retains stage before restoring rollback",
			phase: RestorePhasePreviousPreserved, hadPrevious: true,
			files: recoveryFiles(RestoreFileMissing, RestoreFileCandidate, RestoreFilePrevious, RestoreFileMissing),
			wantAction: RestoreRecoveryRestorePrevious, wantPreserveStaging: true, wantRestoreRollback: true,
		},
		{
			name: "published with prior database quarantines candidate then rolls back",
			phase: RestorePhasePublished, hadPrevious: true,
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFilePrevious, RestoreFileMissing),
			wantAction: RestoreRecoveryRestorePrevious, wantPreserveTarget: true, wantRestoreRollback: true,
		},
		{
			name: "reopening after candidate quarantine restores rollback",
			phase: RestorePhaseReopening, hadPrevious: true,
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFilePrevious, RestoreFileCandidate),
			wantAction: RestoreRecoveryRestorePrevious, wantRestoreRollback: true,
		},
		{
			name: "published without prior database preserves candidate for manual review",
			phase: RestorePhasePublished, hadPrevious: false,
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
			wantAction: RestoreRecoveryPreserveForManualReview,
		},
		{
			name: "committed with prior database verifies candidate and retains rollback",
			phase: RestorePhaseCommitted, hadPrevious: true,
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFilePrevious, RestoreFileMissing),
			wantAction: RestoreRecoveryVerifyCommitted, wantVerifyTarget: true,
		},
		{
			name: "committed first database verifies candidate",
			phase: RestorePhaseCommitted, hadPrevious: false,
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
			wantAction: RestoreRecoveryVerifyCommitted, wantVerifyTarget: true,
		},
		{
			name: "rolled back verifies previous target and retains candidate quarantine",
			phase: RestorePhaseRolledBack, hadPrevious: true,
			files: recoveryFiles(RestoreFilePrevious, RestoreFileMissing, RestoreFileMissing, RestoreFileCandidate),
			wantAction: RestoreRecoveryVerifyRolledBack, wantVerifyTarget: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := PlanRestoreRecovery(recoveryTestJournal(test.phase, test.hadPrevious), test.files)
			if err != nil {
				t.Fatalf("PlanRestoreRecovery() error = %v", err)
			}
			if plan.Action != test.wantAction ||
				plan.RemoveStaging != test.wantRemoveStaging ||
				plan.PreserveStaging != test.wantPreserveStaging ||
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
			files: recoveryFiles(RestoreFileMissing, RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "rollback missing after publication",
			journal: recoveryTestJournal(RestorePhasePublished, true),
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "target unexpectedly present after previous-preserved",
			journal: recoveryTestJournal(RestorePhasePreviousPreserved, true),
			files: recoveryFiles(RestoreFileCandidate, RestoreFileCandidate, RestoreFilePrevious, RestoreFileMissing),
		},
		{
			name: "first publication missing candidate",
			journal: recoveryTestJournal(RestorePhasePublished, false),
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "committed candidate missing",
			journal: recoveryTestJournal(RestorePhaseCommitted, false),
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "unknown phase",
			journal: recoveryTestJournal(RestorePhase("unknown"), false),
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "unsupported journal version",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.Version++
				return j
			}(),
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "empty target path",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.TargetPath = ""
				return j
			}(),
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "missing previous metadata",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.PreviousDatabaseExisted = nil
				return j
			}(),
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "unexpected rollback path for first database",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.RollbackPath = filepath.Join("state", ".unexpected-rollback.db")
				return j
			}(),
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "unverified partial staging artifact is never deleted",
			journal: recoveryTestJournal(RestorePhaseStaging, true),
			files: recoveryFiles(RestoreFilePrevious, RestoreFileUnverified, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "candidate content mismatch",
			journal: recoveryTestJournal(RestorePhasePublished, true),
			files: recoveryFiles(RestoreFileMismatch, RestoreFilePrevious, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "previous content mismatch",
			journal: recoveryTestJournal(RestorePhasePublished, true),
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMismatch, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "unreadable candidate",
			journal: recoveryTestJournal(RestorePhasePublished, true),
			files: recoveryFiles(RestoreFileUnreadable, RestoreFilePrevious, RestoreFileMissing, RestoreFileMissing),
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
		{
			name: "quarantine equals target",
			journal: func() RestoreJournal {
				j := recoveryTestJournal(RestorePhasePrepared, false)
				j.CandidateQuarantinePath = j.TargetPath
				return j
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := PlanRestoreRecovery(test.journal, recoveryFiles(
				RestoreFilePrevious, RestoreFileMissing, RestoreFilePrevious, RestoreFileMissing,
			))
			if !errors.Is(err, ErrRestoreRecoveryAmbiguous) {
				t.Fatalf("PlanRestoreRecovery() = %#v, %v; want ErrRestoreRecoveryAmbiguous", plan, err)
			}
			if plan != (RestoreRecoveryPlan{}) {
				t.Fatalf("colliding paths returned actionable plan: %#v", plan)
			}
		})
	}
}


func TestPlanRestoreRecoveryIntentPhaseTruthTables(t *testing.T) {
	tests := []struct {
		name string
		phase RestorePhase
		hadPrevious bool
		files RestoreRecoveryFiles
		wantPreservePrevious bool
		wantPublishCandidate bool
		wantPreserveTarget bool
		wantPreserveStaging bool
		wantRestoreRollback bool
	}{
		{
			name: "preserving previous before move",
			phase: RestorePhasePreservingPrevious, hadPrevious: true,
			files: recoveryFiles(RestoreFilePrevious, RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing),
			wantPreservePrevious: true,
		},
		{
			name: "preserving previous after move",
			phase: RestorePhasePreservingPrevious, hadPrevious: true,
			files: recoveryFiles(RestoreFileMissing, RestoreFileCandidate, RestoreFilePrevious, RestoreFileMissing),
		},
		{
			name: "publishing with previous database before move",
			phase: RestorePhasePublishing, hadPrevious: true,
			files: recoveryFiles(RestoreFileMissing, RestoreFileCandidate, RestoreFilePrevious, RestoreFileMissing),
			wantPublishCandidate: true,
		},
		{
			name: "publishing with previous database after move",
			phase: RestorePhasePublishing, hadPrevious: true,
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFilePrevious, RestoreFileMissing),
		},
		{
			name: "first publication before move",
			phase: RestorePhasePublishing, hadPrevious: false,
			files: recoveryFiles(RestoreFileMissing, RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing),
			wantPublishCandidate: true,
		},
		{
			name: "first publication after move",
			phase: RestorePhasePublishing, hadPrevious: false,
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "restoring previous before candidate quarantine",
			phase: RestorePhaseRestoringPrevious, hadPrevious: true,
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFilePrevious, RestoreFileMissing),
			wantPreserveTarget: true, wantRestoreRollback: true,
		},
		{
			name: "restoring previous preserves staged candidate before quarantine",
			phase: RestorePhaseRestoringPrevious, hadPrevious: true,
			files: recoveryFiles(RestoreFileMissing, RestoreFileCandidate, RestoreFilePrevious, RestoreFileMissing),
			wantPreserveStaging: true, wantRestoreRollback: true,
		},
		{
			name: "restoring previous after candidate quarantine",
			phase: RestorePhaseRestoringPrevious, hadPrevious: true,
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFilePrevious, RestoreFileCandidate),
			wantRestoreRollback: true,
		},
		{
			name: "restoring previous after rollback move",
			phase: RestorePhaseRestoringPrevious, hadPrevious: true,
			files: recoveryFiles(RestoreFilePrevious, RestoreFileMissing, RestoreFileMissing, RestoreFileCandidate),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := PlanRestoreRecovery(recoveryTestJournal(test.phase, test.hadPrevious), test.files)
			if err != nil {
				t.Fatalf("PlanRestoreRecovery() error = %v", err)
			}
			if plan.Action != RestoreRecoveryResumeIntent ||
				plan.PreservePrevious != test.wantPreservePrevious ||
				plan.PublishCandidate != test.wantPublishCandidate ||
				plan.PreserveTarget != test.wantPreserveTarget ||
				plan.PreserveStaging != test.wantPreserveStaging ||
				plan.RestoreRollback != test.wantRestoreRollback {
				t.Fatalf("PlanRestoreRecovery() = %#v", plan)
			}
		})
	}
}

func TestPlanRestoreRecoveryIntentPhasesRejectContradictoryStates(t *testing.T) {
	tests := []struct {
		name string
		journal RestoreJournal
		files RestoreRecoveryFiles
	}{
		{
			name: "preserving previous has both target and rollback",
			journal: recoveryTestJournal(RestorePhasePreservingPrevious, true),
			files: recoveryFiles(RestoreFilePrevious, RestoreFileCandidate, RestoreFilePrevious, RestoreFileMissing),
		},
		{
			name: "preserving previous has neither target nor rollback",
			journal: recoveryTestJournal(RestorePhasePreservingPrevious, true),
			files: recoveryFiles(RestoreFileMissing, RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "publishing candidate appears at both paths",
			journal: recoveryTestJournal(RestorePhasePublishing, true),
			files: recoveryFiles(RestoreFileCandidate, RestoreFileCandidate, RestoreFilePrevious, RestoreFileMissing),
		},
		{
			name: "publishing candidate missing",
			journal: recoveryTestJournal(RestorePhasePublishing, false),
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name: "restoring previous candidate appears at staging and quarantine",
			journal: recoveryTestJournal(RestorePhaseRestoringPrevious, true),
			files: recoveryFiles(RestoreFileMissing, RestoreFileCandidate, RestoreFilePrevious, RestoreFileCandidate),
		},
		{
			name: "restoring previous candidate lost",
			journal: recoveryTestJournal(RestorePhaseRestoringPrevious, true),
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFilePrevious, RestoreFileMissing),
		},
		{
			name: "restoring previous has previous at both paths",
			journal: recoveryTestJournal(RestorePhaseRestoringPrevious, true),
			files: recoveryFiles(RestoreFilePrevious, RestoreFileMissing, RestoreFilePrevious, RestoreFileCandidate),
		},
		{
			name: "first database cannot enter restoring previous",
			journal: recoveryTestJournal(RestorePhaseRestoringPrevious, false),
			files: recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := PlanRestoreRecovery(test.journal, test.files)
			if !errors.Is(err, ErrRestoreRecoveryAmbiguous) {
				t.Fatalf("PlanRestoreRecovery() = %#v, %v; want ErrRestoreRecoveryAmbiguous", plan, err)
			}
			if plan != (RestoreRecoveryPlan{}) {
				t.Fatalf("contradictory state returned an actionable plan: %#v", plan)
			}
		})
	}
}
