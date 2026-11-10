//go:build linux

package sqlite

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func moveRestoreFileNoReplacePlatform(source, destination string) error {
	directory := filepath.Dir(source)
	// Pin the containing directory and reject symlinks in every path component.
	// If openat2 is unavailable, fail closed rather than falling back to a
	// path-based rename whose parent resolution can be redirected.
	dirfd, err := unix.Openat2(unix.AT_FDCWD, directory, &unix.OpenHow{
		Flags:   uint64(unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC),
		Resolve: unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		return err
	}
	defer unix.Close(dirfd)
	if err := verifyRestoreDirectoryPrivateFD(dirfd); err != nil {
		return fmt.Errorf("validate private SQLite restore parent directory: %w", err)
	}

	sourceName := filepath.Base(source)
	destinationName := filepath.Base(destination)
	if err := unix.Renameat2(dirfd, sourceName, dirfd, destinationName, unix.RENAME_NOREPLACE); err != nil {
		return err
	}
	// The rename changes the directory namespace. Sync the already-pinned
	// parent handle before reporting the move as durable; never reopen by path.
	if err := syncRestoreDirectoryHandle(dirfd); err != nil {
		return fmt.Errorf("sync SQLite restore parent directory after no-replace move: %w", err)
	}
	return nil
}

func syncRestoreDirectoryHandle(directoryFD int) error {
	if err := unix.Fsync(directoryFD); err != nil {
		return fmt.Errorf("fsync SQLite restore parent directory: %w", err)
	}
	return nil
}
