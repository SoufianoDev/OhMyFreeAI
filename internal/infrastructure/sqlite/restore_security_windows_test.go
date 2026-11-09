//go:build windows

package sqlite

import (
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestValidateRestoreVolume(t *testing.T) {
	tests := []struct {
		name string
		info restoreVolumeInfo
		wantErr bool
	}{
		{name: "local NTFS", info: restoreVolumeInfo{DriveType: windows.DRIVE_FIXED, FileSystem: "NTFS", PersistentACLs: true}},
		{name: "network drive", info: restoreVolumeInfo{DriveType: windows.DRIVE_REMOTE, FileSystem: "NTFS", PersistentACLs: true}, wantErr: true},
		{name: "removable drive", info: restoreVolumeInfo{DriveType: windows.DRIVE_REMOVABLE, FileSystem: "NTFS", PersistentACLs: true}, wantErr: true},
		{name: "unknown drive", info: restoreVolumeInfo{DriveType: windows.DRIVE_UNKNOWN, FileSystem: "NTFS", PersistentACLs: true}, wantErr: true},
		{name: "ReFS not yet qualified", info: restoreVolumeInfo{DriveType: windows.DRIVE_FIXED, FileSystem: "ReFS", PersistentACLs: true}, wantErr: true},
		{name: "FAT without persistent ACLs", info: restoreVolumeInfo{DriveType: windows.DRIVE_FIXED, FileSystem: "FAT32", PersistentACLs: false}, wantErr: true},
		{name: "NTFS without persistent ACLs", info: restoreVolumeInfo{DriveType: windows.DRIVE_FIXED, FileSystem: "NTFS", PersistentACLs: false}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRestoreVolume(test.info)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateRestoreVolume() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestIsBroadWritableRestoreACE(t *testing.T) {
	tests := []struct {
		name string
		sid string
		aceType uint8
		mask uint32
		want bool
	}{
		{name: "Everyone can add files", sid: "S-1-1-0", aceType: windows.ACCESS_ALLOWED_ACE_TYPE, mask: 0x2, want: true},
		{name: "Everyone can modify directory metadata", sid: "S-1-1-0", aceType: windows.ACCESS_ALLOWED_ACE_TYPE, mask: 0x110, want: true},
		{name: "Authenticated Users can delete children", sid: "S-1-5-11", aceType: windows.ACCESS_ALLOWED_ACE_TYPE, mask: 0x40, want: true},
		{name: "Builtin Users generic write", sid: "S-1-5-32-545", aceType: windows.ACCESS_ALLOWED_ACE_TYPE, mask: 0x40000000, want: true},
		{name: "Everyone read only", sid: "S-1-1-0", aceType: windows.ACCESS_ALLOWED_ACE_TYPE, mask: 0x1, want: false},
		{name: "Everyone deny write", sid: "S-1-1-0", aceType: windows.ACCESS_DENIED_ACE_TYPE, mask: 0x2, want: false},
		{name: "ordinary user write", sid: "S-1-5-21-1-2-3-1001", aceType: windows.ACCESS_ALLOWED_ACE_TYPE, mask: 0x2, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sid, err := windows.StringToSid(test.sid)
			if err != nil {
				t.Fatal(err)
			}
			if got := isBroadWritableRestoreACE(test.aceType, test.mask, sid); got != test.want {
				t.Fatalf("isBroadWritableRestoreACE() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRestoreDirectorySecurityAcceptsPrivateTempDirectory(t *testing.T) {
	directory := t.TempDir()
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(directory),
		windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)

	if err := verifyRestoreDirectorySecurity(handle); err != nil {
		t.Fatalf("verifyRestoreDirectorySecurity() rejected private temp directory: %v", err)
	}
}

func TestRestoreDirectoryTraversalAcceptsQualifiedLocalVolume(t *testing.T) {
	directory := t.TempDir()
	absolute, err := filepath.Abs(directory)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.VolumeName(absolute) + string(filepath.Separator)
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(root),
		windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES|windows.FILE_TRAVERSE|windows.SYNCHRONIZE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)

	if err := verifyRestoreVolume(handle, root); err != nil {
		t.Fatalf("verifyRestoreVolume() rejected local test volume: %v", err)
	}
}
