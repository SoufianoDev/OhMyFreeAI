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
	from, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return fmt.Errorf("encode temporary SQLite restore journal path: %w", err)
	}
	to, err := windows.UTF16PtrFromString(destination)
	if err != nil {
		return fmt.Errorf("encode SQLite restore journal path: %w", err)
	}
	return windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}

func removeRestoreJournal(string, string) error {
	return errRestoreJournalRemovalUnsupported
}
