package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
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

	// Resolve only the parent directory. This normalizes Windows path casing
	// (and equivalent directory aliases) to the same spelling used by
	// canonicalStateDirectory, while deliberately leaving the final leaf
	// unresolved so a symlink/reparse-point leaf is still rejected by the
	// platform-specific, handle-based open.
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return "", "", fmt.Errorf("%w: resolve restore artifact parent: %v", ErrRestoreArtifactInvalid, err)
	}
	candidate := filepath.Join(parent, filepath.Base(absolute))

	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == "." || !filepath.IsLocal(relative) {
		return "", "", fmt.Errorf("%w: artifact path is outside the application state directory", ErrRestoreArtifactInvalid)
	}
	return root, relative, nil
}
