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
	defer windows.CloseHandle(sourceHandle)

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
	defer windows.CloseHandle(directoryHandle)

	if err := os.Rename(source, parked); err != nil {
		t.Fatalf("move opened source name aside: %v", err)
	}
	if err := os.WriteFile(source, []byte("replacement at old path"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := renameRestoreHandleNoReplace(sourceHandle, directoryHandle, filepath.Base(destination)); err != nil {
		t.Fatalf("rename opened source handle: %v", err)
	}

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
	defer windows.CloseHandle(sourceHandle)

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
	defer windows.CloseHandle(directoryHandle)

	if err := renameRestoreHandleNoReplace(sourceHandle, directoryHandle, filepath.Base(destination)); err == nil {
		t.Fatal("handle-bound rename replaced an existing destination")
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
