package judge

import (
	"testing"

	"codegym/backend/internal/model"
)

func TestGradeForScore(t *testing.T) {
	cases := []struct {
		score float64
		want  string
	}{
		{100, "S"}, {90, "S"}, {89, "A"}, {75, "A"}, {74, "B"}, {60, "B"},
		{59, "C"}, {40, "C"}, {39, "D"}, {0, "D"},
	}
	for _, c := range cases {
		if got := GradeForScore(c.score); got != c.want {
			t.Errorf("GradeForScore(%v) = %q, want %q", c.score, got, c.want)
		}
	}
}

func TestIsCorrect(t *testing.T) {
	if !IsCorrect("x%2==0", "x % 2 == 0") {
		t.Error("whitespace-insensitive match should pass")
	}
	if IsCorrect("x%2==1", "x % 2 == 0") {
		t.Error("different expression must not pass")
	}
	if !IsPartiallyCorrect("x**2 * 1", "x**2") {
		t.Error("answer containing the expected text should be partially correct")
	}
	if IsPartiallyCorrect("n%2==0", "x % 2 == 0") {
		t.Error("unrelated expression must not be partially correct")
	}
}

func TestGradeFillHintPenalty(t *testing.T) {
	ex := model.Exercise{
		ExpectedOutput: "42",
		FillPoints: model.FillPointsConfig{
			Points: []model.FillPoint{
				{ID: "a", OriginalText: "x", WeightScore: 1},
			},
		},
	}
	points := []model.FillPoint{{ID: "a", OriginalText: "x", WeightScore: 1}}

	// no hint -> correct answer scores full
	results, total := GradeFill(ex, points, map[string]string{"FILL_a": "x"}, "42", "42", map[string]int{})
	if total != 100 {
		t.Errorf("expected 100 without hint, got %v", total)
	}
	if !results[0].IsCorrect {
		t.Error("correct answer should be marked correct")
	}

	// level 2 hint -> -5 penalty
	_, total = GradeFill(ex, points, map[string]string{"FILL_a": "x"}, "42", "42", map[string]int{"a": 2})
	if total != 95 {
		t.Errorf("expected 95 after L2 hint, got %v", total)
	}
}

func TestGradeFillWrongAnswer(t *testing.T) {
	ex := model.Exercise{
		ExpectedOutput: "42",
		FillPoints: model.FillPointsConfig{
			Points: []model.FillPoint{
				{ID: "a", OriginalText: "x", WeightScore: 1},
			},
		},
	}
	points := []model.FillPoint{{ID: "a", OriginalText: "x", WeightScore: 1}}
	results, total := GradeFill(ex, points, map[string]string{"FILL_a": "zzz"}, "43", "42", map[string]int{})
	if total != 0 {
		t.Errorf("expected 0 for wrong answer, got %v", total)
	}
	if results[0].IsCorrect {
		t.Error("wrong answer must not be correct")
	}
}

func TestGradeBugfix(t *testing.T) {
	ex := model.Exercise{
		FullCode: "def f():\n    return 1\n",
		TestCases: []model.TestCase{
			{Output: "120"},
			{Output: "6"},
		},
	}
	// all tests pass, code matches expected -> 100
	score, passed, _ := GradeBugfix(ex, "def f():\n    return 1\n", "120\n6", map[string]int{})
	if passed != 2 {
		t.Errorf("expected 2 passed, got %d", passed)
	}
	if score != 100 {
		t.Errorf("expected 100, got %v", score)
	}

	// broken fix -> low score, 0 passed
	score, passed, _ = GradeBugfix(ex, "def f():\n    return 0\n", "0\n0", map[string]int{})
	if passed != 0 {
		t.Errorf("expected 0 passed, got %d", passed)
	}
	if score != 0 {
		t.Errorf("expected 0, got %v", score)
	}
}
