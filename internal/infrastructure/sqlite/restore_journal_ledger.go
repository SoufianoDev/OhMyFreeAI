package sqlite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type restoreJournalLedger struct {
	Version      int              `json:"version"`
	Transactions []RestoreJournal `json:"transactions"`
}

func readRestoreJournalLedger(root string) (restoreJournalLedger, error) {
	var empty restoreJournalLedger
	path := filepath.Join(root, restoreJournalFileName)
	info, err := os.Lstat(path)
	if err != nil {
		return empty, err
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxRestoreJournalBytes {
		return empty, fmt.Errorf("%w: journal ledger is not a bounded regular file", ErrRestoreJournalInvalid)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return empty, fmt.Errorf("read SQLite restore journal ledger: %w", err)
	}

	var shape map[string]json.RawMessage
	if err := json.Unmarshal(data, &shape); err != nil {
		return empty, fmt.Errorf("%w: decode journal document: %v", ErrRestoreJournalInvalid, err)
	}
	if _, isLedger := shape["transactions"]; !isLedger {
		// Accept the previous single-record v3 format. The next successful
		// write upgrades it atomically to the retained ledger envelope.
		var legacy RestoreJournal
		if err := decodeStrictRestoreJournal(data, &legacy); err != nil {
			return empty, fmt.Errorf("%w: decode legacy journal: %v", ErrRestoreJournalInvalid, err)
		}
		legacy, err = canonicalizeRestoreJournal(root, legacy)
		if err != nil {
			return empty, err
		}
		return validateRestoreJournalLedger(root, restoreJournalLedger{
			Version:      restoreJournalLedgerVersion,
			Transactions: []RestoreJournal{legacy},
		})
	}

	var ledger restoreJournalLedger
	if err := decodeStrictRestoreJournal(data, &ledger); err != nil {
		return empty, fmt.Errorf("%w: decode journal ledger: %v", ErrRestoreJournalInvalid, err)
	}
	return validateRestoreJournalLedger(root, ledger)
}

func decodeStrictRestoreJournal(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("trailing journal data")
	}
	return nil
}

func validateRestoreJournalLedger(root string, ledger restoreJournalLedger) (restoreJournalLedger, error) {
	invalid := func(reason string) (restoreJournalLedger, error) {
		return restoreJournalLedger{}, fmt.Errorf("%w: %s", ErrRestoreJournalInvalid, reason)
	}
	if ledger.Version != restoreJournalLedgerVersion {
		return invalid("journal ledger version is unsupported")
	}
	if len(ledger.Transactions) == 0 || len(ledger.Transactions) > maxRestoreJournalTransactions {
		return invalid("journal ledger transaction count is invalid")
	}

	seenIDs := make(map[string]struct{}, len(ledger.Transactions))
	artifactPaths := make([]string, 0, len(ledger.Transactions)*3)
	targetPath := ""
	activeTransactions := 0
	for i := range ledger.Transactions {
		journal, err := canonicalizeRestoreJournal(root, ledger.Transactions[i])
		if err != nil {
			return restoreJournalLedger{}, err
		}
		ledger.Transactions[i] = journal
		if _, exists := seenIDs[journal.TransactionID]; exists {
			return invalid("journal ledger contains a duplicate transaction ID")
		}
		seenIDs[journal.TransactionID] = struct{}{}

		if targetPath == "" {
			targetPath = journal.TargetPath
		} else if !samePath(targetPath, journal.TargetPath) {
			return invalid("journal ledger contains different canonical target paths")
		}
		for _, artifact := range []string{journal.StagingPath, journal.RollbackPath, journal.CandidateQuarantinePath} {
			if artifact == "" {
				continue
			}
			if samePath(artifact, targetPath) {
				return invalid("journal artifact path collides with the canonical target")
			}
			for _, previous := range artifactPaths {
				if samePath(artifact, previous) {
					return invalid("journal artifact paths are reused across retained transactions")
				}
			}
			artifactPaths = append(artifactPaths, artifact)
		}
		if !terminalRestorePhase(journal.Phase) {
			activeTransactions++
		}
	}
	if activeTransactions > 1 {
		return invalid("journal ledger contains more than one active transaction")
	}
	return ledger, nil
}

func currentRestoreJournal(ledger restoreJournalLedger) (RestoreJournal, error) {
	var active *RestoreJournal
	for i := range ledger.Transactions {
		if terminalRestorePhase(ledger.Transactions[i].Phase) {
			continue
		}
		if active != nil {
			return RestoreJournal{}, fmt.Errorf("%w: more than one active transaction", ErrRestoreJournalInvalid)
		}
		active = &ledger.Transactions[i]
	}
	if active != nil {
		return *active, nil
	}
	if len(ledger.Transactions) == 0 {
		return RestoreJournal{}, os.ErrNotExist
	}
	return ledger.Transactions[len(ledger.Transactions)-1], nil
}

func terminalRestorePhase(phase RestorePhase) bool {
	return phase == RestorePhaseCommitted || phase == RestorePhaseRolledBack
}
