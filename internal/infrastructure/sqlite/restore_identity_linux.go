//go:build linux

package sqlite

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)


func restoreArtifactMetadataUnchanged(before, after os.FileInfo) (bool, error) {
	beforeStat, ok := before.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("unexpected Linux stat metadata type %T", before.Sys())
	}
	afterStat, ok := after.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("unexpected Linux stat metadata type %T", after.Sys())
	}

	// ctime is inode status-change time, not creation time. Compare the
	// pinned file object's identity and metadata around hashing; this catches
	// chmod/chown/link-count and content updates even when size/mtime are
	// restored. It is detection, not a filesystem-independent lock.
	return beforeStat.Dev == afterStat.Dev &&
		beforeStat.Ino == afterStat.Ino &&
		beforeStat.Mode == afterStat.Mode &&
		beforeStat.Nlink == afterStat.Nlink &&
		beforeStat.Uid == afterStat.Uid &&
		beforeStat.Gid == afterStat.Gid &&
		beforeStat.Size == afterStat.Size &&
		beforeStat.Mtim == afterStat.Mtim &&
		beforeStat.Ctim == afterStat.Ctim, nil
}

func openRestoreArtifactFile(stateDirectory, path string) (*os.File, error) {
	root, relative, err := restoreArtifactRelativePath(stateDirectory, path)
	if err != nil {
		return nil, err
	}

	rootFD, err := unix.Openat2(unix.AT_FDCWD, root, &unix.OpenHow{
		Flags:   uint64(unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC | unix.O_NOFOLLOW),
		Resolve: unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		return nil, fmt.Errorf("open canonical restore state directory without symlinks: %w", err)
	}
	defer unix.Close(rootFD)

	fd, err := unix.Openat2(rootFD, relative, &unix.OpenHow{
		Flags:   uint64(unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK),
		Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		return nil, fmt.Errorf("open restore artifact beneath pinned state directory: %w", err)
	}

	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("wrap opened restore artifact")
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("inspect opened restore artifact: %w", err)
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, fmt.Errorf("restore artifact is not a regular file")
	}
	return file, nil
}

// normalizeWindowsRestorePath is unreachable on Linux; it keeps the shared
// path-validation file buildable without importing Windows APIs.
func normalizeWindowsRestorePath(path string) (string, error) { return path, nil }
