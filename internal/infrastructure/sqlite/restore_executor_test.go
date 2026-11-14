package sqlite

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	domainconfig "github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
	"github.com/SoufianoDev/OhMyFreeAI/migrations"
)

func newRestoreExecutorFixture(t *testing.T) (*RestoreTransactionExecutor, *StoreController, string, string, domainconfig.Config, domainconfig.Config) {
	t.Helper()
	ctx := context.Background()
	root := restoreTestDir(t)
	targetStore, err := Open(ctx, filepath.Join(root, "state.db"))
	if err != nil {
		t.Fatalf("open previous SQLite database: %v", err)
	}
	if err := targetStore.Migrate(ctx, migrations.Files, "."); err != nil {
		_ = targetStore.Close()
		t.Fatalf("migrate previous SQLite database: %v", err)
	}
	previous := domainconfig.Defaults()
	previous.Gateway.Port = 18180
	previous.UI.Language = "ar"
	if err := targetStore.Save(ctx, previous); err != nil {
		_ = targetStore.Close()
		t.Fatalf("save previous SQLite configuration: %v", err)
	}
	controller := NewStoreController(targetStore)
	archivePath, archiveStore, candidate := createRestoreCandidateSnapshot(t)
	t.Cleanup(func() {
		_ = archiveStore.Close()
		_ = controller.Close(context.Background())
	})
	executor, err := NewRestoreTransactionExecutor(controller, root)
	if err != nil {
		t.Fatalf("create Restore transaction executor: %v", err)
	}
	return executor, controller, root, archivePath, previous, candidate
}

func restoreExecutorRequest(archivePath string) RestoreTransactionRequest {
	return RestoreTransactionRequest{
		ArchivePath: archivePath,
		MigrationFS: migrations.Files,
		MigrationDir: ".",
	}
}

func loadControllerConfiguration(t *testing.T, controller *StoreController) domainconfig.Config {
	t.Helper()
	var configuration domainconfig.Config
	if err := controller.WithShared(context.Background(), func(store *Store) error {
		var err error
		configuration, err = store.Load(context.Background())
		return err
	}); err != nil {
		t.Fatalf("load configuration through StoreController: %v", err)
	}
	return configuration
}

func TestRestoreTransactionExecutorPublishesVerifiedCandidateAndRetainsRollback(t *testing.T) {
	executor, controller, root, archivePath, _, expected := newRestoreExecutorFixture(t)
	if err := executor.Execute(context.Background(), restoreExecutorRequest(archivePath)); err != nil {
		t.Fatalf("execute Restore transaction: %v", err)
	}
	if got := loadControllerConfiguration(t, controller); !reflect.DeepEqual(got, expected) {
		t.Fatalf("active configuration = %#v, want candidate configuration %#v", got, expected)
	}

	journal, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read committed Restore journal: %v", err)
	}
	if journal.Phase != RestorePhaseCommitted {
		t.Fatalf("Restore journal phase = %q, want committed", journal.Phase)
	}
	rollbackIdentity, err := readRestoreContentIdentityPlatform(journal.RollbackPath)
	if err != nil {
		t.Fatalf("read retained rollback identity: %v", err)
	}
	if !rollbackIdentity.matches(journal.PreviousSHA256, journal.PreviousSize) {
		t.Fatalf("retained rollback identity = %#v, want %s/%d", rollbackIdentity, journal.PreviousSHA256, journal.PreviousSize)
	}
	for _, path := range []string{journal.StagingPath, journal.CandidateQuarantinePath} {
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("unexpected transaction artifact %s remains: %v", path, err)
		}
	}
}

func TestRestoreTransactionExecutorQuarantinesCandidateWhenReopenFails(t *testing.T) {
	executor, controller, root, archivePath, previous, _ := newRestoreExecutorFixture(t)
	injected := errors.New("injected candidate reopen failure")
	realOpen := executor.operations.openStore
	var calls atomic.Int32
	executor.operations.openStore = func(ctx context.Context, path string) (*Store, error) {
		if calls.Add(1) == 1 {
			return nil, injected
		}
		return realOpen(ctx, path)
	}

	err := executor.Execute(context.Background(), restoreExecutorRequest(archivePath))
	if !errors.Is(err, injected) {
		t.Fatalf("Execute() error = %v, want injected reopen failure", err)
	}
	if got := loadControllerConfiguration(t, controller); !reflect.DeepEqual(got, previous) {
		t.Fatalf("active configuration after rollback = %#v, want previous %#v", got, previous)
	}

	journal, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read rolled-back Restore journal: %v", err)
	}
	if journal.Phase != RestorePhaseRolledBack {
		t.Fatalf("Restore journal phase = %q, want rolled-back", journal.Phase)
	}
	quarantineIdentity, err := readRestoreContentIdentityPlatform(journal.CandidateQuarantinePath)
	if err != nil {
		t.Fatalf("read quarantined candidate identity: %v", err)
	}
	if !quarantineIdentity.matches(journal.CandidateSHA256, journal.CandidateSize) {
		t.Fatalf("quarantined candidate identity = %#v, want %s/%d", quarantineIdentity, journal.CandidateSHA256, journal.CandidateSize)
	}
	if _, err := os.Lstat(journal.RollbackPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rollback path remains after successful rollback: %v", err)
	}
}

func TestRestoreTransactionExecutorPreservesUnexpectedRollbackCollision(t *testing.T) {
	executor, controller, root, archivePath, previous, _ := newRestoreExecutorFixture(t)
	target := filepath.Join(root, "state.db")
	sentinel := []byte("unexpected rollback evidence")
	realMove := executor.operations.move
	executor.operations.move = func(source, destination string, expected RestoreContentIdentity) error {
		if samePath(source, target) {
			if err := os.WriteFile(destination, sentinel, 0o600); err != nil {
				t.Fatalf("create concurrent rollback collision: %v", err)
			}
		}
		return realMove(source, destination, expected)
	}

	err := executor.Execute(context.Background(), restoreExecutorRequest(archivePath))
	if err == nil {
		t.Fatal("Restore unexpectedly replaced a colliding rollback artifact")
	}
	if got := loadControllerConfiguration(t, controller); !reflect.DeepEqual(got, previous) {
		t.Fatalf("active configuration after collision = %#v, want previous %#v", got, previous)
	}
	journal, readErr := ReadRestoreJournal(root)
	if readErr != nil {
		t.Fatalf("read journal after rollback collision: %v", readErr)
	}
	if journal.Phase != RestorePhasePreservingPrevious {
		t.Fatalf("journal phase = %q, want preserving-previous", journal.Phase)
	}
	got, readErr := os.ReadFile(journal.RollbackPath)
	if readErr != nil {
		t.Fatalf("read preserved collision evidence: %v", readErr)
	}
	if !reflect.DeepEqual(got, sentinel) {
		t.Fatalf("collision evidence changed: %q", got)
	}
}

func TestRestoreTransactionExecutorHoldsJournalAndStoreLeasesDuringBuild(t *testing.T) {
	executor, controller, root, archivePath, _, _ := newRestoreExecutorFixture(t)
	realBuild := executor.operations.buildCandidate
	buildReady := make(chan struct{})
	releaseBuild := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseBuild) }) }
	defer release()
	executor.operations.buildCandidate = func(ctx context.Context, source, state, staging string, migrationsFS fs.FS, migrationDir string) (RestoreCandidate, error) {
		candidate, err := realBuild(ctx, source, state, staging, migrationsFS, migrationDir)
		close(buildReady)
		<-releaseBuild
		return candidate, err
	}
	done := make(chan error, 1)
	go func() {
		done <- executor.Execute(context.Background(), restoreExecutorRequest(archivePath))
	}()
	select {
	case <-buildReady:
	case <-time.After(15 * time.Second):
		t.Fatal("candidate build did not reach the lease assertion point")
	}

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
		t.Fatalf("start second journal process: %v", err)
	}
	childExited := false
	t.Cleanup(func() {
		if !childExited {
			_ = command.Process.Kill()
			_ = command.Wait()
		}
	})
	select {
	case <-time.After(200 * time.Millisecond):
		if _, err := os.Lstat(readyPath); err == nil {
			t.Fatal("second process acquired the journal lease while Restore was active")
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("inspect second-process lease marker: %v", err)
		}
	}

	storeContext, cancelStore := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelStore()
	err := controller.WithShared(storeContext, func(*Store) error { return nil })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("shared Store operation error = %v, want context deadline exceeded", err)
	}

	release()
	if err := <-done; err != nil {
		t.Fatalf("Restore transaction after releasing build gate: %v", err)
	}
	waitForRestoreJournalLeaseMarker(t, readyPath, 10*time.Second)
	if err := os.WriteFile(releasePath, []byte("release"), 0o600); err != nil {
		t.Fatalf("release second journal process: %v", err)
	}
	if err := command.Wait(); err != nil {
		childExited = true
		t.Fatalf("second journal process failed: %v\n%s", err, output.String())
	}
	childExited = true
}
