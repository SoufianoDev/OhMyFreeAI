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
		{name: "candidate opened before commit record", boundary: restoreBoundaryStoreCandidateOpened},
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
