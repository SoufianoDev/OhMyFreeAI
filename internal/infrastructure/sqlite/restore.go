package sqlite

import (
	"context"
	"errors"
	"fmt"
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
	if err := validateRestoreTarget(absoluteTarget); err != nil {
		return "", err
	}

	targetDirectory := filepath.Dir(absoluteTarget)
	if err := os.MkdirAll(targetDirectory, 0o700); err != nil {
		return "", fmt.Errorf("create SQLite restore directory: %w", err)
	}
	if err := os.Chmod(targetDirectory, 0o700); err != nil {
		return "", fmt.Errorf("restrict SQLite restore directory: %w", err)
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

	temporaryStore, err := Open(ctx, temporaryPath)
	if err != nil {
		return "", fmt.Errorf("open temporary SQLite restore database: %w", err)
	}
	storeClosed := false
	defer func() {
		if !storeClosed {
			_ = temporaryStore.Close()
		}
	}()

	if err := temporaryStore.restoreFromPath(ctx, absoluteSource); err != nil {
		return "", err
	}
	if migrationDir == "" {
		migrationDir = "."
	}
	if err := temporaryStore.Migrate(ctx, migrationFS, migrationDir); err != nil {
		return "", fmt.Errorf("migrate restored SQLite database: %w", err)
	}
	if err := verifyBackupDatabase(ctx, temporaryPath); err != nil {
		return "", fmt.Errorf("verify restored SQLite database: %w", err)
	}
	if err := checkpointStore(ctx, temporaryStore); err != nil {
		return "", err
	}
	if err := temporaryStore.Close(); err != nil {
		storeClosed = true
		return "", fmt.Errorf("close restored SQLite database: %w", err)
	}
	storeClosed = true
	if err := removeEmptySQLiteSidecars(temporaryPath); err != nil {
		return "", err
	}
	if err := syncSQLiteFile(temporaryPath); err != nil {
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
		return strings.EqualFold(left, right)
	}
	return left == right
}
