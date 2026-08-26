package seed

import (
	"os/exec"
	"strings"
	"testing"
)

// runPython executes a complete program with python3 (via -c).
func runPython(t *testing.T, code string) string {
	t.Helper()
	out, err := exec.Command("python3", "-c", code).CombinedOutput()
	if err != nil {
		t.Fatalf("python run failed:\n%s\ncode:\n%s", out, code)
	}
	return strings.TrimSpace(string(out))
}

func norm(s string) string {
	// strip all whitespace for comparison
	var b strings.Builder
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// TestPythonBasicsRun executes every Python basic fill exercise and verifies
// the printed output matches the expected value, so the whole bank is valid
// and the seeding is correct.
func TestPythonBasicsRun(t *testing.T) {
	exs, err := buildBank("", "python")
	if err != nil {
		t.Fatalf("build python bank: %v", err)
	}
	if len(exs) != 50 {
		t.Fatalf("expected 50 Python basic questions, got %d", len(exs))
	}
	for _, ex := range exs {
		t.Run(ex.ID, func(t *testing.T) {
			if ex.ExerciseType == "bugfix" {
				return // bugfix uses test cases
			}
			got := runPython(t, ex.FullCode)
			if norm(got) != norm(ex.ExpectedOutput) {
				t.Errorf("output = %q, want %q", got, ex.ExpectedOutput)
			}
		})
	}
}
