package bankwatch

import (
	"os"
	"path/filepath"
	"testing"

	"codegym/backend/internal/store"
)

func bankJSON(id, title, code string) string {
	return `[
  {
    "id": "` + id + `",
    "title": "` + title + `",
    "description": "d",
    "exercise_type": "fill",
    "tech_stack": "testing",
    "weight": "easy",
    "category": "c",
    "subcategory": "pytest",
    "full_code": "` + code + `",
    "expected": "",
    "points": [
      {"id": "p1", "text": "` + code + `", "category": "c", "min_weight": "beginner", "weight_score": 1, "hint": "h", "fill_type": "text"}
    ],
    "hints": {"level_1": "a", "level_2": "b", "level_3": "c", "level_4": "d"},
    "explanation": "e"
  }
]`
}

func TestScanOnceLoadsChangedFiles(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(t.TempDir(), "w.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	file := filepath.Join(dir, "hot-bank.json")
	if err := os.WriteFile(file, []byte(bankJSON("hot-1", "第一版", "print(1)")), 0o644); err != nil {
		t.Fatal(err)
	}

	w := New(dir, 0, st)
	w.SetLogger(func(string, ...any) {})

	if n := w.ScanOnce(); n != 1 {
		t.Fatalf("first scan loaded %d files, want 1", n)
	}
	ex, err := st.GetExercise("hot-1")
	if err != nil || ex.Title != "第一版" {
		t.Fatalf("exercise = %+v (err %v)", ex, err)
	}
	if _, err := st.GetBankSource("hot-bank", "dir"); err != nil {
		t.Fatalf("bank source not recorded: %v", err)
	}

	// unchanged file must not reload
	if n := w.ScanOnce(); n != 0 {
		t.Fatalf("second scan loaded %d files, want 0", n)
	}

	// changed content -> reload with new title
	if err := os.WriteFile(file, []byte(bankJSON("hot-1", "第二版更长的标题", "print(2)")), 0o644); err != nil {
		t.Fatal(err)
	}
	if n := w.ScanOnce(); n != 1 {
		t.Fatalf("third scan loaded %d files, want 1", n)
	}
	ex, _ = st.GetExercise("hot-1")
	if ex.Title != "第二版更长的标题" {
		t.Fatalf("exercise title = %q, want 第二版更长的标题", ex.Title)
	}

	// broken file is skipped, then retried once fixed
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if n := w.ScanOnce(); n != 0 {
		t.Fatalf("broken scan loaded %d files, want 0", n)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte(bankJSON("hot-2", "修好了", "print(3)")), 0o644); err != nil {
		t.Fatal(err)
	}
	if n := w.ScanOnce(); n != 1 {
		t.Fatalf("fixed scan loaded %d files, want 1", n)
	}
	if _, err := st.GetExercise("hot-2"); err != nil {
		t.Fatalf("fixed file not loaded: %v", err)
	}
}
