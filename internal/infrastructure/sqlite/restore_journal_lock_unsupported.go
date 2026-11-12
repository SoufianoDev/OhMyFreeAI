//go:build !linux && !windows

package sqlite

import (
	"context"
	"fmt"
)

func acquireRestoreJournalLockPlatform(context.Context, string) (func() error, error) {
	return nil, fmt.Errorf("cross-process SQLite Restore journal locking is unsupported on this platform")
}
