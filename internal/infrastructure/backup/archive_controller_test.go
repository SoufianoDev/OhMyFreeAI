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

	lockStore, err := sqlite.Open(ctx, databasePath)
	if err != nil {
		_ = store.Close()
		t.Fatalf("open SQLite lock store: %v", err)
	}
	var journalMode string
	if err := store.DB().QueryRowContext(ctx, "PRAGMA journal_mode=DELETE").Scan(&journalMode); err != nil {
		_ = lockStore.Close()
		_ = store.Close()
		t.Fatalf("switch test database to rollback journal mode: %v", err)
	}
	if journalMode != "delete" {
		_ = lockStore.Close()
		_ = store.Close()
		t.Fatalf("journal mode = %q, want delete", journalMode)
	}
	lockConn, err := lockStore.DB().Conn(ctx)
	if err != nil {
		_ = lockStore.Close()
		_ = store.Close()
		t.Fatalf("acquire SQLite lock connection: %v", err)
	}
	if _, err := lockConn.ExecContext(ctx, "BEGIN EXCLUSIVE"); err != nil {
		_ = lockConn.Close()
		_ = lockStore.Close()
		_ = store.Close()
		t.Fatalf("begin exclusive SQLite transaction: %v", err)
	}

	controller := sqlite.NewStoreController(store)
	var releaseOnce sync.Once
	var releaseErr error
	releaseLock := func() error {
		releaseOnce.Do(func() {
			_, releaseErr = lockConn.ExecContext(context.Background(), "ROLLBACK")
		})
		return releaseErr
	}
	t.Cleanup(func() {
		_ = releaseLock()
		_ = lockConn.Close()
		_ = lockStore.Close()
		if err := controller.Close(context.Background()); err != nil && !errors.Is(err, sqlite.ErrStoreUnavailable) {
			t.Errorf("close StoreController during cleanup: %v", err)
		}
	})

	manager := Manager{ApplicationVersion: "0.1.0", Migrations: migrations.Files, MigrationDir: "."}
	archivePath := filepath.Join(directory, "backup.omfaibackup")
	exportDone := make(chan error, 1)
	go func() {
		exportDone <- manager.CreateWithStoreController(ctx, controller, archivePath)
	}()

	deadline := time.Now().Add(3 * time.Second)
	for store.DB().Stats().InUse == 0 {
		if time.Now().After(deadline) {
			t.Fatal("backup did not reserve a SQLite source connection")
		}
		time.Sleep(time.Millisecond)
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- controller.Close(context.Background())
	}()
	select {
	case err := <-closeDone:
		t.Fatalf("controller close completed while backup export was active: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	if err := releaseLock(); err != nil {
		t.Fatalf("release SQLite lock: %v", err)
	}
	select {
	case err := <-exportDone:
		if err != nil {
			t.Fatalf("create backup with StoreController: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("backup export did not finish after SQLite lock release")
	}
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("controller close after backup export: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("controller close did not drain the completed backup export")
	}

	info, err := os.Stat(archivePath)
	if err != nil {
		t.Fatalf("stat completed backup archive: %v", err)
	}
	if info.Size() <= 0 {
		t.Fatal("backup archive is empty")
	}
}
