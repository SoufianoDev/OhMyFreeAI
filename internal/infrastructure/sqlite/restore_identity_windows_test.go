//go:build windows

package sqlite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestInspectRestoreArtifactAcceptsWindowsShortPathAlias(t *testing.T) {
	stateDirectory := t.TempDir()
	path := filepath.Join(stateDirectory, "candidate.sqlite")
	content := []byte("verified SQLite restore candidate")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	longRequired, err := windows.GetLongPathName(pathPtr, nil, 0)
	if err != nil {
		t.Fatalf("query long Windows path: %v", err)
	}
	longBuffer := make([]uint16, longRequired)
	longLength, err := windows.GetLongPathName(pathPtr, &longBuffer[0], uint32(len(longBuffer)))
	if err != nil {
		t.Fatalf("resolve long Windows path: %v", err)
	}
	longPath := windows.UTF16ToString(longBuffer[:longLength])
	longPathPtr, err := windows.UTF16PtrFromString(longPath)
	if err != nil {
		t.Fatal(err)
	}
	required, err := windows.GetShortPathName(longPathPtr, nil, 0)
	if err != nil {
		t.Fatalf("query short Windows path: %v", err)
	}
	if required == 0 {
		t.Fatal("Windows returned an empty short path")
	}
	buffer := make([]uint16, required)
	length, err := windows.GetShortPathName(longPathPtr, &buffer[0], uint32(len(buffer)))
	if err != nil {
		t.Fatalf("resolve short Windows path: %v", err)
	}
	shortPath := windows.UTF16ToString(buffer[:length])
	if strings.EqualFold(filepath.Clean(shortPath), filepath.Clean(longPath)) {
		t.Skip("the filesystem did not provide a distinct 8.3 short path")
	}

	got, err := inspectRestoreArtifact(stateDirectory, shortPath)
	if err != nil {
		t.Fatalf("inspectRestoreArtifact() rejected a valid 8.3 path alias: %v", err)
	}
	if got.Size != uint64(len(content)) {
		t.Fatalf("artifact size = %d, want %d", got.Size, len(content))
	}
}
