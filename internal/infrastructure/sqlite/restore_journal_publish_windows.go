//go:build windows

package sqlite

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func publishRestoreJournal(source, destination string) error {
	if !samePath(filepath.Dir(source), filepath.Dir(destination)) {
		return fmt.Errorf("SQLite restore journal publication must stay within one directory")
	}
	directoryHandle, err := openRestoreDirectoryNoReparse(filepath.Dir(destination))
	if err != nil {
		return fmt.Errorf("open SQLite restore journal directory: %w", err)
	}
	defer windows.CloseHandle(directoryHandle)

	guardHandle, directoryPath, err := openRestoreDirectoryWriteThroughGuard(directoryHandle)
	if err != nil {
		return fmt.Errorf("guard SQLite restore journal directory: %w", err)
	}
	defer windows.CloseHandle(guardHandle)

	sourceHandle, err := openRestoreSourceWriteThrough(directoryPath, filepath.Base(source))
	if err != nil {
		return fmt.Errorf("open SQLite restore journal temporary file: %w", err)
	}
	defer windows.CloseHandle(sourceHandle)

	if err := renameRestoreHandleReplace(sourceHandle, guardHandle, filepath.Base(destination)); err != nil {
		return fmt.Errorf("publish SQLite restore journal with write-through handle rename: %w", err)
	}
	return nil
}

func removeRestoreJournal(string, string) error {
	return errRestoreJournalRemovalUnsupported
}
