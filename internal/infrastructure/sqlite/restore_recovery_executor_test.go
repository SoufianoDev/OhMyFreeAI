package sqlite

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	domainconfig "github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
)

func crashRestoreTransactionForRecovery(t *testing.T, boundary string, rollback bool) (string, domainconfig.Config, domainconfig.Config) {
	t.Helper()
	_, controller, root, archivePath, previous, candidate := newRestoreExecutorFixture(t)
	if err := controller.Close(context.Background()); err != nil {
		t.Fatalf("close parent Store before recovery crash fixture: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRestoreTransactionExecutorProcessDeathHelper$")
	rollbackValue := "0"
	if rollback {
		rollbackValue = "1"
	}
	command.Env = append(os.Environ(),
		restoreExecutorCrashChildEnv+"=1",
		restoreExecutorCrashRootEnv+"="+root,
		restoreExecutorCrashArchiveEnv+"="+archivePath,
		restoreExecutorCrashAtEnv+"="+boundary,
		restoreExecutorCrashRollbackEnv+"="+rollbackValue,
	)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("transaction crash child timed out at %q: %v\n%s", boundary, ctx.Err(), output.String())
	}
	exitError, ok := err.(*exec.ExitError)
	if !ok || exitError.ExitCode() != 86 {
		t.Fatalf("transaction crash child exit = %v, want deliberate exit 86 at %q\n%s", err, boundary, output.String())
	}
	return root, previous, candidate
}

func TestRestoreRecoveryExecutorReconcilesInterruptedTransactionBoundaries(t *testing.T) {
	tests := []struct {
		name string
		boundary string
		rollback bool
		wantCandidate bool
	}{
		{name: "previous preservation intent", boundary: restoreBoundaryJournalPreservingPrevious},
		{name: "previous moved before completion", boundary: restoreBoundaryMovePreviousToRollback},
		{name: "previous preservation completed", boundary: restoreBoundaryJournalPreviousPreserved},
		{name: "candidate publication intent", boundary: restoreBoundaryJournalPublishing},
		{name: "candidate moved before completion", boundary: restoreBoundaryMoveStagingToTarget},
		{name: "candidate publication completed", boundary: restoreBoundaryJournalPublished},
		{name: "candidate reopen intent", boundary: restoreBoundaryJournalReopening},
		{name: "candidate validated read-only before commit", boundary: restoreBoundaryCandidateVerified},
		{name: "candidate opened after commit record", boundary: restoreBoundaryStoreCandidateOpened, wantCandidate: true},
		{name: "rollback intent", boundary: restoreBoundaryJournalRestoringPrevious, rollback: true},
		{name: "candidate quarantined", boundary: restoreBoundaryMoveTargetToQuarantine, rollback: true},
		{name: "staged candidate quarantined", boundary: restoreBoundaryMoveStagingToQuarantine, rollback: true},
		{name: "previous restored before completion record", boundary: restoreBoundaryMoveRollbackToTarget, rollback: true},
		{name: "previous reopened before completion record", boundary: restoreBoundaryStorePreviousOpened, rollback: true},
		{name: "rollback completion record", boundary: restoreBoundaryJournalRolledBack, rollback: true},
		{name: "commit completion record", boundary: restoreBoundaryJournalCommitted, wantCandidate: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, previous, candidate := crashRestoreTransactionForRecovery(t, test.boundary, test.rollback)
			executor, err := NewRestoreRecoveryExecutor(root)
			if err != nil {
				t.Fatalf("create recovery executor: %v", err)
			}
			store, err := executor.Recover(context.Background())
			if err != nil {
				t.Fatalf("reconcile interrupted transaction at %q: %v", test.boundary, err)
			}
			if store == nil {
				t.Fatalf("recovery at %q returned no validated Store", test.boundary)
			}
			got, loadErr := store.Load(context.Background())
			closeErr := store.Close()
			if err := errors.Join(loadErr, closeErr); err != nil {
				t.Fatalf("load or close recovered Store: %v", err)
			}
			want := previous
			wantPhase := RestorePhaseRolledBack
			wantAction := RestoreRecoveryVerifyRolledBack
			if test.wantCandidate {
				want = candidate
				wantPhase = RestorePhaseCommitted
				wantAction = RestoreRecoveryVerifyCommitted
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("recovered configuration = %#v, want %#v", got, want)
			}

			journal, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatalf("read reconciled journal: %v", err)
			}
			if journal.Phase != wantPhase {
				t.Fatalf("reconciled journal phase = %q, want %q", journal.Phase, wantPhase)
			}
			files, err := inspectRestoreRecoveryFiles(root, journal)
			if err != nil {
				t.Fatalf("inspect reconciled artifacts: %v", err)
			}
			plan, err := PlanRestoreRecovery(journal, files)
			if err != nil {
				t.Fatalf("plan reconciled terminal state: %v", err)
			}
			if plan.Action != wantAction {
				t.Fatalf("terminal recovery action = %q, want %q", plan.Action, wantAction)
			}
		})
	}
}

func TestRestoreRecoveryExecutorPreservesStateOnJournalAndMoveFailures(t *testing.T) {
	tests := []struct {
		name string
		fail func(*RestoreRecoveryExecutor, error)
		check func(*testing.T, string)
	}{
		{
			name: "rollback intent journal failure",
			fail: func(executor *RestoreRecoveryExecutor, injected error) {
				realWrite := executor.operations.writeJournal
				executor.operations.writeJournal = func(lease *restoreJournalLease, journal RestoreJournal, boundary string) error {
					if boundary == restoreRecoveryBoundaryJournalRestoringPrevious {
						return injected
					}
					return realWrite(lease, journal, boundary)
				}
			},
			check: func(t *testing.T, root string) {
				t.Helper()
				journal, err := ReadRestoreJournal(root)
				if err != nil {
					t.Fatal(err)
				}
				if journal.Phase != RestorePhasePublished {
					t.Fatalf("journal phase = %q, want published", journal.Phase)
				}
				files, err := inspectRestoreRecoveryFiles(root, journal)
				if err != nil {
					t.Fatal(err)
				}
				if files.Target != RestoreFileCandidate || files.Rollback != RestoreFilePrevious {
					t.Fatalf("artifacts changed before rollback intent: %#v", files)
				}
			},
		},
		{
			name: "pre-move failure",
			fail: func(executor *RestoreRecoveryExecutor, injected error) {
				realMove := executor.operations.move
				executor.operations.move = func(source, destination string, expected RestoreContentIdentity) error {
					if samePath(source, filepath.Join(executor.stateDirectory, "state.db")) {
						return injected
					}
					return realMove(source, destination, expected)
				}
			},
			check: func(t *testing.T, root string) {
				t.Helper()
				journal, err := ReadRestoreJournal(root)
				if err != nil {
					t.Fatal(err)
				}
				if journal.Phase != RestorePhaseRestoringPrevious {
					t.Fatalf("journal phase = %q, want restoring-previous", journal.Phase)
				}
				files, err := inspectRestoreRecoveryFiles(root, journal)
				if err != nil {
					t.Fatal(err)
				}
				if files.Target != RestoreFileCandidate || files.Rollback != RestoreFilePrevious || files.Quarantine != RestoreFileMissing {
					t.Fatalf("pre-move failure changed artifacts: %#v", files)
				}
			},
		},
		{
			name: "post-namespace move failure",
			fail: func(executor *RestoreRecoveryExecutor, injected error) {
				realMove := executor.operations.move
				executor.operations.move = func(source, destination string, expected RestoreContentIdentity) error {
					if samePath(source, filepath.Join(executor.stateDirectory, "state.db")) {
						return moveRestoreFileNoReplaceVerifiedWithHook(source, destination, expected, func() error {
							return injected
						})
					}
					return realMove(source, destination, expected)
				}
			},
			check: func(t *testing.T, root string) {
				t.Helper()
				journal, err := ReadRestoreJournal(root)
				if err != nil {
					t.Fatal(err)
				}
				if journal.Phase != RestorePhaseRestoringPrevious {
					t.Fatalf("journal phase = %q, want restoring-previous", journal.Phase)
				}
				files, err := inspectRestoreRecoveryFiles(root, journal)
				if err != nil {
					t.Fatal(err)
				}
				if files.Target != RestoreFileMissing || files.Rollback != RestoreFilePrevious || files.Quarantine != RestoreFileCandidate {
					t.Fatalf("post-namespace failure did not preserve moved candidate: %#v", files)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, _, _ := crashRestoreTransactionForRecovery(t, restoreBoundaryJournalPublished, false)
			executor, err := NewRestoreRecoveryExecutor(root)
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected recovery failure: " + test.name)
			test.fail(executor, injected)
			store, err := executor.Recover(context.Background())
			if store != nil {
				_ = store.Close()
				t.Fatal("recovery returned a Store after an injected failure")
			}
			if !errors.Is(err, injected) {
				t.Fatalf("Recover() error = %v, want injected error", err)
			}
			test.check(t, root)

			retry, err := NewRestoreRecoveryExecutor(root)
			if err != nil {
				t.Fatal(err)
			}
			recovered, err := retry.Recover(context.Background())
			if err != nil {
				t.Fatalf("retry recovery after preserved failure state: %v", err)
			}
			if recovered == nil {
				t.Fatal("retry recovery returned no Store")
			}
			if err := recovered.Close(); err != nil {
				t.Fatal(err)
			}
			journal, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatal(err)
			}
			if journal.Phase != RestorePhaseRolledBack {
				t.Fatalf("retry journal phase = %q, want rolled-back", journal.Phase)
			}
		})
	}
}


const (
	restoreRecoveryCrashChildEnv = "OHMYFREEAI_RESTORE_RECOVERY_CRASH_CHILD"
	restoreRecoveryCrashRootEnv = "OHMYFREEAI_RESTORE_RECOVERY_CRASH_ROOT"
	restoreRecoveryCrashBoundaryEnv = "OHMYFREEAI_RESTORE_RECOVERY_CRASH_BOUNDARY"
	restoreRecoveryCrashExitCode = 87
)

func TestRestoreRecoveryExecutorProcessDeathHelper(t *testing.T) {
	if os.Getenv(restoreRecoveryCrashChildEnv) != "1" {
		return
	}
	root := os.Getenv(restoreRecoveryCrashRootEnv)
	boundary := os.Getenv(restoreRecoveryCrashBoundaryEnv)
	executor, err := NewRestoreRecoveryExecutor(root)
	if err != nil {
		t.Fatalf("create child recovery executor: %v", err)
	}
	executor.operations.afterBoundary = func(actual string) {
		if actual == boundary {
			os.Exit(restoreRecoveryCrashExitCode)
		}
	}
	store, err := executor.Recover(context.Background())
	if err != nil {
		t.Fatalf("child recovery failed before crash boundary %q: %v", boundary, err)
	}
	if store != nil {
		if err := store.Close(); err != nil {
			t.Fatalf("close child recovered Store without reaching boundary %q: %v", boundary, err)
		}
	}
	t.Fatalf("child recovery completed without reaching boundary %q", boundary)
}

func crashRestoreRecoveryAtBoundary(t *testing.T, root, boundary string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRestoreRecoveryExecutorProcessDeathHelper$")
	command.Env = append(os.Environ(),
		restoreRecoveryCrashChildEnv+"=1",
		restoreRecoveryCrashRootEnv+"="+root,
		restoreRecoveryCrashBoundaryEnv+"="+boundary,
	)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("recovery crash child timed out at %q: %v\n%s", boundary, ctx.Err(), output.String())
	}
	exitError, ok := err.(*exec.ExitError)
	if !ok || exitError.ExitCode() != restoreRecoveryCrashExitCode {
		t.Fatalf("recovery crash child exit = %v, want deliberate exit %d at %q\n%s", err, restoreRecoveryCrashExitCode, boundary, output.String())
	}
}

func TestRestoreRecoveryExecutorRestartsAfterEveryRecoveryBoundary(t *testing.T) {
	tests := []struct {
		name string
		initialBoundary string
		recoveryBoundary string
		wantCandidate bool
	}{
		{
			name: "previous preservation move",
			initialBoundary: restoreBoundaryJournalPreservingPrevious,
			recoveryBoundary: restoreRecoveryBoundaryMoveTargetToRollback,
		},
		{
			name: "previous preservation completion",
			initialBoundary: restoreBoundaryJournalPreservingPrevious,
			recoveryBoundary: restoreRecoveryBoundaryJournalPreviousPreserved,
		},
		{
			name: "rollback intent",
			initialBoundary: restoreBoundaryJournalPublished,
			recoveryBoundary: restoreRecoveryBoundaryJournalRestoringPrevious,
		},
		{
			name: "candidate target quarantine",
			initialBoundary: restoreBoundaryJournalPublished,
			recoveryBoundary: restoreRecoveryBoundaryMoveTargetToQuarantine,
		},
		{
			name: "staging candidate quarantine",
			initialBoundary: restoreBoundaryJournalPreservingPrevious,
			recoveryBoundary: restoreRecoveryBoundaryMoveStagingToQuarantine,
		},
		{
			name: "previous database restored",
			initialBoundary: restoreBoundaryJournalPublished,
			recoveryBoundary: restoreRecoveryBoundaryMoveRollbackToTarget,
		},
		{
			name: "previous Store opened",
			initialBoundary: restoreBoundaryJournalPublished,
			recoveryBoundary: restoreRecoveryBoundaryStoreOpened,
		},
		{
			name: "rollback completion journal",
			initialBoundary: restoreBoundaryJournalPublished,
			recoveryBoundary: restoreRecoveryBoundaryJournalRolledBack,
		},
		{
			name: "committed Store opened",
			initialBoundary: restoreBoundaryJournalCommitted,
			recoveryBoundary: restoreRecoveryBoundaryStoreCommitted,
			wantCandidate: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, previous, candidate := crashRestoreTransactionForRecovery(t, test.initialBoundary, false)
			crashRestoreRecoveryAtBoundary(t, root, test.recoveryBoundary)

			executor, err := NewRestoreRecoveryExecutor(root)
			if err != nil {
				t.Fatalf("create restarted recovery executor: %v", err)
			}
			store, err := executor.Recover(context.Background())
			if err != nil {
				t.Fatalf("restart recovery after process death at %q: %v", test.recoveryBoundary, err)
			}
			if store == nil {
				t.Fatalf("restart recovery at %q returned no validated Store", test.recoveryBoundary)
			}
			got, loadErr := store.Load(context.Background())
			closeErr := store.Close()
			if err := errors.Join(loadErr, closeErr); err != nil {
				t.Fatalf("load or close restarted Store: %v", err)
			}

			want := previous
			wantPhase := RestorePhaseRolledBack
			wantAction := RestoreRecoveryVerifyRolledBack
			if test.wantCandidate {
				want = candidate
				wantPhase = RestorePhaseCommitted
				wantAction = RestoreRecoveryVerifyCommitted
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("restarted Store configuration = %#v, want %#v", got, want)
			}
			journal, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatalf("read journal after restarted recovery: %v", err)
			}
			if journal.Phase != wantPhase {
				t.Fatalf("journal phase after restarted recovery = %q, want %q", journal.Phase, wantPhase)
			}
			files, err := inspectRestoreRecoveryFiles(root, journal)
			if err != nil {
				t.Fatalf("inspect artifacts after restarted recovery: %v", err)
			}
			plan, err := PlanRestoreRecovery(journal, files)
			if err != nil {
				t.Fatalf("plan terminal state after restarted recovery: %v", err)
			}
			if plan.Action != wantAction {
				t.Fatalf("terminal action after restarted recovery = %q, want %q", plan.Action, wantAction)
			}
		})
	}
}


func TestRestoreRecoveryExecutorRetriesAllInjectedFailureBoundaries(t *testing.T) {
	type failureCase struct {
		name string
		initialBoundary string
		journalBoundary string
		moveSource string
		moveDestination string
		moveAfterNamespace bool
		failOpen bool
		wantPhase RestorePhase
		wantFiles RestoreRecoveryFiles
		wantCandidate bool
	}
	tests := []failureCase{
		{
			name: "rollback intent journal",
			initialBoundary: restoreBoundaryJournalPublished,
			journalBoundary: restoreRecoveryBoundaryJournalRestoringPrevious,
			wantPhase: RestorePhasePublished,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileCandidate, Staging: RestoreFileMissing, Rollback: RestoreFilePrevious, Quarantine: RestoreFileMissing},
		},
		{
			name: "target quarantine pre-move",
			initialBoundary: restoreBoundaryJournalPublished,
			moveSource: "target",
			moveDestination: "quarantine",
			wantPhase: RestorePhaseRestoringPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileCandidate, Staging: RestoreFileMissing, Rollback: RestoreFilePrevious, Quarantine: RestoreFileMissing},
		},
		{
			name: "target quarantine post-namespace",
			initialBoundary: restoreBoundaryJournalPublished,
			moveSource: "target",
			moveDestination: "quarantine",
			moveAfterNamespace: true,
			wantPhase: RestorePhaseRestoringPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileMissing, Staging: RestoreFileMissing, Rollback: RestoreFilePrevious, Quarantine: RestoreFileCandidate},
		},
		{
			name: "rollback target pre-move",
			initialBoundary: restoreBoundaryJournalPublished,
			moveSource: "rollback",
			moveDestination: "target",
			wantPhase: RestorePhaseRestoringPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileMissing, Staging: RestoreFileMissing, Rollback: RestoreFilePrevious, Quarantine: RestoreFileCandidate},
		},
		{
			name: "rollback target post-namespace",
			initialBoundary: restoreBoundaryJournalPublished,
			moveSource: "rollback",
			moveDestination: "target",
			moveAfterNamespace: true,
			wantPhase: RestorePhaseRestoringPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFilePrevious, Staging: RestoreFileMissing, Rollback: RestoreFileMissing, Quarantine: RestoreFileCandidate},
		},
		{
			name: "rollback completion journal",
			initialBoundary: restoreBoundaryJournalPublished,
			journalBoundary: restoreRecoveryBoundaryJournalRolledBack,
			wantPhase: RestorePhaseRestoringPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFilePrevious, Staging: RestoreFileMissing, Rollback: RestoreFileMissing, Quarantine: RestoreFileCandidate},
		},
		{
			name: "rollback Store open",
			initialBoundary: restoreBoundaryJournalPublished,
			failOpen: true,
			wantPhase: RestorePhaseRestoringPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFilePrevious, Staging: RestoreFileMissing, Rollback: RestoreFileMissing, Quarantine: RestoreFileCandidate},
		},
		{
			name: "committed Store open",
			initialBoundary: restoreBoundaryJournalCommitted,
			failOpen: true,
			wantPhase: RestorePhaseCommitted,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileCandidate, Staging: RestoreFileMissing, Rollback: RestoreFilePrevious, Quarantine: RestoreFileMissing},
			wantCandidate: true,
		},
		{
			name: "previous rollback pre-move",
			initialBoundary: restoreBoundaryJournalPreservingPrevious,
			moveSource: "target",
			moveDestination: "rollback",
			wantPhase: RestorePhasePreservingPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFilePrevious, Staging: RestoreFileCandidate, Rollback: RestoreFileMissing, Quarantine: RestoreFileMissing},
		},
		{
			name: "previous rollback post-namespace",
			initialBoundary: restoreBoundaryJournalPreservingPrevious,
			moveSource: "target",
			moveDestination: "rollback",
			moveAfterNamespace: true,
			wantPhase: RestorePhasePreservingPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileMissing, Staging: RestoreFileCandidate, Rollback: RestoreFilePrevious, Quarantine: RestoreFileMissing},
		},
		{
			name: "previous-preserved journal",
			initialBoundary: restoreBoundaryJournalPreservingPrevious,
			journalBoundary: restoreRecoveryBoundaryJournalPreviousPreserved,
			wantPhase: RestorePhasePreservingPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileMissing, Staging: RestoreFileCandidate, Rollback: RestoreFilePrevious, Quarantine: RestoreFileMissing},
		},
		{
			name: "restoring-previous journal",
			initialBoundary: restoreBoundaryJournalPreservingPrevious,
			journalBoundary: restoreRecoveryBoundaryJournalRestoringPrevious,
			wantPhase: RestorePhasePreviousPreserved,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileMissing, Staging: RestoreFileCandidate, Rollback: RestoreFilePrevious, Quarantine: RestoreFileMissing},
		},
		{
			name: "staging quarantine pre-move",
			initialBoundary: restoreBoundaryJournalPreservingPrevious,
			moveSource: "staging",
			moveDestination: "quarantine",
			wantPhase: RestorePhaseRestoringPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileMissing, Staging: RestoreFileCandidate, Rollback: RestoreFilePrevious, Quarantine: RestoreFileMissing},
		},
		{
			name: "staging quarantine post-namespace",
			initialBoundary: restoreBoundaryJournalPreservingPrevious,
			moveSource: "staging",
			moveDestination: "quarantine",
			moveAfterNamespace: true,
			wantPhase: RestorePhaseRestoringPrevious,
			wantFiles: RestoreRecoveryFiles{Target: RestoreFileMissing, Staging: RestoreFileMissing, Rollback: RestoreFilePrevious, Quarantine: RestoreFileCandidate},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, previous, candidate := crashRestoreTransactionForRecovery(t, test.initialBoundary, false)
			journal, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatalf("read interrupted transaction journal: %v", err)
			}
			executor, err := NewRestoreRecoveryExecutor(root)
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected recovery boundary failure: " + test.name)

			if test.journalBoundary != "" {
				realWrite := executor.operations.writeJournal
				executor.operations.writeJournal = func(lease *restoreJournalLease, next RestoreJournal, boundary string) error {
					if boundary == test.journalBoundary {
						return injected
					}
					return realWrite(lease, next, boundary)
				}
			}
			if test.moveSource != "" {
				sourcePath := journal.TargetPath
				switch test.moveSource {
				case "staging":
					sourcePath = journal.StagingPath
				case "rollback":
					sourcePath = journal.RollbackPath
				}
				destinationPath := journal.TargetPath
				switch test.moveDestination {
				case "rollback":
					destinationPath = journal.RollbackPath
				case "quarantine":
					destinationPath = journal.CandidateQuarantinePath
				}
				realMove := executor.operations.move
				executor.operations.move = func(source, destination string, expected RestoreContentIdentity) error {
					if samePath(source, sourcePath) && samePath(destination, destinationPath) {
						if test.moveAfterNamespace {
							return moveRestoreFileNoReplaceVerifiedWithHook(source, destination, expected, func() error {
								return injected
							})
						}
						return injected
					}
					return realMove(source, destination, expected)
				}
			}
			if test.failOpen {
				realOpen := executor.operations.openStore
				executor.operations.openStore = func(ctx context.Context, path string) (*Store, error) {
					if samePath(path, journal.TargetPath) {
						return nil, injected
					}
					return realOpen(ctx, path)
				}
			}

			store, err := executor.Recover(context.Background())
			if store != nil {
				_ = store.Close()
				t.Fatal("Recover returned a Store after an injected failure")
			}
			if !errors.Is(err, injected) {
				t.Fatalf("Recover() error = %v, want injected failure", err)
			}
			current, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatalf("read journal after injected failure: %v", err)
			}
			if current.Phase != test.wantPhase {
				t.Fatalf("journal phase after injected failure = %q, want %q", current.Phase, test.wantPhase)
			}
			files, err := inspectRestoreRecoveryFiles(root, current)
			if err != nil {
				t.Fatalf("inspect artifacts after injected failure: %v", err)
			}
			if files != test.wantFiles {
				t.Fatalf("artifact identities after injected failure = %#v, want %#v", files, test.wantFiles)
			}

			retry, err := NewRestoreRecoveryExecutor(root)
			if err != nil {
				t.Fatal(err)
			}
			recovered, err := retry.Recover(context.Background())
			if err != nil {
				t.Fatalf("retry recovery after %q: %v", test.name, err)
			}
			if recovered == nil {
				t.Fatal("retry recovery returned no validated Store")
			}
			got, loadErr := recovered.Load(context.Background())
			closeErr := recovered.Close()
			if err := errors.Join(loadErr, closeErr); err != nil {
				t.Fatalf("load or close recovered Store: %v", err)
			}
			want := previous
			wantPhase := RestorePhaseRolledBack
			if test.wantCandidate {
				want = candidate
				wantPhase = RestorePhaseCommitted
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("recovered configuration = %#v, want %#v", got, want)
			}
			terminal, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatalf("read terminal recovery journal: %v", err)
			}
			if terminal.Phase != wantPhase {
				t.Fatalf("terminal journal phase = %q, want %q", terminal.Phase, wantPhase)
			}
		})
	}
}


func TestRestoreRecoveryExecutorDetectsSourceLeafSubstitutionAfterMove(t *testing.T) {
	root, _, _ := crashRestoreTransactionForRecovery(t, restoreBoundaryJournalPublished, false)
	journal, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatal(err)
	}
	rollbackBefore, err := os.ReadFile(journal.RollbackPath)
	if err != nil {
		t.Fatal(err)
	}
	substituted := []byte("untrusted replacement source leaf")
	executor, err := NewRestoreRecoveryExecutor(root)
	if err != nil {
		t.Fatal(err)
	}
	openCalls := 0
	executor.operations.openStore = func(context.Context, string) (*Store, error) {
		openCalls++
		return nil, errors.New("unexpected SQLite open after source substitution")
	}
	realMove := executor.operations.move
	executor.operations.move = func(source, destination string, expected RestoreContentIdentity) error {
		if samePath(source, journal.TargetPath) && samePath(destination, journal.CandidateQuarantinePath) {
			if err := os.Remove(source); err != nil {
				return err
			}
			if err := os.WriteFile(source, substituted, 0o600); err != nil {
				return err
			}
		}
		return realMove(source, destination, expected)
	}

	store, err := executor.Recover(context.Background())
	if store != nil {
		_ = store.Close()
		t.Fatal("recovery returned a Store after source substitution")
	}
	if err == nil {
		t.Fatal("recovery accepted a substituted source leaf")
	}
	if openCalls != 0 {
		t.Fatalf("SQLite open calls = %d, want zero", openCalls)
	}
	after, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatal(err)
	}
	if after.Phase != RestorePhaseRestoringPrevious {
		t.Fatalf("journal phase = %q, want restoring-previous", after.Phase)
	}
	if _, err := os.Lstat(journal.TargetPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("substituted source unexpectedly remains at target: %v", err)
	}
	quarantined, err := os.ReadFile(journal.CandidateQuarantinePath)
	if err != nil {
		t.Fatalf("read preserved substituted artifact: %v", err)
	}
	if !bytes.Equal(quarantined, substituted) {
		t.Fatalf("quarantined artifact = %q, want substituted bytes %q", quarantined, substituted)
	}
	rollbackAfter, err := os.ReadFile(journal.RollbackPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rollbackAfter, rollbackBefore) {
		t.Fatal("previous rollback database changed after source substitution")
	}
	actual, err := readRestoreContentIdentityPlatform(journal.CandidateQuarantinePath)
	if err != nil {
		t.Fatal(err)
	}
	if actual.matches(journal.CandidateSHA256, journal.CandidateSize) {
		t.Fatal("substituted quarantined artifact unexpectedly matches the journal identity")
	}
}

func TestRestoreRecoveryExecutorDoesNotOverwriteRacedDestination(t *testing.T) {
	root, _, _ := crashRestoreTransactionForRecovery(t, restoreBoundaryJournalPublished, false)
	journal, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatal(err)
	}
	targetBefore, err := os.ReadFile(journal.TargetPath)
	if err != nil {
		t.Fatal(err)
	}
	rollbackBefore, err := os.ReadFile(journal.RollbackPath)
	if err != nil {
		t.Fatal(err)
	}
	collision := []byte("destination created after recovery inspection")
	executor, err := NewRestoreRecoveryExecutor(root)
	if err != nil {
		t.Fatal(err)
	}
	openCalls := 0
	executor.operations.openStore = func(context.Context, string) (*Store, error) {
		openCalls++
		return nil, errors.New("unexpected SQLite open after destination collision")
	}
	realMove := executor.operations.move
	executor.operations.move = func(source, destination string, expected RestoreContentIdentity) error {
		if samePath(source, journal.TargetPath) && samePath(destination, journal.CandidateQuarantinePath) {
			if err := os.WriteFile(destination, collision, 0o600); err != nil {
				return err
			}
		}
		return realMove(source, destination, expected)
	}

	store, err := executor.Recover(context.Background())
	if store != nil {
		_ = store.Close()
		t.Fatal("recovery returned a Store after destination collision")
	}
	if err == nil {
		t.Fatal("recovery accepted a destination that appeared after inspection")
	}
	if openCalls != 0 {
		t.Fatalf("SQLite open calls = %d, want zero", openCalls)
	}
	after, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatal(err)
	}
	if after.Phase != RestorePhaseRestoringPrevious {
		t.Fatalf("journal phase = %q, want restoring-previous", after.Phase)
	}
	for path, want := range map[string][]byte{
		journal.TargetPath:             targetBefore,
		journal.RollbackPath:           rollbackBefore,
		journal.CandidateQuarantinePath: collision,
	} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read preserved artifact %s: %v", path, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("artifact %s changed: got %q, want %q", path, got, want)
		}
	}
}



func TestRestoreRecoveryExecutorStopsBeforeMutationForUnverifiedArtifacts(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, string)
		verify  func(*testing.T, string)
	}{
		{
			name: "mismatched regular file",
			prepare: func(t *testing.T, target string) {
				if err := os.WriteFile(target, []byte("unverified replacement bytes"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			verify: func(t *testing.T, target string) {
				got, err := os.ReadFile(target)
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != "unverified replacement bytes" {
					t.Fatalf("mismatched target changed: %q", got)
				}
			},
		},
		{
			name: "directory substituted for candidate",
			prepare: func(t *testing.T, target string) {
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(target, 0o700); err != nil {
					t.Fatal(err)
				}
			},
			verify: func(t *testing.T, target string) {
				info, err := os.Lstat(target)
				if err != nil {
					t.Fatal(err)
				}
				if !info.IsDir() {
					t.Fatalf("unreadable target mode = %v, want directory", info.Mode())
				}
			},
		},
		{
			name: "symlink substituted for candidate",
			prepare: func(t *testing.T, target string) {
				external := target + ".external"
				if err := os.WriteFile(external, []byte("external data must remain untouched"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(external, target); err != nil {
					t.Skipf("symlink creation is unavailable in this environment: %v", err)
				}
			},
			verify: func(t *testing.T, target string) {
				info, err := os.Lstat(target)
				if err != nil {
					t.Fatal(err)
				}
				if info.Mode()&os.ModeSymlink == 0 {
					t.Fatalf("preserved target mode = %v, want symlink", info.Mode())
				}
				link, err := os.Readlink(target)
				if err != nil {
					t.Fatal(err)
				}
				if link != target+".external" {
					t.Fatalf("preserved symlink target = %q, want %q", link, target+".external")
				}
				got, err := os.ReadFile(target + ".external")
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != "external data must remain untouched" {
					t.Fatalf("external target changed: %q", got)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, _, _ := crashRestoreTransactionForRecovery(t, restoreBoundaryJournalPublished, false)
			journal, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatal(err)
			}
			rollbackBefore, err := os.ReadFile(journal.RollbackPath)
			if err != nil {
				t.Fatal(err)
			}
			test.prepare(t, journal.TargetPath)

			executor, err := NewRestoreRecoveryExecutor(root)
			if err != nil {
				t.Fatal(err)
			}
			openCalls, moveCalls, journalWrites := 0, 0, 0
			realWrite := executor.operations.writeJournal
			executor.operations.openStore = func(context.Context, string) (*Store, error) {
				openCalls++
				return nil, errors.New("unverified recovery artifact must not open SQLite")
			}
			executor.operations.move = func(string, string, RestoreContentIdentity) error {
				moveCalls++
				return errors.New("unverified recovery artifact must not move")
			}
			executor.operations.writeJournal = func(lease *restoreJournalLease, next RestoreJournal, boundary string) error {
				journalWrites++
				return realWrite(lease, next, boundary)
			}

			store, err := executor.Recover(context.Background())
			if store != nil {
				_ = store.Close()
				t.Fatal("recovery returned a Store for an unverified target")
			}
			if !errors.Is(err, ErrRestoreRecoveryAmbiguous) {
				t.Fatalf("Recover() error = %v, want ErrRestoreRecoveryAmbiguous", err)
			}
			if openCalls != 0 || moveCalls != 0 || journalWrites != 0 {
				t.Fatalf("recovery performed operations for an unverified target: open=%d move=%d journal-write=%d", openCalls, moveCalls, journalWrites)
			}

			after, err := ReadRestoreJournal(root)
			if err != nil {
				t.Fatal(err)
			}
			if after.Phase != RestorePhasePublished {
				t.Fatalf("journal phase changed to %q, want %q", after.Phase, RestorePhasePublished)
			}
			rollbackAfter, err := os.ReadFile(journal.RollbackPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(rollbackAfter, rollbackBefore) {
				t.Fatal("rollback database changed while rejecting an unverified target")
			}
			test.verify(t, journal.TargetPath)
		})
	}
}
