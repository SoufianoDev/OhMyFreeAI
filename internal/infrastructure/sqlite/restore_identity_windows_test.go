//go:build windows

package sqlite

import (
	"os"
	"path/filepath"
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
	required, err := windows.GetShortPathName(pathPtr, nil, 0)
	if err != nil {
		t.Fatalf("query short Windows path: %v", err)
	}
	if required == 0 {
		t.Fatal("Windows returned an empty short path")
	}
	buffer := make([]uint16, required)
	length, err := windows.GetShortPathName(pathPtr, &buffer[0], uint32(len(buffer)))
	if err != nil {
		t.Fatalf("resolve short Windows path: %v", err)
	}
	shortPath := windows.UTF16ToString(buffer[:length])
	if filepath.Clean(shortPath) == filepath.Clean(path) {
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
