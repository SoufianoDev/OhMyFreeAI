package restore

import (
	"context"
	"fmt"

	apprestore "github.com/SoufianoDev/OhMyFreeAI/internal/application/restore"
	"github.com/SoufianoDev/OhMyFreeAI/internal/infrastructure/backup"
	"github.com/SoufianoDev/OhMyFreeAI/internal/infrastructure/sqlite"
)

type backupCandidateBuilder struct {
	manager backup.Manager
}

func (builder backupCandidateBuilder) BuildCandidate(
	ctx context.Context,
	archivePath, stateDirectory, candidatePath string,
) (apprestore.Candidate, apprestore.ArchiveInfo, error) {
	if builder.manager.Migrations == nil {
		return apprestore.Candidate{}, apprestore.ArchiveInfo{}, backup.ErrBackupManagerConfig
	}

	candidate, manifest, err := builder.manager.BuildRestoreCandidateAtPath(ctx, archivePath, stateDirectory, candidatePath)
	info, infoErr := archiveInfoFromManifest(manifest)
	if err != nil {
		return appCandidate(candidate), info, err
	}
	if infoErr != nil {
		return appCandidate(candidate), apprestore.ArchiveInfo{}, infoErr
	}
	return appCandidate(candidate), info, nil
}

func appCandidate(candidate sqlite.RestoreCandidate) apprestore.Candidate {
	return apprestore.Candidate{
		Path:   candidate.Path,
		Size:   candidate.Identity.Size,
		SHA256: candidate.Identity.SHA256,
	}
}

func archiveInfoFromManifest(manifest backup.Manifest) (apprestore.ArchiveInfo, error) {
	if len(manifest.Files) != 1 || manifest.Files[0].Path != "state.sqlite" {
		return apprestore.ArchiveInfo{}, fmt.Errorf("%w: expected exactly one state.sqlite entry", backup.ErrInvalidBackupArchive)
	}
	record := manifest.Files[0]
	return apprestore.ArchiveInfo{
		FormatVersion:      manifest.FormatVersion,
		ApplicationVersion: manifest.ApplicationVersion,
		SchemaVersion:      manifest.SchemaVersion,
		CreatedAt:          manifest.CreatedAt,
		DatabaseSizeBytes:  record.Size,
		DatabaseSHA256:     record.SHA256,
	}, nil
}

type sqliteTransactionExecutor struct {
	executor *sqlite.RestoreTransactionExecutor
}

func (adapter sqliteTransactionExecutor) Execute(ctx context.Context, request apprestore.TransactionRequest) error {
	if adapter.executor == nil || request.BuildCandidate == nil {
		return apprestore.ErrServiceUnavailable
	}
	return adapter.executor.Execute(ctx, sqlite.RestoreTransactionRequest{
		ArchivePath: request.ArchivePath,
		BuildCandidate: func(ctx context.Context, archivePath, stateDirectory, candidatePath string) (sqlite.RestoreCandidate, error) {
			candidate, err := request.BuildCandidate(ctx, archivePath, stateDirectory, candidatePath)
			return sqlite.RestoreCandidate{
				Path: candidate.Path,
				Identity: sqlite.RestoreContentIdentity{
					Size:   candidate.Size,
					SHA256: candidate.SHA256,
				},
			}, err
		},
	})
}

// NewApplicationService composes the real .omfaibackup candidate builder with
// the isolated SQLite transaction executor. The returned service is not wired
// to Wails, GUI, CLI, or startup recovery.
func NewApplicationService(
	manager backup.Manager,
	controller *sqlite.StoreController,
	stateDirectory string,
) (apprestore.Service, error) {
	if manager.Migrations == nil {
		return apprestore.Service{}, backup.ErrBackupManagerConfig
	}
	executor, err := sqlite.NewRestoreTransactionExecutor(controller, stateDirectory)
	if err != nil {
		return apprestore.Service{}, fmt.Errorf("create isolated SQLite Restore transaction executor: %w", err)
	}
	return apprestore.NewService(
		backupCandidateBuilder{manager: manager},
		sqliteTransactionExecutor{executor: executor},
	), nil
}
