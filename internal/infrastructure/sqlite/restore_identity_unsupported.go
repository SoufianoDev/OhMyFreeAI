//go:build !linux && !windows

package sqlite

import (
	"fmt"
	"os"
)


// restoreArtifactMetadataUnchanged relies on the shared size and modification
// time checks on Windows. Windows handle sharing prevents concurrent writers
// from opening a conflicting write handle during inspection.
func restoreArtifactMetadataUnchanged(before, after os.FileInfo) (bool, error) {
	return true, nil
}

func openRestoreArtifactFile(stateDirectory, path string) (*os.File, error) {
	return nil, fmt.Errorf("%w: platform-specific safe artifact inspection is not implemented", ErrRestoreArtifactInvalid)
}

// normalizeWindowsRestorePath is unreachable on unsupported platforms.
func normalizeWindowsRestorePath(path string) (string, error) { return path, nil }
