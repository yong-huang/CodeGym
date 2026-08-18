package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"codegym/backend/internal/api"
	"codegym/backend/internal/engine"
	"codegym/backend/internal/model"
	"codegym/backend/internal/seed"
	"codegym/backend/internal/store"
)

func main() {
	addr := flag.String("addr", ":8090", "listen address")
	dbPath := flag.String("db", "data/codegym.db", "sqlite database path")
	pyBin := flag.String("python", firstNonEmpty(os.Getenv("PYTHON_BIN"), "python3"), "python interpreter")
	libDir := flag.String("library", "data/library", "内置题库 JSON 目录（可指向外部目录；留空则用内置构建器）")
	flag.Parse()

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	if err := seedExercises(st, *libDir); err != nil {
		log.Printf("seed: %v", err)
	}
	if err := seedCategories(st); err != nil {
		log.Printf("seed categories: %v", err)
	}
	if err := seedEnvironments(st); err != nil {
		log.Printf("seed environments: %v", err)
	}

	eng := engine.New(*pyBin)
	server := api.New(st, eng)

	log.Printf("CodeGym API listening on %s (db=%s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}

// seedExercises loads the question bank: prefer the JSON library directory
// (editable without recompiling), falling back to the built-in builders.
// Only the exercises table is touched; progress/favorites are never modified.
func seedExercises(st *store.Store, libDir string) error {
	exs, err := seed.LoadLibrary(libDir)
	if err != nil {
		return err
	}
	if len(exs) == 0 {
		exs, err = seed.All()
		if err != nil {
			return err
		}
		log.Printf("library dir 为空，使用内置题库构建器")
	} else {
		log.Printf("从题库 JSON 目录加载 %d 题（%s）", len(exs), libDir)
	}
	for _, ex := range exs {
		if err := st.UpsertExercise(ex); err != nil {
			return err
		}
	}
	log.Printf("seeded %d exercises", len(exs))
	return nil
}

// seedCategories creates the default managed subcategories (idempotent).
func seedCategories(st *store.Store) error {
	for _, c := range seed.CategoryPlan() {
		if err := st.CreateCategory(c.Tech, c.Sub); err != nil {
			return err
		}
	}
	log.Printf("seeded %d categories", len(seed.CategoryPlan()))
	return nil
}

// seedEnvironments creates the default execution environments (idempotent).
func seedEnvironments(st *store.Store) error {
	defaults := []model.Environment{
		{Name: "python-3.10", Type: "local", Config: map[string]any{"interpreter": "python3", "timeout": "30s"}, HealthStatus: "healthy", MaxConcurrent: 5, Enabled: true},
		{Name: "go-1.26", Type: "local", Config: map[string]any{"toolchain": "go", "timeout": "30s"}, HealthStatus: "healthy", MaxConcurrent: 5, Enabled: true},
		{Name: "answer-match", Type: "local", Config: map[string]any{"description": "command/config/非 python-go 答案匹配判题"}, HealthStatus: "healthy", MaxConcurrent: 10, Enabled: true},
	}
	for _, e := range defaults {
		if err := st.UpsertEnvironment(e); err != nil {
			return err
		}
	}
	log.Printf("seeded %d environments", len(defaults))
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
