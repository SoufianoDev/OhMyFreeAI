package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/SoufianoDev/OhMyFreeAI/internal/infrastructure/sqlite"
)

const (
	ArchiveExtension = ".omfaibackup"
	archiveFormat    = "OhMyFreeAI"
	archiveVersion   = 1
	manifestName     = "manifest.json"
	databaseName     = "state.sqlite"

	maxManifestBytes = 64 << 10
	maxDatabaseBytes = 1 << 30
	maxArchiveBytes  = maxDatabaseBytes + (16 << 20)
)

var (
	ErrInvalidBackupPath    = errors.New("invalid OhMyFreeAI backup path")
	ErrInvalidBackupArchive = errors.New("invalid OhMyFreeAI backup archive")
	ErrBackupArchiveExists  = errors.New("OhMyFreeAI backup archive already exists")
	ErrBackupManagerConfig  = errors.New("backup manager is not configured")
)

type FileRecord struct {
	Path   string `json:"path"`
	Size   uint64 `json:"size_bytes"`
	SHA256 string `json:"sha256"`
}

type Manifest struct {
	Format             string       `json:"format"`
	FormatVersion      int          `json:"format_version"`
	ApplicationVersion string       `json:"application_version"`
	SchemaVersion      int          `json:"schema_version"`
	CreatedAt          time.Time    `json:"created_at"`
	Files              []FileRecord `json:"files"`
}

// Manager packages verified SQLite snapshots into the user-selected
// .omfaibackup format. The archive contains only the manifest and SQLite
// metadata database; it never bundles model weights, engine binaries or
// secret values.
type Manager struct {
	ApplicationVersion string
	Migrations         fs.FS
	MigrationDir       string
}

func (m Manager) Create(ctx context.Context, store *sqlite.Store, destinationPath string) error {
	return m.create(ctx, store, destinationPath, nil)
}

func (m Manager) create(ctx context.Context, store *sqlite.Store, destinationPath string, beforePublish func()) error {
	if store == nil || strings.TrimSpace(m.ApplicationVersion) == "" || m.Migrations == nil {
		return ErrBackupManagerConfig
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateArchiveExtension(destinationPath); err != nil {
		return err
	}
	absoluteDestination, err := filepath.Abs(destinationPath)
	if err != nil {
		return fmt.Errorf("resolve backup archive destination: %w", err)
	}
	absoluteDestination = filepath.Clean(absoluteDestination)
	if sameArchivePath(absoluteDestination, store.DatabasePath()) {
		return ErrInvalidBackupPath
	}
	directory := filepath.Dir(absoluteDestination)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create backup archive directory: %w", err)
	}
	if _, err := os.Lstat(absoluteDestination); err == nil {
		return ErrBackupArchiveExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect backup archive destination: %w", err)
	}

	stagingDirectory, err := os.MkdirTemp(directory, ".ohmyfreeai-backup-stage-*")
	if err != nil {
		return fmt.Errorf("create backup staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDirectory)

	snapshotPath := filepath.Join(stagingDirectory, databaseName)
	if err := store.BackupTo(ctx, snapshotPath); err != nil {
		return fmt.Errorf("create SQLite snapshot: %w", err)
	}
	schemaVersion, err := sqlite.SnapshotSchemaVersion(ctx, snapshotPath)
	if err != nil {
		return fmt.Errorf("verify SQLite snapshot schema: %w", err)
	}
	databaseInfo, err := os.Stat(snapshotPath)
	if err != nil {
		return fmt.Errorf("stat SQLite snapshot: %w", err)
	}
	if databaseInfo.Size() <= 0 || uint64(databaseInfo.Size()) > maxDatabaseBytes {
		return fmt.Errorf("%w: SQLite snapshot size exceeds the archive limit", ErrInvalidBackupArchive)
	}
	digest, size, err := hashFile(ctx, snapshotPath)
	if err != nil {
		return fmt.Errorf("hash SQLite snapshot: %w", err)
	}
	if size == 0 || size > maxDatabaseBytes {
		return fmt.Errorf("%w: hashed SQLite snapshot size exceeds the archive limit", ErrInvalidBackupArchive)
	}

	manifest := Manifest{
		Format:             archiveFormat,
		FormatVersion:      archiveVersion,
		ApplicationVersion: m.ApplicationVersion,
		SchemaVersion:      schemaVersion,
		CreatedAt:          time.Now().UTC(),
		Files: []FileRecord{{
			Path:   databaseName,
			Size:   size,
			SHA256: digest,
		}},
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode backup manifest: %w", err)
	}
	if len(manifestBytes) > maxManifestBytes {
		return fmt.Errorf("%w: manifest exceeds size limit", ErrInvalidBackupArchive)
	}
	manifestPath := filepath.Join(stagingDirectory, manifestName)
	if err := os.WriteFile(manifestPath, manifestBytes, 0o600); err != nil {
		return fmt.Errorf("write backup manifest: %w", err)
	}

	archiveFile, err := os.CreateTemp(directory, ".ohmyfreeai-backup-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary backup archive: %w", err)
	}
	temporaryArchivePath := archiveFile.Name()
	published := false
	defer func() {
		if !published {
			_ = os.Remove(temporaryArchivePath)
		}
	}()

	zipWriter := zip.NewWriter(archiveFile)
	if err := addArchiveFile(ctx, zipWriter, manifestName, manifestPath); err != nil {
		_ = zipWriter.Close()
		_ = archiveFile.Close()
		return err
	}
	if err := addArchiveFile(ctx, zipWriter, databaseName, snapshotPath); err != nil {
		_ = zipWriter.Close()
		_ = archiveFile.Close()
		return err
	}
	if err := zipWriter.Close(); err != nil {
		_ = archiveFile.Close()
		return fmt.Errorf("finalize backup archive: %w", err)
	}
	if err := archiveFile.Sync(); err != nil {
		_ = archiveFile.Close()
		return fmt.Errorf("sync backup archive: %w", err)
	}
	if err := archiveFile.Close(); err != nil {
		return fmt.Errorf("close backup archive: %w", err)
	}

	verifiedSnapshot := filepath.Join(stagingDirectory, "verified.sqlite")
	if _, err := extractArchive(ctx, temporaryArchivePath, verifiedSnapshot); err != nil {
		return fmt.Errorf("verify completed backup archive: %w", err)
	}
	_ = os.Remove(verifiedSnapshot)
	if beforePublish != nil {
		beforePublish()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, err := os.Lstat(absoluteDestination); err == nil {
		return ErrBackupArchiveExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("recheck backup archive destination: %w", err)
	}
	if err := os.Rename(temporaryArchivePath, absoluteDestination); err != nil {
		return fmt.Errorf("publish backup archive: %w", err)
	}
	published = true
	return nil
}

// CreateWithStoreController holds a shared lease until the complete archive is published.
func (m Manager) CreateWithStoreController(ctx context.Context, controller *sqlite.StoreController, destinationPath string) error {
	return m.createWithStoreController(ctx, controller, destinationPath, nil)
}

func (m Manager) createWithStoreController(ctx context.Context, controller *sqlite.StoreController, destinationPath string, beforePublish func()) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return controller.WithShared(ctx, func(store *sqlite.Store) error {
		return m.create(ctx, store, destinationPath, beforePublish)
	})
}

// Inspect validates the complete archive and returns its manifest without
// changing the live application database. Extraction is isolated in a
// temporary directory that is removed before the method returns.
func (m Manager) Inspect(ctx context.Context, archivePath string) (Manifest, error) {
	var empty Manifest
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	if err := validateArchiveExtension(archivePath); err != nil {
		return empty, err
	}
	absoluteArchive, err := filepath.Abs(archivePath)
	if err != nil {
		return empty, fmt.Errorf("resolve backup archive: %w", err)
	}
	stagingDirectory, err := os.MkdirTemp("", ".ohmyfreeai-backup-inspect-*")
	if err != nil {
		return empty, fmt.Errorf("create backup inspection directory: %w", err)
	}
	defer os.RemoveAll(stagingDirectory)

	snapshotPath := filepath.Join(stagingDirectory, databaseName)
	manifest, err := extractArchive(ctx, absoluteArchive, snapshotPath)
	if err != nil {
		return empty, err
	}
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	return manifest, nil
}

func (m Manager) Restore(ctx context.Context, archivePath, targetDatabasePath string) (string, error) {
	if m.Migrations == nil {
		return "", ErrBackupManagerConfig
	}
	if err := validateArchiveExtension(archivePath); err != nil {
		return "", err
	}
	absoluteArchive, err := filepath.Abs(archivePath)
	if err != nil {
		return "", fmt.Errorf("resolve backup archive: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	absoluteTarget, err := filepath.Abs(targetDatabasePath)
	if err != nil {
		return "", fmt.Errorf("resolve restore database target: %w", err)
	}
	if sameArchivePath(absoluteArchive, absoluteTarget) {
		return "", ErrInvalidBackupPath
	}
	targetDirectory := filepath.Dir(absoluteTarget)
	if err := os.MkdirAll(targetDirectory, 0o700); err != nil {
		return "", fmt.Errorf("create restore staging directory: %w", err)
	}
	stagingDirectory, err := os.MkdirTemp(targetDirectory, ".ohmyfreeai-archive-restore-*")
	if err != nil {
		return "", fmt.Errorf("create restore staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDirectory)

	snapshotPath := filepath.Join(stagingDirectory, databaseName)
	if _, err := extractArchive(ctx, absoluteArchive, snapshotPath); err != nil {
		return "", err
	}
	rollbackPath, err := sqlite.RestoreDatabase(ctx, snapshotPath, absoluteTarget, m.Migrations, m.MigrationDir)
	if err != nil {
		return "", fmt.Errorf("restore OhMyFreeAI backup: %w", err)
	}
	return rollbackPath, nil
}

func extractArchive(ctx context.Context, archivePath, snapshotPath string) (Manifest, error) {
	var manifest Manifest
	if err := ctx.Err(); err != nil {
		return manifest, err
	}
	archiveInfo, err := os.Lstat(archivePath)
	if err != nil {
		return manifest, fmt.Errorf("%w: inspect archive: %v", ErrInvalidBackupArchive, err)
	}
	if !archiveInfo.Mode().IsRegular() || archiveInfo.Size() <= 0 || archiveInfo.Size() > maxArchiveBytes {
		return manifest, fmt.Errorf("%w: archive is not a regular file or exceeds the size limit", ErrInvalidBackupArchive)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return manifest, fmt.Errorf("%w: open ZIP container: %v", ErrInvalidBackupArchive, err)
	}
	defer reader.Close()
	if len(reader.File) != 2 {
		return manifest, fmt.Errorf("%w: expected exactly two entries", ErrInvalidBackupArchive)
	}

	entries := make(map[string]*zip.File, 2)
	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() || !entry.Mode().IsRegular() || entry.Mode()&os.ModeSymlink != 0 {
			return manifest, fmt.Errorf("%w: only regular files are allowed", ErrInvalidBackupArchive)
		}
		if entry.Flags&1 != 0 {
			return manifest, fmt.Errorf("%w: encrypted entries are not supported", ErrInvalidBackupArchive)
		}
		if entry.Method != zip.Store && entry.Method != zip.Deflate {
			return manifest, fmt.Errorf("%w: unsupported compression method", ErrInvalidBackupArchive)
		}
		if entry.Name != manifestName && entry.Name != databaseName {
			return manifest, fmt.Errorf("%w: unexpected archive entry %q", ErrInvalidBackupArchive, entry.Name)
		}
		if _, exists := entries[entry.Name]; exists {
			return manifest, fmt.Errorf("%w: duplicate archive entry %q", ErrInvalidBackupArchive, entry.Name)
		}
		entries[entry.Name] = entry
	}
	manifestEntry, hasManifest := entries[manifestName]
	databaseEntry, hasDatabase := entries[databaseName]
	if !hasManifest || !hasDatabase {
		return manifest, fmt.Errorf("%w: manifest or database entry is missing", ErrInvalidBackupArchive)
	}

	manifestBytes, err := readArchiveEntry(manifestEntry, maxManifestBytes)
	if err != nil {
		return manifest, err
	}
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, fmt.Errorf("%w: decode manifest: %v", ErrInvalidBackupArchive, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return manifest, fmt.Errorf("%w: manifest contains trailing data", ErrInvalidBackupArchive)
	}
	if err := validateManifest(manifest); err != nil {
		return manifest, err
	}
	fileRecord := manifest.Files[0]
	if databaseEntry.UncompressedSize64 != fileRecord.Size || databaseEntry.UncompressedSize64 > maxDatabaseBytes {
		return manifest, fmt.Errorf("%w: database size does not match manifest or exceeds limit", ErrInvalidBackupArchive)
	}

	entryReader, err := databaseEntry.Open()
	if err != nil {
		return manifest, fmt.Errorf("%w: open database entry: %v", ErrInvalidBackupArchive, err)
	}
	defer entryReader.Close()
	output, err := os.OpenFile(snapshotPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return manifest, fmt.Errorf("create extracted SQLite snapshot: %w", err)
	}
	outputClosed := false
	defer func() {
		if !outputClosed {
			_ = output.Close()
		}
	}()

	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(output, hash), io.LimitReader(contextReader{ctx: ctx, reader: entryReader}, maxDatabaseBytes+1))
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return manifest, contextErr
		}
		return manifest, fmt.Errorf("%w: extract database entry: %v", ErrInvalidBackupArchive, err)
	}
	if written <= 0 || uint64(written) != fileRecord.Size || uint64(written) > maxDatabaseBytes {
		return manifest, fmt.Errorf("%w: extracted database size mismatch", ErrInvalidBackupArchive)
	}
	if hex.EncodeToString(hash.Sum(nil)) != fileRecord.SHA256 {
		return manifest, fmt.Errorf("%w: database SHA-256 mismatch", ErrInvalidBackupArchive)
	}
	if err := output.Sync(); err != nil {
		return manifest, fmt.Errorf("sync extracted SQLite snapshot: %w", err)
	}
	if err := output.Close(); err != nil {
		return manifest, fmt.Errorf("close extracted SQLite snapshot: %w", err)
	}
	outputClosed = true

	schemaVersion, err := sqlite.SnapshotSchemaVersion(ctx, snapshotPath)
	if err != nil {
		return manifest, fmt.Errorf("%w: verify extracted SQLite snapshot: %v", ErrInvalidBackupArchive, err)
	}
	if schemaVersion != manifest.SchemaVersion {
		return manifest, fmt.Errorf("%w: manifest schema version %d does not match database version %d", ErrInvalidBackupArchive, manifest.SchemaVersion, schemaVersion)
	}
	return manifest, nil
}

func validateManifest(manifest Manifest) error {
	if manifest.Format != archiveFormat || manifest.FormatVersion != archiveVersion {
		return fmt.Errorf("%w: unsupported backup format/version", ErrInvalidBackupArchive)
	}
	if strings.TrimSpace(manifest.ApplicationVersion) == "" || manifest.SchemaVersion < 1 || manifest.CreatedAt.IsZero() {
		return fmt.Errorf("%w: required manifest metadata is missing", ErrInvalidBackupArchive)
	}
	if len(manifest.Files) != 1 || manifest.Files[0].Path != databaseName || manifest.Files[0].Size == 0 || manifest.Files[0].Size > maxDatabaseBytes {
		return fmt.Errorf("%w: manifest file list is invalid", ErrInvalidBackupArchive)
	}
	checksum, err := hex.DecodeString(manifest.Files[0].SHA256)
	if err != nil || len(checksum) != sha256.Size {
		return fmt.Errorf("%w: manifest SHA-256 is invalid", ErrInvalidBackupArchive)
	}
	return nil
}

func readArchiveEntry(entry *zip.File, limit uint64) ([]byte, error) {
	if entry.UncompressedSize64 > limit {
		return nil, fmt.Errorf("%w: manifest exceeds size limit", ErrInvalidBackupArchive)
	}
	reader, err := entry.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: open manifest entry: %v", ErrInvalidBackupArchive, err)
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read manifest entry: %v", ErrInvalidBackupArchive, err)
	}
	if uint64(len(data)) != entry.UncompressedSize64 || uint64(len(data)) > limit {
		return nil, fmt.Errorf("%w: manifest size mismatch", ErrInvalidBackupArchive)
	}
	return data, nil
}

func addArchiveFile(ctx context.Context, writer *zip.Writer, archiveName, sourcePath string) error {
	header := &zip.FileHeader{Name: archiveName, Method: zip.Deflate}
	header.SetMode(0o600)
	header.Modified = time.Now().UTC()
	entryWriter, err := writer.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("create archive entry %s: %w", archiveName, err)
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open archive source %s: %w", archiveName, err)
	}
	defer source.Close()
	if _, err := io.Copy(entryWriter, contextReader{ctx: ctx, reader: source}); err != nil {
		return fmt.Errorf("write archive entry %s: %w", archiveName, err)
	}
	return nil
}

func hashFile(ctx context.Context, path string) (string, uint64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, contextReader{ctx: ctx, reader: file})
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hash.Sum(nil)), uint64(size), nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(buffer []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(buffer)
	}
}

func validateArchiveExtension(path string) error {
	if strings.TrimSpace(path) == "" || !strings.EqualFold(filepath.Ext(path), ArchiveExtension) {
		return ErrInvalidBackupPath
	}
	return nil
}

func sameArchivePath(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
