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
	RestoreRecoveryVerifyRolledBack RestoreRecoveryAction = "verify-rolled-back"
	RestoreRecoveryResumeIntent RestoreRecoveryAction = "resume-intent"
)

type RestoreFileIdentity string

const (
	RestoreFileUnverified RestoreFileIdentity = "unverified"
	RestoreFileMissing RestoreFileIdentity = "missing"
	RestoreFileCandidate RestoreFileIdentity = "candidate"
	RestoreFilePrevious RestoreFileIdentity = "previous"
	RestoreFileMismatch RestoreFileIdentity = "mismatch"
	RestoreFileUnreadable RestoreFileIdentity = "unreadable"
)

// RestoreRecoveryFiles contains identities verified against the journal, not
// mere path existence. The executor must hash each regular file before planning.
type RestoreRecoveryFiles struct {
	Target RestoreFileIdentity
	Staging RestoreFileIdentity
	Rollback RestoreFileIdentity
	Quarantine RestoreFileIdentity
}

// RestoreRecoveryPlan is a conservative instruction set for a later executor.
// A manual-review plan deliberately leaves every artifact untouched.
type RestoreRecoveryPlan struct {
	Action RestoreRecoveryAction
	RemoveStaging bool
	PreserveStaging bool
	PreservePrevious bool
	PublishCandidate bool
	PreserveTarget bool
	RestoreRollback bool
	VerifyTarget bool
	Reason string
}

// PlanRestoreRecovery decides what may safely happen after an interrupted
// restore. It never infers ownership from file presence alone.
func PlanRestoreRecovery(journal RestoreJournal, files RestoreRecoveryFiles) (RestoreRecoveryPlan, error) {
	ambiguous := func(reason string) (RestoreRecoveryPlan, error) {
		return RestoreRecoveryPlan{}, fmt.Errorf("%w: %s", ErrRestoreRecoveryAmbiguous, reason)
	}
	if journal.Version != restoreJournalVersion || !validRestorePhase(journal.Phase) ||
		strings.TrimSpace(journal.TransactionID) == "" ||
		strings.TrimSpace(journal.TargetPath) == "" ||
		strings.TrimSpace(journal.StagingPath) == "" ||
		strings.TrimSpace(journal.CandidateQuarantinePath) == "" {
		return ambiguous("journal version, phase, transaction ID, or required path is invalid")
	}
	if journal.PreviousDatabaseExisted == nil {
		return ambiguous("journal does not record whether a previous database existed")
	}
	if samePath(journal.TargetPath, journal.StagingPath) ||
		samePath(journal.TargetPath, journal.CandidateQuarantinePath) ||
		samePath(journal.StagingPath, journal.CandidateQuarantinePath) {
		return ambiguous("transaction paths collide")
	}
	hadPrevious := *journal.PreviousDatabaseExisted
	if hadPrevious {
		if strings.TrimSpace(journal.RollbackPath) == "" ||
			samePath(journal.RollbackPath, journal.TargetPath) ||
			samePath(journal.RollbackPath, journal.StagingPath) ||
			samePath(journal.RollbackPath, journal.CandidateQuarantinePath) {
			return ambiguous("previous database has a missing or colliding rollback path")
		}
		if err := validateRestoreContentIdentity(journal.PreviousSHA256, journal.PreviousSize, "previous database", true); err != nil {
			return ambiguous("previous database identity is invalid")
		}
	} else if journal.RollbackPath != "" || journal.PreviousSHA256 != "" || journal.PreviousSize != 0 ||
		files.Rollback != RestoreFileMissing {
		return ambiguous("unexpected previous database identity or rollback artifact")
	}
	candidateRequired := journal.Phase == RestorePhasePublishing ||
		journal.Phase == RestorePhasePublished ||
		journal.Phase == RestorePhaseReopening ||
		journal.Phase == RestorePhaseRestoringPrevious ||
		journal.Phase == RestorePhaseCommitted ||
		journal.Phase == RestorePhaseRolledBack
	if candidateRequired || journal.CandidateSHA256 != "" || journal.CandidateSize != 0 {
		if err := validateRestoreContentIdentity(journal.CandidateSHA256, journal.CandidateSize, "candidate", true); err != nil {
			return ambiguous("candidate identity is invalid")
		}
	}
	for name, state := range map[string]RestoreFileIdentity{
		"target": files.Target, "staging": files.Staging,
		"rollback": files.Rollback, "quarantine": files.Quarantine,
	} {
		switch state {
		case RestoreFileMissing, RestoreFileCandidate, RestoreFilePrevious,
			RestoreFileMismatch, RestoreFileUnreadable, RestoreFileUnverified:
		default:
			return ambiguous(name + " file identity state is invalid")
		}
		if state == RestoreFileMismatch || state == RestoreFileUnreadable || state == RestoreFileUnverified {
			return ambiguous(name + " file is present but its identity is not verified")
		}
	}
	if files.Target == RestoreFileCandidate && journal.CandidateSHA256 == "" {
		return ambiguous("target candidate identity is not recorded")
	}
	if files.Staging == RestoreFileCandidate && journal.CandidateSHA256 == "" {
		return ambiguous("staging candidate identity is not recorded")
	}
	if files.Quarantine == RestoreFileCandidate && journal.CandidateSHA256 == "" {
		return ambiguous("quarantined candidate identity is not recorded")
	}
	if files.Target == RestoreFilePrevious && !hadPrevious {
		return ambiguous("unexpected previous database at target")
	}
	if files.Rollback == RestoreFilePrevious && !hadPrevious {
		return ambiguous("unexpected previous database at rollback")
	}

	switch journal.Phase {
	case RestorePhasePrepared, RestorePhaseStaging:
		wantTarget := RestoreFileMissing
		if hadPrevious {
			wantTarget = RestoreFilePrevious
		}
		if files.Target != wantTarget || files.Rollback != RestoreFileMissing ||
			files.Quarantine != RestoreFileMissing ||
			(files.Staging != RestoreFileMissing && files.Staging != RestoreFileCandidate) {
			return ambiguous("verified file identities conflict with a pre-publication phase")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryCleanupStaging,
			RemoveStaging: files.Staging == RestoreFileCandidate,
			Reason: "publication has not begun; preserve the verified original target",
		}, nil

	case RestorePhasePreviousPreserved:
		if !hadPrevious || files.Target != RestoreFileMissing ||
			files.Rollback != RestoreFilePrevious || files.Staging != RestoreFileCandidate ||
			files.Quarantine != RestoreFileMissing {
			return ambiguous("previous-preserved phase requires verified rollback and staged candidate identities")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryRestorePrevious,
			PreserveStaging: true,
			RestoreRollback: true,
			Reason: "preserve the verified staged candidate, then restore the verified previous database",
		}, nil

	case RestorePhasePublished, RestorePhaseReopening:
		if hadPrevious {
			if files.Rollback != RestoreFilePrevious || files.Staging != RestoreFileMissing {
				return ambiguous("uncommitted publication requires a verified rollback and no staging artifact")
			}
			switch {
			case files.Target == RestoreFileCandidate && files.Quarantine == RestoreFileMissing:
				return RestoreRecoveryPlan{
					Action: RestoreRecoveryRestorePrevious,
					PreserveTarget: true,
					RestoreRollback: true,
					Reason: "quarantine the verified published candidate, then restore the previous database",
				}, nil
			case files.Target == RestoreFileMissing && files.Quarantine == RestoreFileCandidate:
				return RestoreRecoveryPlan{
					Action: RestoreRecoveryRestorePrevious,
					RestoreRollback: true,
					Reason: "candidate is already quarantined; restore the verified previous database",
				}, nil
			default:
				return ambiguous("candidate must be verified at exactly one of target or quarantine")
			}
		}
		if files.Rollback != RestoreFileMissing || files.Staging != RestoreFileMissing ||
			!((files.Target == RestoreFileCandidate && files.Quarantine == RestoreFileMissing) ||
				(files.Target == RestoreFileMissing && files.Quarantine == RestoreFileCandidate)) {
			return ambiguous("first publication has no verified candidate to preserve")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryPreserveForManualReview,
			Reason: "no previous database existed; preserve the verified candidate and require manual review",
		}, nil

	case RestorePhasePreservingPrevious:
		if !hadPrevious || files.Quarantine != RestoreFileMissing ||
			files.Staging != RestoreFileMissing && files.Staging != RestoreFileCandidate {
			return ambiguous("preserving-previous requires a prior database, an untouched quarantine path, and no unverified staging artifact")
		}
		switch {
		case files.Target == RestoreFilePrevious && files.Rollback == RestoreFileMissing:
			return RestoreRecoveryPlan{
				Action: RestoreRecoveryResumeIntent,
				PreservePrevious: true,
				Reason: "the previous database is still at target; preserve it at the recorded rollback path",
			}, nil
		case files.Target == RestoreFileMissing && files.Rollback == RestoreFilePrevious:
			return RestoreRecoveryPlan{
				Action: RestoreRecoveryResumeIntent,
				Reason: "the previous database is already preserved at rollback; complete the intent transition",
			}, nil
		default:
			return ambiguous("preserving-previous requires the previous database at exactly one of target or rollback")
		}

	case RestorePhasePublishing:
		wantRollback := RestoreFileMissing
		if hadPrevious {
			wantRollback = RestoreFilePrevious
		}
		if files.Rollback != wantRollback || files.Quarantine != RestoreFileMissing ||
			!((files.Target == RestoreFileMissing && files.Staging == RestoreFileCandidate) ||
				(files.Target == RestoreFileCandidate && files.Staging == RestoreFileMissing)) {
			return ambiguous("publishing requires the candidate at exactly one of staging or target and the expected rollback state")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryResumeIntent,
			PublishCandidate: files.Staging == RestoreFileCandidate,
			Reason: "publish the verified staged candidate if the no-replace move has not completed",
		}, nil

	case RestorePhaseRestoringPrevious:
		if !hadPrevious {
			return ambiguous("restoring-previous requires an existing previous database")
		}
		switch {
		case files.Target == RestoreFileMissing && files.Staging == RestoreFileCandidate &&
			files.Rollback == RestoreFilePrevious && files.Quarantine == RestoreFileMissing:
			return RestoreRecoveryPlan{
				Action: RestoreRecoveryResumeIntent,
				PreserveStaging: true,
				RestoreRollback: true,
				Reason: "candidate remains staged; quarantine it before restoring the previous database",
			}, nil
		case files.Staging != RestoreFileMissing:
			return ambiguous("restoring-previous has an unexpected staging artifact")
		case files.Target == RestoreFileCandidate && files.Rollback == RestoreFilePrevious && files.Quarantine == RestoreFileMissing:
			return RestoreRecoveryPlan{
				Action: RestoreRecoveryResumeIntent,
				PreserveTarget: true,
				RestoreRollback: true,
				Reason: "quarantine the verified candidate, then restore the verified previous database",
			}, nil
		case files.Target == RestoreFileMissing && files.Rollback == RestoreFilePrevious && files.Quarantine == RestoreFileCandidate:
			return RestoreRecoveryPlan{
				Action: RestoreRecoveryResumeIntent,
				RestoreRollback: true,
				Reason: "candidate is already quarantined; restore the verified previous database",
			}, nil
		case files.Target == RestoreFilePrevious && files.Rollback == RestoreFileMissing && files.Quarantine == RestoreFileCandidate:
			return RestoreRecoveryPlan{
				Action: RestoreRecoveryResumeIntent,
				Reason: "previous database is already restored; complete the rollback intent transition",
			}, nil
		default:
			return ambiguous("restoring-previous requires one verified candidate at target or quarantine and one verified previous database at rollback or target")
		}

	case RestorePhaseCommitted:
		wantRollback := RestoreFileMissing
		if hadPrevious {
			wantRollback = RestoreFilePrevious
		}
		if files.Target != RestoreFileCandidate || files.Staging != RestoreFileMissing ||
			files.Rollback != wantRollback || files.Quarantine != RestoreFileMissing {
			return ambiguous("committed transaction has inconsistent verified file identities")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryVerifyCommitted,
			VerifyTarget: true,
			Reason: "verify the committed candidate before journal cleanup or rollback retention changes",
		}, nil

	case RestorePhaseRolledBack:
		if !hadPrevious || files.Target != RestoreFilePrevious ||
			files.Staging != RestoreFileMissing || files.Rollback != RestoreFileMissing ||
			files.Quarantine != RestoreFileCandidate {
			return ambiguous("rolled-back transaction requires the verified previous target and retained candidate quarantine")
		}
		return RestoreRecoveryPlan{
			Action: RestoreRecoveryVerifyRolledBack,
			VerifyTarget: true,
			Reason: "verify the restored previous database; retain the quarantined candidate",
		}, nil

	default:
		// Intent phases need the executor's exhaustive move-by-move truth tables.
		return ambiguous("intent phase is not executable by the current recovery planner")
	}
}
