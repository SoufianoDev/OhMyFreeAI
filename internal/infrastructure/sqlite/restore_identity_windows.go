//go:build windows

package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// readRestoreContentIdentityPlatform opens the leaf relative to a retained,
// security-checked directory handle and hashes from that same no-reparse handle.
func readRestoreContentIdentityPlatform(path string) (RestoreContentIdentity, error) {
	var identity RestoreContentIdentity
	if err := validateRestoreIdentityLeaf(filepath.Base(path)); err != nil {
		return identity, err
	}
	directory, err := openRestoreDirectoryNoReparse(filepath.Dir(path))
	if err != nil {
		return identity, fmt.Errorf("open SQLite restore identity parent: %w", err)
	}
	defer windows.CloseHandle(directory)

	leaf := filepath.Base(path)
	if err := validateRestoreMoveLeaf(leaf); err != nil {
		return identity, err
	}
	objectName, err := windows.NewNTUnicodeString(leaf)
	if err != nil {
		return identity, fmt.Errorf("encode SQLite restore identity leaf: %w", err)
	}
	attributes := windows.OBJECT_ATTRIBUTES{
		RootDirectory: directory,
		ObjectName:    objectName,
		Attributes:    windows.OBJ_CASE_INSENSITIVE,
	}
	attributes.Length = uint32(unsafe.Sizeof(attributes))
	var ioStatus windows.IO_STATUS_BLOCK
	var handle windows.Handle
	err = windows.NtCreateFile(
		&handle,
		windows.FILE_READ_DATA|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE,
		&attributes,
		&ioStatus,
		nil,
		windows.FILE_ATTRIBUTE_NORMAL,
		windows.FILE_SHARE_READ,
		windows.FILE_OPEN,
		windows.FILE_NON_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT,
		0,
		0,
	)
	runtime.KeepAlive(directory)
	runtime.KeepAlive(objectName)
	runtime.KeepAlive(&attributes)
	runtime.KeepAlive(&ioStatus)
	if err != nil {
		return identity, fmt.Errorf("open SQLite restore identity leaf without following reparse points: %w", err)
	}
	defer windows.CloseHandle(handle)

	var before windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &before); err != nil {
		return identity, fmt.Errorf("inspect opened SQLite restore identity: %w", err)
	}
	if before.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DEVICE) != 0 {
		return identity, fmt.Errorf("SQLite restore identity is not a regular non-reparse file")
	}
	size := uint64(before.FileSizeHigh)<<32 | uint64(before.FileSizeLow)
	if size == 0 {
		return identity, fmt.Errorf("SQLite restore identity is empty")
	}

	hash := sha256.New()
	buffer := make([]byte, 128*1024)
	var total uint64
	for total < size {
		chunk := buffer
		if remaining := size - total; remaining < uint64(len(chunk)) {
			chunk = chunk[:int(remaining)]
		}
		var read uint32
		if err := windows.ReadFile(handle, chunk, &read, nil); err != nil {
			return identity, fmt.Errorf("read opened SQLite restore identity: %w", err)
		}
		if read == 0 {
			return identity, fmt.Errorf("SQLite restore identity ended before its recorded size")
		}
		_, _ = hash.Write(chunk[:read])
		total += uint64(read)
	}
	var after windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &after); err != nil {
		return identity, fmt.Errorf("reinspect opened SQLite restore identity: %w", err)
	}
	afterSize := uint64(after.FileSizeHigh)<<32 | uint64(after.FileSizeLow)
	if after.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DEVICE) != 0 ||
		after.VolumeSerialNumber != before.VolumeSerialNumber ||
		after.FileIndexHigh != before.FileIndexHigh || after.FileIndexLow != before.FileIndexLow ||
		afterSize != size || total != size ||
		after.LastWriteTime != before.LastWriteTime {
		return identity, fmt.Errorf("SQLite restore identity changed while being read")
	}
	identity.Size = total
	identity.SHA256 = hex.EncodeToString(hash.Sum(nil))
	return identity, nil
}
