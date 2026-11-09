package sqlite

import (
	"errors"
	"testing"
)

// PreservingPrevious is entered only after the candidate has been completely
// staged and its identity recorded. Losing staging in this phase is not a
// recoverable "resume intent" state: it would strand the previous DB in the
// rollback slot without a candidate to publish.
func TestPlanRestoreRecoveryPreservingPreviousRequiresStagedCandidate(t *testing.T) {
	tests := []struct {
		name  string
		files RestoreRecoveryFiles
	}{
		{
			name:  "previous still at target but candidate disappeared",
			files: recoveryFiles(RestoreFilePrevious, RestoreFileMissing, RestoreFileMissing, RestoreFileMissing),
		},
		{
			name:  "previous preserved but candidate disappeared",
			files: recoveryFiles(RestoreFileMissing, RestoreFileMissing, RestoreFilePrevious, RestoreFileMissing),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			journal := recoveryTestJournal(RestorePhasePreservingPrevious, true)
			plan, err := PlanRestoreRecovery(journal, test.files)
			if !errors.Is(err, ErrRestoreRecoveryAmbiguous) {
				t.Fatalf("PlanRestoreRecovery() = %#v, %v; want ErrRestoreRecoveryAmbiguous", plan, err)
			}
			if plan != (RestoreRecoveryPlan{}) {
				t.Fatalf("ambiguous state returned actionable plan: %#v", plan)
			}
		})
	}
}
