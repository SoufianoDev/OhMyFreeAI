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

	_ "modernc.org/sqlite"
)

const (
	driverName    = "sqlite"
	busyTimeoutMS = 5000
	maxOpenConns  = 4
	maxIdleConns  = 4
)

var (
	ErrInvalidDatabasePath = errors.New("invalid SQLite database path")
	ErrUnsupportedPath     = errors.New("SQLite database path is not supported")
	ErrStoreClosed         = errors.New("SQLite store is closed")
)

// Store owns the process-local SQLite connection pool.
type Store struct {
	db   *sql.DB
	path string
}

// Open creates or opens a private local database file and configures every
// connection with foreign-key enforcement and a bounded busy timeout.
func Open(ctx context.Context, databasePath string) (*Store, error) {
	if databasePath == "" || strings.TrimSpace(databasePath) == "" || databasePath == ":memory:" {
		return nil, ErrInvalidDatabasePath
	}

	absolutePath, err := filepath.Abs(databasePath)
	if err != nil {
		return nil, fmt.Errorf("resolve SQLite database path: %w", err)
	}
	absolutePath = filepath.Clean(absolutePath)

	uriPath := filepath.ToSlash(absolutePath)
	if runtime.GOOS == "windows" && strings.HasPrefix(uriPath, "//") {
		return nil, fmt.Errorf("%w: UNC/network paths are not supported", ErrUnsupportedPath)
	}
	if info, err := os.Lstat(absolutePath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("%w: database path must be a regular file, not a symlink or special file", ErrInvalidDatabasePath)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect SQLite database path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o700); err != nil {
		return nil, fmt.Errorf("create SQLite database directory: %w", err)
	}
	file, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create SQLite database file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close newly created SQLite database file: %w", err)
	}

	dsn, err := dataSourceName(absolutePath)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLite connection pool: %w", err)
	}
	closeOnError := func(err error) (*Store, error) {
		_ = db.Close()
		return nil, err
	}

	// Set WAL once before allowing the pool to create additional connections.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.PingContext(ctx); err != nil {
		return closeOnError(fmt.Errorf("ping SQLite database: %w", err))
	}

	var journalMode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode=WAL").Scan(&journalMode); err != nil {
		return closeOnError(fmt.Errorf("enable SQLite WAL mode: %w", err))
	}
	if !strings.EqualFold(journalMode, "wal") {
		return closeOnError(fmt.Errorf("SQLite refused WAL mode; active journal mode is %q", journalMode))
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	return &Store{db: db, path: absolutePath}, nil
}

func dataSourceName(absolutePath string) (string, error) {
	uriPath := filepath.ToSlash(absolutePath)
	if runtime.GOOS == "windows" && len(uriPath) >= 2 && uriPath[1] == ':' {
		uriPath = "/" + uriPath
	}
	u := &url.URL{Scheme: "file", Path: uriPath}
	query := url.Values{}
	query.Add("_pragma", "foreign_keys(1)")
	query.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busyTimeoutMS))
	query.Add("_pragma", "synchronous(NORMAL)")
	u.RawQuery = query.Encode()
	return u.String(), nil
}

// DB exposes the database/sql handle to infrastructure repositories.
func (s *Store) DB() *sql.DB {
	if s == nil {
		return nil
	}
	return s.db
}

// Path returns the canonical local database path.
func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

// Close releases the SQLite connection pool.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close SQLite database: %w", err)
	}
	return nil
}
