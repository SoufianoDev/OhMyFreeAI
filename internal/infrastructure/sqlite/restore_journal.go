package sqlite

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	restoreJournalFileName = "restore-journal.json"
	restoreJournalVersion  = 1
	maxRestoreJournalBytes = 16 << 10
)

var ErrRestoreJournalInvalid = errors.New("SQLite restore journal is invalid")

type RestorePhase string

const (
	RestorePhasePrepared          RestorePhase = "prepared"
	RestorePhaseStaging           RestorePhase = "staging"
	RestorePhasePreviousPreserved RestorePhase = "previous-preserved"
	RestorePhasePublished         RestorePhase = "published"
	RestorePhaseReopening         RestorePhase = "reopening"
	RestorePhaseCommitted         RestorePhase = "committed"
)

// RestoreJournal records only recovery metadata. Archive identity is content
// based so the selected archive itself need not live inside the state directory.
type RestoreJournal struct {
	Version        int          `json:"version"`
	TransactionID  string       `json:"transaction_id"`
	Phase          RestorePhase `json:"phase"`
	TargetPath     string       `json:"target_path"`
	StagingPath    string       `json:"staging_path"`
	RollbackPath   string       `json:"rollback_path,omitempty"`
	ArchiveSHA256  string       `json:"archive_sha256"`
	ArchiveSize    uint64       `json:"archive_size_bytes"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// WriteRestoreJournal durably replaces the journal using a same-directory
// temporary file, file sync, atomic rename and directory sync.
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
	data, err := json.Marshal(journal)
	if err != nil {
		return fmt.Errorf("encode SQLite restore journal: %w", err)
	}
	if len(data) > maxRestoreJournalBytes {
		return fmt.Errorf("%w: journal exceeds size limit", ErrRestoreJournalInvalid)
	}

	journalPath := filepath.Join(root, restoreJournalFileName)
	if info, statErr := os.Lstat(journalPath); statErr == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%w: journal path is not a regular file", ErrRestoreJournalInvalid)
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("inspect SQLite restore journal: %w", statErr)
	}

	temporary, err := os.CreateTemp(root, ".restore-journal-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary SQLite restore journal: %w", err)
	}
	temporaryPath := temporary.Name()
	published := false
	defer func() {
		if !published {
			_ = os.Remove(temporaryPath)
		}
	}()
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
	if err := os.Rename(temporaryPath, journalPath); err != nil {
		return fmt.Errorf("publish SQLite restore journal: %w", err)
	}
	published = true
	if err := syncRestoreDirectory(root); err != nil {
		return err
	}
	return nil
}

// ReadRestoreJournal strictly decodes and validates the durable journal.
// Missing journals are reported as os.ErrNotExist.
func ReadRestoreJournal(stateDirectory string) (RestoreJournal, error) {
	var empty RestoreJournal
	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return empty, err
	}
	journalPath := filepath.Join(root, restoreJournalFileName)
	info, err := os.Lstat(journalPath)
	if err != nil {
		return empty, err
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxRestoreJournalBytes {
		return empty, fmt.Errorf("%w: journal is not a bounded regular file", ErrRestoreJournalInvalid)
	}
	data, err := os.ReadFile(journalPath)
	if err != nil {
		return empty, fmt.Errorf("read SQLite restore journal: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var journal RestoreJournal
	if err := decoder.Decode(&journal); err != nil {
		return empty, fmt.Errorf("%w: decode journal: %v", ErrRestoreJournalInvalid, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return empty, fmt.Errorf("%w: trailing journal data", ErrRestoreJournalInvalid)
	}
	journal, err = canonicalizeRestoreJournal(root, journal)
	if err != nil {
		return empty, err
	}
	return journal, nil
}

// RemoveRestoreJournal removes the committed/recovered journal and syncs the
// containing directory. Callers must first finish recovery and retain rollback
// artifacts according to the application cleanup policy.
func RemoveRestoreJournal(stateDirectory string) error {
	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return err
	}
	journalPath := filepath.Join(root, restoreJournalFileName)
	if info, statErr := os.Lstat(journalPath); statErr == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%w: journal path is not a regular file", ErrRestoreJournalInvalid)
		}
	} else if errors.Is(statErr, os.ErrNotExist) {
		return nil
	} else {
		return fmt.Errorf("inspect SQLite restore journal: %w", statErr)
	}
	if err := os.Remove(journalPath); err != nil {
		return fmt.Errorf("remove SQLite restore journal: %w", err)
	}
	return syncRestoreDirectory(root)
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
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("%w: state directory is not a directory", ErrRestoreJournalInvalid)
	}
	// Persist the resolved spelling consistently. Windows may expose the same
	// directory through both a long path and an 8.3 alias; mixing those forms
	// across journal writes and reads breaks exact round-trip comparisons.
	return filepath.Clean(resolved), nil
}

func canonicalizeRestoreJournal(root string, journal RestoreJournal) (RestoreJournal, error) {
	if journal.Version != restoreJournalVersion ||
		strings.TrimSpace(journal.TransactionID) == "" ||
		len(journal.TransactionID) > 128 ||
		!validRestorePhase(journal.Phase) ||
		journal.ArchiveSize == 0 ||
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
	if journal.TargetPath == journal.StagingPath {
		return RestoreJournal{}, fmt.Errorf("%w: target and staging paths are identical", ErrRestoreJournalInvalid)
	}
	if journal.RollbackPath != "" {
		journal.RollbackPath, err = canonicalStatePath(root, journal.RollbackPath)
		if err != nil {
			return RestoreJournal{}, err
		}
		if journal.RollbackPath == journal.TargetPath || journal.RollbackPath == journal.StagingPath {
			return RestoreJournal{}, fmt.Errorf("%w: rollback path collides with another transaction path", ErrRestoreJournalInvalid)
		}
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
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return "", fmt.Errorf("%w: resolve transaction path parent: %v", ErrRestoreJournalInvalid, err)
	}
	resolvedCandidate := filepath.Join(parent, filepath.Base(absolute))
	if info, statErr := os.Lstat(absolute); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("%w: transaction path must not be a symbolic link", ErrRestoreJournalInvalid)
	} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return "", fmt.Errorf("%w: inspect transaction path: %v", ErrRestoreJournalInvalid, statErr)
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedCandidate)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("%w: transaction path escapes the application state directory", ErrRestoreJournalInvalid)
	}
	return resolvedCandidate, nil
}

func validRestorePhase(phase RestorePhase) bool {
	switch phase {
	case RestorePhasePrepared, RestorePhaseStaging, RestorePhasePreviousPreserved,
		RestorePhasePublished, RestorePhaseReopening, RestorePhaseCommitted:
		return true
	default:
		return false
	}
}

func syncRestoreDirectory(path string) error {
	// Windows does not expose directory handles that support File.Sync through
	// the standard library; file contents are still synced before rename.
	if runtime.GOOS == "windows" {
		return nil
	}
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open SQLite restore state directory for sync: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync SQLite restore state directory: %w", err)
	}
	return nil
}
