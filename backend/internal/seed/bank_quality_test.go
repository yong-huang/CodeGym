package seed

import (
	"testing"
)

// TestBankPointIDsUnique verifies no exercise ships duplicate fill point ids.
// Answers are keyed by FILL_<id>, so duplicated ids would make two blanks
// share one answer key (grading and code rebuild both break).
func TestBankPointIDsUnique(t *testing.T) {
	all, err := All()
	if err != nil {
		t.Fatalf("load bank: %v", err)
	}
	for _, ex := range all {
		seen := map[string]bool{}
		for _, p := range ex.FillPoints.Points {
			if seen[p.ID] {
				t.Errorf("%s: duplicate fill point id %q", ex.ID, p.ID)
			}
			seen[p.ID] = true
		}
	}
}

// TestBankExecutableHasExpected guards the passability of runnable questions:
// a python/go fill exercise without expected_output can never be passed.
func TestBankExecutableHasExpected(t *testing.T) {
	all, err := All()
	if err != nil {
		t.Fatalf("load bank: %v", err)
	}
	for _, ex := range all {
		runnable := (ex.TechStack == "python" || ex.TechStack == "go") &&
			ex.ExerciseType == "fill" && ex.Subcategory != "FastAPI"
		if runnable && ex.ExpectedOutput == "" {
			t.Errorf("%s (%s): executable fill without expected_output", ex.ID, ex.TechStack)
		}
	}
}
