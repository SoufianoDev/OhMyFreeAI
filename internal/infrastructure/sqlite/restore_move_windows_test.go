//go:build windows

package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestRenameRestoreHandleNoReplaceUsesWriteThroughOpenedSourceHandle(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "candidate.db")
	parked := filepath.Join(directory, "parked.db")
	destination := filepath.Join(directory, "published.db")
	if err := os.WriteFile(source, []byte("original candidate"), 0o600); err != nil {
		t.Fatal(err)
	}

	directoryHandle, err := openRestoreDirectoryNoReparse(directory)
	if err != nil {
		t.Fatalf("open qualified restore directory: %v", err)
	}
	defer windows.CloseHandle(directoryHandle)
	guardHandle, directoryPath, err := openRestoreDirectoryWriteThroughGuard(directoryHandle)
	if err != nil {
		t.Fatalf("open write-through directory guard: %v", err)
	}
	defer func() {
		if guardHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(guardHandle)
		}
	}()

	sourceHandle, err := openRestoreSourceWriteThrough(directoryPath, filepath.Base(source))
	if err != nil {
		t.Fatalf("open write-through source handle: %v", err)
	}
	defer func() {
		if sourceHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(sourceHandle)
		}
	}()

	if err := os.Rename(source, parked); err != nil {
		t.Fatalf("move opened source name aside: %v", err)
	}
	if err := os.WriteFile(source, []byte("replacement at old path"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := renameRestoreHandleNoReplace(sourceHandle, guardHandle, filepath.Base(destination)); err != nil {
		t.Fatalf("rename opened source handle: %v", err)
	}
	if err := windows.CloseHandle(sourceHandle); err != nil {
		t.Fatal(err)
	}
	sourceHandle = windows.InvalidHandle
	if err := windows.CloseHandle(guardHandle); err != nil {
		t.Fatal(err)
	}
	guardHandle = windows.InvalidHandle

	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original candidate" {
		t.Fatalf("published content = %q, want original candidate", got)
	}
	got, err = os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "replacement at old path" {
		t.Fatalf("replacement source content = %q, want replacement at old path", got)
	}
	if _, err := os.Stat(parked); !os.IsNotExist(err) {
		t.Fatalf("parked original path still exists or cannot be inspected: %v", err)
	}
}

func TestRenameRestoreHandleNoReplaceRejectsDestinationCollision(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "candidate.db")
	destination := filepath.Join(directory, "published.db")
	if err := os.WriteFile(source, []byte("candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	directoryHandle, err := openRestoreDirectoryNoReparse(directory)
	if err != nil {
		t.Fatalf("open qualified restore directory: %v", err)
	}
	defer windows.CloseHandle(directoryHandle)
	guardHandle, directoryPath, err := openRestoreDirectoryWriteThroughGuard(directoryHandle)
	if err != nil {
		t.Fatalf("open write-through directory guard: %v", err)
	}
	defer windows.CloseHandle(guardHandle)
	sourceHandle, err := openRestoreSourceWriteThrough(directoryPath, filepath.Base(source))
	if err != nil {
		t.Fatalf("open write-through source handle: %v", err)
	}
	defer windows.CloseHandle(sourceHandle)

	if err := renameRestoreHandleNoReplace(sourceHandle, guardHandle, filepath.Base(destination)); !errors.Is(err, os.ErrExist) {
		t.Fatalf("handle-bound no-replace rename error = %v, want os.ErrExist", err)
	}
	for path, want := range map[string]string{
		source:      "candidate",
		destination: "existing",
	} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s content = %q, want %q", path, got, want)
		}
	}
}

func TestOpenRestoreDirectoryWriteThroughGuardPreventsRename(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "state")
	movedDirectory := filepath.Join(root, "state-moved")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}

	directoryHandle, err := openRestoreDirectoryNoReparse(directory)
	if err != nil {
		t.Fatalf("open qualified restore directory: %v", err)
	}
	defer windows.CloseHandle(directoryHandle)
	guardHandle, guardedPath, err := openRestoreDirectoryWriteThroughGuard(directoryHandle)
	if err != nil {
		t.Fatalf("open write-through directory guard: %v", err)
	}
	if guardedPath == "" {
		t.Fatal("directory guard returned an empty path")
	}
	if err := os.Rename(directory, movedDirectory); err == nil {
		_ = windows.CloseHandle(guardHandle)
		t.Fatal("directory was renamed while its no-delete-share guard was held")
	}
	if err := windows.CloseHandle(guardHandle); err != nil {
		t.Fatalf("close directory guard: %v", err)
	}
	if err := os.Rename(directory, movedDirectory); err != nil {
		t.Fatalf("rename directory after releasing guard: %v", err)
	}
}


func TestOpenRestoreDirectoryRelativeUsesPinnedParentHandle(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "state")
	movedDirectory := filepath.Join(root, "state-moved")
	nested := filepath.Join(directory, "nested")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}

	rootHandle, err := openRestoreDirectoryNoReparse(root)
	if err != nil {
		t.Fatalf("open root directory without reparse traversal: %v", err)
	}

	directoryHandle, err := openRestoreDirectoryRelative(rootHandle, "state")
	if err != nil {
		_ = windows.CloseHandle(rootHandle)
		t.Fatalf("open state directory relative to pinned root: %v", err)
	}
	if err := windows.CloseHandle(rootHandle); err != nil {
		_ = windows.CloseHandle(directoryHandle)
		t.Fatalf("close root directory handle: %v", err)
	}
	defer windows.CloseHandle(directoryHandle)

	expectedNestedHandle, err := openRestoreDirectoryRelative(directoryHandle, "nested")
	if err != nil {
		t.Fatalf("open original nested directory: %v", err)
	}

	var expected windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(expectedNestedHandle, &expected); err != nil {
		_ = windows.CloseHandle(expectedNestedHandle)
		t.Fatalf("inspect original nested directory: %v", err)
	}
	if err := windows.CloseHandle(expectedNestedHandle); err != nil {
		t.Fatalf("close original nested directory handle: %v", err)
	}

	if err := os.Rename(directory, movedDirectory); err != nil {
		t.Fatalf("move opened parent directory aside: %v", err)
	}
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatalf("create replacement directory tree: %v", err)
	}

	pinnedNestedHandle, err := openRestoreDirectoryRelative(directoryHandle, "nested")
	if err != nil {
		t.Fatalf("open nested directory relative to pinned parent: %v", err)
	}
	defer windows.CloseHandle(pinnedNestedHandle)

	var got windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(pinnedNestedHandle, &got); err != nil {
		t.Fatalf("inspect nested directory opened from pinned parent: %v", err)
	}
	if got.VolumeSerialNumber != expected.VolumeSerialNumber ||
		got.FileIndexHigh != expected.FileIndexHigh ||
		got.FileIndexLow != expected.FileIndexLow {
		t.Fatal("relative directory open followed the replacement path instead of the pinned parent handle")
	}
}

func TestOpenRestoreDirectoryNoReparseRejectsSymlinkComponent(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	link := filepath.Join(root, "link")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create a Windows directory symlink in this environment: %v", err)
	}

	if _, err := openRestoreDirectoryNoReparse(filepath.Join(link, "child")); err == nil {
		t.Fatal("directory traversal accepted a reparse-point component")
	}
}

func TestOpenRestoreDirectoryNoReparseRejectsUNCPath(t *testing.T) {
	if _, err := openRestoreDirectoryNoReparse(`\\server\share\state`); err == nil {
		t.Fatal("directory traversal accepted an unqualified UNC path")
	}
}
