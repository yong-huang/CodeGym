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
	bankDir := flag.String("bank", "data/bank", "题库 JSON 目录（唯一数据源，可用 $CODEGYM_BANK_DIR 或留空自动定位）")
	adminToken := flag.String("admin-token", os.Getenv("CODEGYM_ADMIN_TOKEN"), "教练端接口令牌（启用后所有 /api/v1/admin/* 需携带 X-Admin-Token 或 Bearer；为空不鉴权，仅建议本地开发使用）")
	flag.Parse()

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	if err := seedExercises(st, *bankDir); err != nil {
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
	server.SetAdminToken(*adminToken)

	log.Printf("CodeGym API listening on %s (db=%s, admin-auth=%v)", *addr, *dbPath, *adminToken != "")
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}

// seedExercises loads the whole question bank from the JSON library directory
// (the single source of truth; editable without recompiling) and upserts it.
// Only the exercises table is touched; progress/favorites are never modified.
func seedExercises(st *store.Store, bankDir string) error {
	exs, err := seed.BuildAll(bankDir)
	if err != nil {
		return err
	}
	log.Printf("从题库 JSON 目录加载 %d 题（%s）", len(exs), bankDir)
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
