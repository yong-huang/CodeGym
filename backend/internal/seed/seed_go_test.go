package seed

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"codegym/backend/internal/model"
	"codegym/backend/internal/variant"
)

// compile exercises a full Go program by staging a temp module and running
// `go build .` against the given source (buggyCode when present).
func compile(t *testing.T, src string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module codegymverify\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed:\n%s\ncode:\n%s", out, src)
	}
}

// TestGoBasicsCompile verifies every runnable Go exercise (fill + bugfix, plus
// their buggy variants) actually compiles, so the bank is guaranteed valid.
func TestGoBasicsCompile(t *testing.T) {
	exs, err := buildBank("", "go")
	if err != nil {
		t.Fatalf("build GoBasics: %v", err)
	}
	if len(exs) != 50 {
		t.Fatalf("expected 50 Go basic questions, got %d", len(exs))
	}
		for _, ex := range exs {
			if ex.TechStack != "go" {
				continue
			}
			t.Run(ex.ID, func(t *testing.T) {
				compile(t, ex.FullCode)
				if ex.ExerciseType == model.ExerciseTypeBugfix && ex.BuggyCode != "" {
					compile(t, ex.BuggyCode)
				}
			})
		}
}

// runGoResult stages a module and runs `go run .`, returning stdout.
func runGoResult(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeFile("go.mod", "module codegymrun\n\ngo 1.22\n")
	writeFile("main.go", src)
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run failed:\n%s\ncode:\n%s", out, src)
	}
	return strings.TrimSpace(string(out))
}

// TestGoConcurrencyRun executes every Go concurrency fill/bugfix full code
// and verifies the printed (deterministic) output matches the expected value.
func TestGoConcurrencyRun(t *testing.T) {
	exs, err := buildBank("", "go-concurrency")
	if err != nil {
		t.Fatalf("build GoConcurrency: %v", err)
	}
	if len(exs) != 50 {
		t.Fatalf("expected 50 Go concurrency questions, got %d", len(exs))
	}
	for _, ex := range exs {
		t.Run(ex.ID, func(t *testing.T) {
			got := runGoResult(t, ex.FullCode)
			if strings.TrimSpace(strings.Join(strings.Fields(got), " ")) != strings.TrimSpace(strings.Join(strings.Fields(ex.ExpectedOutput), " ")) {
				t.Errorf("output = %q, want %q", got, ex.ExpectedOutput)
			}
		})
	}
}

// TestEveryLevelHasFill checks that every exercise yields at least one fill
// point at EVERY weight level (no "empty" broken questions when switching).
func TestEveryLevelHasFill(t *testing.T) {
	all, err := All()
	if err != nil {
		t.Fatal(err)
	}
	levels := []string{"beginner", "easy", "medium", "hard", "expert"}
	for _, ex := range all {
		for _, lvl := range levels {
			points := variant.SelectPoints(ex.FillPoints, lvl)
			if len(points) == 0 {
				t.Errorf("%s has no fill at level %s (total points=%d)", ex.ID, lvl, len(ex.FillPoints.Points))
			}
		}
	}
}

// TestWeightSwitchChangesFills verifies that at least for multi-point
// exercises, switching to a harder level can fill more points, i.e. the
// weight control actually does something.
func TestWeightSwitchChangesFills(t *testing.T) {
	all, err := All()
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	multi := 0
	for _, ex := range all {
		if len(ex.FillPoints.Points) < 2 {
			continue
		}
		multi++
		beginner := len(variant.SelectPoints(ex.FillPoints, "beginner"))
		hard := len(variant.SelectPoints(ex.FillPoints, "hard"))
		if hard > beginner {
			changed++
		}
	}
	if changed == 0 {
		t.Fatalf("no multi-point exercise changes fill count between beginner and hard (multi=%d)", multi)
	}
	t.Logf("multi-point exercises=%d, of which weight-change increases fills=%d", multi, changed)
}
