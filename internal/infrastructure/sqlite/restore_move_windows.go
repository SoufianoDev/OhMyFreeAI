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
	directory := filepath.Dir(destination)
	directoryPtr, err := windows.UTF16PtrFromString(directory)
	if err != nil {
		return fmt.Errorf("encode SQLite restore destination directory: %w", err)
	}
	directoryHandle, err := windows.CreateFile(
		directoryPtr,
		windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return fmt.Errorf("open SQLite restore destination directory: %w", err)
	}
	defer windows.CloseHandle(directoryHandle)

	sourceHandle, err := openRestoreSourceRelative(directoryHandle, filepath.Base(source))
	if err != nil {
		return fmt.Errorf("open SQLite restore source relative to pinned directory: %w", err)
	}
	defer windows.CloseHandle(sourceHandle)

	return renameRestoreHandleNoReplace(sourceHandle, directoryHandle, filepath.Base(destination))
}

func openRestoreSourceRelative(directory windows.Handle, sourceLeaf string) (windows.Handle, error) {
	if err := validateRestoreMoveLeaf(sourceLeaf); err != nil {
		return windows.InvalidHandle, err
	}
	objectName, err := windows.NewNTUnicodeString(sourceLeaf)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode SQLite restore source filename: %w", err)
	}
	objectAttributes := windows.OBJECT_ATTRIBUTES{
		RootDirectory: directory,
		ObjectName:    objectName,
		Attributes:    windows.OBJ_CASE_INSENSITIVE,
	}
	objectAttributes.Length = uint32(unsafe.Sizeof(objectAttributes))

	var ioStatus windows.IO_STATUS_BLOCK
	var source windows.Handle
	err = windows.NtCreateFile(
		&source,
		windows.DELETE|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE,
		&objectAttributes,
		&ioStatus,
		nil,
		windows.FILE_ATTRIBUTE_NORMAL,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		windows.FILE_OPEN,
		windows.FILE_NON_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT,
		nil,
		0,
	)
	runtime.KeepAlive(objectName)
	runtime.KeepAlive(&objectAttributes)
	runtime.KeepAlive(&ioStatus)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("NtCreateFile relative to SQLite restore directory: %w", err)
	}
	return source, nil
}

func validateRestoreMoveLeaf(leaf string) error {
	if leaf == "" || leaf == "." || leaf == ".." || strings.ContainsAny(leaf, `/\\:`) {
		return fmt.Errorf("SQLite restore filename must be a single path component")
	}
	return nil
}

func renameRestoreHandleNoReplace(source, destinationDirectory windows.Handle, destinationLeaf string) error {
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

	var directoryInfo windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(destinationDirectory, &directoryInfo); err != nil {
		return fmt.Errorf("inspect opened SQLite restore destination directory: %w", err)
	}
	if directoryInfo.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 ||
		directoryInfo.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return fmt.Errorf("opened SQLite restore destination is not a regular directory")
	}

	encodedName, err := windows.UTF16FromString(destinationLeaf)
	if err != nil {
		return fmt.Errorf("encode SQLite restore destination filename: %w", err)
	}
	encodedName = encodedName[:len(encodedName)-1]
	nameBytes := len(encodedName) * 2
	if nameBytes == 0 || uint64(nameBytes) > uint64(^uint32(0)) {
		return fmt.Errorf("SQLite restore destination filename length is invalid")
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
	binary.LittleEndian.PutUint32(buffer[0:4], 0)
	binary.LittleEndian.PutUint32(buffer[4:8], 0)
	binary.LittleEndian.PutUint64(buffer[8:16], uint64(destinationDirectory))
	binary.LittleEndian.PutUint32(buffer[16:20], uint32(nameBytes))
	for i, unit := range encodedName {
		binary.LittleEndian.PutUint16(buffer[fixedHeaderSize+i*2:fixedHeaderSize+i*2+2], unit)
	}

	var ioStatus windows.IO_STATUS_BLOCK
	err = windows.NtSetInformationFile(
		source,
		&ioStatus,
		&buffer[0],
		uint32(len(buffer)),
		fileRenameInformationExClass,
	)
	runtime.KeepAlive(buffer)
	runtime.KeepAlive(source)
	runtime.KeepAlive(destinationDirectory)
	runtime.KeepAlive(&ioStatus)
	if err != nil {
		if errors.Is(err, windows.STATUS_OBJECT_NAME_COLLISION) {
			return fmt.Errorf("%w: %v", os.ErrExist, err)
		}
		return fmt.Errorf("NtSetInformationFile(FileRenameInformationEx): %w", err)
	}
	return nil
}
