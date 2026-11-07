package sqlite

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrStoreControllerRequired = errors.New("SQLite store controller requires a store")
	ErrStoreUnavailable        = errors.New("SQLite store is unavailable")
)

// StoreController is the sole owner of the active Store. Shared operations
// drain before an exclusive operation begins; waiting writers prevent reader
// starvation. Callers must not retain the Store passed to an operation.
type StoreController struct {
	mu            sync.Mutex
	store         *Store
	readers       int
	writer        bool
	waitingWriters int
	changed       chan struct{}
}

func NewStoreControllerChecked(store *Store) (*StoreController, error) {
	if store == nil {
		return nil, ErrStoreControllerRequired
	}
	return &StoreController{store: store, changed: make(chan struct{})}, nil
}

func NewStoreController(store *Store) *StoreController {
	controller, err := NewStoreControllerChecked(store)
	if err != nil {
		panic(err)
	}
	return controller
}

func (c *StoreController) signalLocked() {
	close(c.changed)
	c.changed = make(chan struct{})
}

func (c *StoreController) acquireShared(ctx context.Context) (*Store, func(), error) {
	if c == nil {
		return nil, nil, ErrStoreUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		c.mu.Lock()
		if c.store == nil {
			c.mu.Unlock()
			return nil, nil, ErrStoreUnavailable
		}
		if !c.writer && c.waitingWriters == 0 {
			store := c.store
			c.readers++
			c.mu.Unlock()
			var once sync.Once
			release := func() {
				once.Do(func() {
					c.mu.Lock()
					c.readers--
					c.signalLocked()
					c.mu.Unlock()
				})
			}
			return store, release, nil
		}
		changed := c.changed
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-changed:
		}
	}
}

func (c *StoreController) acquireExclusive(ctx context.Context) (*Store, func(), error) {
	if c == nil {
		return nil, nil, ErrStoreUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}

	c.mu.Lock()
	c.waitingWriters++
	c.signalLocked()
	c.mu.Unlock()

	waiting := true
	removeWaiting := func() {
		if !waiting {
			return
		}
		c.mu.Lock()
		c.waitingWriters--
		waiting = false
		c.signalLocked()
		c.mu.Unlock()
	}

	for {
		c.mu.Lock()
		if c.store == nil {
			c.waitingWriters--
			waiting = false
			c.signalLocked()
			c.mu.Unlock()
			return nil, nil, ErrStoreUnavailable
		}
		if !c.writer && c.readers == 0 {
			c.waitingWriters--
			waiting = false
			c.writer = true
			store := c.store
			c.mu.Unlock()
			var once sync.Once
			release := func() {
				once.Do(func() {
					c.mu.Lock()
					c.writer = false
					c.signalLocked()
					c.mu.Unlock()
				})
			}
			return store, release, nil
		}
		changed := c.changed
		c.mu.Unlock()

		select {
		case <-ctx.Done():
			removeWaiting()
			return nil, nil, ctx.Err()
		case <-changed:
		}
	}
}

// WithShared runs an operation while holding a shared lease. The Store must
// not escape the callback.
func (c *StoreController) WithShared(ctx context.Context, operation func(*Store) error) error {
	if operation == nil {
		return errors.New("SQLite shared operation is required")
	}
	store, release, err := c.acquireShared(ctx)
	if err != nil {
		return err
	}
	defer release()
	return operation(store)
}

// WithExclusive drains all shared operations and prevents new ones while the
// callback runs. The callback returns the Store that remains active. It must
// return the original Store on a non-destructive operation.
func (c *StoreController) WithExclusive(ctx context.Context, operation func(*Store) (*Store, error)) error {
	if operation == nil {
		return errors.New("SQLite exclusive operation is required")
	}
	store, release, err := c.acquireExclusive(ctx)
	if err != nil {
		return err
	}
	defer release()

	next, err := operation(store)
	if err != nil {
		return err
	}
	if next == nil {
		return ErrStoreUnavailable
	}
	c.mu.Lock()
	c.store = next
	c.signalLocked()
	c.mu.Unlock()
	return nil
}

// Load and Save make StoreController a stable configuration repository. Each
// operation holds its lease for the full SQLite call.
func (c *StoreController) Load(ctx context.Context) (Config, error) {
	var result Config
	err := c.WithShared(ctx, func(store *Store) error {
		var err error
		result, err = store.Load(ctx)
		return err
	})
	return result, err
}

func (c *StoreController) Save(ctx context.Context, cfg Config) error {
	return c.WithShared(ctx, func(store *Store) error {
		return store.Save(ctx, cfg)
	})
}
