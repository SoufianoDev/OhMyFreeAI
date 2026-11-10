//go:build windows

package sqlite

import (
	"errors"
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
		sourceHandle, err := openRestoreSourceWriteThrough(directoryHandle, filepath.Base(source))
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

	if err := renameRestoreHandleNoReplace(sourceHandle, directoryHandle, filepath.Base(destination)); err != nil {
		t.Fatalf("rename opened source handle: %v", err)
	}
	if err := windows.CloseHandle(sourceHandle); err != nil {
		t.Fatal(err)
	}
	sourceHandle = windows.InvalidHandle

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
		sourceHandle, err := openRestoreSourceWriteThrough(directoryHandle, filepath.Base(source))
	if err != nil {
		t.Fatalf("open write-through source handle: %v", err)
	}
	defer func() {
		if sourceHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(sourceHandle)
		}
	}()

	if err := renameRestoreHandleNoReplace(sourceHandle, directoryHandle, filepath.Base(destination)); !errors.Is(err, os.ErrExist) {
		t.Fatalf("handle-bound no-replace rename error = %v, want os.ErrExist", err)
	}
	if err := windows.CloseHandle(sourceHandle); err != nil {
		t.Fatalf("close source handle after collision: %v", err)
	}
	sourceHandle = windows.InvalidHandle
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

func TestRenameRestoreHandleNoReplaceStaysBoundToPinnedDirectoryAfterAncestorRename(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	directory := filepath.Join(parent, "state")
	movedParent := filepath.Join(root, "parent-moved")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(directory, "candidate.db")
	if err := os.WriteFile(source, []byte("original candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	directoryHandle, err := openRestoreDirectoryNoReparse(directory)
	if err != nil {
		t.Fatalf("open qualified restore directory: %v", err)
	}
	defer windows.CloseHandle(directoryHandle)
	sourceHandle, err := openRestoreSourceWriteThrough(directoryHandle, filepath.Base(source))
	if err != nil {
		t.Fatalf("open write-through source relative to pinned directory: %v", err)
	}
	defer func() {
		if sourceHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(sourceHandle)
		}
	}()
	if err := os.Rename(parent, movedParent); err != nil {
		t.Fatalf("rename ancestor of pinned directory: %v", err)
	}
	replacementDirectory := filepath.Join(parent, "state")
	if err := os.MkdirAll(replacementDirectory, 0o700); err != nil {
		t.Fatalf("create replacement ancestor tree: %v", err)
	}
	decoySource := filepath.Join(replacementDirectory, "candidate.db")
	if err := os.WriteFile(decoySource, []byte("decoy candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := renameRestoreHandleNoReplace(sourceHandle, directoryHandle, "published.db"); err != nil {
		t.Fatalf("rename opened source relative to pinned directory handle: %v", err)
	}
	if err := windows.CloseHandle(sourceHandle); err != nil {
		t.Fatal(err)
	}
	sourceHandle = windows.InvalidHandle
	published := filepath.Join(movedParent, "state", "published.db")
	got, err := os.ReadFile(published)
	if err != nil {
		t.Fatalf("read destination in original pinned directory: %v", err)
	}
	if string(got) != "original candidate" {
		t.Fatalf("published content = %q, want original candidate", got)
	}
	got, err = os.ReadFile(decoySource)
	if err != nil {
		t.Fatalf("read decoy source: %v", err)
	}
	if string(got) != "decoy candidate" {
		t.Fatalf("decoy source content = %q, want decoy candidate", got)
	}
	if _, err := os.Stat(filepath.Join(replacementDirectory, "published.db")); !os.IsNotExist(err) {
		t.Fatalf("rename escaped pinned directory into replacement tree: %v", err)
	}
}

func TestValidateRestoreSourceLocationRejectsAncestorReplacementDecoy(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	directory := filepath.Join(parent, "state")
	movedParent := filepath.Join(root, "parent-moved")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	directoryHandle, err := openRestoreDirectoryNoReparse(directory)
	if err != nil {
		t.Fatalf("open qualified restore directory: %v", err)
	}
	defer windows.CloseHandle(directoryHandle)
	if err := os.Rename(parent, movedParent); err != nil {
		t.Fatalf("rename pinned directory ancestor: %v", err)
	}
	replacementDirectory := filepath.Join(parent, "state")
	if err := os.MkdirAll(replacementDirectory, 0o700); err != nil {
		t.Fatalf("create replacement directory tree: %v", err)
	}
	decoy := filepath.Join(replacementDirectory, "candidate.db")
	if err := os.WriteFile(decoy, []byte("decoy"), 0o600); err != nil {
		t.Fatal(err)
	}
	decoyPointer, err := windows.UTF16PtrFromString(decoy)
	if err != nil {
		t.Fatal(err)
	}
	decoyHandle, err := windows.CreateFile(
		decoyPointer,
		windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		t.Fatalf("open replacement-path decoy: %v", err)
	}
	defer windows.CloseHandle(decoyHandle)
	if err := validateRestoreSourceLocation(decoyHandle, directoryHandle, "candidate.db"); err == nil {
		t.Fatal("source location validation accepted a decoy from the replacement ancestor tree")
	}
}

func TestRenameRestoreHandleReplaceReplacesOnlyPinnedDirectoryLeaf(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "journal.tmp")
	destination := filepath.Join(directory, "journal.json")
	if err := os.WriteFile(source, []byte("new journal"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("old journal"), 0o600); err != nil {
		t.Fatal(err)
	}
	directoryHandle, err := openRestoreDirectoryNoReparse(directory)
	if err != nil {
		t.Fatalf("open qualified journal directory: %v", err)
	}
	defer windows.CloseHandle(directoryHandle)
	sourceHandle, err := openRestoreSourceWriteThrough(directoryHandle, filepath.Base(source))
	if err != nil {
		t.Fatalf("open write-through journal source: %v", err)
	}
	defer func() {
		if sourceHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(sourceHandle)
		}
	}()
	if err := renameRestoreHandleReplace(sourceHandle, directoryHandle, filepath.Base(destination)); err != nil {
		t.Fatalf("replace journal relative to pinned directory: %v", err)
	}
	if err := windows.CloseHandle(sourceHandle); err != nil {
		t.Fatal(err)
	}
	sourceHandle = windows.InvalidHandle
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new journal" {
		t.Fatalf("published journal = %q, want new journal", got)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("temporary journal still exists or cannot be inspected: %v", err)
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
