package sqlite

import (
	"errors"
	"fmt"
	"strings"
)

var ErrRestoreRecoveryAmbiguous = errors.New("SQLite restore recovery state is ambiguous")

type RestoreRecoveryAction string

const (
	RestoreRecoveryCleanupStaging RestoreRecoveryAction = "cleanup-staging"
	RestoreRecoveryRestorePrevious RestoreRecoveryAction = "restore-previous"
	RestoreRecoveryPreserveForManualReview RestoreRecoveryAction = "preserve-for-manual-review"
	RestoreRecoveryVerifyCommitted RestoreRecoveryAction = "verify-committed"
)

// RestoreRecoveryFiles is a read-only snapshot of transaction-owned file
// presence. The planner never performs filesystem operations.
type RestoreRecoveryFiles struct {
	TargetExists   bool
	StagingExists  bool
	RollbackExists bool
}

// RestoreRecoveryPlan is a conservative instruction set for a later executor.
// A manual-review plan deliberately leaves every artifact untouched.
type RestoreRecoveryPlan struct {
	Action          RestoreRecoveryAction
	RemoveStaging   bool
	PreserveTarget  bool
	RestoreRollback bool
	VerifyTarget    bool
	Reason          string
}

// PlanRestoreRecovery decides what may safely happen after an interrupted
// restore. It does not inspect or mutate the filesystem and fails closed when
// the observed files cannot be explained by the durable journal phase.
func PlanRestoreRecovery(journal RestoreJournal, files RestoreRecoveryFiles) (RestoreRecoveryPlan, error) {
	ambiguous := func(reason string) (RestoreRecoveryPlan, error) {
		return RestoreRecoveryPlan{}, fmt.Errorf("%w: %s", ErrRestoreRecoveryAmbiguous, reason)
	}
	if journal.Version != restoreJournalVersion || !validRestorePhase(journal.Phase) ||
		strings.TrimSpace(journal.TransactionID) == "" ||
		strings.TrimSpace(journal.TargetPath) == "" ||
		strings.TrimSpace(journal.StagingPath) == "" {
		return ambiguous("journal version, phase, transaction ID, or required path is invalid")
	}
	if journal.PreviousDatabaseExisted == nil {
		return ambiguous("journal does not record whether a previous database existed")
	}
	hadPrevious := *journal.PreviousDatabaseExisted
	if hadPrevious && journal.RollbackPath == "" {
		return ambiguous("previous database existed but no rollback path is recorded")
	}
	if !hadPrevious && (files.RollbackExists || journal.RollbackPath != "") {
		return ambiguous("journal or filesystem contains an unexpected rollback path")
	}

	switch journal.Phase {
	case RestorePhasePrepared, RestorePhaseStaging:
		if hadPrevious != files.TargetExists || files.RollbackExists {
			return ambiguous("target/rollback presence conflicts with a pre-publication phase")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryCleanupStaging,
			RemoveStaging: files.StagingExists,
			Reason: "publication has not begun; preserve the original target",
		}, nil

	case RestorePhasePreviousPreserved:
		if !hadPrevious || files.TargetExists || !files.RollbackExists {
			return ambiguous("previous-preserved phase requires a missing target and an existing rollback")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryRestorePrevious,
			RemoveStaging: files.StagingExists,
			RestoreRollback: true,
			Reason: "restore the preserved database before opening the store",
		}, nil

	case RestorePhasePublished, RestorePhaseReopening:
		if hadPrevious {
			if !files.RollbackExists {
				return ambiguous("published transaction lost its required rollback database")
			}
			return RestoreRecoveryPlan{
				Action: RestoreRecoveryRestorePrevious,
				RemoveStaging: files.StagingExists,
				PreserveTarget: files.TargetExists,
				RestoreRollback: true,
				Reason: "uncommitted publication must roll back to the preserved database",
			}, nil
		}
		if !files.TargetExists {
			return ambiguous("published transaction has neither a previous database nor a published target")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryPreserveForManualReview,
			Reason: "no previous database existed; preserve the published candidate and require verification",
		}, nil

	case RestorePhaseCommitted:
		if !files.TargetExists || files.StagingExists || hadPrevious != files.RollbackExists {
			return ambiguous("committed transaction has inconsistent target, staging, or rollback files")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryVerifyCommitted,
			VerifyTarget: true,
			Reason: "verify the committed database before journal cleanup or rollback retention changes",
		}, nil

	default:
		return ambiguous("journal phase is unknown")
	}
}
