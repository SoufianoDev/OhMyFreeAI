//go:build linux

package sqlite

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// readRestoreJournalFilePlatform reads the ledger from one no-follow file
// descriptor. The private parent is pinned before the leaf is opened.
func readRestoreJournalFilePlatform(path string, maxBytes int) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("%w: journal ledger size limit is invalid", ErrRestoreJournalInvalid)
	}
	if err := validateRestoreIdentityLeaf(filepath.Base(path)); err != nil {
		return nil, fmt.Errorf("%w: journal ledger filename is invalid: %v", ErrRestoreJournalInvalid, err)
	}

	parent, err := unix.Openat2(unix.AT_FDCWD, filepath.Dir(path), &unix.OpenHow{
		Flags:   uint64(unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC),
		Resolve: unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		if errors.Is(err, unix.ENOENT) {
			return nil, fmt.Errorf("open SQLite restore journal parent: %w", os.ErrNotExist)
		}
		return nil, fmt.Errorf("open SQLite restore journal parent without symlink traversal: %w", err)
	}
	defer unix.Close(parent)
	if err := verifyRestoreDirectoryPrivateFD(parent); err != nil {
		return nil, fmt.Errorf("validate private SQLite restore journal directory: %w", err)
	}

	fd, err := unix.Openat2(parent, filepath.Base(path), &unix.OpenHow{
		Flags:   uint64(unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK),
		Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		if errors.Is(err, unix.ENOENT) {
			return nil, fmt.Errorf("open SQLite restore journal ledger: %w", os.ErrNotExist)
		}
		if errors.Is(err, unix.ELOOP) {
			return nil, fmt.Errorf("%w: journal ledger must not be a symbolic link", ErrRestoreJournalInvalid)
		}
		return nil, fmt.Errorf("open SQLite restore journal ledger without following links: %w", err)
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("wrap SQLite restore journal descriptor")
	}
	defer file.Close()

	var before unix.Stat_t
	if err := unix.Fstat(fd, &before); err != nil {
		return nil, fmt.Errorf("inspect opened SQLite restore journal ledger: %w", err)
	}
	if before.Mode&unix.S_IFMT != unix.S_IFREG || before.Size <= 0 ||
		before.Size > int64(maxBytes) || before.Nlink != 1 ||
		before.Uid != uint32(unix.Geteuid()) || before.Mode&0o7777 != 0o600 {
		return nil, fmt.Errorf("%w: journal ledger must be a bounded, private, single-link regular file", ErrRestoreJournalInvalid)
	}

	data := make([]byte, int(before.Size))
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, fmt.Errorf("read opened SQLite restore journal ledger: %w", err)
	}
	var after unix.Stat_t
	if err := unix.Fstat(fd, &after); err != nil {
		return nil, fmt.Errorf("reinspect opened SQLite restore journal ledger: %w", err)
	}
	if after.Mode&unix.S_IFMT != unix.S_IFREG || after.Dev != before.Dev ||
		after.Ino != before.Ino || after.Size != before.Size ||
		after.Nlink != before.Nlink || after.Uid != before.Uid ||
		after.Mode&0o7777 != before.Mode&0o7777 ||
		after.Mtim != before.Mtim || after.Ctim != before.Ctim {
		return nil, fmt.Errorf("%w: journal ledger changed while being read", ErrRestoreJournalInvalid)
	}
	return data, nil
}
