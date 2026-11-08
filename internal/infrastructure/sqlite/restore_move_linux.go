//go:build linux

package sqlite

import (
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

	sourceName := filepath.Base(source)
	destinationName := filepath.Base(destination)
	return unix.Renameat2(dirfd, sourceName, dirfd, destinationName, unix.RENAME_NOREPLACE)
}
