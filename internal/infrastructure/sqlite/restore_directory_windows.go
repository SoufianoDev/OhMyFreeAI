//go:build windows

package sqlite

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// canonicalRestoreDirectoryPathPlatform expands Windows 8.3 aliases so paths
// returned by CreateTemp and paths supplied by callers use one stable spelling.
func canonicalRestoreDirectoryPathPlatform(path string) (string, error) {
	input, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", fmt.Errorf("encode SQLite restore directory path: %w", err)
	}
	size := uint32(260)
	for size <= 32768 {
		buffer := make([]uint16, size)
		n, callErr := windows.GetLongPathName(input, &buffer[0], size)
		if callErr != nil {
			return "", fmt.Errorf("expand SQLite restore directory long path: %w", callErr)
		}
		if n == 0 {
			return "", fmt.Errorf("expand SQLite restore directory long path returned an empty path")
		}
		if n < size {
			return filepath.Clean(windows.UTF16ToString(buffer[:n])), nil
		}
		size = n + 1
	}
	return "", fmt.Errorf("SQLite restore directory path exceeds the Windows path-length limit")
}

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
