package sqlite

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

const restoreJournalLockFileName = ".restore-journal.lock"

var errRestoreJournalLeaseClosed = errors.New("SQLite restore journal lease is closed")

// restoreJournalLease serializes journal access across processes. The executor
// must hold one lease for the whole transaction and use its Read/Write methods
// for every phase update rather than reacquiring the lock.
type restoreJournalLease struct {
	root    string
	release func() error
	mu      sync.Mutex
}

func acquireRestoreJournalLease(ctx context.Context, stateDirectory string) (*restoreJournalLease, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return nil, err
	}
	release, err := acquireRestoreJournalLockPlatform(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("acquire cross-process SQLite restore journal lock: %w", err)
	}
	if err := ctx.Err(); err != nil {
		_ = release()
		return nil, err
	}
	return &restoreJournalLease{root: root, release: release}, nil
}

func (lease *restoreJournalLease) Read() (RestoreJournal, error) {
	var empty RestoreJournal
	if lease == nil {
		return empty, errRestoreJournalLeaseClosed
	}
	lease.mu.Lock()
	defer lease.mu.Unlock()
	if lease.release == nil {
		return empty, errRestoreJournalLeaseClosed
	}
	return readRestoreJournalLocked(lease.root)
}

func (lease *restoreJournalLease) Write(journal RestoreJournal) error {
	if lease == nil {
		return errRestoreJournalLeaseClosed
	}
	lease.mu.Lock()
	defer lease.mu.Unlock()
	if lease.release == nil {
		return errRestoreJournalLeaseClosed
	}
	return writeRestoreJournalLocked(lease.root, journal)
}

func (lease *restoreJournalLease) Close() error {
	if lease == nil {
		return nil
	}
	lease.mu.Lock()
	defer lease.mu.Unlock()
	if lease.release == nil {
		return nil
	}
	release := lease.release
	lease.release = nil
	return release()
}
