package store

import (
	"path/filepath"
	"testing"

	"codegym/backend/internal/model"
)

func TestStreakNotInflatedBySameDayAttempts(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	ex := model.Exercise{
		ID: "ex-test", Title: "t", FullCode: "print(1)", TechStack: "python",
		Weight: "easy", ExerciseType: model.ExerciseTypeFill, Enabled: true,
		TimeLimitSecs: 30, Subcategory: "基础",
		FillPoints: model.FillPointsConfig{Points: []model.FillPoint{}},
	}
	if err := st.UpsertExercise(ex); err != nil {
		t.Fatal(err)
	}

	// many attempts on the same (single) day
	for i := 0; i < 50; i++ {
		if err := st.UpdateProgress(ex, 100, "S", 0, true, "u1"); err != nil {
			t.Fatal(err)
		}
	}
	p, err := st.GetProgress("u1")
	if err != nil {
		t.Fatal(err)
	}
	if p.TotalAttempts != 50 {
		t.Errorf("total_attempts = %d, want 50", p.TotalAttempts)
	}
	if p.CurrentStreak != 1 {
		t.Errorf("current_streak = %d, want 1 (same-day attempts must not inflate)", p.CurrentStreak)
	}
	if p.LongestStreak != 1 {
		t.Errorf("longest_streak = %d, want 1", p.LongestStreak)
	}
}
