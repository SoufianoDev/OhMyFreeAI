//go:build !linux && !windows

package sqlite

import "errors"

var ErrRestoreMoveUnsupported = errors.New("atomic no-replace SQLite restore moves are unsupported on this platform")

func moveRestoreFileNoReplacePlatform(string, string) error {
	return ErrRestoreMoveUnsupported
}

func validateRestoreStateDirectoryPlatform(string) error {
	return ErrRestoreMoveUnsupported
}
