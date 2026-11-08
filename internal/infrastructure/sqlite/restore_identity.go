package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
)

var ErrRestoreArtifactInvalid = errors.New("SQLite restore artifact is not a safely inspectable regular file")

// RestoreContentIdentity binds an artifact's content to its SHA-256 digest and
// byte length. The digest is obtained from the opened file object.
type RestoreContentIdentity struct {
	SHA256 string
	Size   uint64
}

func inspectRestoreArtifact(stateDirectory, path string) (RestoreContentIdentity, error) {
	var identity RestoreContentIdentity
	file, err := openRestoreArtifactFile(stateDirectory, path)
	if err != nil {
		return identity, fmt.Errorf("%w: open artifact: %v", ErrRestoreArtifactInvalid, err)
	}
	defer file.Close()

	before, err := file.Stat()
	if err != nil {
		return identity, fmt.Errorf("%w: stat opened artifact: %v", ErrRestoreArtifactInvalid, err)
	}
	if !before.Mode().IsRegular() || before.Size() <= 0 {
		return identity, fmt.Errorf("%w: artifact must be a non-empty regular file", ErrRestoreArtifactInvalid)
	}

	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return identity, fmt.Errorf("%w: hash opened artifact: %v", ErrRestoreArtifactInvalid, err)
	}
	after, err := file.Stat()
	if err != nil {
		return identity, fmt.Errorf("%w: restat opened artifact: %v", ErrRestoreArtifactInvalid, err)
	}
	if !after.Mode().IsRegular() || size <= 0 || uint64(size) != uint64(after.Size()) ||
		before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return identity, fmt.Errorf("%w: artifact changed while its identity was being read", ErrRestoreArtifactInvalid)
	}

	return RestoreContentIdentity{
		SHA256: hex.EncodeToString(hash.Sum(nil)),
		Size:   uint64(size),
	}, nil
}

func restoreArtifactRelativePath(stateDirectory, path string) (string, string, error) {
	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return "", "", err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("resolve SQLite restore artifact path: %w", err)
	}
	absolute = filepath.Clean(absolute)

	var relative string
	if runtime.GOOS == "windows" {
		var ok bool
		relative, ok = windowsRestoreArtifactRelativePath(root, absolute)
		if !ok {
			return "", "", fmt.Errorf("%w: artifact path is outside the application state directory", ErrRestoreArtifactInvalid)
		}
	} else {
		relative, err = filepath.Rel(root, absolute)
		if err != nil || relative == "." || !filepath.IsLocal(relative) {
			return "", "", fmt.Errorf("%w: artifact path is outside the application state directory", ErrRestoreArtifactInvalid)
		}
	}
	if relative == "." || !filepath.IsLocal(relative) {
		return "", "", fmt.Errorf("%w: artifact path is outside the application state directory", ErrRestoreArtifactInvalid)
	}
	return root, relative, nil
}

// Windows paths are case-insensitive. filepath.Rel performs a lexical
// comparison, so compare path components with EqualFold while preserving the
// caller's spelling in the relative path. Do not EvalSymlinks on the artifact
// parent: the platform opener must reject every reparse/symlink component
// relative to the pinned state-directory handle.
func windowsRestoreArtifactRelativePath(root, candidate string) (string, bool) {
	var err error
	root, err = normalizeWindowsRestorePath(root)
	if err != nil {
		return "", false
	}
	candidate, err = normalizeWindowsRestorePath(candidate)
	if err != nil {
		return "", false
	}
	root = filepath.Clean(root)
	candidate = filepath.Clean(candidate)

	rootVolume := filepath.VolumeName(root)
	candidateVolume := filepath.VolumeName(candidate)
	if !strings.EqualFold(rootVolume, candidateVolume) {
		return "", false
	}

	rootComponents := windowsRestorePathComponents(strings.TrimPrefix(root, rootVolume))
	candidateComponents := windowsRestorePathComponents(strings.TrimPrefix(candidate, candidateVolume))
	if len(candidateComponents) <= len(rootComponents) {
		return "", false
	}
	for i, component := range rootComponents {
		if !strings.EqualFold(component, candidateComponents[i]) {
			return "", false
		}
	}
	return filepath.Join(candidateComponents[len(rootComponents):]...), true
}

func windowsRestorePathComponents(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '\\' || r == '/'
	})
}
