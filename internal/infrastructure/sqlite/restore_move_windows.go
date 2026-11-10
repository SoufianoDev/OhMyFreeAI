//go:build windows

package sqlite

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func moveRestoreFileNoReplacePlatform(source, destination string) error {
	directoryHandle, err := openRestoreDirectoryNoReparse(filepath.Dir(destination))
	if err != nil {
		return fmt.Errorf("open SQLite restore destination directory without reparse traversal: %w", err)
	}
	defer windows.CloseHandle(directoryHandle)

	sourceHandle, err := openRestoreSourceWriteThrough(directoryHandle, filepath.Base(source))
	if err != nil {
		return fmt.Errorf("open SQLite restore source with write-through semantics: %w", err)
	}
	defer windows.CloseHandle(sourceHandle)

	return renameRestoreHandleNoReplace(sourceHandle, directoryHandle, filepath.Base(destination))
}

func openRestoreDirectoryNoReparse(path string) (windows.Handle, error) {
	if strings.TrimSpace(path) == "" {
		return windows.InvalidHandle, fmt.Errorf("SQLite restore directory path is empty")
	}
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("resolve SQLite restore directory: %w", err)
	}
	volume := filepath.VolumeName(absolute)
	if len(volume) != 2 || volume[1] != ':' || len(absolute) < 3 || absolute[2] != filepath.Separator {
		return windows.InvalidHandle, fmt.Errorf("SQLite restore directory must be an absolute local drive path")
	}

	rootPath := volume + string(filepath.Separator)
	rootPointer, err := windows.UTF16PtrFromString(rootPath)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode SQLite restore volume root: %w", err)
	}
	current, err := windows.CreateFile(
		rootPointer,
		windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE|windows.SYNCHRONIZE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("open SQLite restore volume root: %w", err)
	}
	if err := verifyRestoreDirectoryHandle(current); err != nil {
		_ = windows.CloseHandle(current)
		return windows.InvalidHandle, fmt.Errorf("validate SQLite restore volume root: %w", err)
	}
	if err := verifyRestoreDirectoryACLWithMask(current, restoreVolumeRootNamespaceMask); err != nil {
		_ = windows.CloseHandle(current)
		return windows.InvalidHandle, fmt.Errorf("validate SQLite restore volume-root ACL: %w", err)
	}

	remainder := strings.TrimLeft(absolute[len(volume):], string(filepath.Separator))
	if remainder == "" {
		_ = windows.CloseHandle(current)
		return windows.InvalidHandle, fmt.Errorf("SQLite restore state directory must not be a volume root")
	}
	for _, component := range strings.Split(remainder, string(filepath.Separator)) {
		next, err := openRestoreDirectoryRelative(current, component)
		if err != nil {
			_ = windows.CloseHandle(current)
			return windows.InvalidHandle, fmt.Errorf("open SQLite restore path component %q: %w", component, err)
		}
		if err := windows.CloseHandle(current); err != nil {
			_ = windows.CloseHandle(next)
			return windows.InvalidHandle, fmt.Errorf("close previous SQLite restore directory handle: %w", err)
		}
		current = next
	}
	if err := verifyRestoreDirectorySecurity(current); err != nil {
		_ = windows.CloseHandle(current)
		return windows.InvalidHandle, fmt.Errorf("validate private SQLite restore state directory: %w", err)
	}
	if err := verifyRestoreVolume(current, rootPath); err != nil {
		_ = windows.CloseHandle(current)
		return windows.InvalidHandle, fmt.Errorf("qualify SQLite restore state volume: %w", err)
	}
	return current, nil
}

func openRestoreDirectoryRelative(directory windows.Handle, leaf string) (windows.Handle, error) {
	if directory == windows.InvalidHandle {
		return windows.InvalidHandle, fmt.Errorf("SQLite restore parent directory handle is invalid")
	}
	if err := validateRestoreMoveLeaf(leaf); err != nil {
		return windows.InvalidHandle, err
	}
	objectName, err := windows.NewNTUnicodeString(leaf)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode SQLite restore directory component: %w", err)
	}
	objectAttributes := windows.OBJECT_ATTRIBUTES{
		RootDirectory: directory,
		ObjectName:    objectName,
		Attributes:    windows.OBJ_CASE_INSENSITIVE,
	}
	objectAttributes.Length = uint32(unsafe.Sizeof(objectAttributes))

	var ioStatus windows.IO_STATUS_BLOCK
	var child windows.Handle
	err = windows.NtCreateFile(
		&child,
		windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE|windows.SYNCHRONIZE,
		&objectAttributes,
		&ioStatus,
		nil,
		windows.FILE_ATTRIBUTE_DIRECTORY,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		windows.FILE_OPEN,
		windows.FILE_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT,
		0,
		0,
	)
	runtime.KeepAlive(directory)
	runtime.KeepAlive(objectName)
	runtime.KeepAlive(&objectAttributes)
	runtime.KeepAlive(&ioStatus)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("NtCreateFile relative to SQLite restore parent directory: %w", err)
	}
	if err := verifyRestoreDirectoryHandle(child); err != nil {
		_ = windows.CloseHandle(child)
		return windows.InvalidHandle, err
	}
	if err := verifyRestoreDirectoryACL(child); err != nil {
		_ = windows.CloseHandle(child)
		return windows.InvalidHandle, fmt.Errorf("validate SQLite restore ancestor ACL: %w", err)
	}
	return child, nil
}

func verifyRestoreDirectoryHandle(directory windows.Handle) error {
	if directory == windows.InvalidHandle {
		return fmt.Errorf("SQLite restore directory handle is invalid")
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(directory, &info); err != nil {
		return fmt.Errorf("inspect opened SQLite restore directory: %w", err)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 ||
		info.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DEVICE) != 0 {
		return fmt.Errorf("opened SQLite restore path component is not a regular non-reparse directory")
	}
	return nil
}

func restoreWindowsPathFromHandle(handle windows.Handle) (string, error) {
	const maxPathUnits = 32768
	buffer := make([]uint16, maxPathUnits)
	written, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], uint32(len(buffer)), 0)
	if err != nil {
		return "", fmt.Errorf("resolve path from opened SQLite restore handle: %w", err)
	}
	if written == 0 || written >= uint32(len(buffer)) {
		return "", fmt.Errorf("resolved SQLite restore path is empty or exceeds the supported path length")
	}
	return filepath.Clean(windows.UTF16ToString(buffer[:written])), nil
}

func openRestoreSourceWriteThrough(directory windows.Handle, sourceLeaf string) (windows.Handle, error) {
	if err := validateRestoreMoveLeaf(sourceLeaf); err != nil {
		return windows.InvalidHandle, err
	}
	objectName, err := windows.NewNTUnicodeString(sourceLeaf)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode SQLite restore source leaf: %w", err)
	}
	objectAttributes := windows.OBJECT_ATTRIBUTES{
		RootDirectory: directory,
		ObjectName:    objectName,
		Attributes:    windows.OBJ_CASE_INSENSITIVE,
	}
	objectAttributes.Length = uint32(unsafe.Sizeof(objectAttributes))

	// NtCreateFile uses the native FILE_WRITE_THROUGH CreateOptions bit (0x2).
	// This is distinct from Win32 FILE_FLAG_WRITE_THROUGH (0x80000000). Opening
	// relative to RootDirectory binds the source lookup to the pinned directory
	// object instead of resolving a path reconstructed from its current name.
	const fileWriteThroughCreateOption = 0x00000002
	createOptions := uint32(windows.FILE_OPEN_REPARSE_POINT |
		windows.FILE_NON_DIRECTORY_FILE |
		windows.FILE_SYNCHRONOUS_IO_NONALERT |
		fileWriteThroughCreateOption)
	var ioStatus windows.IO_STATUS_BLOCK
	var source windows.Handle
	err = windows.NtCreateFile(
		&source,
		windows.GENERIC_WRITE|windows.DELETE|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE,
		&objectAttributes,
		&ioStatus,
		nil,
		windows.FILE_ATTRIBUTE_NORMAL,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		windows.FILE_OPEN,
		createOptions,
		0,
		0,
	)
	runtime.KeepAlive(directory)
	runtime.KeepAlive(objectName)
	runtime.KeepAlive(&objectAttributes)
	runtime.KeepAlive(&ioStatus)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("open SQLite restore source relative to pinned directory with FILE_WRITE_THROUGH: %w", err)
	}

	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(source, &info); err != nil {
		_ = windows.CloseHandle(source)
		return windows.InvalidHandle, fmt.Errorf("inspect opened SQLite restore source: %w", err)
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DEVICE) != 0 {
		_ = windows.CloseHandle(source)
		return windows.InvalidHandle, fmt.Errorf("opened SQLite restore source is not a regular non-reparse file")
	}
	if err := validateRestoreSourceLocation(source, directory, sourceLeaf); err != nil {
		_ = windows.CloseHandle(source)
		return windows.InvalidHandle, err
	}
	// Flush source content before the namespace operation. The subsequent
	// handle-relative no-replace rename uses this same FILE_WRITE_THROUGH object.
	if err := windows.FlushFileBuffers(source); err != nil {
		_ = windows.CloseHandle(source)
		return windows.InvalidHandle, fmt.Errorf("flush SQLite restore source before rename: %w", err)
	}
	return source, nil
}

// validateRestoreSourceLocation rejects a path-open race where the source was
// resolved through a replacement ancestor rather than the pinned directory.
func validateRestoreSourceLocation(source, directory windows.Handle, sourceLeaf string) error {
	sourcePath, err := restoreWindowsPathFromHandle(source)
	if err != nil {
		return fmt.Errorf("resolve opened SQLite restore source: %w", err)
	}
	directoryPath, err := restoreWindowsPathFromHandle(directory)
	if err != nil {
		return fmt.Errorf("resolve pinned SQLite restore directory: %w", err)
	}
	if !equalRestoreMovePath(filepath.Dir(sourcePath), directoryPath) ||
		!equalRestoreMovePath(filepath.Base(sourcePath), sourceLeaf) {
		return fmt.Errorf("opened SQLite restore source is not the expected leaf in the pinned directory")
	}
	return nil
}

func validateRestoreSourceParent(source, directory windows.Handle) error {
	sourcePath, err := restoreWindowsPathFromHandle(source)
	if err != nil {
		return fmt.Errorf("resolve opened SQLite restore source: %w", err)
	}
	directoryPath, err := restoreWindowsPathFromHandle(directory)
	if err != nil {
		return fmt.Errorf("resolve pinned SQLite restore directory: %w", err)
	}
	if !equalRestoreMovePath(filepath.Dir(sourcePath), directoryPath) {
		return fmt.Errorf("opened SQLite restore source is no longer in the pinned destination directory")
	}
	return nil
}

func validateRestoreMoveLeaf(leaf string) error {
	if leaf == "" || leaf == "." || leaf == ".." || strings.ContainsAny(leaf, `/\\:`) {
		return fmt.Errorf("SQLite restore filename must be a single path component")
	}
	return nil
}

func renameRestoreHandleNoReplace(source, destinationDirectory windows.Handle, destinationLeaf string) error {
	return renameRestoreHandle(source, destinationDirectory, destinationLeaf, false)
}

func renameRestoreHandleReplace(source, destinationDirectory windows.Handle, destinationLeaf string) error {
	return renameRestoreHandle(source, destinationDirectory, destinationLeaf, true)
}

// renameRestoreHandle issues FILE_RENAME_INFORMATION_EX through the exact
// source handle opened with FILE_FLAG_WRITE_THROUGH. RootDirectory anchors the
// destination to the retained directory object; FileName is a single relative
// leaf, so no ancestor path is re-resolved during the rename.
func renameRestoreHandle(source, destinationDirectory windows.Handle, destinationLeaf string, replace bool) error {
	if err := validateRestoreMoveLeaf(destinationLeaf); err != nil {
		return err
	}
	var sourceInfo windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(source, &sourceInfo); err != nil {
		return fmt.Errorf("inspect opened SQLite restore source: %w", err)
	}
	if sourceInfo.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DEVICE) != 0 {
		return fmt.Errorf("opened SQLite restore source is not a regular non-reparse file")
	}
	if err := verifyRestoreDirectoryHandle(destinationDirectory); err != nil {
		return fmt.Errorf("validate pinned SQLite restore destination directory: %w", err)
	}
	if err := validateRestoreSourceParent(source, destinationDirectory); err != nil {
		return err
	}

	encodedName, err := windows.UTF16FromString(destinationLeaf)
	if err != nil {
		return fmt.Errorf("encode SQLite restore destination leaf: %w", err)
	}
	encodedName = encodedName[:len(encodedName)-1]
	nameBytes := len(encodedName) * 2
	if nameBytes == 0 || uint64(nameBytes) > uint64(^uint32(0)) {
		return fmt.Errorf("SQLite restore destination leaf length is invalid")
	}

	const fixedHeaderSize = 20
	const minimumInformationSize = 24
	const fileRenameInformationExClass = 65
	bufferSize := fixedHeaderSize + nameBytes
	if bufferSize < minimumInformationSize {
		bufferSize = minimumInformationSize
	}
	if uint64(bufferSize) > uint64(^uint32(0)) {
		return fmt.Errorf("SQLite restore rename information is too large")
	}
	buffer := make([]byte, bufferSize)
	var renameFlags uint32
	if replace {
		const fileRenameReplaceIfExists = 0x00000001
		const fileRenamePOSIXSemantics = 0x00000002
		renameFlags = fileRenameReplaceIfExists | fileRenamePOSIXSemantics
	}
	binary.LittleEndian.PutUint32(buffer[0:4], renameFlags)
	binary.LittleEndian.PutUint32(buffer[4:8], 0)
	binary.LittleEndian.PutUint64(buffer[8:16], uint64(destinationDirectory))
	binary.LittleEndian.PutUint32(buffer[16:20], uint32(nameBytes))
	for i, unit := range encodedName {
		binary.LittleEndian.PutUint16(buffer[fixedHeaderSize+i*2:fixedHeaderSize+i*2+2], unit)
	}

	var ioStatus windows.IO_STATUS_BLOCK
	err = windows.NtSetInformationFile(source, &ioStatus, &buffer[0], uint32(len(buffer)), fileRenameInformationExClass)
	runtime.KeepAlive(buffer)
	runtime.KeepAlive(source)
	runtime.KeepAlive(destinationDirectory)
	runtime.KeepAlive(&ioStatus)
	if err != nil {
		var status windows.NTStatus
		if errors.As(err, &status) && uint32(status) == 0xC0000035 { // STATUS_OBJECT_NAME_COLLISION
			return fmt.Errorf("%w: %v", os.ErrExist, err)
		}
		return fmt.Errorf("NtSetInformationFile(FileRenameInformationEx): %w", err)
	}
	return nil
}
