//go:build linux

package sqlite

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestMoveRestoreFileNoReplaceRejectsSymlinkedContainingDirectory(t *testing.T) {
	root := t.TempDir()
	realDirectory := filepath.Join(root, "real")
	if err := os.Mkdir(realDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(realDirectory, "source.db")
	if err := os.WriteFile(source, []byte("candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(realDirectory, alias); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skip("creating symlinks is not permitted")
		}
		t.Fatal(err)
	}

	err := moveRestoreFileNoReplace(filepath.Join(alias, "source.db"), filepath.Join(alias, "target.db"))
	if err == nil {
		t.Fatal("moveRestoreFileNoReplace() accepted a symlinked containing directory")
	}
	if _, err := os.Lstat(source); err != nil {
		t.Fatalf("real source was changed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(realDirectory, "target.db")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination unexpectedly exists: %v", err)
	}
}

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

func TestMoveRestoreFileNoReplaceRejectsNonPrivateContainingDirectory(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "source.db")
	destination := filepath.Join(directory, "target.db")
	if err := os.WriteFile(source, []byte("candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(directory, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := moveRestoreFileNoReplace(source, destination); err == nil {
		t.Fatal("accepted a containing directory with group permissions")
	}
	if _, err := os.Lstat(source); err != nil {
		t.Fatalf("source changed despite rejected directory: %v", err)
	}
	if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination unexpectedly exists: %v", err)
	}
}



func TestMoveRestoreFileNoReplaceVerifiedOnTmpfs(t *testing.T) {
	directory, err := os.MkdirTemp("/dev/shm", ".ohmyfreeai-restore-tmpfs-")
	if err != nil {
		if os.Getenv("CI") == "true" {
			t.Fatalf("GitHub CI must provide the tmpfs fixture: %v", err)
		}
		t.Skipf("private tmpfs fixture is unavailable: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })
	if err := os.Chmod(directory, 0o700); err != nil {
		t.Fatalf("restrict tmpfs fixture directory: %v", err)
	}

	var filesystem unix.Statfs_t
	if err := unix.Statfs(directory, &filesystem); err != nil {
		t.Fatalf("inspect tmpfs fixture filesystem: %v", err)
	}
	if uint64(filesystem.Type) != uint64(unix.TMPFS_MAGIC) {
		if os.Getenv("CI") == "true" {
			t.Fatalf("GitHub CI /dev/shm must be tmpfs (filesystem magic 0x%x)", uint64(filesystem.Type))
		}
		t.Skipf("/dev/shm is not tmpfs (filesystem magic 0x%x)", uint64(filesystem.Type))
	}

	source := filepath.Join(directory, "candidate.staging")
	destination := filepath.Join(directory, "candidate.db")
	content := []byte("verified candidate on a distinct tmpfs provider")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := moveRestoreFileNoReplaceVerified(source, destination, restoreIdentityForTest(content)); err != nil {
		t.Fatalf("verified no-replace move on tmpfs: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read verified tmpfs destination: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("tmpfs destination = %q, want %q", got, content)
	}
	if _, err := os.Lstat(source); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("tmpfs source remains after move: %v", err)
	}

	collisionSource := filepath.Join(directory, "collision-source.staging")
	collisionDestination := filepath.Join(directory, "collision-destination.db")
	sourceContent := []byte("source must remain after tmpfs collision")
	destinationContent := []byte("existing destination must not be replaced")
	if err := os.WriteFile(collisionSource, sourceContent, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(collisionDestination, destinationContent, 0o600); err != nil {
		t.Fatal(err)
	}
	err = moveRestoreFileNoReplaceVerified(collisionSource, collisionDestination, restoreIdentityForTest(sourceContent))
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("tmpfs collision error = %v, want os.ErrExist", err)
	}
	for path, want := range map[string][]byte{
		collisionSource:      sourceContent,
		collisionDestination: destinationContent,
	} {
		got, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read preserved tmpfs artifact %s: %v", path, readErr)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("tmpfs artifact %s = %q, want %q", path, got, want)
		}
	}
}

func TestMoveRestoreFileNoReplaceFailsClosedWhenRenameat2IsUnsupported(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "filesystem does not support RENAME_NOREPLACE", err: unix.EOPNOTSUPP},
		{name: "filesystem rejects RENAME_NOREPLACE flag", err: unix.EINVAL},
		{name: "kernel does not provide renameat2", err: unix.ENOSYS},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := restoreTestDir(t)
			source := filepath.Join(directory, "candidate.staging")
			destination := filepath.Join(directory, "candidate.db")
			content := []byte("candidate must remain untouched when no-replace is unsupported")
			if err := os.WriteFile(source, content, 0o600); err != nil {
				t.Fatal(err)
			}

			renameCalls := 0
			afterNamespaceCalls := 0
			err := moveRestoreFileNoReplacePlatformWithRenameat2(
				source,
				destination,
				func(oldDirFD int, oldPath string, newDirFD int, newPath string, flags uint) error {
					renameCalls++
					if flags != unix.RENAME_NOREPLACE {
						t.Fatalf("rename flags = %d, want RENAME_NOREPLACE", flags)
					}
					return test.err
				},
				func() error {
					afterNamespaceCalls++
					return nil
				},
			)
			if !errors.Is(err, test.err) {
				t.Fatalf("move error = %v, want %v", err, test.err)
			}
			if renameCalls != 1 {
				t.Fatalf("renameat2 calls = %d, want exactly one", renameCalls)
			}
			if afterNamespaceCalls != 0 {
				t.Fatalf("post-namespace hook calls = %d, want zero after a rejected rename", afterNamespaceCalls)
			}

			got, err := os.ReadFile(source)
			if err != nil {
				t.Fatalf("read preserved source: %v", err)
			}
			if !bytes.Equal(got, content) {
				t.Fatalf("source changed: got %q, want %q", got, content)
			}
			if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("destination exists after unsupported no-replace operation: %v", err)
			}
		})
	}
}
