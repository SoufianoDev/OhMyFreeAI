package sqlite

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const (
	restoreJournalLeaseChildEnv   = "OMFAI_RESTORE_JOURNAL_LEASE_CHILD"
	restoreJournalLeaseRootEnv    = "OMFAI_RESTORE_JOURNAL_LEASE_ROOT"
	restoreJournalLeaseReadyEnv   = "OMFAI_RESTORE_JOURNAL_LEASE_READY"
	restoreJournalLeaseReleaseEnv = "OMFAI_RESTORE_JOURNAL_LEASE_RELEASE"
)

func TestWriteRestoreJournalSerializesAcrossProcesses(t *testing.T) {
	root := restoreTestDir(t)
	readyPath := filepath.Join(root, "lease-ready")
	releasePath := filepath.Join(root, "lease-release")

	command := exec.Command(os.Args[0], "-test.run=^TestRestoreJournalLeaseProcessHelper$")
	command.Env = append(os.Environ(),
		restoreJournalLeaseChildEnv+"=1",
		restoreJournalLeaseRootEnv+"="+root,
		restoreJournalLeaseReadyEnv+"="+readyPath,
		restoreJournalLeaseReleaseEnv+"="+releasePath,
	)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Start(); err != nil {
		t.Fatalf("start journal-lock child process: %v", err)
	}
	childExited := false
	t.Cleanup(func() {
		if !childExited {
			_ = command.Process.Kill()
			_ = command.Wait()
		}
	})
	waitForRestoreJournalLeaseMarker(t, readyPath, 10*time.Second)

	waitContext, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	lease, err := acquireRestoreJournalLease(waitContext, root)
	cancel()
	if lease != nil {
		_ = lease.Close()
		t.Fatal("a second process acquired the exclusive Restore journal lease while the child held it")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting for the held Restore journal lease returned %v, want context deadline exceeded", err)
	}

	writeDone := make(chan error, 1)
	go func() {
		writeDone <- WriteRestoreJournal(root, validTestRestoreJournal(root))
	}()
	select {
	case err := <-writeDone:
		t.Fatalf("WriteRestoreJournal completed while another process held the journal lease: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	if err := os.WriteFile(releasePath, []byte("release"), 0o600); err != nil {
		t.Fatalf("release child journal lease: %v", err)
	}
	select {
	case err := <-writeDone:
		if err != nil {
			t.Fatalf("WriteRestoreJournal failed after the child released its lease: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("WriteRestoreJournal did not continue after the child released its lease")
	}

	if err := command.Wait(); err != nil {
		childExited = true
		t.Fatalf("journal-lock child process failed: %v\n%s", err, output.String())
	}
	childExited = true

	got, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read journal after cross-process serialization test: %v", err)
	}
	if got.TransactionID != "test-transaction-001" || got.Phase != RestorePhasePrepared {
		t.Fatalf("journal after serialized write = %q/%q, want test-transaction-001/prepared", got.TransactionID, got.Phase)
	}
}

func TestRestoreJournalLeaseProcessHelper(t *testing.T) {
	if os.Getenv(restoreJournalLeaseChildEnv) != "1" {
		t.Skip("cross-process journal lease helper")
	}
	root := os.Getenv(restoreJournalLeaseRootEnv)
	readyPath := os.Getenv(restoreJournalLeaseReadyEnv)
	releasePath := os.Getenv(restoreJournalLeaseReleaseEnv)
	if root == "" || readyPath == "" || releasePath == "" {
		t.Fatal("cross-process journal lease helper arguments are incomplete")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	lease, err := acquireRestoreJournalLease(ctx, root)
	if err != nil {
		t.Fatalf("child process acquire Restore journal lease: %v", err)
	}
	defer lease.Close()

	if err := os.WriteFile(readyPath, []byte("ready"), 0o600); err != nil {
		t.Fatalf("signal acquired Restore journal lease: %v", err)
	}
	waitForRestoreJournalLeaseMarker(t, releasePath, 15*time.Second)
}

func TestRestoreJournalLeaseRejectsSymlinkLockFile(t *testing.T) {
	root := restoreTestDir(t)
	target := filepath.Join(root, "lock-target")
	if err := os.WriteFile(target, []byte("not a lock"), 0o600); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(root, restoreJournalLockFileName)
	if err := os.Symlink(target, lockPath); err != nil {
		if runtime.GOOS == "windows" && errors.Is(err, os.ErrPermission) {
			t.Skip("creating a Windows symbolic link requires an unavailable privilege")
		}
		t.Fatalf("create Restore journal lock symlink: %v", err)
	}
	lease, err := acquireRestoreJournalLease(context.Background(), root)
	if lease != nil {
		_ = lease.Close()
		t.Fatal("Restore journal lease accepted a symbolic-link lock file")
	}
	if err == nil {
		t.Fatal("Restore journal lease accepted a symbolic-link lock file")
	}
}

func TestRestoreJournalLeaseRejectsHardLinkLockFile(t *testing.T) {
	root := restoreTestDir(t)
	target := filepath.Join(root, "lock-target")
	if err := os.WriteFile(target, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(root, restoreJournalLockFileName)
	if err := os.Link(target, lockPath); err != nil {
		t.Fatalf("create hard-linked Restore journal lock: %v", err)
	}

	lease, err := acquireRestoreJournalLease(context.Background(), root)
	if lease != nil {
		_ = lease.Close()
		t.Fatal("Restore journal lease accepted a hard-linked lock file")
	}
	if err == nil {
		t.Fatal("Restore journal lease accepted a hard-linked lock file")
	}
}

func TestRestoreJournalLeaseRejectsNonEmptyLockFile(t *testing.T) {
	root := restoreTestDir(t)
	lockPath := filepath.Join(root, restoreJournalLockFileName)
	if err := os.WriteFile(lockPath, []byte("unexpected"), 0o600); err != nil {
		t.Fatal(err)
	}
	lease, err := acquireRestoreJournalLease(context.Background(), root)
	if lease != nil {
		_ = lease.Close()
		t.Fatal("Restore journal lease accepted a lock file containing data")
	}
	if err == nil {
		t.Fatal("Restore journal lease accepted a lock file containing data")
	}
}

func waitForRestoreJournalLeaseMarker(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(path); err == nil {
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("inspect Restore journal lease marker %s: %v", path, err)
		}
		select {
		case <-deadline.C:
			t.Fatalf("timed out waiting for Restore journal lease marker %s", path)
		case <-ticker.C:
		}
	}
}

func TestRestoreJournalLeaseReadsAndWritesWhileHeld(t *testing.T) {
	root := restoreTestDir(t)
	lease, err := acquireRestoreJournalLease(context.Background(), root)
	if err != nil {
		t.Fatalf("acquire Restore journal lease: %v", err)
	}
	defer lease.Close()

	journal := validTestRestoreJournal(root)
	if err := lease.Write(journal); err != nil {
		t.Fatalf("write through Restore journal lease: %v", err)
	}
	got, err := lease.Read()
	if err != nil {
		t.Fatalf("read through Restore journal lease: %v", err)
	}
	if got.TransactionID != journal.TransactionID || got.Phase != journal.Phase {
		t.Fatalf("leased journal = %q/%q, want %q/%q", got.TransactionID, got.Phase, journal.TransactionID, journal.Phase)
	}

	if err := lease.Close(); err != nil {
		t.Fatalf("close Restore journal lease: %v", err)
	}
	if _, err := lease.Read(); !errors.Is(err, errRestoreJournalLeaseClosed) {
		t.Fatalf("read after lease close error = %v, want closed-lease error", err)
	}
	if err := lease.Write(journal); !errors.Is(err, errRestoreJournalLeaseClosed) {
		t.Fatalf("write after lease close error = %v, want closed-lease error", err)
	}
}
