package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"codegym/backend/internal/api"
	"codegym/backend/internal/engine"
	"codegym/backend/internal/seed"
	"codegym/backend/internal/store"
)

func main() {
	addr := flag.String("addr", ":8090", "listen address")
	dbPath := flag.String("db", "data/codegym.db", "sqlite database path")
	pyBin := flag.String("python", firstNonEmpty(os.Getenv("PYTHON_BIN"), "python3"), "python interpreter")
	flag.Parse()

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	if err := seedExercises(st); err != nil {
		log.Printf("seed: %v", err)
	}
	if err := seedCategories(st); err != nil {
		log.Printf("seed categories: %v", err)
	}

	eng := engine.New(*pyBin)
	server := api.New(st, eng)

	log.Printf("CodeGym API listening on %s (db=%s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}

// seedExercises loads the built-in exercise library on first boot (idempotent).
func seedExercises(st *store.Store) error {
	exs, err := seed.All()
	if err != nil {
		return err
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

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
