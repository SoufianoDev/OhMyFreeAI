//go:build windows

package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestRenameRestoreHandleNoReplaceUsesOpenedSourceHandle(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "candidate.db")
	parked := filepath.Join(directory, "parked.db")
	destination := filepath.Join(directory, "published.db")
	if err := os.WriteFile(source, []byte("original candidate"), 0o600); err != nil {
		t.Fatal(err)
	}

	sourceHandle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(source),
		windows.DELETE|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if sourceHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(sourceHandle)
		}
	}()

	directoryHandle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(directory),
		windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if directoryHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(directoryHandle)
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
	if err := windows.CloseHandle(directoryHandle); err != nil {
		t.Fatal(err)
	}
	directoryHandle = windows.InvalidHandle

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

	sourceHandle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(source),
		windows.DELETE|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if sourceHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(sourceHandle)
		}
	}()

	directoryHandle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(directory),
		windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if directoryHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(directoryHandle)
		}
	}()

	if err := renameRestoreHandleNoReplace(sourceHandle, directoryHandle, filepath.Base(destination)); err == nil {
		t.Fatal("handle-bound rename replaced an existing destination")
	}
	if err := windows.CloseHandle(sourceHandle); err != nil {
		t.Fatal(err)
	}
	sourceHandle = windows.InvalidHandle
	if err := windows.CloseHandle(directoryHandle); err != nil {
		t.Fatal(err)
	}
	directoryHandle = windows.InvalidHandle
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


func TestOpenRestoreSourceRelativeUsesPinnedDirectoryHandle(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "state")
	movedDirectory := filepath.Join(root, "state-moved")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(directory, "candidate.db")
	if err := os.WriteFile(source, []byte("original candidate"), 0o600); err != nil {
		t.Fatal(err)
	}

	directoryHandle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(directory),
		windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if directoryHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(directoryHandle)
		}
	}()

	if err := os.Rename(directory, movedDirectory); err != nil {
		t.Fatalf("move opened directory path aside: %v", err)
	}
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "candidate.db"), []byte("replacement at old directory path"), 0o600); err != nil {
		t.Fatal(err)
	}

	sourceHandle, err := openRestoreSourceRelative(directoryHandle, "candidate.db")
	if err != nil {
		t.Fatalf("open source relative to pinned directory handle: %v", err)
	}
	defer func() {
		if sourceHandle != windows.InvalidHandle {
			_ = windows.CloseHandle(sourceHandle)
		}
	}()

	if err := renameRestoreHandleNoReplace(sourceHandle, directoryHandle, "published.db"); err != nil {
		t.Fatalf("rename opened source handle: %v", err)
	}
	if err := windows.CloseHandle(sourceHandle); err != nil {
		t.Fatal(err)
	}
	sourceHandle = windows.InvalidHandle
	if err := windows.CloseHandle(directoryHandle); err != nil {
		t.Fatal(err)
	}
	directoryHandle = windows.InvalidHandle

	got, err := os.ReadFile(filepath.Join(movedDirectory, "published.db"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original candidate" {
		t.Fatalf("published content = %q, want original candidate", got)
	}
	got, err = os.ReadFile(filepath.Join(directory, "candidate.db"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "replacement at old directory path" {
		t.Fatalf("replacement content = %q, want replacement at old directory path", got)
	}
}
