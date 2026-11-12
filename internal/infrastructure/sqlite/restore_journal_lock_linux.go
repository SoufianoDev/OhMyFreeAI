//go:build linux

package sqlite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

func acquireRestoreJournalLockPlatform(ctx context.Context, root string) (func() error, error) {
	directory, err := unix.Openat2(unix.AT_FDCWD, root, &unix.OpenHow{
		Flags:   uint64(unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC),
		Resolve: unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		return nil, fmt.Errorf("open Restore journal lock directory without symlink traversal: %w", err)
	}
	if err := verifyRestoreDirectoryPrivateFD(directory); err != nil {
		_ = unix.Close(directory)
		return nil, fmt.Errorf("validate Restore journal lock directory: %w", err)
	}

	lockFD, err := openRestoreJournalLockAt(directory)
	if err != nil {
		_ = unix.Close(directory)
		return nil, err
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			_ = unix.Close(lockFD)
			_ = unix.Close(directory)
			return nil, err
		}
		err := unix.Flock(lockFD, unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			break
		}
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if !errors.Is(err, unix.EWOULDBLOCK) && !errors.Is(err, unix.EAGAIN) {
			_ = unix.Close(lockFD)
			_ = unix.Close(directory)
			return nil, fmt.Errorf("lock Restore journal ledger: %w", err)
		}
		select {
		case <-ctx.Done():
			_ = unix.Close(lockFD)
			_ = unix.Close(directory)
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}

	if err := ctx.Err(); err != nil {
		_ = unix.Flock(lockFD, unix.LOCK_UN)
		_ = unix.Close(lockFD)
		_ = unix.Close(directory)
		return nil, err
	}

	return func() error {
		unlockErr := unix.Flock(lockFD, unix.LOCK_UN)
		closeFileErr := unix.Close(lockFD)
		closeDirectoryErr := unix.Close(directory)
		var result error
		if unlockErr != nil {
			result = errors.Join(result, fmt.Errorf("unlock Restore journal ledger: %w", unlockErr))
		}
		if closeFileErr != nil {
			result = errors.Join(result, fmt.Errorf("close Restore journal lock file: %w", closeFileErr))
		}
		if closeDirectoryErr != nil {
			result = errors.Join(result, fmt.Errorf("close Restore journal lock directory: %w", closeDirectoryErr))
		}
		return result
	}, nil
}

func openRestoreJournalLockAt(directory int) (int, error) {
	resolve := uint64(unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS)
	for attempt := 0; attempt < 4; attempt++ {
		created, err := unix.Openat2(directory, restoreJournalLockFileName, &unix.OpenHow{
			Flags:   uint64(unix.O_RDWR | unix.O_CREAT | unix.O_EXCL | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK),
			Mode:    0o600,
			Resolve: resolve,
		})
		if err == nil {
			if err := unix.Fchmod(created, 0o600); err != nil {
				_ = unix.Close(created)
				return -1, fmt.Errorf("restrict Restore journal lock file permissions: %w", err)
			}
			if err := validateRestoreJournalLockFD(created); err != nil {
				_ = unix.Close(created)
				return -1, err
			}
			return created, nil
		}
		if !errors.Is(err, unix.EEXIST) {
			return -1, fmt.Errorf("create Restore journal lock file without following links: %w", err)
		}

		probe, err := unix.Openat2(directory, restoreJournalLockFileName, &unix.OpenHow{
			Flags:   uint64(unix.O_PATH | unix.O_CLOEXEC | unix.O_NOFOLLOW),
			Resolve: resolve,
		})
		if errors.Is(err, unix.ENOENT) {
			continue
		}
		if err != nil {
			return -1, fmt.Errorf("inspect existing Restore journal lock without following links: %w", err)
		}
		var probed unix.Stat_t
		if err := unix.Fstat(probe, &probed); err != nil {
			_ = unix.Close(probe)
			return -1, fmt.Errorf("inspect existing Restore journal lock: %w", err)
		}
		if err := validateRestoreJournalLockInfo(&probed); err != nil {
			_ = unix.Close(probe)
			return -1, err
		}

		opened, err := unix.Openat2(directory, restoreJournalLockFileName, &unix.OpenHow{
			Flags:   uint64(unix.O_RDWR | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK),
			Resolve: resolve,
		})
		_ = unix.Close(probe)
		if errors.Is(err, unix.ENOENT) {
			continue
		}
		if err != nil {
			return -1, fmt.Errorf("open existing Restore journal lock without following links: %w", err)
		}
		var openedInfo unix.Stat_t
		if err := unix.Fstat(opened, &openedInfo); err != nil {
			_ = unix.Close(opened)
			return -1, fmt.Errorf("inspect opened Restore journal lock: %w", err)
		}
		if err := validateRestoreJournalLockInfo(&openedInfo); err != nil {
			_ = unix.Close(opened)
			return -1, err
		}
		if openedInfo.Dev != probed.Dev || openedInfo.Ino != probed.Ino {
			_ = unix.Close(opened)
			return -1, fmt.Errorf("Restore journal lock file changed while it was being opened")
		}
		return opened, nil
	}
	return -1, fmt.Errorf("Restore journal lock file changed repeatedly while it was being opened")
}

func validateRestoreJournalLockFD(fd int) error {
	var info unix.Stat_t
	if err := unix.Fstat(fd, &info); err != nil {
		return fmt.Errorf("inspect Restore journal lock file: %w", err)
	}
	return validateRestoreJournalLockInfo(&info)
}

func validateRestoreJournalLockInfo(info *unix.Stat_t) error {
	if info.Mode&unix.S_IFMT != unix.S_IFREG || info.Nlink != 1 {
		return fmt.Errorf("Restore journal lock is not a single-link regular file")
	}
	if info.Uid != uint32(unix.Geteuid()) || info.Mode&0o7777 != 0o600 {
		return fmt.Errorf("Restore journal lock owner or permissions are not private to the current user")
	}
	if info.Size != 0 {
		return fmt.Errorf("Restore journal lock file contains unexpected data")
	}
	return nil
}
