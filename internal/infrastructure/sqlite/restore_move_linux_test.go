//go:build linux

package sqlite

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestSyncRestoreDirectoryHandle(t *testing.T) {
	directoryFD, err := unix.Openat2(unix.AT_FDCWD, t.TempDir(), &unix.OpenHow{
		Flags:   uint64(unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC),
		Resolve: unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		t.Fatalf("open pinned test directory: %v", err)
	}
	defer unix.Close(directoryFD)

	if err := syncRestoreDirectoryHandle(directoryFD); err != nil {
		t.Fatalf("syncRestoreDirectoryHandle() error = %v", err)
	}
}

func TestSyncRestoreDirectoryHandleRejectsInvalidDescriptor(t *testing.T) {
	if err := syncRestoreDirectoryHandle(-1); err == nil {
		t.Fatal("syncRestoreDirectoryHandle() accepted an invalid directory descriptor")
	}
}
