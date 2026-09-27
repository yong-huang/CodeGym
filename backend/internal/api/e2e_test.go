package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codegym/backend/internal/api"
	"codegym/backend/internal/engine"
	"codegym/backend/internal/model"
	"codegym/backend/internal/seed"
	"codegym/backend/internal/store"
)

// newTestServer spins up a real server (SQLite temp DB + local python engine).
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "e2e.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	exs, err := seed.All()
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	for _, ex := range exs {
		if err := st.UpsertExercise(ex); err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}
	for _, env := range []model.Environment{
		{Name: "python-3.10", Type: "local", Config: map[string]any{"interpreter": "python3"}, HealthStatus: "healthy", MaxConcurrent: 5, Enabled: true},
		{Name: "go-1.26", Type: "local", Config: map[string]any{"toolchain": "go"}, HealthStatus: "healthy", MaxConcurrent: 5, Enabled: true},
	} {
		if err := st.UpsertEnvironment(env); err != nil {
			t.Fatalf("upsert env: %v", err)
		}
	}
	srv := api.New(st, engine.New("python3"))
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(func() {
		ts.Close()
		st.Close()
	})
	return ts
}

func doJSON(t *testing.T, method, url string, body any, out any) *http.Response {
	t.Helper()
	return doJSONAuth(t, method, url, body, out, "")
}

// doJSONAuth is doJSON with a session token (Authorization: Bearer).
func doJSONAuth(t *testing.T, method, url string, body any, out any, token string) *http.Response {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, rd)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do %s %s: %v", method, url, err)
	}
	if out != nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode %s %s: %v", method, url, err)
		}
	} else {
		// drain and close so the server finishes its writes
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	return resp
}

// registerUser registers a fresh trainee and returns (user id, session token).
func registerUser(t *testing.T, ts *httptest.Server) (string, string) {
	t.Helper()
	var out struct {
		Token string `json:"token"`
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	username := fmt.Sprintf("u%d", time.Now().UnixNano()%1_000_000_000_000)
	resp := doJSON(t, "POST", ts.URL+"/api/v1/auth/register",
		map[string]any{"username": username, "password": "secret123"}, &out)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: status %d", username, resp.StatusCode)
	}
	if out.Token == "" || out.User.ID == "" {
		t.Fatalf("register returned empty token/user: %+v", out)
	}
	return out.User.ID, out.Token
}

func TestE2E_BareAnswerKeysAccepted(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// The browser UI submits answers keyed by bare point id (no FILL_ prefix).
	// These must be graded correctly, not treated as empty/missing.
	var res struct {
		Score        float64 `json:"score"`
		Grade        string  `json:"grade"`
		IsCorrect    bool    `json:"is_correct"`
		WeakAnalysis string  `json:"weak_analysis"`
		FillResults  []struct {
			PointID    string `json:"point_id"`
			UserAnswer string `json:"user_answer"`
			IsCorrect  bool   `json:"is_correct"`
		} `json:"fill_results"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-palindrome-002",
		"weight_level": "medium",
		"fill_answers": map[string]string{
			"return_expr": "s[::-1]",
			"return_type": "-> bool",
			"compare":     "s == s[::-1]",
			"call_arg":    "'hello'",
		},
	}, &res)
	if res.Score != 100 || res.Grade != "S" || !res.IsCorrect {
		t.Errorf("bare keys => score=%v grade=%q correct=%v, want 100/S/true", res.Score, res.Grade, res.IsCorrect)
	}
	for _, fr := range res.FillResults {
		if fr.UserAnswer == "" {
			t.Errorf("fill %s: answer lost (empty user_answer)", fr.PointID)
		}
		if !fr.IsCorrect {
			t.Errorf("fill %s not correct with correct bare-keyed answer", fr.PointID)
		}
	}

	// A stale hint_usage keyed by point id must also keep working.
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "kubectl-getpods-005",
		"weight_level": "easy",
		"fill_answers": map[string]string{"flag": "-A"},
		"hint_usage":   map[string]int{"flag": 4},
	}, &res)
	if res.Score != 70 {
		t.Errorf("bare key + hint => score=%v, want 70", res.Score)
	}

	// A broken fill (e.g. quotes omitted on a string arg) makes the run fail;
	// the response must still carry a graded score, grade and weak analysis.
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-palindrome-002",
		"weight_level": "hard",
		"fill_answers": map[string]string{
			"return_expr": "s[::-1]",
			"return_type": "-> bool",
			"compare":     "s == s[::-1]",
			"call_arg":    "hello", // missing quotes -> NameError
		},
	}, &res)
	if res.Score == 0 || res.IsCorrect {
		t.Errorf("broken-run grading => score=%v correct=%v, want partial score + not correct", res.Score, res.IsCorrect)
	}
	if res.Grade == "" {
		t.Error("grade must be set even when execution fails")
	}
	if res.WeakAnalysis == "" {
		t.Error("weak analysis must be present when a fill is wrong")
	}
}

func TestE2E_OpenShift(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// exactly 50 openshift basics in the openshift tech stack
	var list []struct {
		ID          string `json:"id"`
		Subcategory string `json:"subcategory"`
	}
	doJSON(t, "GET", base+"/exercises?tech_stack=openshift", nil, &list)
	nOc := 0
	for _, it := range list {
		if strings.HasPrefix(it.ID, "oc-basics-") {
			nOc++
		}
	}
	if nOc != 50 {
		t.Fatalf("oc-basics count = %d, want 50", nOc)
	}

	// a command question (answer-match, alias for subcommand)
	var res struct {
		Score     float64 `json:"score"`
		IsCorrect bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "oc-basics-001",
		"weight_level": "beginner",
		"fill_answers": map[string]string{"FILL_verb": "whoami"},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("oc command => score=%v correct=%v", res.Score, res.IsCorrect)
	}

	// a config YAML question
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "oc-basics-042",
		"weight_level": "medium",
		"fill_answers": map[string]string{"FILL_host": "myapp.example.com", "FILL_svc": "myapp", "FILL_port": "8080"},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("oc config => score=%v correct=%v", res.Score, res.IsCorrect)
	}

	// OpenShift 对比 Kubernetes bank: exactly 30 under 对比 subcategory
	var diffList []struct {
		ID          string `json:"id"`
		Subcategory string `json:"subcategory"`
	}
	doJSON(t, "GET", base+"/exercises?tech_stack=openshift", nil, &diffList)
	nDiff := 0
	for _, it := range diffList {
		if it.Subcategory == "对比" {
			nDiff++
		}
	}
	if nDiff != 30 {
		t.Fatalf("openshift 对比 count = %d, want 30", nDiff)
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "oc-diff-001",
		"weight_level": "beginner",
		"fill_answers": map[string]string{"FILL_res": "dc"},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("oc-diff-001 => score=%v correct=%v", res.Score, res.IsCorrect)
	}
}

func TestE2E_Categories(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1/admin"

	// list includes 基础 and the default plan
	var cats []model.Category
	doJSON(t, "GET", base+"/categories", nil, &cats)
	if len(cats) == 0 {
		t.Fatal("no categories returned")
	}
	foundBasis := false
	for _, c := range cats {
		if c.TechStack == "go" && c.Subcategory == "基础" {
			foundBasis = true
		}
	}
	if !foundBasis {
		t.Error("go/基础 category missing")
	}

	// create an empty category
	doJSON(t, "POST", base+"/categories", map[string]any{"tech_stack": "go", "subcategory": "框架"}, nil)

	// rename it and verify exercises move along
	doJSON(t, "PUT", base+"/categories", map[string]any{"tech_stack": "go", "subcategory": "基础", "new_subcategory": "入门"}, nil)

	// a go exercise should now be under 入门
	var list []struct {
		ID          string `json:"id"`
		Subcategory string `json:"subcategory"`
	}
	doJSON(t, "GET", ts.URL+"/api/v1/exercises?tech_stack=go", nil, &list)
	moved := 0
	for _, it := range list {
		if it.Subcategory == "入门" {
			moved++
		}
	}
	if moved == 0 {
		t.Errorf("rename did not move go exercises to 入门")
	}

	// delete 资源配置 (k8s) -> its exercises move back to 基础
	doJSON(t, "DELETE", base+"/categories", map[string]any{"tech_stack": "kubernetes", "subcategory": "资源配置"}, nil)
	doJSON(t, "GET", ts.URL+"/api/v1/exercises?tech_stack=kubernetes", nil, &list)
	stillRes := 0
	for _, it := range list {
		if it.Subcategory == "资源配置" {
			stillRes++
		}
	}
	if stillRes != 0 {
		t.Errorf("delete category did not reassign k8s exercises, still %d", stillRes)
	}

	// deleting 基础 is rejected
	resp := doJSON(t, "DELETE", base+"/categories", map[string]any{"tech_stack": "go", "subcategory": "基础"}, nil)
	if resp.StatusCode == 200 {
		t.Error("deleting default 基础 should be rejected")
	}
}

func TestE2E_Favorites(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"
	_, token := registerUser(t, ts)

	// toggle on
	var ack struct {
		Favorited bool `json:"favorited"`
	}
	doJSONAuth(t, "POST", base+"/favorites", map[string]any{"exercise_id": "go-basics-001"}, &ack, token)
	if !ack.Favorited {
		t.Error("first toggle should favorite")
	}
	// toggle off
	doJSONAuth(t, "POST", base+"/favorites", map[string]any{"exercise_id": "go-basics-001"}, &ack, token)
	if ack.Favorited {
		t.Error("second toggle should unfavorite")
	}
	// favorite two, then list
	doJSONAuth(t, "POST", base+"/favorites", map[string]any{"exercise_id": "go-basics-001"}, &ack, token)
	doJSONAuth(t, "POST", base+"/favorites", map[string]any{"exercise_id": "py-lc-001"}, &ack, token)

	// list must honor the session token (per-user isolation)
	req, _ := http.NewRequest("GET", base+"/favorites", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var items []model.FavoriteItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("favorites = %d, want 2", len(items))
	}
	found := false
	for _, it := range items {
		if it.ExerciseID == "go-basics-001" && it.Title != "" {
			found = true
		}
	}
	if !found {
		t.Error("go-basics-001 not in favorites or missing title")
	}

	// a different user must see no favorites (isolation)
	_, otherToken := registerUser(t, ts)
	req2, _ := http.NewRequest("GET", base+"/favorites", nil)
	req2.Header.Set("Authorization", "Bearer "+otherToken)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("get other-user favorites: %v", err)
	}
	defer resp2.Body.Close()
	var other []model.FavoriteItem
	json.NewDecoder(resp2.Body).Decode(&other)
	if len(other) != 0 {
		t.Errorf("other user favorites = %d, want 0", len(other))
	}
}

func TestE2E_ExerciseListAndVariant(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// 1. list exercises
	var list []struct {
		ID           string `json:"id"`
		Title        string `json:"title"`
		ExerciseType string `json:"exercise_type"`
		TechStack    string `json:"tech_stack"`
		PointCount   int    `json:"point_count"`
	}
	doJSON(t, "GET", base+"/exercises", nil, &list)
	if len(list) != 1466 {
		t.Fatalf("expected 1466 exercises (816 + 180 AWS + 90 Python 进阶 + 200 OpenStack + 180 测试框架), got %d", len(list))
	}
	var ids []string
	for _, ex := range list {
		if ex.ID == "" || ex.Title == "" || ex.PointCount == 0 {
			t.Errorf("list item incomplete: %+v", ex)
		}
		ids = append(ids, ex.ID)
	}
	if !contains(ids, "py-factorial-bug-003") {
		t.Error("bugfix exercise missing from list")
	}
	if !contains(ids, "go-concurrent-106") {
		t.Error("Go exercise library missing from list")
	}

	// 2. filter by tech stack
	var pyList []map[string]any
	doJSON(t, "GET", base+"/exercises?tech_stack=python", nil, &pyList)
	if len(pyList) != 264 {
		t.Errorf("expected 264 python exercises (54 基础 + 30 FastAPI + 20 协程 + 20 多线程 + 90 进阶 + 50 Agent 开发), got %d", len(pyList))
	}

	var dsList []map[string]any
	doJSON(t, "GET", base+"/exercises?tech_stack=datastructure", nil, &dsList)
	if len(dsList) != 60 {
		t.Errorf("expected 60 datastructure exercises (6 子分类各 10 题), got %d", len(dsList))
	}

	// 3. variant masking + weight bounds
	for _, lvl := range []string{"beginner", "easy", "medium", "hard", "expert"} {
		var v struct {
			ExerciseID  string `json:"exercise_id"`
			WeightLevel string `json:"weight_level"`
			VariantID   string `json:"variant_id"`
			MarkedCode  string `json:"marked_code"`
			FillPoints  []struct {
				ID           string `json:"id"`
				OriginalText string `json:"original_text"` // must stay empty
				HintLevel4   string `json:"hint_level_4"`
			} `json:"fill_points"`
		}
		doJSON(t, "GET", base+"/exercises/py-lc-001/variant?level="+lvl, nil, &v)
		if v.ExerciseID != "py-lc-001" || v.WeightLevel != lvl || v.VariantID == "" {
			t.Errorf("variant metadata wrong: %+v", v)
		}
		if !strings.Contains(v.MarkedCode, "____FILL_") {
			t.Error("marked_code does not contain fill tokens")
		}
		// answers must NOT be exposed in the variant payload (only in hint_level_4)
		for _, p := range v.FillPoints {
			if p.OriginalText != "" {
				t.Errorf("%s: original_text leaked for %s", lvl, p.ID)
			}
			if !strings.Contains(p.HintLevel4, "正确答案") {
				t.Errorf("%s: hint_level_4 missing answer for %s", lvl, p.ID)
			}
		}
		// number of masks equals number of fill points
		if got := strings.Count(v.MarkedCode, "____FILL_"); got != len(v.FillPoints) {
			t.Errorf("%s: %d tokens but %d fill points", lvl, got, len(v.FillPoints))
		}
		// beginner level must fill at least one point
		if lvl == "beginner" && len(v.FillPoints) < 1 {
			t.Error("beginner level should select at least one fill point")
		}
	}
}

func TestE2E_FillExerciseCorrectFlow(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// execute a fully correct fill attempt
	var res struct {
		ExecutionID string  `json:"execution_id"`
		Success     bool    `json:"success"`
		Output      string  `json:"output"`
		IsCorrect   bool    `json:"is_correct"`
		Score       float64 `json:"score"`
		Grade       string  `json:"grade"`
		FillResults []struct {
			ID        string `json:"id"`
			IsCorrect bool   `json:"is_correct"`
		} `json:"fill_results"`
		ExecutionTimeMS int64 `json:"execution_time_ms"`
	}
	resp := doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-lc-001",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_expr": "x**2", "FILL_iter": "numbers"},
	}, &res)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("execute status = %d", resp.StatusCode)
	}
	if !res.Success {
		t.Error("execute not marked success")
	}
	if res.Output != "[4, 16, 36, 64, 100]" {
		t.Errorf("output = %q", res.Output)
	}
	if !res.IsCorrect {
		t.Error("correct attempt not marked correct")
	}
	if res.Score != 100 {
		t.Errorf("score = %v, want 100", res.Score)
	}
	if res.Grade != "S" {
		t.Errorf("grade = %q, want S", res.Grade)
	}
	if res.ExecutionID == "" {
		t.Error("execution_id missing")
	}
	if res.ExecutionTimeMS < 0 {
		t.Error("negative execution time")
	}
	if len(res.FillResults) != 2 {
		t.Errorf("expected 2 fill results, got %d", len(res.FillResults))
	}
	for _, fr := range res.FillResults {
		if !fr.IsCorrect {
			t.Errorf("fill %s not correct", fr.ID)
		}
	}
}

func TestE2E_FillExerciseWrongAndHint(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// wrong expression -> fail, low score
	var wrong struct {
		Score     float64 `json:"score"`
		IsCorrect bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-lc-001",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_expr": "x*2", "FILL_iter": "numbers"},
	}, &wrong)
	if wrong.IsCorrect {
		t.Error("wrong attempt marked correct")
	}
	if wrong.Score >= 100 {
		t.Errorf("wrong attempt score = %v", wrong.Score)
	}

	// hint usage deducts points from a fully-correct attempt
	var hinted struct {
		Score float64 `json:"score"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-lc-001",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_expr": "x**2", "FILL_iter": "numbers"},
		"hint_usage":   map[string]int{"expr": 2},
	}, &hinted)
	// expr(0.6w, -5) + iter(0.4w, 0) => (95*0.6+100*0.4) = 97
	if hinted.Score != 97 {
		t.Errorf("hint penalty not applied: score = %v, want 97", hinted.Score)
	}

	// L4 answer reveal costs -30 on the single-point command exercise
	var revealed struct {
		Score float64 `json:"score"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "kubectl-getpods-005",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_flag": "-A"},
		"hint_usage":   map[string]int{"flag": 4},
	}, &revealed)
	if revealed.Score != 70 {
		t.Errorf("L4 reveal score = %v, want 70", revealed.Score)
	}
}

func TestE2E_BugfixFlow(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// correct fix -> S / 100
	correctFix := "def factorial(n):\n    result = 1\n    for i in range(1, n + 1):\n        result *= i\n    return result"
	var good struct {
		Output    string  `json:"output"`
		IsCorrect bool    `json:"is_correct"`
		Score     float64 `json:"score"`
		Grade     string  `json:"grade"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-factorial-bug-003",
		"weight_level": "hard",
		"user_code":    correctFix,
	}, &good)
	if good.Output != "120\n6" {
		t.Errorf("output = %q, want 120\\n6", good.Output)
	}
	if !good.IsCorrect || good.Score != 100 || good.Grade != "S" {
		t.Errorf("correct fix => correct=%v score=%v grade=%q", good.IsCorrect, good.Score, good.Grade)
	}

	// original buggy code -> D / 0
	buggyCode := "def factorial(n):\n    result = 1\n    for i in range(n):\n        result *= i\n    return result"
	var bad struct {
		IsCorrect bool    `json:"is_correct"`
		Score     float64 `json:"score"`
		Grade     string  `json:"grade"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-factorial-bug-003",
		"weight_level": "hard",
		"user_code":    buggyCode,
	}, &bad)
	if bad.IsCorrect || bad.Score != 0 || bad.Grade != "D" {
		t.Errorf("broken fix => correct=%v score=%v grade=%q", bad.IsCorrect, bad.Score, bad.Grade)
	}

	// time out -> friendly error
	// create a short-timeout exercise so the test stays fast
	timeoutEx := model.Exercise{
		ID:             "e2e-timeout-001",
		Title:          "无限循环",
		Description:    "用来测试执行超时",
		ExerciseType:   model.ExerciseTypeFill,
		TechStack:      "python",
		Weight:         "easy",
		Category:       "e2e",
		FullCode:       "while True:\n    pass\nprint('never')",
		ExpectedOutput: "never",
		FillPoints: model.FillPointsConfig{
			Points: []model.FillPoint{{
				ID: "p1", OriginalText: "pass", Category: "statement",
				WeightScore: 1, MinWeight: "easy", FillType: "text",
				Location: model.FillLocation{SpanStart: 16, SpanEnd: 20},
			}},
			MinFills: 1, MaxFills: 1,
		},
		Enabled:       true,
		TimeLimitSecs: 2,
	}
	if resp := doJSON(t, "POST", base+"/admin/exercises", timeoutEx, nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("create timeout exercise status = %d", resp.StatusCode)
	}
	var timeout struct {
		Success      bool   `json:"success"`
		ErrorMessage string `json:"error_message"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "e2e-timeout-001",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_p1": "pass"},
	}, &timeout)
	if timeout.Success {
		t.Error("infinite loop should not succeed")
	}
	if !strings.Contains(timeout.ErrorMessage, "超时") {
		t.Errorf("expected timeout message, got %q", timeout.ErrorMessage)
	}
}

func TestE2E_CommandAndConfigAnswerMatch(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// command exercise: correct option
	var cmd struct {
		Score   float64 `json:"score"`
		Grade   string  `json:"grade"`
		Correct bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "kubectl-getpods-005",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_flag": "--all-namespaces"},
	}, &cmd)
	if cmd.Score != 100 || cmd.Grade != "S" || !cmd.Correct {
		t.Errorf("correct flag => score=%v grade=%q correct=%v", cmd.Score, cmd.Grade, cmd.Correct)
	}

	// command exercise: wrong option
	var cmdBad struct {
		Score float64 `json:"score"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "kubectl-getpods-005",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_flag": "-n default"},
	}, &cmdBad)
	if cmdBad.Score != 0 {
		t.Errorf("wrong flag score = %v, want 0", cmdBad.Score)
	}

	// config exercise: all correct
	var cfg struct {
		Score   float64 `json:"score"`
		Correct bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "k8s-nginx-pod-006",
		"weight_level": "medium",
		"fill_answers": map[string]string{
			"FILL_image": "nginx:1.25",
			"FILL_port":  "80",
			"FILL_name":  "nginx-pod",
		},
	}, &cfg)
	if cfg.Score != 100 || !cfg.Correct {
		t.Errorf("correct config => score=%v correct=%v", cfg.Score, cfg.Correct)
	}
}

func TestE2E_ProgressAndWeakPoints(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"
	_, token := registerUser(t, ts)

	// wrong attempt first
	doJSONAuth(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-lc-001",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_expr": "x*2", "FILL_iter": "numbers"},
	}, nil, token)
	// then correct
	doJSONAuth(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-lc-001",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_expr": "x**2", "FILL_iter": "numbers"},
	}, nil, token)

	var prog struct {
		Progress struct {
			TotalAttempts   int                `json:"total_attempts"`
			CorrectAttempts int                `json:"correct_attempts"`
			Completed       map[string]bool    `json:"completed_exercises"`
			WeakPoints      map[string]int     `json:"weak_points"`
			TechStackScores map[string]float64 `json:"tech_stack_scores"`
			CurrentStreak   int                `json:"current_streak"`
		} `json:"progress"`
		Recent []struct {
			ID string `json:"id"`
		} `json:"recent_attempts"`
	}
	req, _ := http.NewRequest("GET", base+"/progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&prog); err != nil {
		t.Fatal(err)
	}
	if prog.Progress.TotalAttempts != 2 || prog.Progress.CorrectAttempts != 1 {
		t.Errorf("attempts = %d/%d, want 2/1", prog.Progress.TotalAttempts, prog.Progress.CorrectAttempts)
	}
	if !prog.Progress.Completed["py-lc-001"] {
		t.Error("exercise not marked completed")
	}
	if prog.Progress.WeakPoints["py-lc-001"] != 1 {
		t.Errorf("weak_points = %v, want 1 failure", prog.Progress.WeakPoints)
	}
	if len(prog.Recent) != 2 {
		t.Errorf("recent attempts = %d, want 2", len(prog.Recent))
	}

	// weak-points endpoint reflects the failure
	var weak []struct {
		ExerciseID string `json:"exercise_id"`
		FailCount  int    `json:"fail_count"`
	}
	req2, err := http.NewRequest("GET", base+"/weak-points", nil)
	if err != nil {
		t.Fatal(err)
	}
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	json.NewDecoder(resp2.Body).Decode(&weak)
	found := false
	for _, w := range weak {
		if w.ExerciseID == "py-lc-001" {
			found = true
			if w.FailCount != 1 {
				t.Errorf("weak fail_count = %d, want 1", w.FailCount)
			}
		}
	}
	if !found {
		t.Error("py-lc-001 not in weak points")
	}
}

func TestE2E_GoFlow(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// variant masking for a Go exercise
	var v struct {
		ExerciseID string `json:"exercise_id"`
		MarkedCode string `json:"marked_code"`
		FillPoints []struct {
			ID string `json:"id"`
		} `json:"fill_points"`
	}
	doJSON(t, "GET", base+"/exercises/go-even-squares-101/variant?level=medium", nil, &v)
	if v.ExerciseID != "go-even-squares-101" || len(v.FillPoints) == 0 {
		t.Fatalf("go variant wrong: %+v", v)
	}
	if strings.Contains(v.MarkedCode, "n%2 == 0") {
		t.Error("go answer leaked into marked_code")
	}

	// correct Go fill -> runs via `go run` and passes
	var good struct {
		Output    string  `json:"output"`
		Score     float64 `json:"score"`
		Grade     string  `json:"grade"`
		IsCorrect bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "go-even-squares-101",
		"weight_level": "medium",
		"fill_answers": map[string]string{
			"FILL_cond":   "n%2 == 0",
			"FILL_square": "n*n",
			"FILL_append": "append",
		},
	}, &good)
	if good.Output != "[4 16 36 64]" {
		t.Errorf("go output = %q, want [4 16 36 64]", good.Output)
	}
	if !good.IsCorrect || good.Score != 100 || good.Grade != "S" {
		t.Errorf("go correct => score=%v grade=%q correct=%v", good.Score, good.Grade, good.IsCorrect)
	}

	// wrong Go fill -> fails
	var bad struct {
		Score     float64 `json:"score"`
		IsCorrect bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "go-even-squares-101",
		"weight_level": "medium",
		"fill_answers": map[string]string{
			"FILL_cond":   "n%2 == 1",
			"FILL_square": "n+n",
			"FILL_append": "append",
		},
	}, &bad)
	if bad.IsCorrect || bad.Score >= 100 {
		t.Errorf("go wrong => score=%v correct=%v", bad.Score, bad.IsCorrect)
	}

	// Go bugfix flow
	correctFix := "package main\n\nimport \"fmt\"\n\nfunc factorial(n int) int {\n\tresult := 1\n\tfor i := 1; i <= n; i++ {\n\t\tresult *= i\n\t}\n\treturn result\n}\n\nfunc main() {\n\tfmt.Println(factorial(5))\n\tfmt.Println(factorial(3))\n}"
	var fix struct {
		Output    string  `json:"output"`
		Score     float64 `json:"score"`
		IsCorrect bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "go-factorial-bug-103",
		"weight_level": "hard",
		"user_code":    correctFix,
	}, &fix)
	if fix.Output != "120\n6" || !fix.IsCorrect || fix.Score != 100 {
		t.Errorf("go bugfix => output=%q score=%v correct=%v", fix.Output, fix.Score, fix.IsCorrect)
	}

	// Go basic bank (50 questions) executes in real via `go run`
	var basics []struct {
		ID string `json:"id"`
	}
	doJSON(t, "GET", base+"/exercises?tech_stack=go", nil, &basics)
	nBasics := 0
	for _, it := range basics {
		if strings.HasPrefix(it.ID, "go-basics-") {
			nBasics++
		}
	}
	if nBasics != 50 {
		t.Fatalf("go-basics count = %d, want 50", nBasics)
	}
	// a basic fill exercise runs correctly
	var bfill struct {
		Output    string  `json:"output"`
		Score     float64 `json:"score"`
		IsCorrect bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "go-basics-026",
		"weight_level": "hard",
		"fill_answers": map[string]string{"FILL_go": "go func", "FILL_wait": "wg.Wait()"},
	}, &bfill)
	if bfill.Output != "55" || !bfill.IsCorrect || bfill.Score != 100 {
		t.Errorf("go-basics-026 => output=%q score=%v correct=%v", bfill.Output, bfill.Score, bfill.IsCorrect)
	}
	// a basic bugfix exercise runs correctly
	gofx := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\ta, b := 1, 2\n\ta, b = b, a\n\tfmt.Println(a, b)\n}"
	var fix2 struct {
		Output    string `json:"output"`
		IsCorrect bool   `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "go-basics-048",
		"weight_level": "easy",
		"user_code":    gofx,
	}, &fix2)
	if fix2.Output != "2 1" || !fix2.IsCorrect {
		t.Errorf("go-basics-048 => output=%q correct=%v", fix2.Output, fix2.IsCorrect)
	}
}

func TestE2E_K8sBasics(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// the bank contains exactly 50 k8s-basics questions
	var list []struct {
		ID string `json:"id"`
	}
	doJSON(t, "GET", base+"/exercises?tech_stack=kubernetes", nil, &list)
	k8sBasics := 0
	for _, it := range list {
		if strings.HasPrefix(it.ID, "k8s-basics-") {
			k8sBasics++
		}
	}
	if k8sBasics != 50 {
		t.Fatalf("k8s-basics count = %d, want 50", k8sBasics)
	}

	// variant must mask answers for a command question
	var v struct {
		MarkedCode string `json:"marked_code"`
	}
	doJSON(t, "GET", base+"/exercises/k8s-basics-001/variant?level=beginner", nil, &v)
	if !strings.Contains(v.MarkedCode, "____FILL_") {
		t.Errorf("k8s command not masked: %q", v.MarkedCode)
	}

	// command answer-match with an acceptable alias
	var res struct {
		Score     float64 `json:"score"`
		Grade     string  `json:"grade"`
		IsCorrect bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "k8s-basics-001",
		"weight_level": "beginner",
		"fill_answers": map[string]string{"FILL_flag": "--all-namespaces"},
	}, &res)
	if res.Score != 100 || res.Grade != "S" || !res.IsCorrect {
		t.Errorf("k8s command => score=%v grade=%q correct=%v", res.Score, res.Grade, res.IsCorrect)
	}

	// config YAML answer-match
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "k8s-basics-051",
		"weight_level": "medium",
		"fill_answers": map[string]string{"FILL_cpu": `"500m"`, "FILL_mem": `"128Mi"`},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("k8s config => score=%v correct=%v", res.Score, res.IsCorrect)
	}

	// K8s 资源配置 series: exactly 30 under the 资源配置 subcategory
	var resList []struct {
		ID          string `json:"id"`
		Subcategory string `json:"subcategory"`
	}
	doJSON(t, "GET", base+"/exercises?tech_stack=kubernetes", nil, &resList)
	nRes := 0
	for _, it := range resList {
		if it.Subcategory == "资源配置" {
			nRes++
		}
	}
	if nRes != 30 {
		t.Fatalf("k8s 资源配置 count = %d, want 30", nRes)
	}

	// a 资源配置 YAML question (Secret) answers correctly
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "k8s-res-001",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_type": "Opaque", "FILL_pwd": "c2VjcmV0"},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("k8s-res-001 => score=%v correct=%v", res.Score, res.IsCorrect)
	}

	// K8s 调度专项: exactly 30 under 调度 subcategory, and one answers correctly
	var schedList []struct {
		ID          string `json:"id"`
		Subcategory string `json:"subcategory"`
	}
	doJSON(t, "GET", base+"/exercises?tech_stack=kubernetes", nil, &schedList)
	nSched := 0
	for _, it := range schedList {
		if it.Subcategory == "调度" {
			nSched++
		}
	}
	if nSched != 30 {
		t.Fatalf("k8s 调度 count = %d, want 30", nSched)
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "k8s-sched-003",
		"weight_level": "beginner",
		"fill_answers": map[string]string{"FILL_verb": "cordon"},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("k8s-sched-003 => score=%v correct=%v", res.Score, res.IsCorrect)
	}

	// Prometheus bank: exactly 30 under devops, a promql + a config question answer correctly
	var promList []struct {
		ID string `json:"id"`
	}
	doJSON(t, "GET", base+"/exercises?tech_stack=devops", nil, &promList)
	nProm := 0
	for _, it := range promList {
		if strings.HasPrefix(it.ID, "prom-") {
			nProm++
		}
	}
	if nProm != 30 {
		t.Fatalf("prometheus count = %d, want 30", nProm)
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "prom-basics-011",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_range": "[5m]"},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("prom promql => score=%v correct=%v", res.Score, res.IsCorrect)
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "prom-basics-023",
		"weight_level": "medium",
		"fill_answers": map[string]string{"FILL_job": "node", "FILL_target": "localhost:9100"},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("prom config => score=%v correct=%v", res.Score, res.IsCorrect)
	}

	// Python 基础 bank: exactly 50 py-basics questions, and one runs correctly
	var pyList []struct {
		ID string `json:"id"`
	}
	doJSON(t, "GET", base+"/exercises?tech_stack=python", nil, &pyList)
	nPy := 0
	for _, it := range pyList {
		if strings.HasPrefix(it.ID, "py-basics-") {
			nPy++
		}
	}
	if nPy != 50 {
		t.Fatalf("py-basics count = %d, want 50", nPy)
	}
	var pyExec struct {
		Output    string  `json:"output"`
		Score     float64 `json:"score"`
		IsCorrect bool    `json:"is_correct"`
	}
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-basics-016",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_expr": "x * x"},
	}, &pyExec)
	if pyExec.Output != "[0, 1, 4, 9, 16]" || !pyExec.IsCorrect || pyExec.Score != 100 {
		t.Errorf("py-basics-016 => output=%q score=%v correct=%v", pyExec.Output, pyExec.Score, pyExec.IsCorrect)
	}

	// Docker 基础 bank: exactly 50 under devops, command + Dockerfile answer-match
	var dkList []struct {
		ID string `json:"id"`
	}
	doJSON(t, "GET", base+"/exercises?tech_stack=devops", nil, &dkList)
	nDk := 0
	for _, it := range dkList {
		if strings.HasPrefix(it.ID, "docker-basics-") {
			nDk++
		}
	}
	if nDk != 50 {
		t.Fatalf("docker-basics count = %d, want 50", nDk)
	}
	// docker command
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "docker-basics-002",
		"weight_level": "beginner",
		"fill_answers": map[string]string{"FILL_flag": "-d"},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("docker command => score=%v correct=%v", res.Score, res.IsCorrect)
	}
	// Dockerfile
	doJSON(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "docker-basics-033",
		"weight_level": "medium",
		"fill_answers": map[string]string{"FILL_base": "python:3.11-slim"},
	}, &res)
	if res.Score != 100 || !res.IsCorrect {
		t.Errorf("docker dockerfile => score=%v correct=%v", res.Score, res.IsCorrect)
	}
}

func TestE2E_ExportImport(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1/admin"

	// export returns the whole library as a JSON array
	resp, err := http.Get(base + "/export")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export status = %d", resp.StatusCode)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "attachment") {
		t.Errorf("export missing attachment header: %q", cd)
	}
	var exported []model.Exercise
	if err := json.NewDecoder(resp.Body).Decode(&exported); err != nil {
		t.Fatal(err)
	}
	if len(exported) != 1466 {
		t.Fatalf("export count = %d, want 1466", len(exported))
	}

	// import: delete one exercise, then restore it from the exported blob
	doJSON(t, "DELETE", base+"/exercises/go-concurrent-106", nil, nil)
	var target model.Exercise
	for _, ex := range exported {
		if ex.ID == "go-concurrent-106" {
			target = ex
		}
	}
	if target.ID == "" {
		t.Fatal("go-concurrent-106 not found in export")
	}
	var imp struct {
		Imported int `json:"imported"`
		Total    int `json:"total"`
	}
	doJSON(t, "POST", base+"/import", []model.Exercise{target}, &imp)
	if imp.Imported != 1 {
		t.Errorf("import restored %d, want 1", imp.Imported)
	}
	// and it is usable again
	var v struct {
		ExerciseID string `json:"exercise_id"`
	}
	doJSON(t, "GET", ts.URL+"/api/v1/exercises/go-concurrent-106/variant?level=hard", nil, &v)
	if v.ExerciseID != "go-concurrent-106" {
		t.Errorf("imported exercise not available: %+v", v)
	}
}

func TestE2E_AdminCRUD(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1/admin"

	// create
	ex := model.Exercise{
		ID:             "e2e-new-001",
		Title:          "E2E 测试训练",
		Description:    "端到端测试创建",
		ExerciseType:   model.ExerciseTypeFill,
		TechStack:      "python",
		Weight:         "easy",
		Category:       "e2e",
		FullCode:       "print(42)",
		ExpectedOutput: "42",
		FillPoints: model.FillPointsConfig{
			Points: []model.FillPoint{{
				ID: "p1", OriginalText: "42", Category: "expression",
				WeightScore: 1, MinWeight: "easy", FillType: "text",
				Location: model.FillLocation{SpanStart: 6, SpanEnd: 8},
			}},
			MinFills: 1, MaxFills: 1,
		},
		Enabled:       true,
		TimeLimitSecs: 30,
	}
	resp := doJSON(t, "POST", base+"/exercises", ex, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create status = %d", resp.StatusCode)
	}

	// it shows up in the list
	var list []struct {
		ID string `json:"id"`
	}
	doJSON(t, "GET", ts.URL+"/api/v1/exercises", nil, &list)
	if !containsIDs(list, "e2e-new-001") {
		t.Error("created exercise not listed")
	}

	// execute it (fill answer matches)
	var exec struct {
		Score float64 `json:"score"`
	}
	doJSON(t, "POST", ts.URL+"/api/v1/execute", map[string]any{
		"exercise_id":  "e2e-new-001",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_p1": "42"},
	}, &exec)
	if exec.Score != 100 {
		t.Errorf("created exercise score = %v, want 100", exec.Score)
	}

	// update
	ex.Title = "E2E 测试训练（改）"
	doJSON(t, "PUT", base+"/exercises/e2e-new-001", ex, nil)
	var updated model.Exercise
	doJSON(t, "GET", ts.URL+"/api/v1/exercises/e2e-new-001", nil, &updated)
	if updated.Title != "E2E 测试训练（改）" {
		t.Errorf("update failed, title = %q", updated.Title)
	}

	// import (array of exercises)
	imported := []model.Exercise{
		{ID: "e2e-import-1", Title: "导入1", FullCode: "print(1)", TechStack: "python", Weight: "easy",
			ExerciseType: model.ExerciseTypeFill, Enabled: true, TimeLimitSecs: 30},
		{ID: "e2e-import-2", Title: "导入2", FullCode: "print(2)", TechStack: "python", Weight: "easy",
			ExerciseType: model.ExerciseTypeFill, Enabled: true, TimeLimitSecs: 30},
	}
	var imp struct {
		Imported int `json:"imported"`
		Total    int `json:"total"`
	}
	doJSON(t, "POST", base+"/import", imported, &imp)
	if imp.Imported != 2 || imp.Total != 2 {
		t.Errorf("import = %d/%d, want 2/2", imp.Imported, imp.Total)
	}

	// stats
	var stats map[string]any
	doJSON(t, "GET", base+"/stats", nil, &stats)
	if stats["exercise_count"].(float64) < 8 {
		t.Errorf("exercise_count = %v", stats["exercise_count"])
	}

	// delete
	doJSON(t, "DELETE", base+"/exercises/e2e-new-001", nil, nil)
	var list2 []struct {
		ID string `json:"id"`
	}
	doJSON(t, "GET", ts.URL+"/api/v1/exercises", nil, &list2)
	if containsIDs(list2, "e2e-new-001") {
		t.Error("deleted exercise still listed")
	}
}

func TestE2E_EnvironmentsAndNotFound(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	var envs []struct {
		Name         string `json:"name"`
		HealthStatus string `json:"health_status"`
	}
	doJSON(t, "GET", base+"/environments", nil, &envs)
	if len(envs) == 0 {
		t.Error("no environments returned")
	}

	resp := doJSON(t, "GET", base+"/exercises/does-not-exist", nil, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("missing exercise status = %d, want 404", resp.StatusCode)
	}

	resp = doJSON(t, "POST", base+"/execute", map[string]any{"exercise_id": "nope"}, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("execute unknown exercise status = %d, want 404", resp.StatusCode)
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func containsIDs(list []struct {
	ID string `json:"id"`
}, needle string) bool {
	for _, item := range list {
		if item.ID == needle {
			return true
		}
	}
	return false
}

// TestE2E_AdminTokenAuth guards the coach-side API: with an admin token
// configured, every /api/v1/admin/* call must present X-Admin-Token (or
// `Authorization: Bearer`), while student endpoints stay open.
func TestE2E_AdminTokenAuth(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "auth.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	srv := api.New(st, engine.New("python3"))
	srv.SetAdminToken("secret-token")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	status := func(path string, hdr map[string]string) int {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+path, nil)
		for k, v := range hdr {
			req.Header.Set(k, v)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp.StatusCode
	}

	if code := status("/api/v1/admin/stats", nil); code != http.StatusUnauthorized {
		t.Errorf("no token => %d, want 401", code)
	}
	if code := status("/api/v1/admin/stats", map[string]string{"X-Admin-Token": "wrong"}); code != http.StatusUnauthorized {
		t.Errorf("wrong token => %d, want 401", code)
	}
	if code := status("/api/v1/admin/export", nil); code != http.StatusUnauthorized {
		t.Errorf("export without token => %d, want 401", code)
	}
	if code := status("/api/v1/admin/stats", map[string]string{"X-Admin-Token": "secret-token"}); code != http.StatusOK {
		t.Errorf("X-Admin-Token => %d, want 200", code)
	}
	if code := status("/api/v1/admin/stats", map[string]string{"Authorization": "Bearer secret-token"}); code != http.StatusOK {
		t.Errorf("Bearer token => %d, want 200", code)
	}
	if code := status("/api/v1/environments", nil); code != http.StatusOK {
		t.Errorf("student endpoint with token configured => %d, want 200", code)
	}
}

// --- 用户体系 / 题库选择 / 动态加载 ---

func TestE2E_AuthFlow(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// register + duplicate + validation
	userID, token := registerUser(t, ts)
	var me struct {
		User model.User `json:"user"`
	}
	if resp := doJSONAuth(t, "GET", base+"/auth/me", nil, &me, token); resp.StatusCode != http.StatusOK {
		t.Fatalf("me => %d", resp.StatusCode)
	}
	if me.User.ID != userID || me.User.Username == "" {
		t.Fatalf("me user = %+v", me.User)
	}
	if resp := doJSON(t, "GET", base+"/auth/me", nil, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous me => %d, want 401", resp.StatusCode)
	}
	if resp := doJSON(t, "POST", base+"/auth/register",
		map[string]any{"username": "ab", "password": "secret123"}, nil); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("short username => %d, want 400", resp.StatusCode)
	}
	if resp := doJSON(t, "POST", base+"/auth/register",
		map[string]any{"username": "e2e_dup_user", "password": "secret123"}, nil); resp.StatusCode != http.StatusCreated {
		t.Fatalf("first register of e2e_dup_user failed")
	}
	if resp := doJSON(t, "POST", base+"/auth/register",
		map[string]any{"username": "E2E_DUP_USER", "password": "secret123"}, nil); resp.StatusCode != http.StatusConflict {
		t.Errorf("duplicate register => %d, want 409", resp.StatusCode)
	}

	// wrong password then successful login issues a new token
	if resp := doJSON(t, "POST", base+"/auth/login",
		map[string]any{"username": "e2e_dup_user", "password": "wrong-pass"}, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong password => %d, want 401", resp.StatusCode)
	}
	var login struct {
		Token string `json:"token"`
	}
	if resp := doJSON(t, "POST", base+"/auth/login",
		map[string]any{"username": "e2e_dup_user", "password": "secret123"}, &login); resp.StatusCode != http.StatusOK {
		t.Fatalf("login => %d", resp.StatusCode)
	}
	if login.Token == "" {
		t.Fatal("login returned empty token")
	}

	// logout revokes the token
	if resp := doJSONAuth(t, "POST", base+"/auth/logout", nil, nil, login.Token); resp.StatusCode != http.StatusOK {
		t.Fatalf("logout => %d", resp.StatusCode)
	}
	if resp := doJSONAuth(t, "GET", base+"/auth/me", nil, nil, login.Token); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("me after logout => %d, want 401", resp.StatusCode)
	}

	// change password: old token stays valid, old password stops working
	var login2 struct {
		Token string `json:"token"`
	}
	if resp := doJSON(t, "POST", base+"/auth/login",
		map[string]any{"username": "e2e_dup_user", "password": "secret123"}, &login2); resp.StatusCode != http.StatusOK {
		t.Fatalf("second login => %d", resp.StatusCode)
	}
	if resp := doJSONAuth(t, "PUT", base+"/auth/me",
		map[string]any{"old_password": "secret123", "new_password": "brand-new-pass"}, nil, login2.Token); resp.StatusCode != http.StatusOK {
		t.Fatalf("update me => %d", resp.StatusCode)
	}
	if resp := doJSONAuth(t, "GET", base+"/auth/me", nil, nil, login2.Token); resp.StatusCode != http.StatusOK {
		t.Errorf("current token after password change => %d, want 200", resp.StatusCode)
	}
	if resp := doJSON(t, "POST", base+"/auth/login",
		map[string]any{"username": "e2e_dup_user", "password": "secret123"}, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("old password after change => %d, want 401", resp.StatusCode)
	}
	if resp := doJSON(t, "POST", base+"/auth/login",
		map[string]any{"username": "e2e_dup_user", "password": "brand-new-pass"}, nil); resp.StatusCode != http.StatusOK {
		t.Errorf("new password login => %d, want 200", resp.StatusCode)
	}
}

func TestE2E_UserIsolationAndBankSelection(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"
	_, tokenA := registerUser(t, ts)
	_, tokenB := registerUser(t, ts)

	// A completes one exercise; B's progress must stay empty
	doJSONAuth(t, "POST", base+"/execute", map[string]any{
		"exercise_id":  "py-lc-001",
		"weight_level": "easy",
		"fill_answers": map[string]string{"FILL_expr": "x**2", "FILL_iter": "numbers"},
	}, nil, tokenA)

	type progressResp struct {
		Progress struct {
			TotalAttempts int `json:"total_attempts"`
		} `json:"progress"`
	}
	var pa, pb progressResp
	doJSONAuth(t, "GET", base+"/progress", nil, &pa, tokenA)
	doJSONAuth(t, "GET", base+"/progress", nil, &pb, tokenB)
	if pa.Progress.TotalAttempts != 1 {
		t.Errorf("user A attempts = %d, want 1", pa.Progress.TotalAttempts)
	}
	if pb.Progress.TotalAttempts != 0 {
		t.Errorf("user B attempts = %d, want 0 (isolation)", pb.Progress.TotalAttempts)
	}

	// bank picker: available list + persisted selection, invalid keys dropped
	var banks struct {
		Available []model.BankInfo `json:"available"`
		Selected  []string         `json:"selected"`
	}
	if resp := doJSONAuth(t, "GET", base+"/me/banks", nil, &banks, tokenA); resp.StatusCode != http.StatusOK {
		t.Fatalf("my banks => %d", resp.StatusCode)
	}
	if len(banks.Available) < 5 || banks.Available[0].ExerciseCount == 0 {
		t.Fatalf("available banks look wrong: %+v", banks.Available[:1])
	}
	var saved struct {
		Selected []string `json:"selected"`
	}
	doJSONAuth(t, "PUT", base+"/me/banks",
		map[string]any{"selected": []string{"testing::pytest", "nope::nope", "testing::pytest"}}, &saved, tokenA)
	if len(saved.Selected) != 1 || saved.Selected[0] != "testing::pytest" {
		t.Fatalf("saved selection = %v", saved.Selected)
	}
	doJSONAuth(t, "GET", base+"/me/banks", nil, &banks, tokenA)
	if len(banks.Selected) != 1 {
		t.Fatalf("selection not persisted: %v", banks.Selected)
	}
	doJSONAuth(t, "GET", base+"/me/banks", nil, &banks, tokenB)
	if len(banks.Selected) != 0 {
		t.Errorf("user B selection = %v, want empty (isolation)", banks.Selected)
	}
}

func TestE2E_AdminUsersAndBankUpload(t *testing.T) {
	ts := newTestServer(t)
	base := ts.URL + "/api/v1"

	// admin creates a user
	var created model.User
	if resp := doJSON(t, "POST", base+"/admin/users",
		map[string]any{"username": "coach_made", "password": "secret123", "display_name": "教练建的"}, &created); resp.StatusCode != http.StatusCreated {
		t.Fatalf("admin create user => %d", resp.StatusCode)
	}
	var users []model.AdminUser
	doJSON(t, "GET", base+"/admin/users", nil, &users)
	found := false
	for _, u := range users {
		if u.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("created user %s not listed", created.ID)
	}

	// disable blocks login, re-enable restores it
	doJSON(t, "PUT", base+"/admin/users/"+created.ID, map[string]any{"disabled": true}, nil)
	if resp := doJSON(t, "POST", base+"/auth/login",
		map[string]any{"username": "coach_made", "password": "secret123"}, nil); resp.StatusCode != http.StatusForbidden {
		t.Errorf("disabled login => %d, want 403", resp.StatusCode)
	}
	doJSON(t, "PUT", base+"/admin/users/"+created.ID, map[string]any{"disabled": false, "password": "reset-pass-1"}, nil)
	if resp := doJSON(t, "POST", base+"/auth/login",
		map[string]any{"username": "coach_made", "password": "reset-pass-1"}, nil); resp.StatusCode != http.StatusOK {
		t.Errorf("reset password login => %d, want 200", resp.StatusCode)
	}
	doJSON(t, "DELETE", base+"/admin/users/"+created.ID, nil, nil)
	if resp := doJSON(t, "POST", base+"/auth/login",
		map[string]any{"username": "coach_made", "password": "reset-pass-1"}, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("deleted user login => %d, want 401", resp.StatusCode)
	}

	// dynamic bank upload: valid bank becomes practiceable immediately
	bank := `[{
		"id": "e2e-upload-001", "title": "上传题", "description": "d",
		"exercise_type": "fill", "tech_stack": "testing", "weight": "easy",
		"category": "c", "subcategory": "pytest", "full_code": "print(1)",
		"expected": "", "points": [{"id": "p1", "text": "print(1)", "category": "c",
		"min_weight": "beginner", "weight_score": 1, "hint": "h", "fill_type": "text"}],
		"hints": {"level_1": "a", "level_2": "b", "level_3": "c", "level_4": "d"},
		"explanation": "e"
	}]`
	var up struct {
		Name     string `json:"name"`
		Imported int    `json:"imported"`
	}
	if resp := doJSON(t, "POST", base+"/admin/banks",
		map[string]any{"name": "e2e-hot-bank", "content": bank}, &up); resp.StatusCode != http.StatusOK {
		t.Fatalf("upload bank => %d", resp.StatusCode)
	}
	if up.Imported != 1 {
		t.Fatalf("imported = %d, want 1", up.Imported)
	}
	var ex model.Exercise
	if resp := doJSON(t, "GET", base+"/exercises/e2e-upload-001", nil, &ex); resp.StatusCode != http.StatusOK {
		t.Fatalf("uploaded exercise not served: %d", resp.StatusCode)
	}
	if ex.TechStack != "testing" || ex.Subcategory != "pytest" {
		t.Errorf("uploaded exercise = %+v", ex)
	}
	var sources []model.BankSource
	doJSON(t, "GET", base+"/admin/banks", nil, &sources)
	if len(sources) != 1 || sources[0].Name != "e2e-hot-bank" || sources[0].Source != "upload" {
		t.Fatalf("bank sources = %+v", sources)
	}
	// invalid payloads rejected
	if resp := doJSON(t, "POST", base+"/admin/banks",
		map[string]any{"name": "bad name", "content": bank}, nil); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bad bank name => %d, want 400", resp.StatusCode)
	}
	if resp := doJSON(t, "POST", base+"/admin/banks",
		map[string]any{"name": "e2e-broken", "content": "{not json"}, nil); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("broken bank => %d, want 400", resp.StatusCode)
	}
}
