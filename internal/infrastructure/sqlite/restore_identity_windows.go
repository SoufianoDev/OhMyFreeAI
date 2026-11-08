//go:build windows

package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// normalizeWindowsRestorePath expands DOS 8.3 aliases and normalizes on-disk
// component spelling without resolving reparse points. The actual open remains
// relative to pinned directory handles and rejects reparse points component by
// component.
func normalizeWindowsRestorePath(path string) (string, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", fmt.Errorf("encode Windows restore path: %w", err)
	}
	required, err := windows.GetLongPathName(pathPtr, nil, 0)
	if err != nil {
		return "", fmt.Errorf("query Windows long restore path: %w", err)
	}
	if required == 0 {
		return "", fmt.Errorf("Windows returned an empty long restore path")
	}
	buffer := make([]uint16, required)
	length, err := windows.GetLongPathName(pathPtr, &buffer[0], uint32(len(buffer)))
	if err != nil {
		return "", fmt.Errorf("normalize Windows restore path: %w", err)
	}
	if length == 0 || length >= uint32(len(buffer)) {
		return "", fmt.Errorf("Windows returned an invalid long restore path length")
	}
	return filepath.Clean(windows.UTF16ToString(buffer[:length])), nil
}

func openRestoreArtifactFile(stateDirectory, path string) (*os.File, error) {
	root, relative, err := restoreArtifactRelativePath(stateDirectory, path)
	if err != nil {
		return nil, err
	}
	components := strings.Split(filepath.Clean(relative), string(filepath.Separator))
	if len(components) == 0 {
		return nil, fmt.Errorf("restore artifact path has no components")
	}
	for _, component := range components {
		if err := validateRestoreMoveLeaf(component); err != nil {
			return nil, fmt.Errorf("invalid restore artifact path component: %w", err)
		}
	}

	rootHandle, err := openRestoreArtifactRoot(root)
	if err != nil {
		return nil, err
	}
	currentDirectory := rootHandle
	ownsCurrentDirectory := false
	closeCurrentDirectory := func() {
		if ownsCurrentDirectory {
			_ = windows.CloseHandle(currentDirectory)
		}
	}

	for _, component := range components[:len(components)-1] {
		next, openErr := openRestoreArtifactDirectoryRelative(currentDirectory, component)
		if openErr != nil {
			closeCurrentDirectory()
			_ = windows.CloseHandle(rootHandle)
			return nil, openErr
		}
		closeCurrentDirectory()
		currentDirectory = next
		ownsCurrentDirectory = true
	}

	fileHandle, err := openRestoreArtifactFileRelative(currentDirectory, components[len(components)-1])
	closeCurrentDirectory()
	_ = windows.CloseHandle(rootHandle)
	if err != nil {
		return nil, err
	}

	file := os.NewFile(uintptr(fileHandle), path)
	if file == nil {
		_ = windows.CloseHandle(fileHandle)
		return nil, fmt.Errorf("wrap opened restore artifact handle")
	}
	return file, nil
}

func openRestoreArtifactRoot(path string) (windows.Handle, error) {
	path = filepath.Clean(path)
	volume := filepath.VolumeName(path)
	// The state directory is local application data. Reject UNC and device
	// namespace paths rather than treating a remote/provider root as trusted.
	if len(volume) != 2 || volume[1] != ':' {
		return windows.InvalidHandle, fmt.Errorf("restore state directory must be on a local drive")
	}

	volumeRoot := volume + string(filepath.Separator)
	rootPtr, err := windows.UTF16PtrFromString(volumeRoot)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode restore volume root: %w", err)
	}
	switch windows.GetDriveType(rootPtr) {
	case windows.DRIVE_FIXED, windows.DRIVE_REMOVABLE, windows.DRIVE_RAMDISK:
	default:
		return windows.InvalidHandle, fmt.Errorf("restore state directory must be on a supported local drive")
	}

	// A drive letter is itself an Object Manager symbolic link. Resolve its
	// current target once, reject SUBST and non-volume mappings, then open the
	// direct volume device path rather than traversing the DOS device alias.
	deviceName, err := windows.UTF16PtrFromString(volume)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode restore drive name: %w", err)
	}
	deviceBuffer := make([]uint16, 1024)
	deviceLength, err := windows.QueryDosDevice(deviceName, &deviceBuffer[0], uint32(len(deviceBuffer)))
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("resolve restore drive device: %w", err)
	}
	devicePath := windows.UTF16ToString(deviceBuffer[:deviceLength])
	if !strings.HasPrefix(devicePath, `\\Device\\`) ||
		strings.Contains(strings.TrimPrefix(devicePath, `\\Device\\`), `\\`) {
		return windows.InvalidHandle, fmt.Errorf("restore drive is not mapped directly to a local volume device")
	}

	volumePath, err := windows.NewNTUnicodeString(strings.TrimRight(devicePath, `\\`) + `\\`)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode restore NT volume path: %w", err)
	}
	attributes := windows.OBJECT_ATTRIBUTES{
		ObjectName: volumePath,
		Attributes: windows.OBJ_CASE_INSENSITIVE,
	}
	attributes.Length = uint32(unsafe.Sizeof(attributes))
	var ioStatus windows.IO_STATUS_BLOCK
	var current windows.Handle
	err = windows.NtCreateFile(
		&current,
		windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE|windows.SYNCHRONIZE,
		&attributes,
		&ioStatus,
		nil,
		windows.FILE_ATTRIBUTE_NORMAL,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		windows.FILE_OPEN,
		windows.FILE_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT,
		0,
		0,
	)
	runtime.KeepAlive(volumePath)
	runtime.KeepAlive(&attributes)
	runtime.KeepAlive(&ioStatus)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("open restore volume root by device path: %w", err)
	}
	if err := validateRestoreArtifactDirectoryHandle(current); err != nil {
		_ = windows.CloseHandle(current)
		return windows.InvalidHandle, fmt.Errorf("validate restore volume root: %w", err)
	}

	relative := strings.TrimLeft(strings.TrimPrefix(path, volume), `\\/`)
	if relative == "" {
		_ = windows.CloseHandle(current)
		return windows.InvalidHandle, fmt.Errorf("restore state directory cannot be a volume root")
	}
	components := strings.FieldsFunc(relative, func(r rune) bool {
		return r == '\\' || r == '/'
	})
	if len(components) == 0 {
		_ = windows.CloseHandle(current)
		return windows.InvalidHandle, fmt.Errorf("restore state directory has no path components")
	}

	// Resolve one component at a time relative to a pinned parent handle.
	// FILE_OPEN_REPARSE_POINT opens each entry itself; validation rejects it
	// before that handle can become the next RootDirectory. This avoids the
	// ancestor traversal that remains possible when CreateFile opens the full
	// canonical path with FILE_FLAG_OPEN_REPARSE_POINT.
	for _, component := range components {
		next, openErr := openRestoreArtifactDirectoryRelative(current, component)
		_ = windows.CloseHandle(current)
		if openErr != nil {
			return windows.InvalidHandle, fmt.Errorf("open restore state directory component %q: %w", component, openErr)
		}
		current = next
	}
	return current, nil
}

func openRestoreArtifactDirectoryRelative(parent windows.Handle, leaf string) (windows.Handle, error) {
	handle, err := openRestoreArtifactChildRelative(
		parent,
		leaf,
		windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE|windows.SYNCHRONIZE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		windows.FILE_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT,
	)
	if err != nil {
		return windows.InvalidHandle, err
	}
	if err := validateRestoreArtifactDirectoryHandle(handle); err != nil {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, err
	}
	return handle, nil
}

func openRestoreArtifactFileRelative(parent windows.Handle, leaf string) (windows.Handle, error) {
	handle, err := openRestoreArtifactChildRelative(
		parent,
		leaf,
		windows.FILE_READ_DATA|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE,
		windows.FILE_SHARE_READ,
		windows.FILE_NON_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT,
	)
	if err != nil {
		return windows.InvalidHandle, err
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("inspect opened restore artifact handle: %w", err)
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DEVICE) != 0 {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("restore artifact is not a regular non-reparse file")
	}
	return handle, nil
}

func openRestoreArtifactChildRelative(
	parent windows.Handle,
	leaf string,
	desiredAccess uint32,
	shareAccess uint32,
	createOptions uint32,
) (windows.Handle, error) {
	if err := validateRestoreMoveLeaf(leaf); err != nil {
		return windows.InvalidHandle, err
	}
	objectName, err := windows.NewNTUnicodeString(leaf)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode restore artifact path component: %w", err)
	}
	attributes := windows.OBJECT_ATTRIBUTES{
		RootDirectory: parent,
		ObjectName:    objectName,
		Attributes:    windows.OBJ_CASE_INSENSITIVE,
	}
	attributes.Length = uint32(unsafe.Sizeof(attributes))

	var ioStatus windows.IO_STATUS_BLOCK
	var handle windows.Handle
	err = windows.NtCreateFile(
		&handle,
		desiredAccess,
		&attributes,
		&ioStatus,
		nil,
		windows.FILE_ATTRIBUTE_NORMAL,
		shareAccess,
		windows.FILE_OPEN,
		createOptions,
		0,
		0,
	)
	runtime.KeepAlive(objectName)
	runtime.KeepAlive(&attributes)
	runtime.KeepAlive(&ioStatus)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("open restore artifact relative to pinned directory: %w", err)
	}
	return handle, nil
}

func validateRestoreArtifactDirectoryHandle(handle windows.Handle) error {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return fmt.Errorf("inspect opened restore directory handle: %w", err)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 ||
		info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return fmt.Errorf("restore path component is not a regular non-reparse directory")
	}
	return nil
}
