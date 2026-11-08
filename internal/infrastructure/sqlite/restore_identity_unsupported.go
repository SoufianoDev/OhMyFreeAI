//go:build !linux && !windows

package sqlite

import (
	"fmt"
	"os"
)

func openRestoreArtifactFile(stateDirectory, path string) (*os.File, error) {
	return nil, fmt.Errorf("%w: platform-specific safe artifact inspection is not implemented", ErrRestoreArtifactInvalid)
}

// normalizeWindowsRestorePath is unreachable on unsupported platforms.
func normalizeWindowsRestorePath(path string) (string, error) { return path, nil }
