package sqlite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

var ErrRestoreRecoveryManualReview = errors.New("SQLite restore recovery requires manual review")

const (
	restoreRecoveryBoundaryJournalRestoringPrevious = "recovery:journal:restoring-previous"
	restoreRecoveryBoundaryJournalPreviousPreserved = "recovery:journal:previous-preserved"
	restoreRecoveryBoundaryMoveTargetToRollback = "recovery:move:target-to-rollback"
	restoreRecoveryBoundaryMoveTargetToQuarantine = "recovery:move:target-to-quarantine"
	restoreRecoveryBoundaryMoveStagingToQuarantine = "recovery:move:staging-to-quarantine"
	restoreRecoveryBoundaryMoveRollbackToTarget = "recovery:move:rollback-to-target"
	restoreRecoveryBoundaryStoreOpened = "recovery:store:opened"
	restoreRecoveryBoundaryJournalRolledBack = "recovery:journal:rolled-back"
	restoreRecoveryBoundaryStoreCommitted = "recovery:store:committed"
)

type restoreRecoveryOperations struct {
	openStore func(context.Context, string) (*Store, error)
	move func(string, string, RestoreContentIdentity) error
	writeJournal func(*restoreJournalLease, RestoreJournal, string) error
	afterBoundary func(string)
}

// RestoreRecoveryExecutor reconciles an interrupted transaction before the
// application opens its canonical Store. Recover returns a validated Store only
// after the target's no-follow content identity has been checked. The caller
// must not publish that Store to a StoreController before Recover succeeds.
type RestoreRecoveryExecutor struct {
	stateDirectory string
	operations restoreRecoveryOperations
}

func NewRestoreRecoveryExecutor(stateDirectory string) (*RestoreRecoveryExecutor, error) {
	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return nil, err
	}
	return &RestoreRecoveryExecutor{
		stateDirectory: root,
		operations: restoreRecoveryOperations{
			openStore: openVerifiedRestoreStore,
			move: moveRestoreFileNoReplaceVerified,
			writeJournal: func(lease *restoreJournalLease, journal RestoreJournal, _ string) error {
				return lease.Write(journal)
			},
		},
	}, nil
}

// Recover holds the cross-process journal lease for the complete reconciliation.
// A missing ledger means there is no interrupted Restore transaction; nil, nil is
// returned so the caller can continue its ordinary first-open path.
func (executor *RestoreRecoveryExecutor) Recover(ctx context.Context) (store *Store, resultErr error) {
	if executor == nil {
		return nil, ErrRestoreRecoveryAmbiguous
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	lease, err := acquireRestoreJournalLease(ctx, executor.stateDirectory)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := lease.Close(); closeErr != nil {
			if store != nil {
				resultErr = errors.Join(resultErr, store.Close())
				store = nil
			}
			resultErr = errors.Join(resultErr, closeErr)
		}
	}()

	journal, err := lease.Read()
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return executor.reconcile(ctx, lease, journal)
}

func (executor *RestoreRecoveryExecutor) reconcile(
	ctx context.Context,
	lease *restoreJournalLease,
	journal RestoreJournal,
) (*Store, error) {
	files, err := inspectRestoreRecoveryFiles(executor.stateDirectory, journal)
	if err != nil {
		return nil, err
	}
	plan, err := PlanRestoreRecovery(journal, files)
	if err != nil {
		return nil, err
	}

	switch journal.Phase {
	case RestorePhaseCommitted:
		if plan.Action == RestoreRecoveryPreserveForManualReview {
			return nil, fmt.Errorf("%w: first-database candidate is preserved for manual review", ErrRestoreRecoveryManualReview)
		}
		if plan.Action != RestoreRecoveryVerifyCommitted {
			return nil, fmt.Errorf("%w: committed journal did not produce a committed verification plan", ErrRestoreRecoveryAmbiguous)
		}
		return executor.openVerifiedTarget(ctx, lease, journal, true, false)
	case RestorePhaseRolledBack:
		if plan.Action != RestoreRecoveryVerifyRolledBack {
			return nil, fmt.Errorf("%w: rolled-back journal did not produce a rollback verification plan", ErrRestoreRecoveryAmbiguous)
		}
		return executor.openVerifiedTarget(ctx, lease, journal, false, false)
	case RestorePhasePreservingPrevious:
		if plan.Action != RestoreRecoveryResumeIntent {
			return nil, fmt.Errorf("%w: preserving-previous journal did not produce a resumable intent", ErrRestoreRecoveryAmbiguous)
		}
		if plan.PreservePrevious {
			if err := executor.moveArtifact(
				journal.TargetPath,
				journal.RollbackPath,
				RestoreContentIdentity{SHA256: journal.PreviousSHA256, Size: journal.PreviousSize},
				restoreRecoveryBoundaryMoveTargetToRollback,
			); err != nil {
				return nil, fmt.Errorf("complete interrupted previous-database preservation: %w", err)
			}
		}
		journal.Phase = RestorePhasePreviousPreserved
		if err := executor.persistJournal(lease, journal, restoreRecoveryBoundaryJournalPreviousPreserved); err != nil {
			return nil, fmt.Errorf("record completed previous-database preservation: %w", err)
		}
		fallthrough
	case RestorePhasePreviousPreserved:
		journal.Phase = RestorePhaseRestoringPrevious
		if err := executor.persistJournal(lease, journal, restoreRecoveryBoundaryJournalRestoringPrevious); err != nil {
			return nil, fmt.Errorf("record interrupted Restore rollback intent: %w", err)
		}
		return executor.restorePrevious(ctx, lease, journal)
	case RestorePhasePublishing:
		if !*journal.PreviousDatabaseExisted {
			return nil, fmt.Errorf("%w: first-database publication must be reviewed without opening SQLite", ErrRestoreRecoveryManualReview)
		}
		if plan.Action != RestoreRecoveryResumeIntent {
			return nil, fmt.Errorf("%w: publishing journal did not produce a resumable intent", ErrRestoreRecoveryAmbiguous)
		}
		journal.Phase = RestorePhaseRestoringPrevious
		if err := executor.persistJournal(lease, journal, restoreRecoveryBoundaryJournalRestoringPrevious); err != nil {
			return nil, fmt.Errorf("record rollback intent after interrupted publication: %w", err)
		}
		return executor.restorePrevious(ctx, lease, journal)
	case RestorePhasePublished, RestorePhaseReopening:
		if plan.Action == RestoreRecoveryPreserveForManualReview || !*journal.PreviousDatabaseExisted {
			return nil, fmt.Errorf("%w: first-database candidate is preserved for manual review", ErrRestoreRecoveryManualReview)
		}
		if plan.Action != RestoreRecoveryRestorePrevious {
			return nil, fmt.Errorf("%w: published candidate did not produce a rollback plan", ErrRestoreRecoveryAmbiguous)
		}
		journal.Phase = RestorePhaseRestoringPrevious
		if err := executor.persistJournal(lease, journal, restoreRecoveryBoundaryJournalRestoringPrevious); err != nil {
			return nil, fmt.Errorf("record rollback intent after interrupted candidate publication: %w", err)
		}
		return executor.restorePrevious(ctx, lease, journal)
	case RestorePhaseRestoringPrevious:
		if plan.Action != RestoreRecoveryResumeIntent {
			return nil, fmt.Errorf("%w: rollback journal did not produce a resumable intent", ErrRestoreRecoveryAmbiguous)
		}
		return executor.restorePrevious(ctx, lease, journal)
	case RestorePhasePrepared, RestorePhaseStaging:
		return nil, fmt.Errorf("%w: pre-publication staging state is retained for a later cleanup protocol", ErrRestoreRecoveryManualReview)
	default:
		return nil, fmt.Errorf("%w: unsupported recovery phase %q", ErrRestoreRecoveryAmbiguous, journal.Phase)
	}
}

func (executor *RestoreRecoveryExecutor) restorePrevious(
	ctx context.Context,
	lease *restoreJournalLease,
	journal RestoreJournal,
) (*Store, error) {
	files, err := inspectRestoreRecoveryFiles(executor.stateDirectory, journal)
	if err != nil {
		return nil, err
	}
	plan, err := PlanRestoreRecovery(journal, files)
	if err != nil {
		return nil, err
	}
	if plan.Action != RestoreRecoveryResumeIntent {
		return nil, fmt.Errorf("%w: rollback planner returned %q", ErrRestoreRecoveryAmbiguous, plan.Action)
	}

	if plan.PreserveTarget {
		if err := executor.moveArtifact(
			journal.TargetPath,
			journal.CandidateQuarantinePath,
			RestoreContentIdentity{SHA256: journal.CandidateSHA256, Size: journal.CandidateSize},
			restoreRecoveryBoundaryMoveTargetToQuarantine,
		); err != nil {
			return nil, fmt.Errorf("quarantine interrupted Restore candidate: %w", err)
		}
	}
	if plan.PreserveStaging {
		if err := executor.moveArtifact(
			journal.StagingPath,
			journal.CandidateQuarantinePath,
			RestoreContentIdentity{SHA256: journal.CandidateSHA256, Size: journal.CandidateSize},
			restoreRecoveryBoundaryMoveStagingToQuarantine,
		); err != nil {
			return nil, fmt.Errorf("quarantine staged Restore candidate: %w", err)
		}
	}

	files, err = inspectRestoreRecoveryFiles(executor.stateDirectory, journal)
	if err != nil {
		return nil, err
	}
	plan, err = PlanRestoreRecovery(journal, files)
	if err != nil {
		return nil, err
	}
	if plan.Action != RestoreRecoveryResumeIntent {
		return nil, fmt.Errorf("%w: rollback planner changed to unexpected action %q", ErrRestoreRecoveryAmbiguous, plan.Action)
	}
	if plan.RestoreRollback {
		if err := executor.moveArtifact(
			journal.RollbackPath,
			journal.TargetPath,
			RestoreContentIdentity{SHA256: journal.PreviousSHA256, Size: journal.PreviousSize},
			restoreRecoveryBoundaryMoveRollbackToTarget,
		); err != nil {
			return nil, fmt.Errorf("restore previous SQLite database: %w", err)
		}
	}

	files, err = inspectRestoreRecoveryFiles(executor.stateDirectory, journal)
	if err != nil {
		return nil, err
	}
	if files.Target != RestoreFilePrevious ||
		files.Staging != RestoreFileMissing ||
		files.Rollback != RestoreFileMissing ||
		files.Quarantine != RestoreFileCandidate {
		return nil, fmt.Errorf("%w: rollback artifacts do not match the verified terminal state", ErrRestoreRecoveryAmbiguous)
	}
	return executor.openVerifiedTarget(ctx, lease, journal, false, true)
}

func (executor *RestoreRecoveryExecutor) openVerifiedTarget(
	ctx context.Context,
	lease *restoreJournalLease,
	journal RestoreJournal,
	candidate bool,
	completeRollback bool,
) (*Store, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	expected := RestoreContentIdentity{SHA256: journal.PreviousSHA256, Size: journal.PreviousSize}
	boundary := restoreRecoveryBoundaryStoreOpened
	if candidate {
		expected = RestoreContentIdentity{SHA256: journal.CandidateSHA256, Size: journal.CandidateSize}
		boundary = restoreRecoveryBoundaryStoreCommitted
	}
	actual, err := readRestoreContentIdentityPlatform(journal.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("verify recovered SQLite target before database open: %w", err)
	}
	if !actual.matches(expected.SHA256, expected.Size) {
		return nil, fmt.Errorf("%w: recovered SQLite target identity changed before database open", ErrRestoreRecoveryAmbiguous)
	}
	store, err := executor.operations.openStore(ctx, journal.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("open and validate recovered SQLite database: %w", err)
	}
	if executor.operations.afterBoundary != nil {
		executor.operations.afterBoundary(boundary)
	}
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(err, store.Close())
	}
	if completeRollback {
		journal.Phase = RestorePhaseRolledBack
		if err := executor.persistJournal(lease, journal, restoreRecoveryBoundaryJournalRolledBack); err != nil {
			return nil, errors.Join(fmt.Errorf("persist recovered rollback completion: %w", err), store.Close())
		}
	}
	return store, nil
}

func (executor *RestoreRecoveryExecutor) persistJournal(
	lease *restoreJournalLease,
	journal RestoreJournal,
	boundary string,
) error {
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.operations.writeJournal(lease, journal, boundary); err != nil {
		return err
	}
	if executor.operations.afterBoundary != nil {
		executor.operations.afterBoundary(boundary)
	}
	return nil
}

func (executor *RestoreRecoveryExecutor) moveArtifact(
	source, destination string,
	expected RestoreContentIdentity,
	boundary string,
) error {
	if err := ensureNoSQLiteSidecars(source); err != nil {
		return err
	}
	if err := ensureNoSQLiteSidecars(destination); err != nil {
		return err
	}
	if err := executor.operations.move(source, destination, expected); err != nil {
		return err
	}
	if executor.operations.afterBoundary != nil {
		executor.operations.afterBoundary(boundary)
	}
	return nil
}
