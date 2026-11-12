//go:build windows

package sqlite

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func acquireRestoreJournalLockPlatform(ctx context.Context, root string) (func() error, error) {
	directory, err := openRestoreDirectoryNoReparse(root)
	if err != nil {
		return nil, fmt.Errorf("open qualified Restore journal lock directory: %w", err)
	}
	lockHandle, err := openRestoreJournalLockFile(directory)
	if err != nil {
		_ = windows.CloseHandle(directory)
		return nil, err
	}

	var overlapped windows.Overlapped
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			_ = windows.CloseHandle(lockHandle)
			_ = windows.CloseHandle(directory)
			return nil, err
		}
		err := windows.LockFileEx(
			lockHandle,
			windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
			0,
			1,
			0,
			&overlapped,
		)
		runtime.KeepAlive(&overlapped)
		if err == nil {
			break
		}
		if !errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			_ = windows.CloseHandle(lockHandle)
			_ = windows.CloseHandle(directory)
			return nil, fmt.Errorf("lock Restore journal ledger: %w", err)
		}
		select {
		case <-ctx.Done():
			_ = windows.CloseHandle(lockHandle)
			_ = windows.CloseHandle(directory)
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}

	if err := ctx.Err(); err != nil {
		_ = windows.UnlockFileEx(lockHandle, 0, 1, 0, &overlapped)
		_ = windows.CloseHandle(lockHandle)
		_ = windows.CloseHandle(directory)
		return nil, err
	}

	return func() error {
		unlockErr := windows.UnlockFileEx(lockHandle, 0, 1, 0, &overlapped)
		closeFileErr := windows.CloseHandle(lockHandle)
		closeDirectoryErr := windows.CloseHandle(directory)
		var result error
		if unlockErr != nil {
			result = errors.Join(result, fmt.Errorf("unlock Restore journal ledger: %w", unlockErr))
		}
		if closeFileErr != nil {
			result = errors.Join(result, fmt.Errorf("close Restore journal lock file: %w", closeFileErr))
		}
		if closeDirectoryErr != nil {
			result = errors.Join(result, fmt.Errorf("close Restore journal lock directory: %w", closeDirectoryErr))
		}
		return result
	}, nil
}

func openRestoreJournalLockFile(directory windows.Handle) (windows.Handle, error) {
	leaf := restoreJournalLockFileName
	if err := validateRestoreMoveLeaf(leaf); err != nil {
		return windows.InvalidHandle, err
	}
	objectName, err := windows.NewNTUnicodeString(leaf)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("encode Restore journal lock filename: %w", err)
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
		windows.GENERIC_READ|windows.GENERIC_WRITE|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE,
		&attributes,
		&ioStatus,
		nil,
		windows.FILE_ATTRIBUTE_NORMAL,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		windows.FILE_OPEN_IF,
		windows.FILE_NON_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT,
		0,
		0,
	)
	runtime.KeepAlive(directory)
	runtime.KeepAlive(objectName)
	runtime.KeepAlive(&attributes)
	runtime.KeepAlive(&ioStatus)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("open or create Restore journal lock relative to the qualified directory handle: %w", err)
	}

	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("inspect Restore journal lock file: %w", err)
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DEVICE) != 0 {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("Restore journal lock is not a regular non-reparse file")
	}
	if info.FileSizeHigh != 0 || info.FileSizeLow != 0 {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("Restore journal lock file contains unexpected data")
	}
	if err := verifyRestoreDirectorySecurity(handle); err != nil {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("validate private Restore journal lock security: %w", err)
	}
	lockPath, err := restoreWindowsPathFromHandle(handle)
	if err != nil {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("resolve opened Restore journal lock path: %w", err)
	}
	directoryPath, err := restoreWindowsPathFromHandle(directory)
	if err != nil {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("resolve qualified Restore journal lock directory: %w", err)
	}
	if !equalRestoreMovePath(filepath.Dir(lockPath), directoryPath) ||
		!equalRestoreMovePath(filepath.Base(lockPath), leaf) {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, fmt.Errorf("Restore journal lock is not the expected leaf in the qualified state directory")
	}
	return handle, nil
}
