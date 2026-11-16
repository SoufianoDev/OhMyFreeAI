package restore

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func validTestCandidate(path string) Candidate {
	return Candidate{Path: path, Size: 128, SHA256: strings.Repeat("a", 64)}
}

func validTestArchiveInfo() ArchiveInfo {
	return ArchiveInfo{
		FormatVersion:      1,
		ApplicationVersion: "0.1.0",
		SchemaVersion:      1,
		CreatedAt:          time.Date(2026, 11, 16, 8, 0, 0, 0, time.UTC),
		DatabaseSizeBytes:  256,
		DatabaseSHA256:     strings.Repeat("b", 64),
	}
}

func TestServiceRestoreBuildsCandidateInsideTransaction(t *testing.T) {
	ctx := context.Background()
	wantInfo := validTestArchiveInfo()
	builderCalls := 0
	builder := CandidateBuilderFunc(func(_ context.Context, archive, stateDirectory, candidatePath string) (Candidate, ArchiveInfo, error) {
		builderCalls++
		if archive != "/backups/selected.omfaibackup" {
			t.Fatalf("archive path = %q", archive)
		}
		if stateDirectory != "/private/state" || candidatePath != "/private/state/.restore-stage.sqlite" {
			t.Fatalf("candidate paths = %q / %q", stateDirectory, candidatePath)
		}
		return validTestCandidate(candidatePath), wantInfo, nil
	})

	executorCalls := 0
	executor := TransactionExecutorFunc(func(_ context.Context, request TransactionRequest) error {
		executorCalls++
		if request.ArchivePath != "/backups/selected.omfaibackup" {
			t.Fatalf("transaction archive path = %q", request.ArchivePath)
		}
		candidate, err := request.BuildCandidate(ctx, request.ArchivePath, "/private/state", "/private/state/.restore-stage.sqlite")
		if err != nil {
			return err
		}
		if candidate.Path != "/private/state/.restore-stage.sqlite" || candidate.Size != 128 || candidate.SHA256 != strings.Repeat("a", 64) {
			t.Fatalf("candidate = %#v", candidate)
		}
		return nil
	})

	service := NewService(builder, executor)
	gotInfo, err := service.Restore(ctx, "/backups/selected.omfaibackup")
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if gotInfo != wantInfo {
		t.Fatalf("archive info = %#v, want %#v", gotInfo, wantInfo)
	}
	if builderCalls != 1 || executorCalls != 1 {
		t.Fatalf("builder/executor calls = %d/%d, want 1/1", builderCalls, executorCalls)
	}
}

func TestServiceRestoreRejectsMissingDependenciesAndArchivePath(t *testing.T) {
	executor := TransactionExecutorFunc(func(context.Context, TransactionRequest) error { return nil })
	builder := CandidateBuilderFunc(func(context.Context, string, string, string) (Candidate, ArchiveInfo, error) {
		return validTestCandidate("/stage"), validTestArchiveInfo(), nil
	})

	for _, test := range []struct {
		name    string
		service Service
		path    string
		want    error
	}{
		{name: "missing builder", service: NewService(nil, executor), path: "/backup.omfaibackup", want: ErrServiceUnavailable},
		{name: "missing executor", service: NewService(builder, nil), path: "/backup.omfaibackup", want: ErrServiceUnavailable},
		{name: "empty path", service: NewService(builder, executor), path: "  ", want: ErrArchivePathRequired},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := test.service.Restore(context.Background(), test.path); !errors.Is(err, test.want) {
				t.Fatalf("Restore() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestServiceRestoreHonorsCancellationBeforeExecutor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	executorCalls := 0
	service := NewService(
		CandidateBuilderFunc(func(context.Context, string, string, string) (Candidate, ArchiveInfo, error) {
			t.Fatal("candidate builder called after cancellation")
			return Candidate{}, ArchiveInfo{}, nil
		}),
		TransactionExecutorFunc(func(context.Context, TransactionRequest) error {
			executorCalls++
			return nil
		}),
	)

	if _, err := service.Restore(ctx, "/backup.omfaibackup"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Restore() error = %v, want context.Canceled", err)
	}
	if executorCalls != 0 {
		t.Fatalf("executor calls = %d, want zero", executorCalls)
	}
}

func TestServiceRestoreRejectsInvalidCandidateIdentity(t *testing.T) {
	service := NewService(
		CandidateBuilderFunc(func(_ context.Context, _, _, candidatePath string) (Candidate, ArchiveInfo, error) {
			return Candidate{Path: candidatePath, Size: 0, SHA256: "not-a-digest"}, ArchiveInfo{}, nil
		}),
		TransactionExecutorFunc(func(ctx context.Context, request TransactionRequest) error {
			_, err := request.BuildCandidate(ctx, request.ArchivePath, "/state", "/state/.stage")
			return err
		}),
	)

	if _, err := service.Restore(context.Background(), "/backup.omfaibackup"); !errors.Is(err, ErrCandidateIdentityInvalid) {
		t.Fatalf("Restore() error = %v, want ErrCandidateIdentityInvalid", err)
	}
}

func TestServiceRestoreRequiresExecutorToInvokeCandidateBuilder(t *testing.T) {
	service := NewService(
		CandidateBuilderFunc(func(context.Context, string, string, string) (Candidate, ArchiveInfo, error) {
			t.Fatal("candidate builder should not be called")
			return Candidate{}, ArchiveInfo{}, nil
		}),
		TransactionExecutorFunc(func(context.Context, TransactionRequest) error { return nil }),
	)

	if _, err := service.Restore(context.Background(), "/backup.omfaibackup"); !errors.Is(err, ErrCandidateBuilderNotInvoked) {
		t.Fatalf("Restore() error = %v, want ErrCandidateBuilderNotInvoked", err)
	}
}

func TestServiceRestoreRejectsSecondCandidateBuilderInvocation(t *testing.T) {
	service := NewService(
		CandidateBuilderFunc(func(_ context.Context, _, _, candidatePath string) (Candidate, ArchiveInfo, error) {
			return validTestCandidate(candidatePath), validTestArchiveInfo(), nil
		}),
		TransactionExecutorFunc(func(ctx context.Context, request TransactionRequest) error {
			for i := 0; i < 2; i++ {
				if _, err := request.BuildCandidate(ctx, request.ArchivePath, "/state", "/state/.stage"); err != nil {
					return err
				}
			}
			return nil
		}),
	)

	if _, err := service.Restore(context.Background(), "/backup.omfaibackup"); !errors.Is(err, ErrCandidateBuilderCalledMoreThanOnce) {
		t.Fatalf("Restore() error = %v, want ErrCandidateBuilderCalledMoreThanOnce", err)
	}
}

func TestServiceRestorePropagatesTransactionFailureAfterCandidateBuild(t *testing.T) {
	injected := errors.New("injected transaction failure")
	wantInfo := validTestArchiveInfo()
	service := NewService(
		CandidateBuilderFunc(func(_ context.Context, _, _, candidatePath string) (Candidate, ArchiveInfo, error) {
			return validTestCandidate(candidatePath), wantInfo, nil
		}),
		TransactionExecutorFunc(func(ctx context.Context, request TransactionRequest) error {
			if _, err := request.BuildCandidate(ctx, request.ArchivePath, "/state", "/state/.stage"); err != nil {
				return err
			}
			return injected
		}),
	)

	gotInfo, err := service.Restore(context.Background(), "/backup.omfaibackup")
	if !errors.Is(err, injected) {
		t.Fatalf("Restore() error = %v, want injected transaction failure", err)
	}
	if gotInfo != wantInfo {
		t.Fatalf("archive info = %#v, want %#v", gotInfo, wantInfo)
	}
}
