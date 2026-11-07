package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"sync"
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


func TestStoreControllerCloseDrainsLeasesAndRejectsFutureOperations(t *testing.T) {
	store := newLifecycleTestStore(t, "state.db")
	controller := NewStoreController(store)

	entered := make(chan struct{})
	releaseShared := make(chan struct{})
	sharedDone := make(chan error, 1)
	go func() {
		sharedDone <- controller.WithShared(context.Background(), func(got *Store) error {
			if got != store {
				t.Errorf("shared lease returned an unexpected store")
			}
			close(entered)
			<-releaseShared
			return nil
		})
	}()
	<-entered

	closeDone := make(chan error, 1)
	go func() { closeDone <- controller.Close(context.Background()) }()
	select {
	case err := <-closeDone:
		t.Fatalf("controller closed while a shared lease was active: %v", err)
	case <-time.After(40 * time.Millisecond):
	}

	close(releaseShared)
	if err := <-sharedDone; err != nil {
		t.Fatalf("shared operation: %v", err)
	}
	if err := <-closeDone; err != nil {
		t.Fatalf("controller close: %v", err)
	}
	if err := controller.WithShared(context.Background(), func(*Store) error {
		t.Fatal("shared callback ran after controller close")
		return nil
	}); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("operation after close error = %v, want ErrStoreUnavailable", err)
	}
	if err := controller.Close(context.Background()); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("second close error = %v, want ErrStoreUnavailable", err)
	}
}

func TestStoreControllerPublishesVerifiedReplacementWhenExclusiveOperationFails(t *testing.T) {
	oldStore := newLifecycleTestStore(t, "old.db")
	newStore := newLifecycleTestStore(t, "rollback.db")
	controller := NewStoreController(oldStore)
	operationErr := errors.New("restore publication failed; rollback verified")

	err := controller.WithExclusive(context.Background(), func(current *Store) (*Store, error) {
		if current != oldStore {
			t.Fatal("exclusive operation received an unexpected current store")
		}
		if closeErr := current.Close(); closeErr != nil {
			return nil, closeErr
		}
		return newStore, operationErr
	})
	if !errors.Is(err, operationErr) {
		t.Fatalf("exclusive operation error = %v, want original operation error", err)
	}

	if err := controller.WithShared(context.Background(), func(current *Store) error {
		if current != newStore {
			t.Fatal("controller exposed the closed pre-restore store instead of the verified rollback store")
		}
		return nil
	}); err != nil {
		t.Fatalf("shared operation after failed restore: %v", err)
	}
	if err := controller.Close(context.Background()); err != nil {
		t.Fatalf("close controller: %v", err)
	}
}

func TestStoreControllerFailsClosedWhenExclusiveOperationReturnsNoSafeStore(t *testing.T) {
	store := newLifecycleTestStore(t, "state.db")
	controller := NewStoreController(store)
	operationErr := errors.New("rollback could not be verified")

	err := controller.WithExclusive(context.Background(), func(current *Store) (*Store, error) {
		if closeErr := current.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, operationErr
	})
	if !errors.Is(err, operationErr) {
		t.Fatalf("exclusive operation error = %v, want original operation error", err)
	}
	if err := controller.WithShared(context.Background(), func(*Store) error {
		t.Fatal("shared operation ran despite the controller having no verified active store")
		return nil
	}); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("shared operation error = %v, want ErrStoreUnavailable", err)
	}
}


func TestStoreControllerFailsClosedWhenExclusiveOperationPanics(t *testing.T) {
	store := newLifecycleTestStore(t, "state.db")
	controller := NewStoreController(store)

	func() {
		defer func() {
			if recovered := recover(); recovered != "injected destructive panic" {
				t.Fatalf("panic = %v, want injected destructive panic", recovered)
			}
		}()
		_ = controller.WithExclusive(context.Background(), func(current *Store) (*Store, error) {
			if err := current.Close(); err != nil {
				t.Fatalf("close current store: %v", err)
			}
			panic("injected destructive panic")
		})
	}()

	if err := controller.WithShared(context.Background(), func(*Store) error {
		t.Fatal("shared operation ran after exclusive callback panicked")
		return nil
	}); !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("shared operation error = %v, want ErrStoreUnavailable", err)
	}
}


func TestStoreControllerReleasesSharedLeaseWhenOperationPanics(t *testing.T) {
	store := newLifecycleTestStore(t, "state.db")
	controller := NewStoreController(store)

	func() {
		defer func() {
			if recovered := recover(); recovered != "injected shared panic" {
				t.Fatalf("panic = %v, want injected shared panic", recovered)
			}
		}()
		_ = controller.WithShared(context.Background(), func(*Store) error {
			panic("injected shared panic")
		})
	}()

	exclusiveRan := false
	if err := controller.WithExclusive(context.Background(), func(current *Store) (*Store, error) {
		exclusiveRan = true
		if current != store {
			t.Fatal("exclusive operation received unexpected store after shared panic")
		}
		return current, nil
	}); err != nil {
		t.Fatalf("exclusive operation after shared panic: %v", err)
	}
	if !exclusiveRan {
		t.Fatal("exclusive operation was blocked by a leaked shared lease")
	}
	if err := controller.Close(context.Background()); err != nil {
		t.Fatalf("close controller: %v", err)
	}
}

func TestStoreControllerCanceledWriterStopsBlockingNewReaders(t *testing.T) {
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

	writerContext, cancelWriter := context.WithCancel(context.Background())
	writerDone := make(chan error, 1)
	go func() {
		writerDone <- controller.WithExclusive(writerContext, func(current *Store) (*Store, error) {
			return current, nil
		})
	}()
	deadline := time.Now().Add(time.Second)
	for {
		controller.mu.Lock()
		waiting := controller.waitingWriters
		controller.mu.Unlock()
		if waiting > 0 {
			break
		}
		if time.Now().After(deadline) {
			cancelWriter()
			t.Fatal("writer did not register before cancellation")
		}
		time.Sleep(time.Millisecond)
	}

	cancelWriter()
	if err := <-writerDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled writer error = %v, want context.Canceled", err)
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
	case <-time.After(time.Second):
		t.Fatal("canceled writer continued blocking readers while another reader held the store")
	}
	if err := <-secondDone; err != nil {
		t.Fatalf("second shared operation: %v", err)
	}

	close(releaseFirst)
	if err := <-firstDone; err != nil {
		t.Fatalf("first shared operation: %v", err)
	}
	if err := controller.Close(context.Background()); err != nil {
		t.Fatalf("close controller: %v", err)
	}
}

func TestStoreControllerCanceledCloseLeavesStoreAvailable(t *testing.T) {
	store := newLifecycleTestStore(t, "state.db")
	controller := NewStoreController(store)

	sharedEntered := make(chan struct{})
	releaseShared := make(chan struct{})
	sharedDone := make(chan error, 1)
	go func() {
		sharedDone <- controller.WithShared(context.Background(), func(*Store) error {
			close(sharedEntered)
			<-releaseShared
			return nil
		})
	}()
	<-sharedEntered

	closeContext, cancelClose := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancelClose()
	if err := controller.Close(closeContext); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("canceled close error = %v, want context deadline exceeded", err)
	}

	// A canceled shutdown must relinquish writer preference and leave the
	// still-open Store published; otherwise the application can deadlock or
	// become unavailable after a user cancels shutdown.
	secondReaderEntered := make(chan struct{})
	secondReaderDone := make(chan error, 1)
	go func() {
		secondReaderDone <- controller.WithShared(context.Background(), func(got *Store) error {
			if got != store {
				t.Error("controller published an unexpected Store after canceled close")
			}
			close(secondReaderEntered)
			return nil
		})
	}()
	select {
	case <-secondReaderEntered:
	case <-time.After(time.Second):
		t.Fatal("canceled close continued blocking shared operations")
	}
	if err := <-secondReaderDone; err != nil {
		t.Fatalf("shared operation after canceled close: %v", err)
	}

	close(releaseShared)
	if err := <-sharedDone; err != nil {
		t.Fatalf("original shared operation: %v", err)
	}
	if err := controller.Close(context.Background()); err != nil {
		t.Fatalf("close controller after canceled close: %v", err)
	}
}

func TestStoreControllerConcurrentLeaseStress(t *testing.T) {
	store := newLifecycleTestStore(t, "state.db")
	controller := NewStoreController(store)

	var activeReaders atomic.Int32
	var activeWriters atomic.Int32
	var violations atomic.Int32
	var workers sync.WaitGroup

	for reader := 0; reader < 12; reader++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := 0; i < 80; i++ {
				err := controller.WithShared(context.Background(), func(got *Store) error {
					if got != store {
						violations.Add(1)
					}
					if activeWriters.Load() != 0 {
						violations.Add(1)
					}
					activeReaders.Add(1)
					runtime.Gosched()
					activeReaders.Add(-1)
					return nil
				})
				if err != nil {
					t.Errorf("shared stress operation: %v", err)
					return
				}
			}
		}()
	}

	for writer := 0; writer < 4; writer++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := 0; i < 40; i++ {
				err := controller.WithExclusive(context.Background(), func(got *Store) (*Store, error) {
					if got != store {
						violations.Add(1)
					}
					writerCount := activeWriters.Add(1)
					if activeReaders.Load() != 0 || writerCount != 1 {
						violations.Add(1)
					}
					runtime.Gosched()
					activeWriters.Add(-1)
					return got, nil
				})
				if err != nil {
					t.Errorf("exclusive stress operation: %v", err)
					return
				}
			}
		}()
	}

	workers.Wait()
	if got := violations.Load(); got != 0 {
		t.Fatalf("detected %d overlapping or stale-store lease violations", got)
	}
	if readers, writers := activeReaders.Load(), activeWriters.Load(); readers != 0 || writers != 0 {
		t.Fatalf("leases remained active after stress test: readers=%d writers=%d", readers, writers)
	}
	if err := controller.Close(context.Background()); err != nil {
		t.Fatalf("close controller: %v", err)
	}
}
