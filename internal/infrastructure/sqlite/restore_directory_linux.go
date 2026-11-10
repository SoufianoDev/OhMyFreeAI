//go:build linux

package sqlite

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func validateRestoreStateDirectoryPlatform(path string) error {
	directoryFD, err := unix.Openat2(unix.AT_FDCWD, path, &unix.OpenHow{
		Flags:   uint64(unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC),
		Resolve: unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		return fmt.Errorf("open SQLite restore state directory without symlink traversal: %w", err)
	}
	defer unix.Close(directoryFD)
	return verifyRestoreDirectoryPrivateFD(directoryFD)
}

func verifyRestoreDirectoryPrivateFD(directoryFD int) error {
	if directoryFD < 0 {
		return fmt.Errorf("SQLite restore directory descriptor is invalid")
	}
	var info unix.Stat_t
	if err := unix.Fstat(directoryFD, &info); err != nil {
		return fmt.Errorf("inspect SQLite restore directory handle: %w", err)
	}
	if info.Mode&unix.S_IFMT != unix.S_IFDIR {
		return fmt.Errorf("SQLite restore parent is not a directory")
	}
	uid := uint32(unix.Geteuid())
	if info.Uid != uid {
		return fmt.Errorf("SQLite restore directory owner UID %d does not match effective UID %d", info.Uid, uid)
	}
	permissions := info.Mode & 0o777
	if permissions&0o077 != 0 {
		return fmt.Errorf("SQLite restore directory mode %04o grants group/other access", permissions)
	}
	return nil
}
