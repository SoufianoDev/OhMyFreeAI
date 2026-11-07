package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrMigrationGap           = errors.New("SQLite migration version gap")
	ErrUntrackedSchema        = errors.New("SQLite schema exists without an applied migration record")
	ErrMigrationHistoryCorrupt = errors.New("SQLite migration history is not contiguous")
	ErrMigrationVersionTooNew = errors.New("SQLite database schema is newer than this application")
	ErrInvalidMigration       = errors.New("invalid SQLite migration")
)

type migrationFile struct {
	name        string
	version     int
	description string
}

// Migrate applies forward SQL migrations in ascending version order. Each
// migration and its schema_migrations record commit in the same transaction.
func (s *Store) Migrate(ctx context.Context, source fs.FS, directory string) error {
	if s == nil || s.db == nil {
		return ErrStoreClosed
	}
	if source == nil {
		return fmt.Errorf("%w: migration filesystem is required", ErrInvalidMigration)
	}
	if directory == "" {
		directory = "."
	}

	names, err := fs.Glob(source, path.Join(directory, "*.sql"))
	if err != nil {
		return fmt.Errorf("list SQLite migrations: %w", err)
	}

	files := make([]migrationFile, 0, len(names))
	seen := make(map[int]struct{}, len(names))
	for _, name := range names {
		base := path.Base(name)
		if strings.HasSuffix(base, ".down.sql") {
			continue
		}
		prefix, description, ok := strings.Cut(base, "_")
		if !ok || !strings.HasSuffix(description, ".sql") {
			return fmt.Errorf("%w: filename %q must use NNN_description.sql", ErrInvalidMigration, base)
		}
		version, err := strconv.Atoi(prefix)
		if err != nil || version < 1 {
			return fmt.Errorf("%w: invalid version in filename %q", ErrInvalidMigration, base)
		}
		if _, duplicate := seen[version]; duplicate {
			return fmt.Errorf("%w: duplicate version %d", ErrInvalidMigration, version)
		}
		seen[version] = struct{}{}
		description = strings.TrimSuffix(description, ".sql")
		files = append(files, migrationFile{name: name, version: version, description: description})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })

	hasHistory, currentVersion, err := currentMigrationVersion(ctx, s.db)
	if err != nil {
		return err
	}
	if hasHistory && currentVersion == 0 {
		return ErrUntrackedSchema
	}
	latestAvailableVersion := 0
	if len(files) > 0 {
		latestAvailableVersion = files[len(files)-1].version
	}
	if currentVersion > latestAvailableVersion {
		return fmt.Errorf("%w: database=%03d application=%03d", ErrMigrationVersionTooNew, currentVersion, latestAvailableVersion)
	}

	for _, migration := range files {
		if migration.version <= currentVersion {
			continue
		}
		if migration.version != currentVersion+1 {
			return fmt.Errorf("%w: expected %03d, found %03d", ErrMigrationGap, currentVersion+1, migration.version)
		}
		if err := s.applyMigration(ctx, source, migration); err != nil {
			return err
		}
		currentVersion = migration.version
	}
	return nil
}

func currentMigrationVersion(ctx context.Context, db *sql.DB) (bool, int, error) {
	var tableCount int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableCount); err != nil {
		return false, 0, fmt.Errorf("inspect SQLite migration history: %w", err)
	}
	if tableCount == 0 {
		return false, 0, nil
	}
	var count, minimum, maximum int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(MIN(version), 0), COALESCE(MAX(version), 0) FROM schema_migrations",
	).Scan(&count, &minimum, &maximum); err != nil {
		return true, 0, fmt.Errorf("read SQLite migration history: %w", err)
	}
	if count == 0 {
		return true, 0, nil
	}
	if minimum != 1 || maximum != count {
		return true, 0, fmt.Errorf("%w: rows=%d min=%d max=%d", ErrMigrationHistoryCorrupt, count, minimum, maximum)
	}
	return true, maximum, nil
}

func (s *Store) applyMigration(ctx context.Context, source fs.FS, migration migrationFile) error {
	script, err := fs.ReadFile(source, migration.name)
	if err != nil {
		return fmt.Errorf("read SQLite migration %s: %w", migration.name, err)
	}
	statements, err := splitSQLStatements(string(script))
	if err != nil {
		return fmt.Errorf("%w in %s: %v", ErrInvalidMigration, migration.name, err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin SQLite migration %s: %w", migration.name, err)
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		return cause
	}
	for index, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return rollback(fmt.Errorf("execute SQLite migration %s statement %d: %w", migration.name, index+1, err))
		}
	}

	var tableCount int
	if err := tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableCount); err != nil {
		return rollback(fmt.Errorf("verify schema_migrations after %s: %w", migration.name, err))
	}
	if tableCount != 1 {
		return rollback(fmt.Errorf("%w: %s did not create schema_migrations", ErrInvalidMigration, migration.name))
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO schema_migrations(version, description, applied_at) VALUES (?, ?, ?)",
		migration.version, migration.description, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return rollback(fmt.Errorf("record SQLite migration %s: %w", migration.name, err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit SQLite migration %s: %w", migration.name, err)
	}
	return nil
}

// splitSQLStatements separates SQL at semicolons outside strings and comments.
// Project migrations must keep statements self-contained; trigger bodies with
// internal semicolons are not supported by this deliberately small splitter.
func splitSQLStatements(script string) ([]string, error) {
	const (
		normal = iota
		singleQuoted
		doubleQuoted
		backtickQuoted
		bracketQuoted
		lineComment
		blockComment
	)
	state := normal
	var current strings.Builder
	statements := make([]string, 0, 16)
	flush := func() {
		statement := strings.TrimSpace(current.String())
		if statement != "" {
			statements = append(statements, statement)
		}
		current.Reset()
	}

	for i := 0; i < len(script); i++ {
		char := script[i]
		next := byte(0)
		if i+1 < len(script) {
			next = script[i+1]
		}

		switch state {
		case normal:
			switch {
			case char == '-' && next == '-':
				state = lineComment
				i++
			case char == '/' && next == '*':
				state = blockComment
				i++
			case char == 39:
				state = singleQuoted
				current.WriteByte(char)
			case char == '"':
				state = doubleQuoted
				current.WriteByte(char)
			case char == 96:
				state = backtickQuoted
				current.WriteByte(char)
			case char == '[':
				state = bracketQuoted
				current.WriteByte(char)
			case char == ';':
				flush()
			default:
				current.WriteByte(char)
			}
		case singleQuoted, doubleQuoted, backtickQuoted:
			current.WriteByte(char)
			var quote byte
			switch state {
			case singleQuoted:
				quote = 39
			case doubleQuoted:
				quote = '"'
			case backtickQuoted:
				quote = 96
			}
			if char == quote {
				if next == quote {
					current.WriteByte(next)
					i++
				} else {
					state = normal
				}
			}
		case bracketQuoted:
			current.WriteByte(char)
			if char == ']' {
				state = normal
			}
		case lineComment:
			if char == 10 {
				current.WriteByte(10)
				state = normal
			}
		case blockComment:
			if char == '*' && next == '/' {
				state = normal
				i++
			}
		}
	}

	if state == singleQuoted || state == doubleQuoted || state == backtickQuoted || state == bracketQuoted || state == blockComment {
		return nil, errors.New("unterminated SQL quote or block comment")
	}
	flush()
	return statements, nil
}
