package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeBackupOperations struct {
	exportPath string
	inspectPath string
	inspection BackupInspection
	err error
}

func (f *fakeBackupOperations) Export(_ context.Context, path string) error {
	f.exportPath = path
	return f.err
}

func (f *fakeBackupOperations) Inspect(_ context.Context, path string) (BackupInspection, error) {
	f.inspectPath = path
	return f.inspection, f.err
}

func TestRunBackupExportUsesSharedOperationAndReportsPath(t *testing.T) {
	operations := &fakeBackupOperations{}
	var output bytes.Buffer
	path := "/tmp/my backup.omfaibackup"
	invocation, err := Parse([]string{"backup", "export", path})
	if err != nil { t.Fatal(err) }
	if err := RunBackup(context.Background(), &output, invocation, operations); err != nil { t.Fatal(err) }
	if operations.exportPath != path { t.Fatalf("export path = %q, want %q", operations.exportPath, path) }
	if !strings.Contains(output.String(), "Backup created: "+path) { t.Fatalf("unexpected output: %q", output.String()) }
}

func TestRunBackupInspectPrintsOnlyValidatedMetadata(t *testing.T) {
	created := time.Date(2026, 11, 7, 20, 0, 0, 0, time.UTC)
	operations := &fakeBackupOperations{inspection: BackupInspection{FormatVersion: 1, ApplicationVersion: "0.1.0", SchemaVersion: 3, CreatedAt: created, DatabaseSizeBytes: 4096, DatabaseSHA256: "abc123"}}
	var output bytes.Buffer
	path := "/tmp/state.omfaibackup"
	invocation, err := Parse([]string{"backup", "inspect", path})
	if err != nil { t.Fatal(err) }
	if err := RunBackup(context.Background(), &output, invocation, operations); err != nil { t.Fatal(err) }
	if operations.inspectPath != path { t.Fatalf("inspect path = %q, want %q", operations.inspectPath, path) }
	for _, expected := range []string{"Backup valid", "Format version: 1", "Application version: 0.1.0", "Database schema: 3", "Created: 2026-11-07T20:00:00Z", "Database size: 4096 bytes", "Database SHA-256: abc123"} {
		if !strings.Contains(output.String(), expected) { t.Fatalf("output missing %q: %s", expected, output.String()) }
	}
}

func TestRunBackupPropagatesOperationFailureWithoutSuccessOutput(t *testing.T) {
	wantErr := errors.New("archive rejected")
	operations := &fakeBackupOperations{err: wantErr}
	var output bytes.Buffer
	invocation, err := Parse([]string{"backup", "inspect", "bad.omfaibackup"})
	if err != nil { t.Fatal(err) }
	if err := RunBackup(context.Background(), &output, invocation, operations); !errors.Is(err, wantErr) { t.Fatalf("error = %v, want %v", err, wantErr) }
	if output.Len() != 0 { t.Fatalf("failure emitted success output: %q", output.String()) }
}

func TestRunBackupRejectsInvalidDependencies(t *testing.T) {
	invocation := Invocation{Command: BackupInspect, Arguments: []string{"backup.omfaibackup"}}
	if err := RunBackup(context.Background(), nil, invocation, &fakeBackupOperations{}); err == nil { t.Fatal("expected nil output rejection") }
	if err := RunBackup(context.Background(), &bytes.Buffer{}, invocation, nil); err == nil { t.Fatal("expected nil operations rejection") }
}
