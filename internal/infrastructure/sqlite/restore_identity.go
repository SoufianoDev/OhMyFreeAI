package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
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

// verifiedRestoreArtifact owns the exact file handle whose bytes were hashed.
// Callers that will later consume or move the artifact must retain this object
// and use file; reopening a path would discard the identity guarantee.
type verifiedRestoreArtifact struct {
	file     *os.File
	identity RestoreContentIdentity
}

func (artifact *verifiedRestoreArtifact) close() error {
	if artifact == nil || artifact.file == nil {
		return nil
	}
	return artifact.file.Close()
}

func inspectRestoreArtifact(stateDirectory, path string) (RestoreContentIdentity, error) {
	artifact, err := openVerifiedRestoreArtifact(stateDirectory, path)
	if err != nil {
		return RestoreContentIdentity{}, err
	}
	defer artifact.close()
	return artifact.identity, nil
}

func openVerifiedRestoreArtifact(stateDirectory, path string) (*verifiedRestoreArtifact, error) {
	file, err := openRestoreArtifactFile(stateDirectory, path)
	if err != nil {
		return nil, fmt.Errorf("%w: open artifact: %v", ErrRestoreArtifactInvalid, err)
	}
	fail := func(err error) (*verifiedRestoreArtifact, error) {
		_ = file.Close()
		return nil, err
	}

	before, err := file.Stat()
	if err != nil {
		return fail(fmt.Errorf("%w: stat opened artifact: %v", ErrRestoreArtifactInvalid, err))
	}
	if !before.Mode().IsRegular() || before.Size() <= 0 {
		return fail(fmt.Errorf("%w: artifact must be a non-empty regular file", ErrRestoreArtifactInvalid))
	}

	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return fail(fmt.Errorf("%w: hash opened artifact: %v", ErrRestoreArtifactInvalid, err))
	}
	after, err := file.Stat()
	if err != nil {
		return fail(fmt.Errorf("%w: restat opened artifact: %v", ErrRestoreArtifactInvalid, err))
	}
	metadataStable, metadataErr := restoreArtifactMetadataUnchanged(before, after)
	if metadataErr != nil {
		return fail(fmt.Errorf("%w: compare artifact metadata: %v", ErrRestoreArtifactInvalid, metadataErr))
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fail(fmt.Errorf("%w: rewind verified artifact: %v", ErrRestoreArtifactInvalid, err))
	}
	if !after.Mode().IsRegular() || size <= 0 || uint64(size) != uint64(after.Size()) ||
		before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) || !metadataStable {
		return fail(fmt.Errorf("%w: artifact changed while its identity was being read", ErrRestoreArtifactInvalid))
	}

	return &verifiedRestoreArtifact{
		file: file,
		identity: RestoreContentIdentity{
			SHA256: hex.EncodeToString(hash.Sum(nil)),
			Size:   uint64(size),
		},
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
