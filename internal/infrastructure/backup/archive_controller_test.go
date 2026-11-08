package backup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/SoufianoDev/OhMyFreeAI/internal/infrastructure/sqlite"
	"github.com/SoufianoDev/OhMyFreeAI/migrations"
)

func TestManagerCreateWithStoreControllerDrainsBeforeClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	directory := t.TempDir()
	databasePath := filepath.Join(directory, "state.db")
	store, err := sqlite.Open(ctx, databasePath)
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	if err := store.Migrate(ctx, migrations.Files, "."); err != nil {
		t.Fatalf("migrate SQLite store: %v", err)
	}

	controller := sqlite.NewStoreController(store)
	manager := Manager{ApplicationVersion: "0.1.0", Migrations: migrations.Files, MigrationDir: "."}
	archivePath := filepath.Join(directory, "backup.omfaibackup")
	archiveReady := make(chan struct{})
	releasePublication := make(chan struct{})
	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() {
			close(releasePublication)
		})
	}
	t.Cleanup(func() {
		release()
		if err := controller.Close(context.Background()); err != nil && !errors.Is(err, sqlite.ErrStoreUnavailable) {
			t.Errorf("close StoreController during cleanup: %v", err)
		}
	})

	exportDone := make(chan error, 1)
	go func() {
		exportDone <- manager.createWithStoreController(ctx, controller, archivePath, func() {
			close(archiveReady)
			<-releasePublication
		})
	}()

	select {
	case <-archiveReady:
	case err := <-exportDone:
		t.Fatalf("backup export returned before reaching the publication barrier: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("backup export did not reach the publication barrier")
	}

	if _, err := os.Lstat(archivePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("archive should not be published while the publication barrier is held; Lstat error = %v", err)
	}

	closeContext, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	closeErr := controller.Close(closeContext)
	closeCancel()
	if !errors.Is(closeErr, context.DeadlineExceeded) {
		t.Fatalf("controller close while backup export was active = %v, want context deadline exceeded", closeErr)
	}

	release()
	select {
	case err := <-exportDone:
		if err != nil {
			t.Fatalf("create backup with StoreController: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("backup export did not finish after releasing the publication barrier")
	}

	if err := controller.WithShared(context.Background(), func(current *sqlite.Store) error {
		if current != store {
			t.Fatal("controller did not retain the active Store after a canceled close")
		}
		return nil
	}); err != nil {
		t.Fatalf("shared operation after canceled close: %v", err)
	}

	if err := controller.Close(context.Background()); err != nil {
		t.Fatalf("controller close after backup export: %v", err)
	}

	info, err := os.Stat(archivePath)
	if err != nil {
		t.Fatalf("stat completed backup archive: %v", err)
	}
	if info.Size() <= 0 {
		t.Fatal("backup archive is empty")
	}
}
