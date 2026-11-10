package sqlite

import (
	"errors"
	"fmt"
	"os"
)

// inspectRestoreRecoveryFiles takes a read-only identity snapshot of the
// transaction artifacts. It does not execute the recovery plan or open SQLite;
// the executor must revalidate each artifact at its mutation boundary.
func inspectRestoreRecoveryFiles(stateDirectory string, journal RestoreJournal) (RestoreRecoveryFiles, error) {
	ambiguous := func(reason string, cause error) (RestoreRecoveryFiles, error) {
		if cause != nil {
			return RestoreRecoveryFiles{}, fmt.Errorf("%w: %s: %w", ErrRestoreRecoveryAmbiguous, reason, cause)
		}
		return RestoreRecoveryFiles{}, fmt.Errorf("%w: %s", ErrRestoreRecoveryAmbiguous, reason)
	}

	root, err := canonicalStateDirectory(stateDirectory)
	if err != nil {
		return ambiguous("resolve canonical state directory", err)
	}
	journal, err = canonicalizeRestoreJournal(root, journal)
	if err != nil {
		return ambiguous("validate recovery journal and transaction paths", err)
	}

	hadPrevious := *journal.PreviousDatabaseExisted
	if hadPrevious &&
		journal.CandidateSHA256 != "" &&
		journal.CandidateSHA256 == journal.PreviousSHA256 &&
		journal.CandidateSize == journal.PreviousSize {
		return ambiguous("candidate and previous database identities are indistinguishable", nil)
	}

	inspect := func(path string) RestoreFileIdentity {
		if path == "" {
			return RestoreFileMissing
		}
		actual, err := readRestoreContentIdentityPlatform(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return RestoreFileMissing
			}
			return RestoreFileUnreadable
		}
		if journal.CandidateSHA256 != "" && actual.matches(journal.CandidateSHA256, journal.CandidateSize) {
			return RestoreFileCandidate
		}
		if hadPrevious && actual.matches(journal.PreviousSHA256, journal.PreviousSize) {
			return RestoreFilePrevious
		}
		return RestoreFileMismatch
	}

	return RestoreRecoveryFiles{
		Target:     inspect(journal.TargetPath),
		Staging:    inspect(journal.StagingPath),
		Rollback:   inspect(journal.RollbackPath),
		Quarantine: inspect(journal.CandidateQuarantinePath),
	}, nil
}
