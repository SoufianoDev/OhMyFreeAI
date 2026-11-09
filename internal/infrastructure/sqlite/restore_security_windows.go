//go:build windows

package sqlite

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type restoreVolumeInfo struct {
	DriveType uint32
	FileSystem string
	PersistentACLs bool
}

const (
	restoreFilePersistentACLs = 0x00000008
	restoreFileRemoteProtocolInfoClass = 13
	restoreBroadWriteMask = uint32(0x00000002 | 0x00000004 | 0x00000040 | 0x00010000 | 0x00040000 | 0x00080000 | 0x10000000 | 0x40000000)
	restoreSecurityOwnerAndDACL = 0x00000001 | 0x00000004
)

var (
	restoreAdvapi32 = windows.NewLazySystemDLL("advapi32.dll")
	restoreGetSecurityInfo = restoreAdvapi32.NewProc("GetSecurityInfo")
	restoreKernel32 = windows.NewLazySystemDLL("kernel32.dll")
	restoreGetVolumeInformationByHandle = restoreKernel32.NewProc("GetVolumeInformationByHandleW")
	restoreGetFileInformationByHandleEx = restoreKernel32.NewProc("GetFileInformationByHandleEx")
)

type restoreRemoteProtocolInfo struct {
	StructureVersion uint16
	StructureSize uint16
	Protocol uint32
	ProtocolMajorVersion uint16
	ProtocolMinorVersion uint16
	ProtocolRevision uint16
	Reserved uint16
	Flags uint32
	GenericReserved [8]uint32
	ProtocolSpecificReserved [16]uint32
	ProtocolSpecific [16]uint32
}

func verifyRestoreDirectorySecurity(directory windows.Handle) error {
	if directory == windows.InvalidHandle {
		return fmt.Errorf("restore state directory handle is invalid")
	}

	var owner *windows.SID
	var dacl *windows.ACL
	var descriptor *windows.SECURITY_DESCRIPTOR
	status, _, _ := restoreGetSecurityInfo.Call(
		uintptr(directory),
		uintptr(windows.SE_FILE_OBJECT),
		uintptr(restoreSecurityOwnerAndDACL),
		uintptr(unsafe.Pointer(&owner)),
		0,
		uintptr(unsafe.Pointer(&dacl)),
		0,
		uintptr(unsafe.Pointer(&descriptor)),
	)
	if status != 0 {
		return fmt.Errorf("GetSecurityInfo for restore state directory failed: %w", syscall.Errno(status))
	}
	if descriptor != nil {
		defer windows.LocalFree(windows.Handle(unsafe.Pointer(descriptor)))
	}
	if owner == nil || !owner.IsValid() {
		return fmt.Errorf("restore state directory has no valid owner SID")
	}
	if dacl == nil {
		return fmt.Errorf("restore state directory has a NULL DACL; access is not restricted")
	}

	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return fmt.Errorf("open current process token: %w", err)
	}
	defer token.Close()
	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return fmt.Errorf("read current process token user: %w", err)
	}
	if tokenUser == nil || tokenUser.User.Sid == nil || !owner.Equals(tokenUser.User.Sid) {
		return fmt.Errorf("restore state directory owner does not match the current user")
	}

	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &ace); err != nil {
			return fmt.Errorf("read restore state directory DACL ACE %d: %w", index, err)
		}
		if ace == nil {
			return fmt.Errorf("restore state directory DACL contains a null ACE")
		}
		switch ace.Header.AceType {
		case windows.ACCESS_ALLOWED_ACE_TYPE, windows.ACCESS_DENIED_ACE_TYPE:
			// These two basic ACE layouts share the same header/mask/SID offset.
		default:
			// Object and callback ACE layouts need additional GUID/condition
			// parsing. Reject them rather than accidentally overlooking a broad
			// allow rule.
			return fmt.Errorf("restore state directory DACL contains unsupported ACE type %d", ace.Header.AceType)
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !sid.IsValid() {
			return fmt.Errorf("restore state directory DACL contains an invalid SID")
		}
		if isBroadWritableRestoreACE(ace.Header.AceType, uint32(ace.Mask), sid) {
			return fmt.Errorf("restore state directory grants broad write or delete access")
		}
	}
	return nil
}

func isBroadWritableRestoreACE(aceType uint8, mask uint32, sid *windows.SID) bool {
	if aceType != windows.ACCESS_ALLOWED_ACE_TYPE || mask&restoreBroadWriteMask == 0 || sid == nil || !sid.IsValid() {
		return false
	}
	for _, wellKnown := range []windows.WELL_KNOWN_SID_TYPE{
		windows.WinWorldSid,
		windows.WinAuthenticatedUserSid,
		windows.WinBuiltinUsersSid,
		windows.WinBuiltinGuestsSid,
		windows.WinBuiltinPowerUsersSid,
		windows.WinBuiltinPreWindows2000CompatibleAccessSid,
		windows.WinInteractiveSid,
		windows.WinNetworkSid,
		windows.WinAnonymousSid,
	} {
		broadSID, err := windows.CreateWellKnownSid(wellKnown)
		if err != nil {
			// A failure to construct a well-known SID must not turn a potentially
			// broad ACE into an accepted one. The caller's allow decision is
			// therefore conservative for this ACE.
			return true
		}
		if sid.Equals(broadSID) {
			return true
		}
	}
	return false
}

func validateRestoreVolume(info restoreVolumeInfo) error {
	if info.DriveType != windows.DRIVE_FIXED {
		return fmt.Errorf("restore state directory must be on a fixed local drive (drive type %d)", info.DriveType)
	}
	if !strings.EqualFold(info.FileSystem, "NTFS") {
		return fmt.Errorf("restore state directory filesystem %q is not qualified; only NTFS is currently supported", info.FileSystem)
	}
	if !info.PersistentACLs {
		return fmt.Errorf("restore state directory filesystem does not report persistent ACL support")
	}
	return nil
}

func verifyRestoreVolume(directory windows.Handle, driveRoot string) error {
	if directory == windows.InvalidHandle {
		return fmt.Errorf("restore state directory handle is invalid")
	}
	root, err := windows.UTF16PtrFromString(driveRoot)
	if err != nil {
		return fmt.Errorf("encode restore drive root: %w", err)
	}
	driveType := windows.GetDriveType(root)

	var volumeName [261]uint16
	var fileSystemName [261]uint16
	var fileSystemFlags uint32
	result, _, callErr := restoreGetVolumeInformationByHandle.Call(
		uintptr(directory),
		uintptr(unsafe.Pointer(&volumeName[0])),
		uintptr(len(volumeName)),
		0,
		0,
		uintptr(unsafe.Pointer(&fileSystemFlags)),
		uintptr(unsafe.Pointer(&fileSystemName[0])),
		uintptr(len(fileSystemName)),
	)
	if result == 0 {
		if callErr == syscall.Errno(0) {
			callErr = syscall.EINVAL
		}
		return fmt.Errorf("GetVolumeInformationByHandleW failed: %w", callErr)
	}

	// FileRemoteProtocolInfo is handle-based, so a mounted remote share cannot
	// pass merely because its drive-letter root is local. On local NTFS, Windows
	// reports ERROR_INVALID_PARAMETER because no remote protocol is associated.
	var remote restoreRemoteProtocolInfo
	remote.StructureVersion = 2
	remote.StructureSize = uint16(unsafe.Sizeof(remote))
	remoteResult, _, remoteErr := restoreGetFileInformationByHandleEx.Call(
		uintptr(directory),
		restoreFileRemoteProtocolInfoClass,
		uintptr(unsafe.Pointer(&remote)),
		unsafe.Sizeof(remote),
	)
	if remoteResult != 0 && remote.Protocol != 0 {
		return fmt.Errorf("restore state directory is backed by remote protocol 0x%x", remote.Protocol)
	}
	if remoteResult == 0 && remoteErr != syscall.Errno(windows.ERROR_INVALID_PARAMETER) &&
		remoteErr != syscall.Errno(windows.ERROR_INVALID_FUNCTION) &&
		remoteErr != syscall.Errno(windows.ERROR_NOT_SUPPORTED) {
		if remoteErr == syscall.Errno(0) {
			remoteErr = syscall.EINVAL
		}
		return fmt.Errorf("query restore volume remote protocol: %w", remoteErr)
	}

	info := restoreVolumeInfo{
		DriveType: driveType,
		FileSystem: windows.UTF16ToString(fileSystemName[:]),
		PersistentACLs: fileSystemFlags&restoreFilePersistentACLs != 0,
	}
	if err := validateRestoreVolume(info); err != nil {
		return err
	}
	return nil
}
