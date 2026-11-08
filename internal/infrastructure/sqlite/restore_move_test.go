//go:build linux || windows

package sqlite

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
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

func TestMoveRestoreFileNoReplaceConcurrentCollision(t *testing.T) {
	directory := t.TempDir()
	destination := filepath.Join(directory, "target.db")
	sources := []string{
		filepath.Join(directory, "candidate-a.db"),
		filepath.Join(directory, "candidate-b.db"),
	}
	for i, source := range sources {
		if err := os.WriteFile(source, []byte{byte('a' + i)}, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	start := make(chan struct{})
	results := make(chan error, len(sources))
	var workers sync.WaitGroup
	for _, source := range sources {
		workers.Add(1)
		go func(source string) {
			defer workers.Done()
			<-start
			results <- moveRestoreFileNoReplace(source, destination)
		}(source)
	}
	close(start)
	workers.Wait()
	close(results)

	successes := 0
	collisions := 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, os.ErrExist):
			collisions++
		default:
			t.Fatalf("unexpected move error: %v", err)
		}
	}
	if successes != 1 || collisions != 1 {
		t.Fatalf("successes = %d, collisions = %d; want 1 each", successes, collisions)
	}

	remainingSources := 0
	for _, source := range sources {
		if _, err := os.Lstat(source); err == nil {
			remainingSources++
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
	}
	if remainingSources != 1 {
		t.Fatalf("remaining source files = %d, want 1", remainingSources)
	}
	if _, err := os.Stat(destination); err != nil {
		t.Fatalf("destination missing after successful move: %v", err)
	}
}

