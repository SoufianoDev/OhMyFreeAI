//go:build !windows

package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
)

func publishRestoreJournal(source, destination string) error {
	if err := os.Rename(source, destination); err != nil {
		return err
	}
	return syncRestoreDirectory(filepath.Dir(destination))
}

func removeRestoreJournal(string, string) error {
	return errRestoreJournalRemovalUnsupported
}

func syncRestoreDirectory(path string) error {
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
