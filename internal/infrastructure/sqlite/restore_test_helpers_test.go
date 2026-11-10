package sqlite

import (
	"os"
	"runtime"
	"testing"
)

// restoreTestDir creates a private fixture directory for tests that exercise
// Restore's state-directory and move security checks. Keep the production
// permission checks strict; make the fixture's intended boundary explicit.
func restoreTestDir(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	if runtime.GOOS == "linux" {
		if err := os.Chmod(directory, 0o700); err != nil {
			t.Fatalf("make private SQLite restore test directory: %v", err)
		}
	}
	return directory
}
