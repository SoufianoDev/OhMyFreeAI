package sqlite

import (
	"bytes"
	"context"
	"errors"
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
	if err := archiveStore.Close(); err != nil {
		_ = controller.Close(context.Background())
		t.Fatalf("close source snapshot Store: %v", err)
	}
	if err := ensureNoSQLiteSidecars(archivePath); err != nil {
		_ = controller.Close(context.Background())
		t.Fatalf("source snapshot is not self-contained after close: %v", err)
	}
	t.Cleanup(func() {
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
		BuildCandidate: func(ctx context.Context, archive, stateDirectory, candidatePath string) (RestoreCandidate, error) {
			return BuildRestoreCandidateAtPath(ctx, archive, stateDirectory, candidatePath, migrations.Files, ".")
		},
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

func TestRestoreTransactionExecutorRejectsRollbackSidecarCollisionBeforeJournal(t *testing.T) {
	executor, controller, root, archivePath, previous, _ := newRestoreExecutorFixture(t)
	executor.operations.newID = func() (string, error) { return "sidecar-collision", nil }
	rollbackPath := filepath.Join(root, ".restore-rollback-sidecar-collision.sqlite")
	sidecarPath := rollbackPath + "-wal"
	sentinel := []byte("unexpected rollback sidecar")
	if err := os.WriteFile(sidecarPath, sentinel, 0o600); err != nil {
		t.Fatalf("create rollback sidecar collision: %v", err)
	}

	err := executor.Execute(context.Background(), restoreExecutorRequest(archivePath))
	if err == nil {
		t.Fatal("Restore accepted an unexpected rollback sidecar")
	}
	if got := loadControllerConfiguration(t, controller); !reflect.DeepEqual(got, previous) {
		t.Fatalf("active configuration after sidecar collision = %#v, want previous %#v", got, previous)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("journal should not be created before sidecar qualification; error = %v", err)
	}
	got, err := os.ReadFile(sidecarPath)
	if err != nil {
		t.Fatalf("read preserved rollback sidecar: %v", err)
	}
	if !reflect.DeepEqual(got, sentinel) {
		t.Fatalf("rollback sidecar evidence changed: %q", got)
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
	request := restoreExecutorRequest(archivePath)
	realBuild := request.BuildCandidate
	buildReady := make(chan struct{})
	releaseBuild := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseBuild) }) }
	defer release()
	request.BuildCandidate = func(ctx context.Context, archive, stateDirectory, staging string) (RestoreCandidate, error) {
		candidate, err := realBuild(ctx, archive, stateDirectory, staging)
		close(buildReady)
		<-releaseBuild
		return candidate, err
	}
	done := make(chan error, 1)
	go func() {
		done <- executor.Execute(context.Background(), request)
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


const (
	restoreExecutorCrashChildEnv   = "OHMYFREEAI_RESTORE_EXECUTOR_CRASH_CHILD"
	restoreExecutorCrashRootEnv    = "OHMYFREEAI_RESTORE_EXECUTOR_CRASH_ROOT"
	restoreExecutorCrashArchiveEnv = "OHMYFREEAI_RESTORE_EXECUTOR_CRASH_ARCHIVE"
	restoreExecutorCrashAtEnv      = "OHMYFREEAI_RESTORE_EXECUTOR_CRASH_AT"
	restoreExecutorCrashRollbackEnv = "OHMYFREEAI_RESTORE_EXECUTOR_CRASH_ROLLBACK"
)

func TestRestoreTransactionExecutorProcessDeathHelper(t *testing.T) {
	if os.Getenv(restoreExecutorCrashChildEnv) != "1" {
		return
	}
	root := os.Getenv(restoreExecutorCrashRootEnv)
	archivePath := os.Getenv(restoreExecutorCrashArchiveEnv)
	crashAt := os.Getenv(restoreExecutorCrashAtEnv)
	if root == "" || archivePath == "" || crashAt == "" {
		t.Fatal("Restore executor crash helper environment is incomplete")
	}

	store, err := OpenExisting(context.Background(), filepath.Join(root, "state.db"))
	if err != nil {
		t.Fatalf("open previous database in crash child: %v", err)
	}
	controller := NewStoreController(store)
	executor, err := NewRestoreTransactionExecutor(controller, root)
	if err != nil {
		t.Fatalf("create Restore executor in crash child: %v", err)
	}
	if os.Getenv(restoreExecutorCrashRollbackEnv) == "1" {
		realOpen := executor.operations.openStore
		var calls atomic.Int32
		executor.operations.openStore = func(ctx context.Context, path string) (*Store, error) {
			if calls.Add(1) == 1 {
				return nil, errors.New("injected candidate reopen failure for crash matrix")
			}
			return realOpen(ctx, path)
		}
	}
	executor.operations.afterBoundary = func(boundary string) {
		if boundary == crashAt {
			os.Exit(86)
		}
	}
	if err := executor.Execute(context.Background(), restoreExecutorRequest(archivePath)); err != nil {
		t.Fatalf("Restore executor exited before crash boundary %q: %v", crashAt, err)
	}
	t.Fatalf("Restore executor completed without reaching crash boundary %q", crashAt)
}

func TestRestoreTransactionExecutorProcessDeathRecoveryPlanning(t *testing.T) {
	tests := []struct {
		name             string
		boundary         string
		rollback         bool
		noJournal        bool
		wantPhase        RestorePhase
		wantAction       RestoreRecoveryAction
		wantAmbiguous    bool
	}{
		{name: "before prepared journal", boundary: "before:prepared", noJournal: true},
		{name: "prepared journal", boundary: "journal:prepared", wantPhase: RestorePhasePrepared, wantAction: RestoreRecoveryCleanupStaging},
		{name: "staging intent", boundary: "journal:staging", wantPhase: RestorePhaseStaging, wantAction: RestoreRecoveryCleanupStaging},
		{name: "candidate built before identity journal", boundary: "candidate:built", wantPhase: RestorePhaseStaging, wantAmbiguous: true},
		{name: "candidate identity journal", boundary: "journal:candidate-identity", wantPhase: RestorePhaseStaging, wantAction: RestoreRecoveryCleanupStaging},
		{name: "previous preservation intent", boundary: "journal:preserving-previous", wantPhase: RestorePhasePreservingPrevious, wantAction: RestoreRecoveryResumeIntent},
		{name: "previous moved before completion journal", boundary: "move:previous-to-rollback", wantPhase: RestorePhasePreservingPrevious, wantAction: RestoreRecoveryResumeIntent},
		{name: "previous preserved completion", boundary: "journal:previous-preserved", wantPhase: RestorePhasePreviousPreserved, wantAction: RestoreRecoveryRestorePrevious},
		{name: "candidate publication intent", boundary: "journal:publishing", wantPhase: RestorePhasePublishing, wantAction: RestoreRecoveryResumeIntent},
		{name: "candidate moved before completion journal", boundary: "move:staging-to-target", wantPhase: RestorePhasePublishing, wantAction: RestoreRecoveryResumeIntent},
		{name: "candidate published completion", boundary: "journal:published", wantPhase: RestorePhasePublished, wantAction: RestoreRecoveryRestorePrevious},
		{name: "candidate reopen intent", boundary: "journal:reopening", wantPhase: RestorePhaseReopening, wantAction: RestoreRecoveryRestorePrevious},
		{name: "candidate opened before commit journal", boundary: "store:candidate-opened", wantPhase: RestorePhaseReopening, wantAction: RestoreRecoveryRestorePrevious},
		{name: "committed journal", boundary: "journal:committed", wantPhase: RestorePhaseCommitted, wantAction: RestoreRecoveryVerifyCommitted},
		{name: "rollback intent", boundary: "journal:restoring-previous", rollback: true, wantPhase: RestorePhaseRestoringPrevious, wantAction: RestoreRecoveryResumeIntent},
		{name: "candidate quarantined before rollback move", boundary: "move:candidate-to-quarantine", rollback: true, wantPhase: RestorePhaseRestoringPrevious, wantAction: RestoreRecoveryResumeIntent},
		{name: "previous moved back before completion", boundary: "move:rollback-to-target", rollback: true, wantPhase: RestorePhaseRestoringPrevious, wantAction: RestoreRecoveryResumeIntent},
		{name: "previous reopened before rollback journal", boundary: "store:previous-opened", rollback: true, wantPhase: RestorePhaseRestoringPrevious, wantAction: RestoreRecoveryResumeIntent},
		{name: "rolled back journal", boundary: "journal:rolled-back", rollback: true, wantPhase: RestorePhaseRolledBack, wantAction: RestoreRecoveryVerifyRolledBack},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, controller, root, archivePath, previous, _ := newRestoreExecutorFixture(t)
			if err := controller.Close(context.Background()); err != nil {
				t.Fatalf("close parent Store before child process: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRestoreTransactionExecutorProcessDeathHelper$")
			rollbackValue := "0"
			if test.rollback {
				rollbackValue = "1"
			}
			command.Env = append(os.Environ(),
				restoreExecutorCrashChildEnv+"=1",
				restoreExecutorCrashRootEnv+"="+root,
				restoreExecutorCrashArchiveEnv+"="+archivePath,
				restoreExecutorCrashAtEnv+"="+test.boundary,
				restoreExecutorCrashRollbackEnv+"="+rollbackValue,
			)
			var output bytes.Buffer
			command.Stdout = &output
			command.Stderr = &output
			err = command.Run()
			if ctx.Err() != nil {
				t.Fatalf("crash child timed out at %q: %v\n%s", test.boundary, ctx.Err(), output.String())
			}
			exitError, ok := err.(*exec.ExitError)
			if !ok || exitError.ExitCode() != 86 {
				t.Fatalf("crash child exit = %v, want deliberate exit 86 at %q\n%s", err, test.boundary, output.String())
			}

			if test.noJournal {
				if _, err := ReadRestoreJournal(root); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("journal before prepared boundary = %v, want not-exist", err)
				}
				reopened, err := OpenExisting(context.Background(), filepath.Join(root, "state.db"))
				if err != nil {
					t.Fatalf("open previous database after pre-journal process death: %v", err)
				}
				got, loadErr := reopened.Load(context.Background())
				closeErr := reopened.Close()
				if err := errors.Join(loadErr, closeErr); err != nil {
					t.Fatalf("read previous configuration after pre-journal process death: %v", err)
				}
				if !reflect.DeepEqual(got, previous) {
					t.Fatalf("previous configuration after pre-journal process death = %#v, want %#v", got, previous)
				}
				return
			}

			journal, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatalf("read journal after process death at %q: %v", test.boundary, err)
			}
			if journal.Phase != test.wantPhase {
				t.Fatalf("journal phase after %q = %q, want %q", test.boundary, journal.Phase, test.wantPhase)
			}
			files, err := inspectRestoreRecoveryFiles(root, journal)
			if err != nil {
				t.Fatalf("inspect artifacts after process death at %q: %v", test.boundary, err)
			}
			plan, planErr := PlanRestoreRecovery(journal, files)
			if test.wantAmbiguous {
				if !errors.Is(planErr, ErrRestoreRecoveryAmbiguous) {
					t.Fatalf("recovery plan after %q error = %v, want ambiguous", test.boundary, planErr)
				}
				if plan != (RestoreRecoveryPlan{}) {
					t.Fatalf("ambiguous recovery plan after %q = %#v, want empty", test.boundary, plan)
				}
				if _, err := os.Lstat(journal.StagingPath); err != nil {
					t.Fatalf("unidentified candidate was not preserved after %q: %v", test.boundary, err)
				}
				return
			}
			if planErr != nil {
				t.Fatalf("recovery plan after %q: %v", test.boundary, planErr)
			}
			if plan.Action != test.wantAction {
				t.Fatalf("recovery action after %q = %q, want %q", test.boundary, plan.Action, test.wantAction)
			}
		})
	}
}

func TestRestoreTransactionExecutorJournalWriteFailuresPreserveRecoveryState(t *testing.T) {
	tests := []struct {
		name          string
		boundary      string
		wantPhase     RestorePhase
		wantAction    RestoreRecoveryAction
		wantAmbiguous bool
		noJournal     bool
	}{
		{name: "prepared", boundary: "journal:prepared", noJournal: true},
		{name: "staging", boundary: "journal:staging", wantPhase: RestorePhasePrepared, wantAction: RestoreRecoveryCleanupStaging},
		{name: "candidate identity", boundary: "journal:candidate-identity", wantPhase: RestorePhaseStaging, wantAmbiguous: true},
		{name: "preserving previous", boundary: "journal:preserving-previous", wantPhase: RestorePhaseStaging, wantAction: RestoreRecoveryCleanupStaging},
		{name: "previous preserved", boundary: "journal:previous-preserved", wantPhase: RestorePhasePreservingPrevious, wantAction: RestoreRecoveryResumeIntent},
		{name: "publishing", boundary: "journal:publishing", wantPhase: RestorePhasePreviousPreserved, wantAction: RestoreRecoveryRestorePrevious},
		{name: "published", boundary: "journal:published", wantPhase: RestorePhasePublishing, wantAction: RestoreRecoveryResumeIntent},
		{name: "reopening", boundary: "journal:reopening", wantPhase: RestorePhasePublished, wantAction: RestoreRecoveryRestorePrevious},
		{name: "committed", boundary: "journal:committed", wantPhase: RestorePhaseReopening, wantAction: RestoreRecoveryRestorePrevious},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor, controller, root, archivePath, _, _ := newRestoreExecutorFixture(t)
			injected := errors.New("injected journal write failure at " + test.boundary)
			realWrite := executor.operations.writeJournal
			executor.operations.writeJournal = func(lease *restoreJournalLease, journal RestoreJournal, boundary string) error {
				if boundary == test.boundary {
					return injected
				}
				return realWrite(lease, journal, boundary)
			}

			err := executor.Execute(context.Background(), restoreExecutorRequest(archivePath))
			if !errors.Is(err, injected) {
				t.Fatalf("Execute() error = %v, want injected failure at %q", err, test.boundary)
			}
			if err := controller.WithShared(context.Background(), func(*Store) error { return nil }); !errors.Is(err, ErrStoreUnavailable) {
				t.Fatalf("controller after %q failure = %v, want Store unavailable", test.boundary, err)
			}

			if test.noJournal {
				if _, err := ReadRestoreJournal(root); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("journal after %q failure = %v, want not-exist", test.boundary, err)
				}
				return
			}
			journal, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatalf("read journal after %q failure: %v", test.boundary, err)
			}
			if journal.Phase != test.wantPhase {
				t.Fatalf("journal phase after %q failure = %q, want %q", test.boundary, journal.Phase, test.wantPhase)
			}
			files, err := inspectRestoreRecoveryFiles(root, journal)
			if err != nil {
				t.Fatalf("inspect artifacts after %q failure: %v", test.boundary, err)
			}
			plan, planErr := PlanRestoreRecovery(journal, files)
			if test.wantAmbiguous {
				if !errors.Is(planErr, ErrRestoreRecoveryAmbiguous) || plan != (RestoreRecoveryPlan{}) {
					t.Fatalf("recovery plan after %q failure = %#v, %v; want empty ambiguous plan", test.boundary, plan, planErr)
				}
				if _, err := os.Lstat(journal.StagingPath); err != nil {
					t.Fatalf("unidentified candidate was not preserved after %q failure: %v", test.boundary, err)
				}
				return
			}
			if planErr != nil {
				t.Fatalf("recovery plan after %q failure: %v", test.boundary, planErr)
			}
			if plan.Action != test.wantAction {
				t.Fatalf("recovery action after %q failure = %q, want %q", test.boundary, plan.Action, test.wantAction)
			}
		})
	}
}
