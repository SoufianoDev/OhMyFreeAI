//go:build linux

package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// readRestoreContentIdentityPlatform pins the parent directory, opens only
// the leaf without following links, and hashes bytes from that same descriptor.
func readRestoreContentIdentityPlatform(path string) (RestoreContentIdentity, error) {
	var identity RestoreContentIdentity
	if err := validateRestoreIdentityLeaf(filepath.Base(path)); err != nil {
		return identity, err
	}
	parent, err := unix.Openat2(unix.AT_FDCWD, filepath.Dir(path), &unix.OpenHow{
		Flags: uint64(unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC),
		Resolve: unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		if errors.Is(err, unix.ENOENT) {
			return identity, fmt.Errorf("open SQLite restore identity parent directory: path does not exist (%v)", err)
		}
		return identity, fmt.Errorf("open SQLite restore identity parent without symlink traversal: %w", err)
	}
	defer unix.Close(parent)
	if err := verifyRestoreDirectoryPrivateFD(parent); err != nil {
		return identity, fmt.Errorf("validate private SQLite restore identity parent directory: %w", err)
	}

	fd, err := unix.Openat2(parent, filepath.Base(path), &unix.OpenHow{
		Flags: uint64(unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW),
		Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		if errors.Is(err, unix.ENOENT) {
			return identity, fmt.Errorf("open SQLite restore identity leaf: %w (%v)", os.ErrNotExist, err)
		}
		return identity, fmt.Errorf("open SQLite restore identity leaf without following links: %w", err)
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		_ = unix.Close(fd)
		return identity, fmt.Errorf("wrap SQLite restore identity descriptor")
	}
	defer file.Close()

	var before unix.Stat_t
	if err := unix.Fstat(fd, &before); err != nil {
		return identity, fmt.Errorf("inspect opened SQLite restore identity: %w", err)
	}
	if before.Mode&unix.S_IFMT != unix.S_IFREG || before.Size <= 0 {
		return identity, fmt.Errorf("SQLite restore identity is not a non-empty regular file")
	}
	hash := sha256.New()
	count, err := io.Copy(hash, file)
	if err != nil {
		return identity, fmt.Errorf("hash opened SQLite restore identity: %w", err)
	}
	var after unix.Stat_t
	if err := unix.Fstat(fd, &after); err != nil {
		return identity, fmt.Errorf("reinspect opened SQLite restore identity: %w", err)
	}
	if after.Mode&unix.S_IFMT != unix.S_IFREG || after.Dev != before.Dev || after.Ino != before.Ino ||
		after.Size != before.Size || count != before.Size || after.Mtim != before.Mtim || after.Ctim != before.Ctim {
		return identity, fmt.Errorf("SQLite restore identity changed while being read")
	}
	identity.Size = uint64(count)
	identity.SHA256 = hex.EncodeToString(hash.Sum(nil))
	return identity, nil
}
