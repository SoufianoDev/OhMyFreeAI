package sqlite

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInspectRestoreRecoveryFilesIdentifiesCandidatePreviousAndMissing(t *testing.T) {
	root := t.TempDir()
	candidate := []byte("candidate database bytes")
	previous := []byte("previous database bytes")
	candidateIdentity := restoreIdentityForTest(candidate)
	previousIdentity := restoreIdentityForTest(previous)

	journal := validTestRestoreJournal(root)
	journal.Phase = RestorePhasePublished
	journal.CandidateSHA256 = candidateIdentity.SHA256
	journal.CandidateSize = candidateIdentity.Size
	journal.PreviousSHA256 = previousIdentity.SHA256
	journal.PreviousSize = previousIdentity.Size

	if err := os.WriteFile(journal.TargetPath, candidate, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(journal.RollbackPath, previous, 0o600); err != nil {
		t.Fatal(err)
	}

	files, err := inspectRestoreRecoveryFiles(root, journal)
	if err != nil {
		t.Fatalf("inspectRestoreRecoveryFiles() error = %v", err)
	}
	want := recoveryFiles(RestoreFileCandidate, RestoreFileMissing, RestoreFilePrevious, RestoreFileMissing)
	if files != want {
		t.Fatalf("inspectRestoreRecoveryFiles() = %#v, want %#v", files, want)
	}
}

func TestInspectRestoreRecoveryFilesMarksUnexpectedContentAsMismatch(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.Phase = RestorePhasePrepared

	if err := os.WriteFile(journal.TargetPath, []byte("unexpected database bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	files, err := inspectRestoreRecoveryFiles(root, journal)
	if err != nil {
		t.Fatalf("inspectRestoreRecoveryFiles() error = %v", err)
	}
	if files.Target != RestoreFileMismatch {
		t.Fatalf("target identity = %q, want %q", files.Target, RestoreFileMismatch)
	}
}

func TestInspectRestoreRecoveryFilesRejectsIndistinguishableCandidateAndPrevious(t *testing.T) {
	root := t.TempDir()
	journal := validTestRestoreJournal(root)
	journal.CandidateSHA256 = journal.PreviousSHA256
	journal.CandidateSize = journal.PreviousSize

	if _, err := inspectRestoreRecoveryFiles(root, journal); !errors.Is(err, ErrRestoreRecoveryAmbiguous) {
		t.Fatalf("inspectRestoreRecoveryFiles() error = %v, want ErrRestoreRecoveryAmbiguous", err)
	}
}

func TestInspectRestoreRecoveryFilesRejectsSymlinkWithoutChangingIt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks may require elevated privileges on Windows")
	}
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.db")
	if err := os.WriteFile(outside, []byte("outside database"), 0o600); err != nil {
		t.Fatal(err)
	}
	journal := validTestRestoreJournal(root)
	if err := os.Symlink(outside, journal.TargetPath); err != nil {
		t.Fatal(err)
	}

	if _, err := inspectRestoreRecoveryFiles(root, journal); !errors.Is(err, ErrRestoreRecoveryAmbiguous) {
		t.Fatalf("inspectRestoreRecoveryFiles() error = %v, want ErrRestoreRecoveryAmbiguous", err)
	}
	info, err := os.Lstat(journal.TargetPath)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink artifact changed during inspection: info=%v err=%v", info, err)
	}
	if data, err := os.ReadFile(outside); err != nil || string(data) != "outside database" {
		t.Fatalf("symlink target changed during inspection: data=%q err=%v", data, err)
	}
}

func TestReadRestoreContentIdentityDoesNotTreatMissingParentAsMissingLeaf(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-parent", "missing.db")
	if _, err := readRestoreContentIdentityPlatform(path); err == nil || errors.Is(err, os.ErrNotExist) {
		t.Fatalf("readRestoreContentIdentityPlatform() error = %v, want a non-NotExist parent error", err)
	}
}

func TestCanonicalStateDirectoryRejectsNonPrivateDirectory(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("POSIX owner/mode checks are Linux-specific")
	}
	root := t.TempDir()
	if err := os.Chmod(root, 0o750); err != nil {
		t.Fatal(err)
	}
	if _, err := canonicalStateDirectory(root); err == nil {
		t.Fatal("accepted a state directory accessible to group members")
	}
}

func TestReadRestoreContentIdentityReportsMissingLeafAsNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.db")
	if _, err := readRestoreContentIdentityPlatform(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("readRestoreContentIdentityPlatform() error = %v, want os.ErrNotExist", err)
	}
}
