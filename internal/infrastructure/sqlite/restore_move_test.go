package sqlite

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMoveRestoreFileNoReplace(t *testing.T) {
	t.Run("moves regular file", func(t *testing.T) {
		directory := t.TempDir()
		source := filepath.Join(directory, "source.db")
		destination := filepath.Join(directory, "target.db")
		if err := os.WriteFile(source, []byte("candidate"), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := moveRestoreFileNoReplace(source, destination); err != nil {
			t.Fatalf("moveRestoreFileNoReplace() error = %v", err)
		}
		if _, err := os.Lstat(source); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("source still exists or could not be inspected: %v", err)
		}
		got, err := os.ReadFile(destination)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "candidate" {
			t.Fatalf("destination content = %q, want candidate", got)
		}
	})

	t.Run("destination collision preserves both files", func(t *testing.T) {
		directory := t.TempDir()
		source := filepath.Join(directory, "source.db")
		destination := filepath.Join(directory, "target.db")
		if err := os.WriteFile(source, []byte("candidate"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, []byte("unexpected"), 0o600); err != nil {
			t.Fatal(err)
		}

		err := moveRestoreFileNoReplace(source, destination)
		if !errors.Is(err, os.ErrExist) {
			t.Fatalf("moveRestoreFileNoReplace() error = %v, want os.ErrExist", err)
		}
		for path, want := range map[string]string{
			source:      "candidate",
			destination: "unexpected",
		} {
			got, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if string(got) != want {
				t.Fatalf("%s content = %q, want %q", path, got, want)
			}
		}
	})

	t.Run("rejects symbolic-link source", func(t *testing.T) {
		directory := t.TempDir()
		realFile := filepath.Join(directory, "real.db")
		source := filepath.Join(directory, "source.db")
		destination := filepath.Join(directory, "target.db")
		if err := os.WriteFile(realFile, []byte("must stay"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(realFile, source); err != nil {
			if errors.Is(err, os.ErrPermission) {
				t.Skip("creating symlinks is not permitted")
			}
			t.Fatal(err)
		}

		if err := moveRestoreFileNoReplace(source, destination); err == nil {
			t.Fatal("moveRestoreFileNoReplace() accepted a symbolic-link source")
		}
		if _, err := os.Lstat(source); err != nil {
			t.Fatalf("symbolic-link source was changed: %v", err)
		}
		if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("destination unexpectedly exists: %v", err)
		}
	})

	t.Run("rejects non-regular source", func(t *testing.T) {
		directory := t.TempDir()
		source := filepath.Join(directory, "source")
		destination := filepath.Join(directory, "target")
		if err := os.Mkdir(source, 0o700); err != nil {
			t.Fatal(err)
		}

		if err := moveRestoreFileNoReplace(source, destination); err == nil {
			t.Fatal("moveRestoreFileNoReplace() accepted a directory")
		}
		if _, err := os.Stat(source); err != nil {
			t.Fatalf("source directory was changed: %v", err)
		}
	})

	t.Run("requires one containing directory", func(t *testing.T) {
		directory := t.TempDir()
		source := filepath.Join(directory, "source.db")
		destinationDirectory := filepath.Join(directory, "other")
		if err := os.Mkdir(destinationDirectory, 0o700); err != nil {
			t.Fatal(err)
		}
		destination := filepath.Join(destinationDirectory, "target.db")
		if err := os.WriteFile(source, []byte("candidate"), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := moveRestoreFileNoReplace(source, destination); err == nil {
			t.Fatal("moveRestoreFileNoReplace() accepted a cross-directory move")
		}
		if _, err := os.Stat(source); err != nil {
			t.Fatalf("source was changed: %v", err)
		}
	})
}
