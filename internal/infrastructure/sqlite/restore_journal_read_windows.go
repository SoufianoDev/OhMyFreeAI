//go:build windows

package sqlite

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// readRestoreJournalFilePlatform reads the ledger from a handle opened
// relative to the validated state directory without following reparse points.
func readRestoreJournalFilePlatform(path string, maxBytes int) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("%w: journal ledger size limit is invalid", ErrRestoreJournalInvalid)
	}
	if err := validateRestoreIdentityLeaf(filepath.Base(path)); err != nil {
		return nil, fmt.Errorf("%w: journal ledger filename is invalid: %v", ErrRestoreJournalInvalid, err)
	}

	directory, err := openRestoreDirectoryNoReparse(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("open qualified SQLite restore journal directory: %w", err)
	}
	defer windows.CloseHandle(directory)

	leaf := filepath.Base(path)
	if err := validateRestoreMoveLeaf(leaf); err != nil {
		return nil, fmt.Errorf("%w: journal ledger filename is invalid: %v", ErrRestoreJournalInvalid, err)
	}
	objectName, err := windows.NewNTUnicodeString(leaf)
	if err != nil {
		return nil, fmt.Errorf("encode SQLite restore journal filename: %w", err)
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
		windows.FILE_READ_DATA|windows.FILE_READ_ATTRIBUTES|windows.READ_CONTROL|windows.SYNCHRONIZE,
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
		var status windows.NTStatus
		if errors.As(err, &status) {
			switch uint32(status) {
			case 0xC000000F, 0xC0000034, 0xC000003A: // STATUS_NO_SUCH_FILE, STATUS_OBJECT_NAME_NOT_FOUND, STATUS_OBJECT_PATH_NOT_FOUND
				return nil, fmt.Errorf("open SQLite restore journal ledger: %w (%v)", os.ErrNotExist, err)
			case 0xC00000BA: // STATUS_FILE_IS_A_DIRECTORY
				return nil, fmt.Errorf("%w: journal ledger path is a directory", ErrRestoreJournalInvalid)
			case 0xC000050B, 0x8000002D: // STATUS_REPARSE_POINT_ENCOUNTERED, STATUS_STOPPED_ON_SYMLINK
				return nil, fmt.Errorf("%w: journal ledger must not be a reparse point", ErrRestoreJournalInvalid)
			}
		}
		return nil, fmt.Errorf("open SQLite restore journal ledger without following reparse points: %w", err)
	}
	defer windows.CloseHandle(handle)

	var before windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &before); err != nil {
		return nil, fmt.Errorf("inspect opened SQLite restore journal ledger: %w", err)
	}
	if before.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DEVICE) != 0 ||
		before.NumberOfLinks != 1 {
		return nil, fmt.Errorf("%w: journal ledger must be a single-link regular non-reparse file", ErrRestoreJournalInvalid)
	}
	if err := verifyRestoreDirectorySecurity(handle); err != nil {
		return nil, fmt.Errorf("%w: journal ledger security validation failed: %v", ErrRestoreJournalInvalid, err)
	}

	openedPath, err := restoreWindowsPathFromHandle(handle)
	if err != nil {
		return nil, fmt.Errorf("resolve opened SQLite restore journal path: %w", err)
	}
	directoryPath, err := restoreWindowsPathFromHandle(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve SQLite restore journal directory path: %w", err)
	}
	if !equalRestoreMovePath(filepath.Dir(openedPath), directoryPath) ||
		!equalRestoreMovePath(filepath.Base(openedPath), leaf) {
		return nil, fmt.Errorf("%w: journal ledger is not the expected leaf in the qualified state directory", ErrRestoreJournalInvalid)
	}

	size := uint64(before.FileSizeHigh)<<32 | uint64(before.FileSizeLow)
	if size == 0 || size > uint64(maxBytes) {
		return nil, fmt.Errorf("%w: journal ledger is empty or exceeds its size limit", ErrRestoreJournalInvalid)
	}
	data := make([]byte, int(size))
	var total uint64
	for total < size {
		var read uint32
		if err := windows.ReadFile(handle, data[int(total):], &read, nil); err != nil {
			return nil, fmt.Errorf("read opened SQLite restore journal ledger: %w", err)
		}
		if read == 0 {
			return nil, fmt.Errorf("read opened SQLite restore journal ledger: unexpected end of file")
		}
		total += uint64(read)
	}

	var after windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &after); err != nil {
		return nil, fmt.Errorf("reinspect opened SQLite restore journal ledger: %w", err)
	}
	afterSize := uint64(after.FileSizeHigh)<<32 | uint64(after.FileSizeLow)
	if after.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DEVICE) != 0 ||
		after.NumberOfLinks != before.NumberOfLinks ||
		after.VolumeSerialNumber != before.VolumeSerialNumber ||
		after.FileIndexHigh != before.FileIndexHigh || after.FileIndexLow != before.FileIndexLow ||
		afterSize != size || total != size ||
		after.CreationTime != before.CreationTime || after.LastWriteTime != before.LastWriteTime {
		return nil, fmt.Errorf("%w: journal ledger changed while being read", ErrRestoreJournalInvalid)
	}
	return data, nil
}
