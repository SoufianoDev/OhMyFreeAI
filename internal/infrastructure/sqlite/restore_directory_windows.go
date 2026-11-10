//go:build windows

package sqlite

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func validateRestoreStateDirectoryPlatform(path string) error {
	handle, err := openRestoreDirectoryNoReparse(path)
	if err != nil {
		return fmt.Errorf("validate private SQLite restore state directory: %w", err)
	}
	if err := windows.CloseHandle(handle); err != nil {
		return fmt.Errorf("close validated SQLite restore state directory handle: %w", err)
	}
	return nil
}
