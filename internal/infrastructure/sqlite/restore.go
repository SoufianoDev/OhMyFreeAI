package sqlite

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	moderncsqlite "modernc.org/sqlite"
)

var (
	ErrRestoreSourceInvalid = errors.New("SQLite restore source is not a regular backup file")
	ErrRestoreTargetInvalid = errors.New("SQLite restore target is not a regular database file")
	ErrRestoreTargetBusy    = errors.New("SQLite restore target is still in use")
)

type onlineRestoreConnection interface {
	NewRestore(sourceURI string) (*moderncsqlite.Backup, error)
}

type RestoreCandidate struct {
	Path     string
	Identity RestoreContentIdentity
}

// BuildRestoreCandidate creates a fully migrated and verified candidate in
// the private state directory without changing the canonical database. The
// returned file is a build artifact: the transaction executor must first
// persist its identity in the journal, then move it to the journal's recorded
// staging path with moveRestoreFileNoReplaceVerified. If an error occurs after
// file creation, Path is returned and the artifact is deliberately preserved.
func BuildRestoreCandidate(ctx context.Context, sourcePath, stateDirectory string, migrationFS fs.FS, migrationDir string) (RestoreCandidate, error) {
	var candidate RestoreCandidate
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return candidate, err
	}
	if migrationFS == nil {
		return candidate, ErrInvalidMigration
	}

	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return candidate, err
	}
	absoluteSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return candidate, fmt.Errorf("resolve SQLite restore candidate source: %w", err)
	}
	absoluteSource = filepath.Clean(absoluteSource)
	sourceInfo, err := os.Lstat(absoluteSource)
	if err != nil {
		return candidate, fmt.Errorf("inspect SQLite restore candidate source: %w", err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return candidate, ErrRestoreSourceInvalid
	}
	if err := verifyBackupDatabase(ctx, absoluteSource); err != nil {
		return candidate, fmt.Errorf("%w: %v", ErrRestoreSourceInvalid, err)
	}

	temporaryFile, err := os.CreateTemp(root, ".ohmyfreeai-restore-candidate-build-*.sqlite")
	if err != nil {
		return candidate, fmt.Errorf("create SQLite restore candidate: %w", err)
	}
	candidate.Path = temporaryFile.Name()
	if err := temporaryFile.Close(); err != nil {
		return restoreCandidateFailure(candidate, "close SQLite restore candidate", err)
	}
	if err := buildRestoreCandidateAtPath(ctx, absoluteSource, candidate.Path, migrationFS, migrationDir); err != nil {
		return restoreCandidateFailure(candidate, "build SQLite restore candidate", err)
	}

	identity, err := readRestoreContentIdentityPlatform(candidate.Path)
	if err != nil {
		return restoreCandidateFailure(candidate, "measure SQLite restore candidate identity", err)
	}
	candidate.Identity = identity
	return candidate, nil
}


// BuildRestoreCandidateAtPath creates a verified candidate at a caller-selected
// path inside the private state directory. The destination is created with
// O_EXCL, so an existing file is never truncated or replaced. The path is
// suitable for a journal-recorded staging location. If construction fails
// after creation, the partial artifact is returned and deliberately preserved.
func BuildRestoreCandidateAtPath(ctx context.Context, sourcePath, stateDirectory, candidatePath string, migrationFS fs.FS, migrationDir string) (RestoreCandidate, error) {
	var candidate RestoreCandidate
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return candidate, err
	}
	if migrationFS == nil {
		return candidate, ErrInvalidMigration
	}

	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return candidate, err
	}
	canonicalCandidate, err := canonicalStatePath(root, candidatePath)
	if err != nil {
		return candidate, err
	}
	if !samePath(filepath.Dir(canonicalCandidate), root) {
		return candidate, fmt.Errorf("%w: candidate must be a direct child of the private state directory", ErrRestoreJournalInvalid)
	}

	file, err := os.OpenFile(canonicalCandidate, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return candidate, fmt.Errorf("create journal-selected SQLite restore candidate without replacement: %w", err)
	}
	candidate.Path = canonicalCandidate
	if err := file.Close(); err != nil {
		return restoreCandidateFailure(candidate, "close journal-selected SQLite restore candidate", err)
	}
	if err := buildRestoreCandidateAtPath(ctx, sourcePath, canonicalCandidate, migrationFS, migrationDir); err != nil {
		return restoreCandidateFailure(candidate, "build journal-selected SQLite restore candidate", err)
	}

	identity, err := readRestoreContentIdentityPlatform(canonicalCandidate)
	if err != nil {
		return restoreCandidateFailure(candidate, "measure journal-selected SQLite restore candidate identity", err)
	}
	candidate.Identity = identity
	return candidate, nil
}

func restoreCandidateFailure(candidate RestoreCandidate, operation string, cause error) (RestoreCandidate, error) {
	if candidate.Path == "" {
		return candidate, fmt.Errorf("%s: %w", operation, cause)
	}
	return candidate, fmt.Errorf("%s; candidate preserved at %s: %w", operation, candidate.Path, cause)
}

// buildRestoreCandidateAtPath fills a caller-created empty regular file. It
// never removes the candidate or its sidecars on failure.
func buildRestoreCandidateAtPath(ctx context.Context, sourcePath, candidatePath string, migrationFS fs.FS, migrationDir string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if migrationFS == nil {
		return ErrInvalidMigration
	}

	absoluteSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("resolve SQLite restore source: %w", err)
	}
	absoluteCandidate, err := filepath.Abs(candidatePath)
	if err != nil {
		return fmt.Errorf("resolve SQLite restore candidate path: %w", err)
	}
	absoluteSource = filepath.Clean(absoluteSource)
	absoluteCandidate = filepath.Clean(absoluteCandidate)
	if samePath(absoluteSource, absoluteCandidate) {
		return ErrInvalidDatabasePath
	}

	sourceInfo, err := os.Lstat(absoluteSource)
	if err != nil {
		return fmt.Errorf("inspect SQLite restore source: %w", err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return ErrRestoreSourceInvalid
	}
	if err := verifyBackupDatabase(ctx, absoluteSource); err != nil {
		return fmt.Errorf("%w: %v", ErrRestoreSourceInvalid, err)
	}

	candidateInfo, err := os.Lstat(absoluteCandidate)
	if err != nil {
		return fmt.Errorf("inspect empty SQLite restore candidate: %w", err)
	}
	if !candidateInfo.Mode().IsRegular() || candidateInfo.Size() != 0 {
		return fmt.Errorf("SQLite restore candidate must be an empty regular file")
	}
	if err := ensureNoSQLiteSidecars(absoluteCandidate); err != nil {
		return err
	}

	candidateStore, err := OpenExisting(ctx, absoluteCandidate)
	if err != nil {
		return fmt.Errorf("open SQLite restore candidate: %w", err)
	}
	storeClosed := false
	defer func() {
		if !storeClosed {
			_ = candidateStore.Close()
		}
	}()

	if err := candidateStore.restoreFromPath(ctx, absoluteSource); err != nil {
		return err
	}
	if migrationDir == "" {
		migrationDir = "."
	}
	if err := candidateStore.Migrate(ctx, migrationFS, migrationDir); err != nil {
		return fmt.Errorf("migrate SQLite restore candidate: %w", err)
	}
	if err := verifyBackupDatabase(ctx, absoluteCandidate); err != nil {
		return fmt.Errorf("verify SQLite restore candidate: %w", err)
	}
	configuration, err := candidateStore.Load(ctx)
	if err != nil {
		return fmt.Errorf("load SQLite restore candidate configuration: %w", err)
	}
	if err := configuration.Validate(); err != nil {
		return fmt.Errorf("validate SQLite restore candidate configuration: %w", err)
	}
	if err := checkpointStore(ctx, candidateStore); err != nil {
		return err
	}
	closeErr := candidateStore.Close()
	storeClosed = true
	if closeErr != nil {
		return fmt.Errorf("close SQLite restore candidate: %w", closeErr)
	}
	if err := removeEmptySQLiteSidecars(absoluteCandidate); err != nil {
		return err
	}
	if err := syncSQLiteFile(absoluteCandidate); err != nil {
		return err
	}
	if err := ensureNoSQLiteSidecars(absoluteCandidate); err != nil {
		return err
	}
	return nil
}

// RestoreDatabase restores a verified SQLite snapshot into targetPath using
// SQLite's online backup API. The target must be closed by the caller. If a
// target database already exists, it is retained beside the new database as
// a private pre-restore rollback file. The returned path is empty when the
// target did not previously exist.
func RestoreDatabase(ctx context.Context, sourcePath, targetPath string, migrationFS fs.FS, migrationDir string) (rollbackPath string, err error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if migrationFS == nil {
		return "", ErrInvalidMigration
	}
	absoluteSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", fmt.Errorf("resolve SQLite restore source: %w", err)
	}
	absoluteTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve SQLite restore target: %w", err)
	}
	absoluteSource = filepath.Clean(absoluteSource)
	absoluteTarget = filepath.Clean(absoluteTarget)
	if samePath(absoluteSource, absoluteTarget) {
		return "", ErrInvalidDatabasePath
	}

	sourceInfo, err := os.Lstat(absoluteSource)
	if err != nil {
		return "", fmt.Errorf("inspect SQLite restore source: %w", err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return "", ErrRestoreSourceInvalid
	}
	if err := verifyBackupDatabase(ctx, absoluteSource); err != nil {
		return "", fmt.Errorf("%w: %v", ErrRestoreSourceInvalid, err)
	}
	if err := validateRestoreTarget(absoluteTarget); err != nil {
		return "", err
	}

	targetDirectory := filepath.Dir(absoluteTarget)
	if err := os.MkdirAll(targetDirectory, 0o700); err != nil {
		return "", fmt.Errorf("create SQLite restore directory: %w", err)
	}

	temporaryFile, err := os.CreateTemp(targetDirectory, ".ohmyfreeai-restore-*.sqlite")
	if err != nil {
		return "", fmt.Errorf("create temporary SQLite restore database: %w", err)
	}
	temporaryPath := temporaryFile.Name()
	if err := temporaryFile.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return "", fmt.Errorf("close temporary SQLite restore database: %w", err)
	}
	published := false
	defer func() {
		if !published {
			removeSQLiteFiles(temporaryPath)
		}
	}()

	if err := buildRestoreCandidateAtPath(ctx, absoluteSource, temporaryPath, migrationFS, migrationDir); err != nil {
		return "", err
	}

	if err := validateRestoreTarget(absoluteTarget); err != nil {
		return "", err
	}
	targetExists := false
	targetInfo, statErr := os.Lstat(absoluteTarget)
	switch {
	case statErr == nil:
		targetExists = true
		if !targetInfo.Mode().IsRegular() {
			return "", ErrRestoreTargetInvalid
		}
		if err := os.Chmod(absoluteTarget, 0o600); err != nil {
			return "", fmt.Errorf("restrict previous SQLite database permissions: %w", err)
		}
		rollbackFile, createErr := os.CreateTemp(targetDirectory, ".ohmyfreeai-pre-restore-*.sqlite")
		if createErr != nil {
			return "", fmt.Errorf("reserve SQLite rollback path: %w", createErr)
		}
		rollbackPath = rollbackFile.Name()
		if closeErr := rollbackFile.Close(); closeErr != nil {
			_ = os.Remove(rollbackPath)
			return "", fmt.Errorf("close SQLite rollback path: %w", closeErr)
		}
		if removeErr := os.Remove(rollbackPath); removeErr != nil {
			return "", fmt.Errorf("prepare SQLite rollback path: %w", removeErr)
		}
		if renameErr := os.Rename(absoluteTarget, rollbackPath); renameErr != nil {
			return "", fmt.Errorf("preserve previous SQLite database: %w", renameErr)
		}
	case errors.Is(statErr, os.ErrNotExist):
		targetExists = false
	default:
		return "", fmt.Errorf("recheck SQLite restore target: %w", statErr)
	}

	if err := os.Rename(temporaryPath, absoluteTarget); err != nil {
		if targetExists {
			if rollbackErr := os.Rename(rollbackPath, absoluteTarget); rollbackErr != nil {
				return "", fmt.Errorf("publish restored database: %v; restore previous database from %s: %w", err, rollbackPath, rollbackErr)
			}
		}
		return "", fmt.Errorf("publish restored SQLite database: %w", err)
	}
	published = true
	return rollbackPath, nil
}

func (s *Store) restoreFromPath(ctx context.Context, sourcePath string) error {
	if s == nil || s.db == nil || s.path == "" {
		return ErrStoreClosed
	}
	sourceURI := readOnlyDataSourceName(sourcePath)
	destinationConnection, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("reserve SQLite restore connection: %w", err)
	}
	defer destinationConnection.Close()

	if err := destinationConnection.Raw(func(rawConnection any) (callbackErr error) {
		driverConnection, ok := rawConnection.(onlineRestoreConnection)
		if !ok {
			return ErrBackupUnsupported
		}
		restore, err := driverConnection.NewRestore(sourceURI)
		if err != nil {
			return fmt.Errorf("initialize SQLite online restore: %w", err)
		}
		defer func() {
			if finishErr := restore.Finish(); callbackErr == nil && finishErr != nil {
				callbackErr = fmt.Errorf("finish SQLite online restore: %w", finishErr)
			}
		}()

		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			done, err := restore.Step(128)
			if err != nil {
				if isRetryableBackupError(err) {
					timer := time.NewTimer(25 * time.Millisecond)
					select {
					case <-ctx.Done():
						if !timer.Stop() {
							<-timer.C
						}
						return ctx.Err()
					case <-timer.C:
						continue
					}
				}
				return fmt.Errorf("copy SQLite restore pages: %w", err)
			}
			if !done {
				return nil
			}
		}
	}); err != nil {
		return err
	}
	return nil
}

func validateRestoreTarget(path string) error {
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect SQLite restore target: %w", err)
	}
	if err == nil && !info.Mode().IsRegular() {
		return ErrRestoreTargetInvalid
	}
	if err == nil {
		if err := verifySQLiteFileHeader(path); err != nil {
			return fmt.Errorf("%w: %v", ErrRestoreTargetInvalid, err)
		}
	}
	return ensureNoSQLiteSidecars(path)
}

func ensureNoSQLiteSidecars(path string) error {
	for _, suffix := range []string{"-wal", "-shm", "-journal"} {
		if _, err := os.Lstat(path + suffix); err == nil {
			return fmt.Errorf("%w: %s exists", ErrRestoreTargetBusy, suffix)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect SQLite sidecar %s: %w", suffix, err)
		}
	}
	return nil
}

func removeEmptySQLiteSidecars(path string) error {
	for _, suffix := range []string{"-wal", "-shm", "-journal"} {
		sidecar := path + suffix
		info, err := os.Lstat(sidecar)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect restored SQLite sidecar: %w", err)
		}
		if !info.Mode().IsRegular() || info.Size() != 0 {
			return fmt.Errorf("%w: restored sidecar %s is not empty", ErrRestoreTargetBusy, suffix)
		}
		if err := os.Remove(sidecar); err != nil {
			return fmt.Errorf("remove empty restored SQLite sidecar: %w", err)
		}
	}
	return nil
}

func checkpointStore(ctx context.Context, store *Store) error {
	var busy, logFrames, checkpointed int
	if err := store.db.QueryRowContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)").Scan(&busy, &logFrames, &checkpointed); err != nil {
		return fmt.Errorf("checkpoint restored SQLite database: %w", err)
	}
	if busy != 0 {
		return fmt.Errorf("%w: WAL checkpoint reported busy=%d log=%d checkpointed=%d", ErrRestoreTargetBusy, busy, logFrames, checkpointed)
	}
	return nil
}

func syncSQLiteFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open restored SQLite database for sync: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync restored SQLite database: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close restored SQLite database: %w", err)
	}
	return nil
}

func samePath(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		if strings.EqualFold(left, right) {
			return true
		}
	} else if left == right {
		return true
	}

	// String comparison alone misses hard links and Windows 8.3 aliases.
	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)
	if leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo) {
		return true
	}
	if runtime.GOOS == "windows" && strings.EqualFold(filepath.Base(left), filepath.Base(right)) {
		leftParent, leftParentErr := os.Stat(filepath.Dir(left))
		rightParent, rightParentErr := os.Stat(filepath.Dir(right))
		if leftParentErr == nil && rightParentErr == nil && os.SameFile(leftParent, rightParent) {
			return true
		}
	}
	return false
}


func verifySQLiteFileHeader(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open database header: %w", err)
	}
	defer file.Close()

	var header [16]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		return fmt.Errorf("read database header: %w", err)
	}
	if string(header[:]) != "SQLite format 3\x00" {
		return errors.New("file does not have the SQLite 3 database header")
	}
	return nil
}
