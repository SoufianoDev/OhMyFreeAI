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
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode restore state directory: %w", err)
	}
	handle, err := windows.CreateFile(
		pathPtr,
		windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("open restore state directory: %w", err)
	}
	if err := validateRestoreArtifactDirectoryHandle(handle); err != nil {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, err
	}
	return handle, nil
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
