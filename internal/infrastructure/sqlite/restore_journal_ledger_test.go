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


func completeLedgerTestTransaction(t *testing.T, root string, journal *RestoreJournal) {
	t.Helper()
	for _, phase := range []RestorePhase{
		RestorePhaseStaging,
		RestorePhasePreservingPrevious,
		RestorePhasePreviousPreserved,
		RestorePhasePublishing,
		RestorePhasePublished,
		RestorePhaseReopening,
		RestorePhaseCommitted,
	} {
		advanceTestRestoreJournal(t, root, journal, phase)
	}
}

func TestRestoreJournalLedgerRetainsSequentialTerminalTransactions(t *testing.T) {
	root := restoreTestDir(t)

	first := uniqueLedgerTestJournal(root, "first-terminal")
	if err := WriteRestoreJournal(root, first); err != nil {
		t.Fatalf("write first transaction: %v", err)
	}
	completeLedgerTestTransaction(t, root, &first)

	second := uniqueLedgerTestJournal(root, "second-terminal")
	if err := WriteRestoreJournal(root, second); err != nil {
		t.Fatalf("start second transaction after terminal first transaction: %v", err)
	}
	active, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read active second transaction: %v", err)
	}
	if active.TransactionID != second.TransactionID || active.Phase != RestorePhasePrepared {
		t.Fatalf("active journal = %q/%q, want %q/prepared", active.TransactionID, active.Phase, second.TransactionID)
	}
	completeLedgerTestTransaction(t, root, &second)

	ledger, err := readRestoreJournalLedger(root)
	if err != nil {
		t.Fatalf("read retained journal ledger: %v", err)
	}
	if len(ledger.Transactions) != 2 {
		t.Fatalf("retained transaction count = %d, want 2", len(ledger.Transactions))
	}
	for i, want := range []RestoreJournal{first, second} {
		got := ledger.Transactions[i]
		if got.TransactionID != want.TransactionID || got.Phase != RestorePhaseCommitted {
			t.Fatalf("retained transaction[%d] = %q/%q, want %q/committed", i, got.TransactionID, got.Phase, want.TransactionID)
		}
	}

	current, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read latest terminal transaction: %v", err)
	}
	if current.TransactionID != second.TransactionID || current.Phase != RestorePhaseCommitted {
		t.Fatalf("latest terminal journal = %q/%q, want %q/committed", current.TransactionID, current.Phase, second.TransactionID)
	}

	if err := RemoveRestoreJournal(root); !errors.Is(err, errRestoreJournalRemovalUnsupported) {
		t.Fatalf("retire terminal journal error = %v, want retained-history error", err)
	}
	afterRetire, err := readRestoreJournalLedger(root)
	if err != nil {
		t.Fatalf("read ledger after rejected retirement: %v", err)
	}
	if len(afterRetire.Transactions) != 2 {
		t.Fatalf("transaction history after retirement attempt = %d, want 2", len(afterRetire.Transactions))
	}
}

func TestRestoreJournalLedgerIgnoresAndPreservesUnpublishedTemporaryFile(t *testing.T) {
	root := restoreTestDir(t)
	journal := uniqueLedgerTestJournal(root, "published")
	if err := WriteRestoreJournal(root, journal); err != nil {
		t.Fatalf("write published journal: %v", err)
	}

	temporaryPath := filepath.Join(root, ".restore-journal-crash-before-publish.tmp")
	if err := os.WriteFile(temporaryPath, []byte("{incomplete"), 0o600); err != nil {
		t.Fatalf("write simulated orphan temporary journal: %v", err)
	}

	got, err := ReadRestoreJournal(root)
	if err != nil {
		t.Fatalf("read durable journal with orphan temporary file: %v", err)
	}
	if got.TransactionID != journal.TransactionID || got.Phase != RestorePhasePrepared {
		t.Fatalf("durable journal = %q/%q, want %q/prepared", got.TransactionID, got.Phase, journal.TransactionID)
	}
	if _, err := os.Lstat(temporaryPath); err != nil {
		t.Fatalf("orphan temporary evidence was removed or became unreadable: %v", err)
	}
}
