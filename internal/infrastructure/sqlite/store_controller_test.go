package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func newLifecycleTestStore(t *testing.T, name string) *Store {
	t.Helper()
	store, err := Open(context.Background(), filepath.Join(t.TempDir(), name))
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	return store
}

func TestStoreControllerExclusiveWaitsForSharedLease(t *testing.T) {
	store := newLifecycleTestStore(t, "state.db")
	controller := NewStoreController(store)

	sharedEntered := make(chan struct{})
	releaseShared := make(chan struct{})
	sharedDone := make(chan error, 1)
	go func() {
		sharedDone <- controller.WithShared(context.Background(), func(got *Store) error {
			if got != store {
				t.Errorf("shared lease returned a different store")
			}
			close(sharedEntered)
			<-releaseShared
			return nil
		})
	}()
	<-sharedEntered

	var exclusiveRan atomic.Bool
	exclusiveDone := make(chan error, 1)
	go func() {
		exclusiveDone <- controller.WithExclusive(context.Background(), func(got *Store) (*Store, error) {
			exclusiveRan.Store(true)
			return got, nil
		})
	}()

	select {
	case err := <-exclusiveDone:
		t.Fatalf("exclusive lease completed before shared lease drained: %v", err)
	case <-time.After(40 * time.Millisecond):
	}
	if exclusiveRan.Load() {
		t.Fatal("exclusive callback ran while a shared lease was active")
	}

	close(releaseShared)
	if err := <-sharedDone; err != nil {
		t.Fatalf("shared operation: %v", err)
	}
	if err := <-exclusiveDone; err != nil {
		t.Fatalf("exclusive operation: %v", err)
	}
	if !exclusiveRan.Load() {
		t.Fatal("exclusive callback did not run")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close test store: %v", err)
	}
}

func TestStoreControllerWriterPreventsNewSharedLeases(t *testing.T) {
	store := newLifecycleTestStore(t, "state.db")
	controller := NewStoreController(store)

	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- controller.WithShared(context.Background(), func(*Store) error {
			close(firstEntered)
			<-releaseFirst
			return nil
		})
	}()
	<-firstEntered

	writerStarted := make(chan struct{})
	writerDone := make(chan error, 1)
	go func() {
		close(writerStarted)
		writerDone <- controller.WithExclusive(context.Background(), func(got *Store) (*Store, error) {
			return got, nil
		})
	}()
	<-writerStarted
	// Wait until the writer has registered itself before attempting a
	// second reader. This makes the writer-preference assertion deterministic.
	deadline := time.Now().Add(time.Second)
	for {
		controller.mu.Lock()
		waiting := controller.waitingWriters
		controller.mu.Unlock()
		if waiting > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("exclusive operation did not register as waiting")
		}
		time.Sleep(time.Millisecond)
	}

	secondEntered := make(chan struct{})
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- controller.WithShared(context.Background(), func(*Store) error {
			close(secondEntered)
			return nil
		})
	}()
	select {
	case <-secondEntered:
		t.Fatal("new shared lease bypassed a waiting exclusive lease")
	case <-time.After(40 * time.Millisecond):
	}

	close(releaseFirst)
	if err := <-firstDone; err != nil {
		t.Fatalf("first shared operation: %v", err)
	}
	if err := <-writerDone; err != nil {
		t.Fatalf("exclusive operation: %v", err)
	}
	select {
	case <-secondEntered:
	case <-time.After(time.Second):
		t.Fatal("second shared operation did not resume after exclusive lease")
	}
	if err := <-secondDone; err != nil {
		t.Fatalf("second shared operation: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close test store: %v", err)
	}
}

func TestStoreControllerCanceledWaitDoesNotBlockFutureOperations(t *testing.T) {
	store := newLifecycleTestStore(t, "state.db")
	controller := NewStoreController(store)

	entered := make(chan struct{})
	release := make(chan struct{})
	sharedDone := make(chan error, 1)
	go func() {
		sharedDone <- controller.WithShared(context.Background(), func(*Store) error {
			close(entered)
			<-release
			return nil
		})
	}()
	<-entered

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	err := controller.WithExclusive(ctx, func(got *Store) (*Store, error) {
		return got, nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("exclusive wait error = %v, want context deadline exceeded", err)
	}

	close(release)
	if err := <-sharedDone; err != nil {
		t.Fatalf("shared operation: %v", err)
	}
	if err := controller.WithShared(context.Background(), func(got *Store) error {
		if got != store {
			t.Fatal("controller lost its store after canceled exclusive wait")
		}
		return nil
	}); err != nil {
		t.Fatalf("shared operation after canceled wait: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close test store: %v", err)
	}
}

func TestStoreControllerRejectsNilStore(t *testing.T) {
	if _, err := NewStoreControllerChecked(nil); err == nil {
		t.Fatal("expected nil store to be rejected")
	}
}

func TestStoreControllerPublishesReplacementOnlyUnderExclusiveLease(t *testing.T) {
	oldStore := newLifecycleTestStore(t, "old.db")
	newStore := newLifecycleTestStore(t, "new.db")
	controller := NewStoreController(oldStore)

	if err := controller.WithExclusive(context.Background(), func(current *Store) (*Store, error) {
		if current != oldStore {
			t.Fatal("exclusive lease received unexpected current store")
		}
		if err := current.Close(); err != nil {
			return nil, err
		}
		return newStore, nil
	}); err != nil {
		t.Fatalf("replace store under exclusive lease: %v", err)
	}

	if err := controller.WithShared(context.Background(), func(current *Store) error {
		if current != newStore {
			t.Fatal("shared operation observed stale store after replacement")
		}
		return nil
	}); err != nil {
		t.Fatalf("shared operation after replacement: %v", err)
	}
	if err := newStore.Close(); err != nil {
		t.Fatalf("close replacement store: %v", err)
	}
}
