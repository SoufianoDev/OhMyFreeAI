package restore

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrServiceUnavailable                 = errors.New("Restore service is not configured")
	ErrArchivePathRequired                = errors.New("Restore archive path is required")
	ErrCandidateIdentityInvalid           = errors.New("Restore candidate builder returned an invalid identity")
	ErrCandidateBuilderNotInvoked         = errors.New("Restore transaction executor did not invoke the candidate builder")
	ErrCandidateBuilderCalledMoreThanOnce = errors.New("Restore transaction executor invoked the candidate builder more than once")
)

// ArchiveInfo contains metadata from the validated .omfaibackup manifest.
// It is safe to display, but it is not a substitute for the candidate's
// separately measured post-build identity.
type ArchiveInfo struct {
	FormatVersion      int
	ApplicationVersion string
	SchemaVersion      int
	CreatedAt          time.Time
	DatabaseSizeBytes  uint64
	DatabaseSHA256     string
}

// Candidate is the application-layer view of a fully built candidate. The
// transaction executor still revalidates this identity from a no-follow
// filesystem handle before it moves or opens the candidate.
type Candidate struct {
	Path   string
	Size   uint64
	SHA256 string
}

type CandidateBuilder interface {
	BuildCandidate(context.Context, string, string, string) (Candidate, ArchiveInfo, error)
}

type CandidateBuilderFunc func(context.Context, string, string, string) (Candidate, ArchiveInfo, error)

func (f CandidateBuilderFunc) BuildCandidate(ctx context.Context, archivePath, stateDirectory, candidatePath string) (Candidate, ArchiveInfo, error) {
	return f(ctx, archivePath, stateDirectory, candidatePath)
}

// TransactionRequest passes the candidate builder into the transaction
// executor. Implementations must call BuildCandidate synchronously exactly
// once during Execute and must not retain the callback after returning.
type TransactionRequest struct {
	ArchivePath    string
	BuildCandidate func(context.Context, string, string, string) (Candidate, error)
}

type TransactionExecutor interface {
	Execute(context.Context, TransactionRequest) error
}

type TransactionExecutorFunc func(context.Context, TransactionRequest) error

func (f TransactionExecutorFunc) Execute(ctx context.Context, request TransactionRequest) error {
	return f(ctx, request)
}

// Service coordinates the archive-aware candidate builder with the isolated
// transaction executor. It deliberately has no GUI, CLI, or startup hooks.
type Service struct {
	builder  CandidateBuilder
	executor TransactionExecutor
}

func NewService(builder CandidateBuilder, executor TransactionExecutor) Service {
	return Service{builder: builder, executor: executor}
}

// Restore executes an already-selected archive through the transaction
// executor. The executor owns locking, journal phases, file moves, rollback,
// and Store replacement; the service only supplies the real candidate builder.
func (s Service) Restore(ctx context.Context, archivePath string) (ArchiveInfo, error) {
	var empty ArchiveInfo
	if s.builder == nil || s.executor == nil {
		return empty, ErrServiceUnavailable
	}
	if strings.TrimSpace(archivePath) == "" {
		return empty, ErrArchivePathRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return empty, err
	}

	var archiveInfo ArchiveInfo
	builderCalls := 0
	request := TransactionRequest{ArchivePath: archivePath}
	request.BuildCandidate = func(buildContext context.Context, archive, stateDirectory, candidatePath string) (Candidate, error) {
		if buildContext == nil {
			buildContext = ctx
		}
		if err := buildContext.Err(); err != nil {
			return Candidate{}, err
		}
		builderCalls++
		if builderCalls > 1 {
			return Candidate{}, ErrCandidateBuilderCalledMoreThanOnce
		}

		candidate, info, err := s.builder.BuildCandidate(buildContext, archive, stateDirectory, candidatePath)
		if err != nil {
			return candidate, err
		}
		if strings.TrimSpace(candidate.Path) == "" || candidate.Size == 0 || !validSHA256(candidate.SHA256) {
			return candidate, ErrCandidateIdentityInvalid
		}
		archiveInfo = info
		return candidate, nil
	}

	if err := s.executor.Execute(ctx, request); err != nil {
		return archiveInfo, fmt.Errorf("execute SQLite Restore transaction: %w", err)
	}
	if builderCalls == 0 {
		return empty, ErrCandidateBuilderNotInvoked
	}
	if builderCalls > 1 {
		return archiveInfo, ErrCandidateBuilderCalledMoreThanOnce
	}
	return archiveInfo, nil
}

func validSHA256(value string) bool {
	if value == "" || value != strings.ToLower(value) {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32
}
