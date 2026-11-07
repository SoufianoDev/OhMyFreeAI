package cli

import (
	"context"
	"fmt"
	"io"
	"time"
)

// BackupInspection contains validated archive metadata safe to display to the user.
type BackupInspection struct {
	FormatVersion      int
	ApplicationVersion string
	SchemaVersion      int
	CreatedAt          time.Time
	DatabaseSizeBytes  uint64
	DatabaseSHA256     string
}

// BackupOperations is implemented by the composition root using the shared archive manager.
type BackupOperations interface {
	Export(context.Context, string) error
	Inspect(context.Context, string) (BackupInspection, error)
}

// RunBackup executes an already-parsed backup command without depending on infrastructure.
func RunBackup(ctx context.Context, output io.Writer, invocation Invocation, operations BackupOperations) error {
	if output == nil || operations == nil || len(invocation.Arguments) != 1 {
		return fmt.Errorf("invalid backup command configuration")
	}
	path := invocation.Arguments[0]
	switch invocation.Command {
	case BackupExport:
		if err := operations.Export(ctx, path); err != nil {
			return fmt.Errorf("export backup: %w", err)
		}
		_, err := fmt.Fprintf(output, "Backup created: %s\n", path)
		return err
	case BackupInspect:
		inspection, err := operations.Inspect(ctx, path)
		if err != nil {
			return fmt.Errorf("inspect backup: %w", err)
		}
		_, err = fmt.Fprintf(output,
			"Backup valid\nFormat version: %d\nApplication version: %s\nDatabase schema: %d\nCreated: %s\nDatabase size: %d bytes\nDatabase SHA-256: %s\n",
			inspection.FormatVersion, inspection.ApplicationVersion, inspection.SchemaVersion,
			inspection.CreatedAt.UTC().Format(time.RFC3339), inspection.DatabaseSizeBytes, inspection.DatabaseSHA256,
		)
		return err
	default:
		return fmt.Errorf("unsupported backup command %q", invocation.Command)
	}
}
