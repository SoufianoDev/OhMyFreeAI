//go:build linux

package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestReadRestoreContentIdentityRejectsFIFOWithoutBlocking(t *testing.T) {
	path := filepath.Join(restoreTestDir(t), "candidate.fifo")
	if err := unix.Mkfifo(path, 0o600); err != nil {
		t.Fatalf("create FIFO fixture: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := readRestoreContentIdentityPlatform(path)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("accepted a FIFO as a restore database artifact")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("opening a raced-in FIFO blocked instead of rejecting the non-regular file")
	}
}
