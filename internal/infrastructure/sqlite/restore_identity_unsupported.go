//go:build !linux && !windows

package sqlite

import (
	"fmt"
	"os"
)


// Unsupported platforms fail closed in openRestoreArtifactFile, so this
// metadata comparison is not reachable during a successful inspection.
func restoreArtifactMetadataUnchanged(before, after os.FileInfo) (bool, error) {
	return true, nil
}

func openRestoreArtifactFile(stateDirectory, path string) (*os.File, error) {
	return nil, fmt.Errorf("%w: platform-specific safe artifact inspection is not implemented", ErrRestoreArtifactInvalid)
}

// normalizeWindowsRestorePath is unreachable on unsupported platforms.
func normalizeWindowsRestorePath(path string) (string, error) { return path, nil }
