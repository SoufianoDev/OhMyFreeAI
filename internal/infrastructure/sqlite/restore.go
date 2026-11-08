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
	ErrRestoreExecutorUnavailable = errors.New("SQLite restore is disabled until the safe filesystem executor is implemented")
)

type onlineRestoreConnection interface {
	NewRestore(sourceURI string) (*moderncsqlite.Backup, error)
}

// RestoreDatabase is intentionally fail-closed. The former path-based restore
// sequence is not safe to expose until the journaled, identity-bound filesystem
// executor and its crash-recovery tests are complete.
func RestoreDatabase(ctx context.Context, sourcePath, targetPath string, migrationFS fs.FS, migrationDir string) (rollbackPath string, err error) {
	return "", ErrRestoreExecutorUnavailable
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
		return strings.EqualFold(left, right)
	}
	return left == right
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
