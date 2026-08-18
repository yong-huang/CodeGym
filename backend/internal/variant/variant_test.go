package variant

import (
	"testing"

	"codegym/backend/internal/model"
)

func testPoints() model.FillPointsConfig {
	return model.FillPointsConfig{
		MinFills: 1,
		MaxFills: 4,
		Points: []model.FillPoint{
			{ID: "p1", MinWeight: "beginner"},
			{ID: "p2", MinWeight: "easy"},
			{ID: "p3", MinWeight: "medium"},
			{ID: "p4", MinWeight: "hard"},
		},
	}
}

func TestSelectPointsBounds(t *testing.T) {
	fp := testPoints()
	available := len(fp.Points)
	for _, lvl := range []string{"beginner", "easy", "medium", "hard", "expert"} {
		for i := 0; i < 50; i++ {
			sel := SelectPoints(fp, lvl)
			if len(sel) == 0 {
				t.Fatalf("%s: no points selected", lvl)
			}
			cfg := model.WeightLevels[lvl]
			lo := cfg.MinFills
			hi := cfg.MaxFills
			if hi > available {
				hi = available // cannot exceed the number of available points
			}
			if lo > available {
				lo = available
			}
			if len(sel) < lo || len(sel) > hi {
				t.Errorf("%s: selected %d, want between %d and %d", lvl, len(sel), lo, hi)
			}
		}
	}
}

func TestSelectPointsEligibility(t *testing.T) {
	fp := testPoints()
	// beginner level should only include points with min_weight beginner
	// (p1 only, unless always_fill)
	sel := SelectPoints(fp, "beginner")
	for _, p := range sel {
		if p.ID != "p1" {
			t.Errorf("beginner selected ineligible point %s", p.ID)
		}
	}

	// hard level should be able to pick every point
	if len(SelectPoints(fp, "hard")) != 4 {
		t.Errorf("hard should select all 4 points")
	}
}

func TestSelectPointsAlwaysFill(t *testing.T) {
	fp := model.FillPointsConfig{
		MinFills: 1,
		MaxFills: 2,
		Points: []model.FillPoint{
			{ID: "always", MinWeight: "expert", AlwaysFill: true},
			{ID: "p1", MinWeight: "medium"},
		},
	}
	for i := 0; i < 20; i++ {
		sel := SelectPoints(fp, "beginner")
		found := false
		for _, p := range sel {
			if p.ID == "always" {
				found = true
			}
		}
		if !found {
			t.Error("always_fill point must be selected at any level")
		}
	}
}

func TestApplyAnswers(t *testing.T) {
	code := "a = foo(x)\nprint(a)"
	points := []model.FillPoint{
		{ID: "fn", Location: model.FillLocation{SpanStart: 4, SpanEnd: 7}, OriginalText: "foo"},
	}
	got := ApplyAnswers(code, points, map[string]string{"FILL_fn": "bar"})
	want := "a = bar(x)\nprint(a)"
	if got != want {
		t.Errorf("ApplyAnswers = %q, want %q", got, want)
	}

	// missing answer falls back to original text
	got = ApplyAnswers(code, points, nil)
	if got != code {
		t.Errorf("ApplyAnswers fallback = %q, want %q", got, code)
	}
}
