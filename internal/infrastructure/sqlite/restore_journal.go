package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	restoreJournalFileName          = "restore-journal.json"
	restoreJournalVersion           = 3
	restoreJournalLedgerVersion     = 1
	maxRestoreJournalBytes          = 1 << 20
	maxRestoreJournalTransactions   = 256
)

var (
	ErrRestoreJournalInvalid = errors.New("SQLite restore journal is invalid")
	errRestoreJournalRemovalUnsupported = errors.New("SQLite restore journal history is retained and cannot be removed automatically")
)

type RestorePhase string

const (
	RestorePhasePrepared           RestorePhase = "prepared"
	RestorePhaseStaging            RestorePhase = "staging"
	RestorePhasePreservingPrevious RestorePhase = "preserving-previous"
	RestorePhasePreviousPreserved  RestorePhase = "previous-preserved"
	RestorePhasePublishing         RestorePhase = "publishing"
	RestorePhasePublished          RestorePhase = "published"
	RestorePhaseReopening           RestorePhase = "reopening"
	RestorePhaseRestoringPrevious  RestorePhase = "restoring-previous"
	RestorePhaseCommitted          RestorePhase = "committed"
	RestorePhaseRolledBack         RestorePhase = "rolled-back"
)

// RestoreJournal records only recovery metadata. Archive identity is content
// based so the selected archive itself need not live inside the state directory.
type RestoreJournal struct {
	Version        int          `json:"version"`
	TransactionID  string       `json:"transaction_id"`
	Phase          RestorePhase `json:"phase"`
	TargetPath     string       `json:"target_path"`
	StagingPath    string       `json:"staging_path"`
	RollbackPath             string       `json:"rollback_path,omitempty"`
	CandidateQuarantinePath  string       `json:"candidate_quarantine_path"`
	PreviousDatabaseExisted  *bool        `json:"previous_database_existed"`
	CandidateSHA256          string       `json:"candidate_sha256,omitempty"`
	CandidateSize            uint64       `json:"candidate_size_bytes,omitempty"`
	PreviousSHA256           string       `json:"previous_sha256,omitempty"`
	PreviousSize             uint64       `json:"previous_size_bytes,omitempty"`
	ArchiveSHA256            string       `json:"archive_sha256"`
	ArchiveSize    uint64       `json:"archive_size_bytes"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// WriteRestoreJournal atomically publishes the latest state of one transaction
// in a bounded ledger. Terminal transaction records are retained so a later
// restore can start without deleting evidence or relying on journal removal.
func WriteRestoreJournal(stateDirectory string, journal RestoreJournal) error {
	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return err
	}
	if journal.UpdatedAt.IsZero() {
		journal.UpdatedAt = time.Now().UTC()
	}
	journal, err = canonicalizeRestoreJournal(root, journal)
	if err != nil {
		return err
	}

	ledger, readErr := readRestoreJournalLedger(root)
	initialPublication := errors.Is(readErr, os.ErrNotExist)
	switch {
	case initialPublication:
		ledger = restoreJournalLedger{Version: restoreJournalLedgerVersion}
	case readErr != nil:
		// A malformed, unreadable, or non-regular existing journal is evidence
		// that must be preserved. Never replace it with a new transaction.
		return fmt.Errorf("%w: refusing to overwrite a journal ledger that cannot be validated: %v", ErrRestoreJournalInvalid, readErr)
	}

	existingIndex := -1
	for i := range ledger.Transactions {
		if ledger.Transactions[i].TransactionID == journal.TransactionID {
			existingIndex = i
			break
		}
	}
	if existingIndex >= 0 {
		if err := validateRestoreJournalTransition(ledger.Transactions[existingIndex], journal); err != nil {
			return err
		}
		ledger.Transactions[existingIndex] = journal
	} else {
		if journal.Phase != RestorePhasePrepared {
			return fmt.Errorf("%w: a new transaction must begin in phase %q", ErrRestoreJournalInvalid, RestorePhasePrepared)
		}
		for _, existing := range ledger.Transactions {
			if !terminalRestorePhase(existing.Phase) {
				return fmt.Errorf("%w: transaction %q is still active in phase %q", ErrRestoreJournalInvalid, existing.TransactionID, existing.Phase)
			}
		}
		if len(ledger.Transactions) >= maxRestoreJournalTransactions {
			return fmt.Errorf("%w: retained transaction history reached its %d-record limit", ErrRestoreJournalInvalid, maxRestoreJournalTransactions)
		}
		ledger.Transactions = append(ledger.Transactions, journal)
	}

	ledger, err = validateRestoreJournalLedger(root, ledger)
	if err != nil {
		return err
	}
	data, err := json.Marshal(ledger)
	if err != nil {
		return fmt.Errorf("encode SQLite restore journal ledger: %w", err)
	}
	if len(data) > maxRestoreJournalBytes {
		return fmt.Errorf("%w: journal ledger exceeds size limit", ErrRestoreJournalInvalid)
	}

	digest := sha256.Sum256(data)
	expectedIdentity := RestoreContentIdentity{Size: uint64(len(data)), SHA256: hex.EncodeToString(digest[:])}

	journalPath := filepath.Join(root, restoreJournalFileName)
	temporary, err := os.CreateTemp(root, ".restore-journal-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary SQLite restore journal: %w", err)
	}
	temporaryPath := temporary.Name()
	// Keep an unpublished journal snapshot when any later step fails. It is
	// evidence for diagnosis/recovery, not disposable scratch space.
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("restrict SQLite restore journal permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write SQLite restore journal: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync SQLite restore journal: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close SQLite restore journal: %w", err)
	}
	var publishErr error
	if initialPublication {
		publishErr = publishRestoreJournalNoReplace(temporaryPath, journalPath, expectedIdentity)
	} else {
		publishErr = publishRestoreJournal(temporaryPath, journalPath)
	}
	if publishErr != nil {
		return fmt.Errorf("publish SQLite restore journal ledger from %s to %s: %w", temporaryPath, journalPath, publishErr)
	}
	return nil
}

// publishRestoreJournalNoReplace verifies the initial ledger from a no-follow handle.
func publishRestoreJournalNoReplace(source, destination string, expected RestoreContentIdentity) error {
	if err := moveRestoreFileNoReplaceVerified(source, destination, expected); err != nil {
		return fmt.Errorf("publish SQLite restore journal without replacement: %w", err)
	}
	return nil
}

// ReadRestoreJournal strictly decodes the retained transaction ledger and
// returns its sole active transaction, or the newest terminal record when idle.
// A legacy single-record v3 journal is accepted and migrated on the next write.
func ReadRestoreJournal(stateDirectory string) (RestoreJournal, error) {
	var empty RestoreJournal
	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return empty, err
	}
	ledger, err := readRestoreJournalLedger(root)
	if err != nil {
		return empty, err
	}
	return currentRestoreJournal(ledger)
}

// RemoveRestoreJournal validates the current state but retains the ledger.
// Terminal records are part of the transaction history and are not deleted.
func RemoveRestoreJournal(stateDirectory string) error {
	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return err
	}
	ledger, err := readRestoreJournalLedger(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	current, err := currentRestoreJournal(ledger)
	if err != nil {
		return err
	}
	if !terminalRestorePhase(current.Phase) {
		return fmt.Errorf("%w: only a committed or rolled-back journal may be retired", ErrRestoreJournalInvalid)
	}
	return removeRestoreJournal(root, filepath.Join(root, restoreJournalFileName))
}

func canonicalStateDirectory(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", ErrRestoreJournalInvalid
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve SQLite state directory: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Clean(absolute))
	if err != nil {
		return "", fmt.Errorf("resolve SQLite state directory: %w", err)
	}
	canonical, err := canonicalRestoreDirectoryPathPlatform(filepath.Clean(resolved))
	if err != nil {
		return "", fmt.Errorf("%w: canonicalize SQLite state directory: %w", ErrRestoreJournalInvalid, err)
	}
	info, err := os.Stat(canonical)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("%w: state directory is not a directory", ErrRestoreJournalInvalid)
	}
	if err := validateRestoreStateDirectoryPlatform(canonical); err != nil {
		return "", fmt.Errorf("%w: state directory security validation failed: %w", ErrRestoreJournalInvalid, err)
	}
	// Persist one normalized spelling. Windows can expose the same directory
	// through both a long path and an 8.3 alias.
	return canonical, nil
}

func canonicalizeRestoreJournal(root string, journal RestoreJournal) (RestoreJournal, error) {
	if journal.Version != restoreJournalVersion ||
		strings.TrimSpace(journal.TransactionID) == "" ||
		len(journal.TransactionID) > 128 ||
		!validRestorePhase(journal.Phase) ||
		journal.ArchiveSize == 0 ||
		journal.PreviousDatabaseExisted == nil ||
		journal.UpdatedAt.IsZero() {
		return RestoreJournal{}, fmt.Errorf("%w: required metadata or phase is invalid", ErrRestoreJournalInvalid)
	}
	digest, err := hex.DecodeString(journal.ArchiveSHA256)
	if err != nil || len(digest) != sha256.Size || strings.ToLower(journal.ArchiveSHA256) != journal.ArchiveSHA256 {
		return RestoreJournal{}, fmt.Errorf("%w: archive SHA-256 is invalid", ErrRestoreJournalInvalid)
	}
	journal.TargetPath, err = canonicalStatePath(root, journal.TargetPath)
	if err != nil {
		return RestoreJournal{}, err
	}
	journal.StagingPath, err = canonicalStatePath(root, journal.StagingPath)
	if err != nil {
		return RestoreJournal{}, err
	}
	if samePath(journal.TargetPath, journal.StagingPath) {
		return RestoreJournal{}, fmt.Errorf("%w: target and staging paths are identical", ErrRestoreJournalInvalid)
	}
	if *journal.PreviousDatabaseExisted {
		if strings.TrimSpace(journal.RollbackPath) == "" {
			return RestoreJournal{}, fmt.Errorf("%w: previous database exists but rollback path is missing", ErrRestoreJournalInvalid)
		}
	} else if journal.RollbackPath != "" {
		return RestoreJournal{}, fmt.Errorf("%w: rollback path is unexpected when no previous database exists", ErrRestoreJournalInvalid)
	}
	if journal.RollbackPath != "" {
		journal.RollbackPath, err = canonicalStatePath(root, journal.RollbackPath)
		if err != nil {
			return RestoreJournal{}, err
		}
		if samePath(journal.RollbackPath, journal.TargetPath) || samePath(journal.RollbackPath, journal.StagingPath) {
			return RestoreJournal{}, fmt.Errorf("%w: rollback path collides with another transaction path", ErrRestoreJournalInvalid)
		}
	}
	journal.CandidateQuarantinePath, err = canonicalStatePath(root, journal.CandidateQuarantinePath)
	if err != nil {
		return RestoreJournal{}, err
	}
	transactionPaths := []string{journal.TargetPath, journal.StagingPath, journal.CandidateQuarantinePath}
	if journal.RollbackPath != "" {
		transactionPaths = append(transactionPaths, journal.RollbackPath)
	}
	for i := range transactionPaths {
		for j := i + 1; j < len(transactionPaths); j++ {
			if samePath(transactionPaths[i], transactionPaths[j]) {
				return RestoreJournal{}, fmt.Errorf("%w: transaction paths collide", ErrRestoreJournalInvalid)
			}
		}
	}
	if err := validateRestoreContentIdentity(journal.CandidateSHA256, journal.CandidateSize, "candidate", false); err != nil {
		return RestoreJournal{}, err
	}
	candidateRequired := journal.Phase == RestorePhasePreservingPrevious ||
		journal.Phase == RestorePhasePreviousPreserved ||
		journal.Phase == RestorePhasePublishing ||
		journal.Phase == RestorePhasePublished ||
		journal.Phase == RestorePhaseReopening ||
		journal.Phase == RestorePhaseRestoringPrevious ||
		journal.Phase == RestorePhaseCommitted ||
		journal.Phase == RestorePhaseRolledBack
	if candidateRequired && journal.CandidateSHA256 == "" {
		return RestoreJournal{}, fmt.Errorf("%w: candidate identity is required in phase %q", ErrRestoreJournalInvalid, journal.Phase)
	}
	previousRequired := journal.Phase == RestorePhasePreservingPrevious ||
		journal.Phase == RestorePhasePreviousPreserved ||
		journal.Phase == RestorePhaseRestoringPrevious ||
		journal.Phase == RestorePhaseRolledBack
	if previousRequired && !*journal.PreviousDatabaseExisted {
		return RestoreJournal{}, fmt.Errorf("%w: phase %q requires a previous database", ErrRestoreJournalInvalid, journal.Phase)
	}
	if !*journal.PreviousDatabaseExisted && (journal.PreviousSHA256 != "" || journal.PreviousSize != 0) {
		return RestoreJournal{}, fmt.Errorf("%w: previous database identity is unexpected when no previous database exists", ErrRestoreJournalInvalid)
	}
	if err := validateRestoreContentIdentity(journal.PreviousSHA256, journal.PreviousSize, "previous database", *journal.PreviousDatabaseExisted); err != nil {
		return RestoreJournal{}, err
	}
	return journal, nil
}

func canonicalStatePath(root, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("%w: transaction path is empty", ErrRestoreJournalInvalid)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("%w: resolve transaction path: %v", ErrRestoreJournalInvalid, err)
	}
	absolute = filepath.Clean(absolute)
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("%w: resolve state directory: %v", ErrRestoreJournalInvalid, err)
	}
	resolvedRoot, err = canonicalRestoreDirectoryPathPlatform(filepath.Clean(resolvedRoot))
	if err != nil {
		return "", fmt.Errorf("%w: canonicalize state directory: %v", ErrRestoreJournalInvalid, err)
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return "", fmt.Errorf("%w: resolve transaction path parent: %v", ErrRestoreJournalInvalid, err)
	}
	parent, err = canonicalRestoreDirectoryPathPlatform(filepath.Clean(parent))
	if err != nil {
		return "", fmt.Errorf("%w: canonicalize transaction path parent: %v", ErrRestoreJournalInvalid, err)
	}
	resolvedCandidate := filepath.Join(parent, filepath.Base(absolute))
	if info, statErr := os.Lstat(absolute); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("%w: transaction path must not be a symbolic link", ErrRestoreJournalInvalid)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("%w: existing transaction path must be a regular file", ErrRestoreJournalInvalid)
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return "", fmt.Errorf("%w: inspect transaction path: %v", ErrRestoreJournalInvalid, statErr)
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedCandidate)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("%w: transaction path escapes the application state directory", ErrRestoreJournalInvalid)
	}
	return resolvedCandidate, nil
}

func validateRestoreContentIdentity(digest string, size uint64, label string, required bool) error {
	if digest == "" && size == 0 && !required {
		return nil
	}
	decoded, err := hex.DecodeString(digest)
	if err != nil || len(decoded) != sha256.Size || strings.ToLower(digest) != digest || size == 0 {
		return fmt.Errorf("%w: %s SHA-256/size identity is invalid", ErrRestoreJournalInvalid, label)
	}
	return nil
}

func validRestorePhase(phase RestorePhase) bool {
	switch phase {
	case RestorePhasePrepared, RestorePhaseStaging, RestorePhasePreservingPrevious,
		RestorePhasePreviousPreserved, RestorePhasePublishing, RestorePhasePublished,
		RestorePhaseReopening, RestorePhaseRestoringPrevious, RestorePhaseCommitted,
		RestorePhaseRolledBack:
		return true
	default:
		return false
	}
}



// validateRestoreJournalTransition prevents a later write from changing the
// transaction's identity or skipping a durable intent/completion phase.
func validateRestoreJournalTransition(current, next RestoreJournal) error {
	invalid := func(reason string) error {
		return fmt.Errorf("%w: %s", ErrRestoreJournalInvalid, reason)
	}
	if current.Version != next.Version ||
		current.TransactionID != next.TransactionID ||
		current.TargetPath != next.TargetPath ||
		current.StagingPath != next.StagingPath ||
		current.RollbackPath != next.RollbackPath ||
		current.CandidateQuarantinePath != next.CandidateQuarantinePath ||
		current.PreviousDatabaseExisted == nil ||
		next.PreviousDatabaseExisted == nil ||
		*current.PreviousDatabaseExisted != *next.PreviousDatabaseExisted ||
		current.ArchiveSHA256 != next.ArchiveSHA256 ||
		current.ArchiveSize != next.ArchiveSize ||
		current.PreviousSHA256 != next.PreviousSHA256 ||
		current.PreviousSize != next.PreviousSize {
		return invalid("transaction metadata and previous-database identity are immutable")
	}
	if current.CandidateSHA256 != "" {
		if next.CandidateSHA256 != current.CandidateSHA256 || next.CandidateSize != current.CandidateSize {
			return invalid("candidate identity cannot be changed or cleared")
		}
	} else if next.CandidateSHA256 != "" && next.Phase != RestorePhaseStaging {
		return invalid("candidate identity may only be recorded during staging")
	}
	if !validRestorePhaseTransition(current.Phase, next.Phase, *current.PreviousDatabaseExisted) {
		return invalid(fmt.Sprintf("illegal phase transition %q -> %q", current.Phase, next.Phase))
	}
	return nil
}

func validRestorePhaseTransition(current, next RestorePhase, hasPrevious bool) bool {
	if current == next {
		return true
	}
	switch current {
	case RestorePhasePrepared:
		return next == RestorePhaseStaging
	case RestorePhaseStaging:
		if hasPrevious {
			return next == RestorePhasePreservingPrevious
		}
		return next == RestorePhasePublishing
	case RestorePhasePreservingPrevious:
		return next == RestorePhasePreviousPreserved
	case RestorePhasePreviousPreserved:
		return next == RestorePhasePublishing || (hasPrevious && next == RestorePhaseRestoringPrevious)
	case RestorePhasePublishing:
		return next == RestorePhasePublished || (hasPrevious && next == RestorePhaseRestoringPrevious)
	case RestorePhasePublished:
		return next == RestorePhaseReopening || (hasPrevious && next == RestorePhaseRestoringPrevious)
	case RestorePhaseReopening:
		return next == RestorePhaseCommitted || (hasPrevious && next == RestorePhaseRestoringPrevious)
	case RestorePhaseRestoringPrevious:
		return next == RestorePhaseRolledBack
	default:
		return false
	}
}
