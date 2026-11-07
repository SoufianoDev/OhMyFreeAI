package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	moderncsqlite "modernc.org/sqlite"
)

var (
	ErrBackupDestinationExists = errors.New("SQLite backup destination already exists")
	ErrBackupUnsupported       = errors.New("SQLite driver does not expose the online backup API")
	ErrBackupIntegrity         = errors.New("SQLite backup failed integrity verification")
)

type onlineBackupConnection interface {
	NewBackup(destinationURI string) (*moderncsqlite.Backup, error)
}

// BackupTo creates a consistent SQLite snapshot using SQLite's online backup
// API. It does not copy the live database file or its WAL sidecar.
func (s *Store) BackupTo(ctx context.Context, destinationPath string) error {
	if s == nil || s.db == nil || s.path == "" {
		return ErrStoreClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(destinationPath) == "" {
		return ErrInvalidDatabasePath
	}

	absoluteDestination, err := filepath.Abs(destinationPath)
	if err != nil {
		return fmt.Errorf("resolve SQLite backup destination: %w", err)
	}
	absoluteDestination = filepath.Clean(absoluteDestination)
	if absoluteDestination == s.path {
		return ErrInvalidDatabasePath
	}
	directory := filepath.Dir(absoluteDestination)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create SQLite backup directory: %w", err)
	}
	if _, err := os.Lstat(absoluteDestination); err == nil {
		return ErrBackupDestinationExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect SQLite backup destination: %w", err)
	}

	temporaryFile, err := os.CreateTemp(directory, ".ohmyfreeai-backup-*.sqlite")
	if err != nil {
		return fmt.Errorf("create temporary SQLite backup: %w", err)
	}
	temporaryPath := temporaryFile.Name()
	if err := temporaryFile.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("close temporary SQLite backup: %w", err)
	}
	published := false
	defer func() {
		if !published {
			removeSQLiteFiles(temporaryPath)
		}
	}()

	destinationURI := backupDataSourceName(temporaryPath)

	sourceConnection, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("reserve SQLite source connection: %w", err)
	}
	defer sourceConnection.Close()

	if err := sourceConnection.Raw(func(rawConnection any) (callbackErr error) {
		driverConnection, ok := rawConnection.(onlineBackupConnection)
		if !ok {
			return ErrBackupUnsupported
		}
		backup, err := driverConnection.NewBackup(destinationURI)
		if err != nil {
			return fmt.Errorf("initialize SQLite online backup: %w", err)
		}
		defer func() {
			if finishErr := backup.Finish(); callbackErr == nil && finishErr != nil {
				callbackErr = fmt.Errorf("finish SQLite online backup: %w", finishErr)
			}
		}()

		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			done, err := backup.Step(128)
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
				return fmt.Errorf("copy SQLite backup pages: %w", err)
			}
			if !done {
				return nil
			}
		}
	}); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if err := verifyBackupDatabase(ctx, temporaryPath); err != nil {
		return err
	}
	if err := os.Chmod(temporaryPath, 0o600); err != nil {
		return fmt.Errorf("restrict SQLite backup permissions: %w", err)
	}
	backupFile, err := os.OpenFile(temporaryPath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open completed SQLite backup: %w", err)
	}
	if err := backupFile.Sync(); err != nil {
		_ = backupFile.Close()
		return fmt.Errorf("sync completed SQLite backup: %w", err)
	}
	if err := backupFile.Close(); err != nil {
		return fmt.Errorf("close completed SQLite backup: %w", err)
	}

	if _, err := os.Lstat(absoluteDestination); err == nil {
		return ErrBackupDestinationExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("recheck SQLite backup destination: %w", err)
	}
	if err := os.Rename(temporaryPath, absoluteDestination); err != nil {
		return fmt.Errorf("publish SQLite backup: %w", err)
	}
	published = true
	return nil
}

func backupDataSourceName(absolutePath string) string {
	uriPath := filepath.ToSlash(absolutePath)
	if runtime.GOOS == "windows" && len(uriPath) >= 2 && uriPath[1] == ':' {
		uriPath = "/" + uriPath
	}
	u := &url.URL{Scheme: "file", Path: uriPath}
	query := url.Values{}
	query.Set("mode", "rwc")
	query.Add("_pragma", "synchronous(FULL)")
	u.RawQuery = query.Encode()
	return u.String()
}

func readOnlyDataSourceName(absolutePath string) string {
	uriPath := filepath.ToSlash(absolutePath)
	if runtime.GOOS == "windows" && len(uriPath) >= 2 && uriPath[1] == ':' {
		uriPath = "/" + uriPath
	}
	u := &url.URL{Scheme: "file", Path: uriPath}
	query := url.Values{}
	query.Set("mode", "ro")
	query.Add("_pragma", "foreign_keys(1)")
	query.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busyTimeoutMS))
	u.RawQuery = query.Encode()
	return u.String()
}

func verifyBackupDatabase(ctx context.Context, path string) error {
	db, err := sql.Open(driverName, readOnlyDataSourceName(path))
	if err != nil {
		return fmt.Errorf("open SQLite backup for verification: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping SQLite backup: %w", err)
	}

	integrityRows, err := db.QueryContext(ctx, "PRAGMA integrity_check")
	if err != nil {
		return fmt.Errorf("run SQLite backup integrity check: %w", err)
	}
	integrityCount := 0
	for integrityRows.Next() {
		var result string
		if err := integrityRows.Scan(&result); err != nil {
			_ = integrityRows.Close()
			return fmt.Errorf("read SQLite integrity result: %w", err)
		}
		integrityCount++
		if result != "ok" {
			_ = integrityRows.Close()
			return fmt.Errorf("%w: %s", ErrBackupIntegrity, result)
		}
	}
	if err := integrityRows.Err(); err != nil {
		_ = integrityRows.Close()
		return fmt.Errorf("iterate SQLite integrity results: %w", err)
	}
	if err := integrityRows.Close(); err != nil {
		return fmt.Errorf("close SQLite integrity results: %w", err)
	}
	if integrityCount == 0 {
		return fmt.Errorf("%w: integrity_check returned no rows", ErrBackupIntegrity)
	}

	foreignKeyRows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return fmt.Errorf("run SQLite backup foreign-key check: %w", err)
	}
	if foreignKeyRows.Next() {
		var table string
		var rowID sql.NullInt64
		var parent string
		var constraintID int
		scanErr := foreignKeyRows.Scan(&table, &rowID, &parent, &constraintID)
		_ = foreignKeyRows.Close()
		if scanErr != nil {
			return fmt.Errorf("read SQLite foreign-key violation: %w", scanErr)
		}
		return fmt.Errorf("%w: table=%s rowid=%v parent=%s constraint=%d", ErrBackupIntegrity, table, rowID, parent, constraintID)
	}
	if err := foreignKeyRows.Err(); err != nil {
		_ = foreignKeyRows.Close()
		return fmt.Errorf("iterate SQLite foreign-key results: %w", err)
	}
	if err := foreignKeyRows.Close(); err != nil {
		return fmt.Errorf("close SQLite foreign-key results: %w", err)
	}

	hasHistory, version, err := currentMigrationVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("verify SQLite backup migration history: %w", err)
	}
	if !hasHistory || version < 1 {
		return fmt.Errorf("%w: backup contains no applied migration history", ErrBackupIntegrity)
	}
	return nil
}

func isRetryableBackupError(err error) bool {
	var sqliteError *moderncsqlite.Error
	if !errors.As(err, &sqliteError) {
		return false
	}
	// SQLite primary result codes SQLITE_BUSY and SQLITE_LOCKED are 5 and 6.
	primaryCode := sqliteError.Code() & 0xff
	return primaryCode == 5 || primaryCode == 6
}

func removeSQLiteFiles(path string) {
	for _, candidate := range []string{path, path + "-journal", path + "-wal", path + "-shm"} {
		_ = os.Remove(candidate)
	}
}


// SnapshotSchemaVersion validates a self-contained SQLite snapshot and returns
// its contiguous applied migration version.
func SnapshotSchemaVersion(ctx context.Context, path string) (int, error) {
	if err := verifyBackupDatabase(ctx, path); err != nil {
		return 0, err
	}
	db, err := sql.Open(driverName, readOnlyDataSourceName(path))
	if err != nil {
		return 0, fmt.Errorf("open SQLite snapshot schema history: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	hasHistory, version, err := currentMigrationVersion(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("read SQLite snapshot schema version: %w", err)
	}
	if !hasHistory || version < 1 {
		return 0, fmt.Errorf("%w: snapshot has no applied schema version", ErrBackupIntegrity)
	}
	return version, nil
}
