package restore

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	domainconfig "github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
	"github.com/SoufianoDev/OhMyFreeAI/internal/infrastructure/backup"
	"github.com/SoufianoDev/OhMyFreeAI/internal/infrastructure/sqlite"
	"github.com/SoufianoDev/OhMyFreeAI/migrations"
)

func openConfiguredStore(t *testing.T, path string, configuration domainconfig.Config) *sqlite.Store {
	t.Helper()
	ctx := context.Background()
	store, err := sqlite.Open(ctx, path)
	if err != nil {
		t.Fatalf("open test SQLite Store: %v", err)
	}
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		_ = store.Close()
		t.Fatalf("migrate test SQLite Store: %v", err)
	}
	if err := store.Save(ctx, configuration); err != nil {
		_ = store.Close()
		t.Fatalf("save test SQLite configuration: %v", err)
	}
	return store
}

func TestApplicationServiceRestoresRealBackupThroughTransactionExecutor(t *testing.T) {
	ctx := context.Background()
	stateDirectory := t.TempDir()

	previousConfiguration := domainconfig.Defaults()
	previousConfiguration.Gateway.Port = 18180
	previousConfiguration.UI.Language = "ar"
	previousStore := openConfiguredStore(t, filepath.Join(stateDirectory, "state.db"), previousConfiguration)
	controller := sqlite.NewStoreController(previousStore)
	t.Cleanup(func() {
		if err := controller.Close(context.Background()); err != nil {
			t.Errorf("close StoreController: %v", err)
		}
	})

	sourceDirectory := t.TempDir()
	candidateConfiguration := domainconfig.Defaults()
	candidateConfiguration.Gateway.Port = 19220
	candidateConfiguration.UI.Language = "en"
	sourceStore := openConfiguredStore(t, filepath.Join(sourceDirectory, "candidate-source.db"), candidateConfiguration)
	manager := backup.Manager{
		ApplicationVersion: "0.1.0",
		Migrations:         migrations.Files,
		MigrationDir:       ".",
	}
	archivePath := filepath.Join(sourceDirectory, "candidate.omfaibackup")
	if err := manager.Create(ctx, sourceStore, archivePath); err != nil {
		_ = sourceStore.Close()
		t.Fatalf("create real .omfaibackup fixture: %v", err)
	}
	if err := sourceStore.Close(); err != nil {
		t.Fatalf("close source Store: %v", err)
	}

	service, err := NewApplicationService(manager, controller, stateDirectory)
	if err != nil {
		t.Fatalf("create isolated application Restore service: %v", err)
	}
	info, err := service.Restore(ctx, archivePath)
	if err != nil {
		t.Fatalf("restore validated .omfaibackup through application service: %v", err)
	}
	if info.FormatVersion != 1 || info.ApplicationVersion != "0.1.0" || info.SchemaVersion != 1 ||
		info.DatabaseSizeBytes == 0 || len(info.DatabaseSHA256) != 64 || info.CreatedAt.IsZero() {
		t.Fatalf("archive info = %#v", info)
	}

	var actualConfiguration domainconfig.Config
	if err := controller.WithShared(ctx, func(store *sqlite.Store) error {
		var loadErr error
		actualConfiguration, loadErr = store.Load(ctx)
		return loadErr
	}); err != nil {
		t.Fatalf("load configuration from restored StoreController: %v", err)
	}
	if !reflect.DeepEqual(actualConfiguration, candidateConfiguration) {
		t.Fatalf("restored configuration = %#v, want %#v", actualConfiguration, candidateConfiguration)
	}

	journal, err := sqlite.ReadRestoreJournal(stateDirectory)
	if err != nil {
		t.Fatalf("read committed Restore journal: %v", err)
	}
	if journal.Phase != sqlite.RestorePhaseCommitted || journal.PreviousDatabaseExisted == nil || !*journal.PreviousDatabaseExisted {
		t.Fatalf("Restore journal = %#v, want committed transaction with a previous database", journal)
	}
	rollbackInfo, err := os.Stat(journal.RollbackPath)
	if err != nil {
		t.Fatalf("inspect retained previous database: %v", err)
	}
	if !rollbackInfo.Mode().IsRegular() || rollbackInfo.Size() == 0 {
		t.Fatalf("retained rollback artifact is not a non-empty regular file: %#v", rollbackInfo)
	}
}
