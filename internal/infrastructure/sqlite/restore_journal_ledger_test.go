package sqlite

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func uniqueLedgerTestJournal(root, transactionID string) RestoreJournal {
	journal := validTestRestoreJournal(root)
	journal.TransactionID = transactionID
	journal.StagingPath = filepath.Join(root, "."+transactionID+"-stage.db")
	journal.RollbackPath = filepath.Join(root, "."+transactionID+"-rollback.db")
	journal.CandidateQuarantinePath = filepath.Join(root, "."+transactionID+"-quarantine.db")
	return journal
}

func TestRestoreJournalLedgerRejectsSecondActiveTransaction(t *testing.T) {
	root := restoreTestDir(t)
	first := validTestRestoreJournal(root)
	if err := WriteRestoreJournal(root, first); err != nil {
		t.Fatal(err)
	}
	second := uniqueLedgerTestJournal(root, "second-active")
	if err := WriteRestoreJournal(root, second); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("second active transaction error = %v, want ErrRestoreJournalInvalid", err)
	}
	ledger, err := readRestoreJournalLedger(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(ledger.Transactions) != 1 || ledger.Transactions[0].TransactionID != first.TransactionID {
		t.Fatalf("ledger after rejected second transaction = %#v", ledger.Transactions)
	}
}

func TestRestoreJournalLedgerRejectsArtifactPathReuseAcrossTransactions(t *testing.T) {
	root := restoreTestDir(t)
	first := validTestRestoreJournal(root)
	if err := WriteRestoreJournal(root, first); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []RestorePhase{
		RestorePhaseStaging,
		RestorePhasePreservingPrevious,
		RestorePhasePreviousPreserved,
		RestorePhasePublishing,
		RestorePhasePublished,
		RestorePhaseReopening,
		RestorePhaseCommitted,
	} {
		advanceTestRestoreJournal(t, root, &first, phase)
	}
	second := uniqueLedgerTestJournal(root, "reuse-stage")
	second.StagingPath = first.StagingPath
	if err := WriteRestoreJournal(root, second); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("reused artifact path error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func TestRestoreJournalLedgerMigratesLegacySingleRecordOnUpdate(t *testing.T) {
	root := restoreTestDir(t)
	journal := validTestRestoreJournal(root)
	legacy, err := json.Marshal(journal)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, restoreJournalFileName), legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	advanceTestRestoreJournal(t, root, &journal, RestorePhaseStaging)

	ledger, err := readRestoreJournalLedger(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(ledger.Transactions) != 1 || ledger.Transactions[0].Phase != RestorePhaseStaging {
		t.Fatalf("migrated ledger = %#v", ledger)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(mustReadFile(t, filepath.Join(root, restoreJournalFileName)), &document); err != nil {
		t.Fatal(err)
	}
	if _, ok := document["transactions"]; !ok {
		t.Fatal("journal write did not migrate the legacy single-record format")
	}
}

func TestRestoreJournalLedgerRejectsMultipleActiveRecords(t *testing.T) {
	root := restoreTestDir(t)
	first := validTestRestoreJournal(root)
	second := uniqueLedgerTestJournal(root, "second-active")
	data, err := json.Marshal(restoreJournalLedger{
		Version:      restoreJournalLedgerVersion,
		Transactions: []RestoreJournal{first, second},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, restoreJournalFileName), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRestoreJournal(root); !errors.Is(err, ErrRestoreJournalInvalid) {
		t.Fatalf("read multiple active records error = %v, want ErrRestoreJournalInvalid", err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
