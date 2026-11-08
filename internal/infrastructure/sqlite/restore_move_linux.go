//go:build linux

package sqlite

import "golang.org/x/sys/unix"

func moveRestoreFileNoReplacePlatform(source, destination string) error {
	return unix.Renameat2(unix.AT_FDCWD, source, unix.AT_FDCWD, destination, unix.RENAME_NOREPLACE)
}
