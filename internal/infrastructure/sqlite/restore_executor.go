package sqlite

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type RestoreCandidateBuilder func(context.Context, string, string, string) (RestoreCandidate, error)

type RestoreTransactionRequest struct {
	ArchivePath    string
	BuildCandidate RestoreCandidateBuilder
}

type RestoreTransactionExecutor struct {
	controller     *StoreController
	stateDirectory string
	operations     restoreTransactionOperations
}

type restoreTransactionOperations struct {
	openStore     func(context.Context, string) (*Store, error)
	move          func(string, string, RestoreContentIdentity) error
	writeJournal  func(*restoreJournalLease, RestoreJournal, string) error
	afterBoundary func(string)
	newID         func() (string, error)
}

const (
	restoreBoundaryBeforePrepared           = "before:prepared"
	restoreBoundaryJournalPrepared          = "journal:prepared"
	restoreBoundaryJournalStaging           = "journal:staging"
	restoreBoundaryCandidateBuilt           = "candidate:built"
	restoreBoundaryJournalCandidateIdentity = "journal:candidate-identity"
	restoreBoundaryJournalPreservingPrevious = "journal:preserving-previous"
	restoreBoundaryMovePreviousToRollback   = "move:previous-to-rollback"
	restoreBoundaryJournalPreviousPreserved = "journal:previous-preserved"
	restoreBoundaryJournalPublishing        = "journal:publishing"
	restoreBoundaryMoveStagingToTarget      = "move:staging-to-target"
	restoreBoundaryJournalPublished         = "journal:published"
	restoreBoundaryJournalReopening         = "journal:reopening"
	restoreBoundaryStoreCandidateOpened     = "store:candidate-opened"
	restoreBoundaryJournalCommitted         = "journal:committed"
	restoreBoundaryJournalRestoringPrevious = "journal:restoring-previous"
	restoreBoundaryMoveTargetToQuarantine   = "move:candidate-to-quarantine"
	restoreBoundaryMoveStagingToQuarantine  = "move:staging-to-quarantine"
	restoreBoundaryMoveRollbackToTarget     = "move:rollback-to-target"
	restoreBoundaryStorePreviousOpened      = "store:previous-opened"
	restoreBoundaryJournalRolledBack        = "journal:rolled-back"
)

func (executor *RestoreTransactionExecutor) reachBoundary(boundary string) {
	if executor.operations.afterBoundary != nil {
		executor.operations.afterBoundary(boundary)
	}
}

func (executor *RestoreTransactionExecutor) persistJournal(
	lease *restoreJournalLease,
	journal RestoreJournal,
	boundary string,
) error {
	if err := executor.operations.writeJournal(lease, journal, boundary); err != nil {
		return err
	}
	executor.reachBoundary(boundary)
	return nil
}

func (executor *RestoreTransactionExecutor) moveArtifact(
	source, destination string,
	expected RestoreContentIdentity,
	boundary string,
) error {
	if err := executor.operations.move(source, destination, expected); err != nil {
		return err
	}
	executor.reachBoundary(boundary)
	return nil
}

func NewRestoreTransactionExecutor(controller *StoreController, stateDirectory string) (*RestoreTransactionExecutor, error) {
	if controller == nil {
		return nil, ErrStoreUnavailable
	}
	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return nil, err
	}
	return &RestoreTransactionExecutor{
		controller:     controller,
		stateDirectory: root,
		operations: restoreTransactionOperations{
			openStore: openVerifiedRestoreStore,
			move:      moveRestoreFileNoReplaceVerified,
			writeJournal: func(lease *restoreJournalLease, journal RestoreJournal, _ string) error {
				return lease.Write(journal)
			},
			newID: newRestoreTransactionID,
		},
	}, nil
}

func (executor *RestoreTransactionExecutor) Execute(ctx context.Context, request RestoreTransactionRequest) error {
	if executor == nil || executor.controller == nil {
		return ErrStoreUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if request.ArchivePath == "" || request.BuildCandidate == nil {
		return ErrRestoreSourceInvalid
	}

	archivePath, err := filepath.Abs(request.ArchivePath)
	if err != nil {
		return fmt.Errorf("resolve SQLite restore archive: %w", err)
	}
	archivePath = filepath.Clean(archivePath)
	archiveIdentity, err := readRestoreArchiveIdentity(archivePath)
	if err != nil {
		return err
	}

	lease, err := acquireRestoreJournalLease(ctx, executor.stateDirectory)
	if err != nil {
		return err
	}
	current, readErr := lease.Read()
	if readErr == nil && !terminalRestorePhase(current.Phase) {
		closeErr := lease.Close()
		return errors.Join(
			fmt.Errorf("%w: transaction %q is still active in phase %q", ErrRestoreRecoveryAmbiguous, current.TransactionID, current.Phase),
			closeErr,
		)
	}
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return errors.Join(readErr, lease.Close())
	}

	var operationErr error
	func() {
		defer func() {
			operationErr = errors.Join(operationErr, lease.Close())
		}()
		operationErr = executor.controller.WithExclusive(ctx, func(store *Store) (*Store, error) {
			return executor.executeLocked(ctx, lease, store, archivePath, archiveIdentity, request)
		})
	}()
	return operationErr
}

func (executor *RestoreTransactionExecutor) executeLocked(
	ctx context.Context,
	lease *restoreJournalLease,
	store *Store,
	archivePath string,
	archiveIdentity RestoreContentIdentity,
	request RestoreTransactionRequest,
) (*Store, error) {
	if store == nil || store.db == nil {
		return nil, ErrStoreClosed
	}
	target, err := filepath.Abs(store.Path())
	if err != nil {
		return store, fmt.Errorf("resolve active SQLite database path: %w", err)
	}
	target = filepath.Clean(target)
	if !samePath(filepath.Dir(target), executor.stateDirectory) {
		return store, fmt.Errorf("%w: active database must be a direct child of the private state directory", ErrRestoreJournalInvalid)
	}
	if samePath(target, archivePath) {
		return store, ErrInvalidDatabasePath
	}
	if err := checkpointStore(ctx, store); err != nil {
		return store, err
	}
	if err := store.Close(); err != nil {
		return executor.reopenPrevious(target, fmt.Errorf("close active SQLite store before Restore: %w", err))
	}

	reopenPrevious := func(cause error) (*Store, error) {
		return executor.reopenPrevious(target, cause)
	}
	if err := ctx.Err(); err != nil {
		return reopenPrevious(err)
	}
	if err := ensureNoSQLiteSidecars(target); err != nil {
		return reopenPrevious(err)
	}
	previousIdentity, err := readRestoreContentIdentityPlatform(target)
	if err != nil {
		return reopenPrevious(fmt.Errorf("identify previous SQLite database: %w", err))
	}

	transactionID, err := executor.operations.newID()
	if err != nil {
		return reopenPrevious(fmt.Errorf("generate SQLite Restore transaction ID: %w", err))
	}
	stagingPath := filepath.Join(executor.stateDirectory, ".restore-stage-"+transactionID+".sqlite")
	rollbackPath := filepath.Join(executor.stateDirectory, ".restore-rollback-"+transactionID+".sqlite")
	quarantinePath := filepath.Join(executor.stateDirectory, ".restore-quarantine-"+transactionID+".sqlite")
	for _, path := range []string{stagingPath, rollbackPath, quarantinePath} {
		if err := requireRestorePathAbsent(path); err != nil {
			return reopenPrevious(err)
		}
		if err := ensureNoSQLiteSidecars(path); err != nil {
			return reopenPrevious(fmt.Errorf("inspect SQLite Restore artifact sidecars: %w", err))
		}
	}

	previousExists := true
	journal := RestoreJournal{
		Version:                  restoreJournalVersion,
		TransactionID:            transactionID,
		Phase:                    RestorePhasePrepared,
		TargetPath:               target,
		StagingPath:              stagingPath,
		RollbackPath:             rollbackPath,
		CandidateQuarantinePath:  quarantinePath,
		PreviousDatabaseExisted:  &previousExists,
		PreviousSHA256:           previousIdentity.SHA256,
		PreviousSize:             previousIdentity.Size,
		ArchiveSHA256:            archiveIdentity.SHA256,
		ArchiveSize:              archiveIdentity.Size,
		UpdatedAt:                time.Now().UTC(),
	}
	executor.reachBoundary(restoreBoundaryBeforePrepared)
	// From the first journal publication attempt onward, reopening SQLite could
	// change the recorded previous-file identity. Keep the controller closed on
	// every error so startup recovery can inspect the exact recorded artifacts.
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalPrepared); err != nil {
		return nil, fmt.Errorf("persist prepared SQLite Restore journal: %w", err)
	}
	journal.Phase = RestorePhaseStaging
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalStaging); err != nil {
		return nil, fmt.Errorf("persist SQLite Restore staging intent: %w", err)
	}

	candidate, err := request.BuildCandidate(ctx, archivePath, executor.stateDirectory, stagingPath)
	if err != nil {
		return nil, fmt.Errorf("build staged SQLite Restore candidate: %w", err)
	}
	executor.reachBoundary(restoreBoundaryCandidateBuilt)
	if !samePath(candidate.Path, stagingPath) || candidate.Identity.Size == 0 || candidate.Identity.SHA256 == "" {
		return nil, fmt.Errorf("%w: candidate builder returned an unexpected path or incomplete identity", ErrRestoreJournalInvalid)
	}
	journal.CandidateSHA256 = candidate.Identity.SHA256
	journal.CandidateSize = candidate.Identity.Size
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalCandidateIdentity); err != nil {
		return nil, fmt.Errorf("persist staged SQLite Restore candidate identity: %w", err)
	}
	if candidate.Identity.matches(previousIdentity.SHA256, previousIdentity.Size) {
		return nil, fmt.Errorf("%w: candidate and previous database identities are indistinguishable", ErrRestoreJournalInvalid)
	}
	archiveAfterBuild, err := readRestoreArchiveIdentity(archivePath)
	if err != nil {
		return nil, err
	}
	if !archiveAfterBuild.matches(archiveIdentity.SHA256, archiveIdentity.Size) {
		return nil, fmt.Errorf("%w: archive changed while its candidate was being built", ErrRestoreSourceInvalid)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	journal.Phase = RestorePhasePreservingPrevious
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalPreservingPrevious); err != nil {
		return nil, fmt.Errorf("persist previous-database preservation intent: %w", err)
	}
	if err := ensureNoSQLiteSidecars(target); err != nil {
		return nil, err
	}
	if err := ensureNoSQLiteSidecars(rollbackPath); err != nil {
		return nil, err
	}
	if err := executor.moveArtifact(target, rollbackPath, previousIdentity, restoreBoundaryMovePreviousToRollback); err != nil {
		return nil, fmt.Errorf("preserve previous SQLite database: %w", err)
	}
	journal.Phase = RestorePhasePreviousPreserved
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalPreviousPreserved); err != nil {
		return nil, fmt.Errorf("persist previous-database preservation completion: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return executor.rollback(lease, journal, err)
	}

	journal.Phase = RestorePhasePublishing
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalPublishing); err != nil {
		return nil, fmt.Errorf("persist SQLite candidate publication intent: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return executor.rollback(lease, journal, err)
	}
	if err := ensureNoSQLiteSidecars(stagingPath); err != nil {
		return executor.rollback(lease, journal, err)
	}
	if err := ensureNoSQLiteSidecars(target); err != nil {
		return executor.rollback(lease, journal, err)
	}
	if err := executor.moveArtifact(stagingPath, target, candidate.Identity, restoreBoundaryMoveStagingToTarget); err != nil {
		// A move error may be reported after the namespace changed but before
		// its durability or identity-verification boundary is confirmed. Do not
		// issue another mutation based on an assumed source/destination state.
		return nil, fmt.Errorf("publish verified SQLite Restore candidate: %w", err)
	}
	journal.Phase = RestorePhasePublished
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalPublished); err != nil {
		return nil, fmt.Errorf("persist SQLite candidate publication completion: %w", err)
	}
	journal.Phase = RestorePhaseReopening
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalReopening); err != nil {
		return nil, fmt.Errorf("persist SQLite candidate reopen intent: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return executor.rollback(lease, journal, err)
	}

	candidateStore, err := executor.operations.openStore(ctx, target)
	if err != nil {
		return executor.rollback(lease, journal, fmt.Errorf("open and validate published SQLite candidate: %w", err))
	}
	executor.reachBoundary(restoreBoundaryStoreCandidateOpened)
	if err := ctx.Err(); err != nil {
		_ = candidateStore.Close()
		return executor.rollback(lease, journal, err)
	}
	journal.Phase = RestorePhaseCommitted
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalCommitted); err != nil {
		_ = candidateStore.Close()
		return nil, fmt.Errorf("persist committed SQLite Restore transaction: %w", err)
	}
	return candidateStore, nil
}

func (executor *RestoreTransactionExecutor) reopenPrevious(target string, cause error) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := executor.operations.openStore(ctx, target)
	if err != nil {
		return nil, errors.Join(cause, fmt.Errorf("reopen previous SQLite database: %w", err))
	}
	return store, cause
}

func (executor *RestoreTransactionExecutor) rollback(
	lease *restoreJournalLease,
	journal RestoreJournal,
	cause error,
) (*Store, error) {
	recoveryCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	journal.Phase = RestorePhaseRestoringPrevious
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalRestoringPrevious); err != nil {
		return nil, errors.Join(cause, fmt.Errorf("persist SQLite Restore rollback intent: %w", err))
	}
	files, err := inspectRestoreRecoveryFiles(executor.stateDirectory, journal)
	if err != nil {
		return nil, errors.Join(cause, err)
	}
	plan, err := PlanRestoreRecovery(journal, files)
	if err != nil {
		return nil, errors.Join(cause, err)
	}
	if plan.Action != RestoreRecoveryResumeIntent {
		return nil, errors.Join(cause, fmt.Errorf("%w: rollback planner returned %q", ErrRestoreRecoveryAmbiguous, plan.Action))
	}
	if plan.PreserveTarget {
		if err := ensureNoSQLiteSidecars(journal.TargetPath); err != nil {
			return nil, errors.Join(cause, err)
		}
		if err := ensureNoSQLiteSidecars(journal.CandidateQuarantinePath); err != nil {
			return nil, errors.Join(cause, err)
		}
		if err := executor.moveArtifact(journal.TargetPath, journal.CandidateQuarantinePath, RestoreContentIdentity{
			Size: journal.CandidateSize,
			SHA256: journal.CandidateSHA256,
		}, restoreBoundaryMoveTargetToQuarantine); err != nil {
			return nil, errors.Join(cause, fmt.Errorf("quarantine failed SQLite Restore candidate: %w", err))
		}
	}
	if plan.PreserveStaging {
		if err := ensureNoSQLiteSidecars(journal.StagingPath); err != nil {
			return nil, errors.Join(cause, err)
		}
		if err := ensureNoSQLiteSidecars(journal.CandidateQuarantinePath); err != nil {
			return nil, errors.Join(cause, err)
		}
		if err := executor.moveArtifact(journal.StagingPath, journal.CandidateQuarantinePath, RestoreContentIdentity{
			Size: journal.CandidateSize,
			SHA256: journal.CandidateSHA256,
		}, restoreBoundaryMoveStagingToQuarantine); err != nil {
			return nil, errors.Join(cause, fmt.Errorf("quarantine staged SQLite Restore candidate: %w", err))
		}
	}
	if plan.RestoreRollback {
		if err := ensureNoSQLiteSidecars(journal.RollbackPath); err != nil {
			return nil, errors.Join(cause, err)
		}
		if err := ensureNoSQLiteSidecars(journal.TargetPath); err != nil {
			return nil, errors.Join(cause, err)
		}
		if err := executor.moveArtifact(journal.RollbackPath, journal.TargetPath, RestoreContentIdentity{
			Size: journal.PreviousSize,
			SHA256: journal.PreviousSHA256,
		}, restoreBoundaryMoveRollbackToTarget); err != nil {
			return nil, errors.Join(cause, fmt.Errorf("restore previous SQLite database: %w", err))
		}
	}

	files, err = inspectRestoreRecoveryFiles(executor.stateDirectory, journal)
	if err != nil {
		return nil, errors.Join(cause, err)
	}
	if files.Target != RestoreFilePrevious ||
		files.Staging != RestoreFileMissing ||
		files.Rollback != RestoreFileMissing ||
		files.Quarantine != RestoreFileCandidate {
		return nil, errors.Join(cause, fmt.Errorf("%w: rollback artifacts do not match the verified terminal state", ErrRestoreRecoveryAmbiguous))
	}
	restoredStore, err := executor.operations.openStore(recoveryCtx, journal.TargetPath)
	if err != nil {
		return nil, errors.Join(cause, fmt.Errorf("open and validate rolled-back SQLite database: %w", err))
	}
	executor.reachBoundary(restoreBoundaryStorePreviousOpened)
	journal.Phase = RestorePhaseRolledBack
	journal.UpdatedAt = time.Now().UTC()
	if err := executor.persistJournal(lease, journal, restoreBoundaryJournalRolledBack); err != nil {
		return restoredStore, errors.Join(cause, fmt.Errorf("persist rolled-back SQLite Restore transaction: %w", err))
	}
	return restoredStore, cause
}

func openVerifiedRestoreStore(ctx context.Context, path string) (*Store, error) {
	if err := verifyBackupDatabase(ctx, path); err != nil {
		return nil, err
	}
	store, err := OpenExisting(ctx, path)
	if err != nil {
		return nil, err
	}
	configuration, err := store.Load(ctx)
	if err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("load reopened SQLite configuration: %w", err)
	}
	if err := configuration.Validate(); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("validate reopened SQLite configuration: %w", err)
	}
	return store, nil
}

func readRestoreArchiveIdentity(path string) (RestoreContentIdentity, error) {
	var identity RestoreContentIdentity
	beforePath, err := os.Lstat(path)
	if err != nil {
		return identity, fmt.Errorf("inspect SQLite Restore archive: %w", err)
	}
	if !beforePath.Mode().IsRegular() {
		return identity, ErrRestoreSourceInvalid
	}
	file, err := os.Open(path)
	if err != nil {
		return identity, fmt.Errorf("open SQLite Restore archive: %w", err)
	}
	before, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return identity, fmt.Errorf("inspect opened SQLite Restore archive: %w", err)
	}
	if !before.Mode().IsRegular() || !os.SameFile(beforePath, before) {
		_ = file.Close()
		return identity, ErrRestoreSourceInvalid
	}
	hash := sha256.New()
	size, copyErr := io.Copy(hash, file)
	after, statErr := file.Stat()
	closeErr := file.Close()
	if copyErr != nil || statErr != nil || closeErr != nil {
		return identity, errors.Join(
			wrapRestoreArchiveError("read SQLite Restore archive identity", copyErr),
			wrapRestoreArchiveError("reinspect SQLite Restore archive", statErr),
			wrapRestoreArchiveError("close SQLite Restore archive", closeErr),
		)
	}
	afterPath, err := os.Lstat(path)
	if err != nil {
		return identity, fmt.Errorf("reinspect SQLite Restore archive path: %w", err)
	}
	if !after.Mode().IsRegular() || !afterPath.Mode().IsRegular() ||
		!os.SameFile(before, after) || !os.SameFile(after, afterPath) ||
		size == 0 || size != after.Size() || before.Size() != after.Size() ||
		!before.ModTime().Equal(after.ModTime()) {
		return identity, fmt.Errorf("%w: archive changed while being read", ErrRestoreSourceInvalid)
	}
	identity.Size = uint64(size)
	identity.SHA256 = hex.EncodeToString(hash.Sum(nil))
	return identity, nil
}

func wrapRestoreArchiveError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func requireRestorePathAbsent(path string) error {
	_, err := os.Lstat(path)
	if err == nil {
		return fmt.Errorf("%w: Restore artifact path already exists: %s", ErrRestoreJournalInvalid, path)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect Restore artifact path %s: %w", path, err)
	}
	return nil
}

func newRestoreTransactionID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
