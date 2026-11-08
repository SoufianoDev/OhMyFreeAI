package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreDatabaseFailsClosedWithoutTouchingFiles(t *testing.T) {
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "candidate.sqlite")
	targetPath := filepath.Join(directory, "state.db")
	source := []byte("candidate bytes")
	target := []byte("existing application database")

	if err := os.WriteFile(sourcePath, source, 0o600); err != nil {
		t.Fatalf("write restore source: %v", err)
	}
	if err := os.WriteFile(targetPath, target, 0o600); err != nil {
		t.Fatalf("write restore target: %v", err)
	}

	if _, err := RestoreDatabase(context.Background(), sourcePath, targetPath, nil, ""); !errors.Is(err, ErrRestoreExecutorUnavailable) {
		t.Fatalf("RestoreDatabase() error = %v, want ErrRestoreExecutorUnavailable", err)
	}

	gotSource, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read source after disabled restore: %v", err)
	}
	if string(gotSource) != string(source) {
		t.Fatalf("source changed after disabled restore: got %q, want %q", gotSource, source)
	}
	gotTarget, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read target after disabled restore: %v", err)
	}
	if string(gotTarget) != string(target) {
		t.Fatalf("target changed after disabled restore: got %q, want %q", gotTarget, target)
	}
}
