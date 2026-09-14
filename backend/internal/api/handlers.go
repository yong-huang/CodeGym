package api

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"time"

	"codegym/backend/internal/engine"
	"codegym/backend/internal/judge"
	"codegym/backend/internal/llm"
	"codegym/backend/internal/model"
	"codegym/backend/internal/store"
	"codegym/backend/internal/variant"
)

// Server wires the store and engine into HTTP handlers.
type Server struct {
	store      *store.Store
	engine     *engine.Engine
	llm        *llm.Client
	adminToken string // 为空表示不启用教练端鉴权（本地开发默认）
	now        func() time.Time
}

// New creates an API server. LLM settings come from the DB when present,
// otherwise fall back to environment/defaults (see internal/llm).
func New(st *store.Store, eng *engine.Engine) *Server {
	ep, _ := st.GetSetting("llm.endpoint")
	md, _ := st.GetSetting("llm.model")
	return &Server{store: st, engine: eng, llm: llm.New(ep, md), now: time.Now}
}

// SetAdminToken enables bearer-token auth for every /api/v1/admin route.
// Passing an empty string disables the guard (local development default).
func (s *Server) SetAdminToken(token string) { s.adminToken = token }

// Handler returns the root http.Handler with all routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/exercises", s.listExercises)
	mux.HandleFunc("GET /api/v1/exercises/{id}", s.getExercise)
	mux.HandleFunc("GET /api/v1/exercises/{id}/variant", s.getVariant)
	mux.HandleFunc("POST /api/v1/execute", s.execute)
	mux.HandleFunc("GET /api/v1/progress", s.getProgress)
	mux.HandleFunc("GET /api/v1/weak-points", s.getWeakPoints)
	mux.HandleFunc("GET /api/v1/favorites", s.listFavorites)
	mux.HandleFunc("POST /api/v1/favorites", s.toggleFavorite)
	mux.HandleFunc("GET /api/v1/environments", s.listEnvironments)

	mux.HandleFunc("POST /api/v1/admin/exercises", s.adminOnly(s.createExercise))
	mux.HandleFunc("PUT /api/v1/admin/exercises/{id}", s.adminOnly(s.updateExercise))
	mux.HandleFunc("DELETE /api/v1/admin/exercises/{id}", s.adminOnly(s.deleteExercise))
	mux.HandleFunc("POST /api/v1/admin/import", s.adminOnly(s.importExercises))
	mux.HandleFunc("GET /api/v1/admin/export", s.adminOnly(s.exportExercises))
	mux.HandleFunc("GET /api/v1/admin/categories", s.adminOnly(s.listCategories))
	mux.HandleFunc("POST /api/v1/admin/categories", s.adminOnly(s.createCategory))
	mux.HandleFunc("PUT /api/v1/admin/categories", s.adminOnly(s.renameCategory))
	mux.HandleFunc("DELETE /api/v1/admin/categories", s.adminOnly(s.deleteCategory))
	mux.HandleFunc("POST /api/v1/admin/ai-generate", s.adminOnly(s.aiGenerate))
	mux.HandleFunc("GET /api/v1/admin/llm-settings", s.adminOnly(s.getLLMSettings))
	mux.HandleFunc("PUT /api/v1/admin/llm-settings", s.adminOnly(s.setLLMSettings))
	mux.HandleFunc("POST /api/v1/admin/llm-settings/test", s.adminOnly(s.testLLMSettings))
	mux.HandleFunc("GET /api/v1/admin/environments", s.adminOnly(s.adminListEnvironments))
	mux.HandleFunc("POST /api/v1/admin/environments", s.adminOnly(s.adminCreateEnvironment))
	mux.HandleFunc("PUT /api/v1/admin/environments/{id}", s.adminOnly(s.adminUpdateEnvironment))
	mux.HandleFunc("DELETE /api/v1/admin/environments/{id}", s.adminOnly(s.adminDeleteEnvironment))
	mux.HandleFunc("POST /api/v1/admin/environments/{id}/test", s.adminOnly(s.adminTestEnvironment))
	mux.HandleFunc("GET /api/v1/admin/stats", s.adminOnly(s.adminStats))

	return s.withCommon(mux)
}

// adminOnly guards coach-side endpoints with the configured admin token.
// With no token set (local dev default) requests pass through unchanged;
// otherwise a valid X-Admin-Token header or `Authorization: Bearer <token>`
// is required (constant-time comparison).
func (s *Server) adminOnly(next http.HandlerFunc) http.HandlerFunc {
	if s.adminToken == "" {
		return next
	}
	want := []byte(s.adminToken)
	return func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Admin-Token")
		if got == "" {
			if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
				got = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if subtle.ConstantTimeCompare([]byte(got), want) != 1 {
			writeErr(w, http.StatusUnauthorized, "需要管理令牌：请求头携带 X-Admin-Token 或 Authorization: Bearer <token>")
			return
		}
		next(w, r)
	}
}

func (s *Server) withCommon(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Id")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// userID extracts the calling user, defaulting to a demo trainee (MVP auth).
func (s *Server) userID(r *http.Request) string {
	if id := r.Header.Get("X-User-Id"); id != "" {
		return id
	}
	return "demo-user"
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// --- 学员端 ---

func (s *Server) listExercises(w http.ResponseWriter, r *http.Request) {
	tech := r.URL.Query().Get("tech_stack")
	exs, err := s.store.ListExercises(tech)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	// strip answer-bearing fields for the list view
	type listItem struct {
		ID           string `json:"id"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		ExerciseType string `json:"exercise_type"`
		TechStack    string `json:"tech_stack"`
		Weight       string `json:"weight"`
		Category     string `json:"category"`
		Subcategory  string `json:"subcategory"`
		PointCount   int    `json:"point_count"`
	}
	out := make([]listItem, 0, len(exs))
	for _, ex := range exs {
		out = append(out, listItem{
			ID: ex.ID, Title: ex.Title, Description: ex.Description,
			ExerciseType: ex.ExerciseType, TechStack: ex.TechStack,
			Weight: ex.Weight, Category: ex.Category,
			Subcategory: ex.Subcategory, PointCount: len(ex.FillPoints.Points),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getExercise(w http.ResponseWriter, r *http.Request) {
	ex, err := s.store.GetExercise(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "exercise not found")
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

// variantResponse is the payload sent to the training UI. It deliberately
// strips original_text so answers are not leaked to the client.
type variantResponse struct {
	ExerciseID     string             `json:"exercise_id"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	ExerciseType   string             `json:"exercise_type"`
	TechStack      string             `json:"tech_stack"`
	Category       string             `json:"category"`
	WeightLevel    string             `json:"weight_level"`
	VariantID      string             `json:"variant_id"`
	MarkedCode     string             `json:"marked_code"`
	CodeForEdit    string             `json:"code_for_edit,omitempty"`
	BugDescription string             `json:"bug_description,omitempty"`
	FillPoints     []variantFillPoint `json:"fill_points"`
	SpotterHints   model.SpotterHints `json:"spotter_hints"`
	TimeLimitSecs  int                `json:"time_limit_seconds"`
}

type variantFillPoint struct {
	ID         string   `json:"id"`
	Category   string   `json:"category"`
	FillType   string   `json:"fill_type"`
	Line       int      `json:"line"`
	Options    []string `json:"options,omitempty"`
	HintLevel1 string   `json:"hint_level_1"`
	HintLevel2 string   `json:"hint_level_2"`
	HintLevel3 string   `json:"hint_level_3"`
	HintLevel4 string   `json:"hint_level_4"`
}

// buildVariantResponse masks the selected fill spans inside full_code.
func buildVariantResponse(ex model.Exercise, level string, points []model.FillPoint, variantID string) variantResponse {
	marked := ex.FullCode
	if ex.ExerciseType == model.ExerciseTypeBugfix {
		marked = ex.BuggyCode
	}
	// mask right-to-left by span position so earlier offsets stay valid
	ordered := make([]model.FillPoint, len(points))
	copy(ordered, points)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Location.SpanStart < ordered[j].Location.SpanStart })

	for i := len(ordered) - 1; i >= 0; i-- {
		p := ordered[i]
		token := "____FILL_" + p.ID + "____"
		if p.Location.SpanStart >= 0 && p.Location.SpanEnd <= len(marked) && p.Location.SpanStart < p.Location.SpanEnd {
			marked = marked[:p.Location.SpanStart] + token + marked[p.Location.SpanEnd:]
		}
	}

	resp := variantResponse{
		ExerciseID:     ex.ID,
		Title:          ex.Title,
		Description:    ex.Description,
		ExerciseType:   ex.ExerciseType,
		TechStack:      ex.TechStack,
		Category:       ex.Category,
		WeightLevel:    level,
		VariantID:      variantID,
		MarkedCode:     marked,
		BugDescription: ex.BugDescription,
		SpotterHints:   ex.SpotterHints,
		TimeLimitSecs:  ex.TimeLimitSecs,
	}
	if ex.ExerciseType == model.ExerciseTypeBugfix {
		resp.CodeForEdit = ex.BuggyCode
	}

	// always send a JSON array so the client never sees null
	resp.FillPoints = []variantFillPoint{}

	for _, p := range points {
		resp.FillPoints = append(resp.FillPoints, variantFillPoint{
			ID:         p.ID,
			Category:   p.Category,
			FillType:   p.FillType,
			Line:       p.Location.Line,
			Options:    p.Options,
			HintLevel1: p.Hint,
			HintLevel2: categoryHint2(p.Category),
			HintLevel3: "👀 接近答案：" + truncate(p.OriginalText, 20),
			HintLevel4: "📝 正确答案：" + p.OriginalText,
		})
	}
	// Bug 修复题是整段改代码：训练点提示没有"空里的内容"可言，
	// 改用每题精心编写的题目级四级辅助（引导定位与修复 bug）。
	if ex.ExerciseType == model.ExerciseTypeBugfix && ex.SpotterHints != (model.SpotterHints{}) {
		for i := range resp.FillPoints {
			resp.FillPoints[i].HintLevel1 = ex.SpotterHints.Level1
			resp.FillPoints[i].HintLevel2 = ex.SpotterHints.Level2
			resp.FillPoints[i].HintLevel3 = ex.SpotterHints.Level3
			resp.FillPoints[i].HintLevel4 = ex.SpotterHints.Level4
		}
	}
	return resp
}

func (s *Server) getVariant(w http.ResponseWriter, r *http.Request) {
	ex, err := s.store.GetExercise(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "exercise not found")
		return
	}
	level := store.ValidateWeightLevel(r.URL.Query().Get("level"))
	points := variant.SelectPoints(ex.FillPoints, level)
	resp := buildVariantResponse(ex, level, points, variant.NewVariantID())
	writeJSON(w, http.StatusOK, resp)
}

// executeRequest mirrors design section 12.3.
type executeRequest struct {
	ExerciseID  string            `json:"exercise_id"`
	Environment string            `json:"environment_id"`
	FillAnswers map[string]string `json:"fill_answers"`
	VariantID   string            `json:"variant_id"`
	WeightLevel string            `json:"weight_level"`
	HintUsage   map[string]int    `json:"hint_usage"`
	UserCode    string            `json:"user_code,omitempty"` // bugfix direct-edit mode
}

// FillAnswersKey returns the canonical key for a fill point answer:
// bare ids sent by lenient clients are normalized to the FILL_ prefix form
// required by the documented API (设计 §12.3).
func FillAnswersKey(id string) string {
	if strings.HasPrefix(id, "FILL_") {
		return id
	}
	return "FILL_" + id
}

// normalizeAnswers accepts answers keyed either by bare point id or the
// documented FILL_<id> form and returns a canonical FILL_-keyed map.
func normalizeAnswers(answers map[string]string) map[string]string {
	if len(answers) == 0 {
		return answers
	}
	out := make(map[string]string, len(answers))
	for k, v := range answers {
		out[FillAnswersKey(k)] = v
	}
	return out
}

func (s *Server) execute(w http.ResponseWriter, r *http.Request) {
	var req executeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	ex, err := s.store.GetExercise(req.ExerciseID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "exercise not found")
		return
	}

	level := store.ValidateWeightLevel(req.WeightLevel)
	hintUsage := req.HintUsage
	if hintUsage == nil {
		hintUsage = map[string]int{}
	}
	// normalize fill answers so both FILL_<id> and bare <id> keys are accepted
	answers := normalizeAnswers(req.FillAnswers)
	req.FillAnswers = answers

	result := model.ExecuteResult{
		ExpectedOutput: ex.ExpectedOutput,
		Logs:           []string{},
		FillResults:    []model.FillResult{},
	}

	// bugfix type: direct code editing
	if ex.ExerciseType == model.ExerciseTypeBugfix {
		userCode := strings.TrimSpace(req.UserCode)
		if userCode == "" {
			// fall back to filling the single bug point into buggy code
			points := variant.SelectPoints(ex.FillPoints, level)
			userCode = variant.ApplyAnswers(ex.BuggyCode, points, req.FillAnswers)
		}
		toRun := userCode + ex.ValidationCode
		run, runErr := s.engine.Run(r.Context(), ex.TechStack, toRun, time.Duration(ex.TimeLimitSecs)*time.Second)
		result.Output = run.Stdout
		result.ExecutionTimeMS = run.TimeMS
		if runErr != nil {
			result.Success = false
			result.ErrorMessage = engine.FormatError(runErr, run.Stderr)
			result.Logs = append(result.Logs, "❌ 执行失败", result.ErrorMessage)
			writeJSON(w, http.StatusOK, result)
			return
		}
		score, passed, fillResults := judge.GradeBugfix(ex, userCode, run.Stdout, hintUsage)
		result.Success = true
		result.Score = score
		result.Grade = judge.GradeForScore(score)
		result.IsCorrect = passed == len(ex.TestCases) && passed > 0
		result.FillResults = fillResults
		result.WeakAnalysis = judge.WeakAnalysis(fillResults, hintUsage)
		result.CorrectAnswer = ex.FullCode
		result.Explanation = ex.Explanation
		result.Logs = append(result.Logs, "✅ 执行成功")
		if result.IsCorrect {
			result.Logs = append(result.Logs, "🎉 恭喜，全部测试通过！")
		} else {
			result.Logs = append(result.Logs, fmt.Sprintf("⚠️ 测试通过 %d/%d", passed, len(ex.TestCases)))
		}
		s.recordAttempt(ex, req, &result, hintUsage, level, s.userID(r))
		writeJSON(w, http.StatusOK, result)
		return
	}

	// fill / command / config types
	points := variant.SelectPoints(ex.FillPoints, level)
	// grade exactly what the user answered so results match the variant they saw
	if answered := variant.FilterForAnswers(ex.FillPoints, req.WeightLevel, req.FillAnswers); len(answered) > 0 {
		points = answered
	}
	rebuilt := variant.ApplyAnswers(ex.FullCode, points, req.FillAnswers)

	// FastAPI（python/FastAPI 子类）是答案匹配题：需要 fastapi 运行时且无
	// stdout 输出，不本地执行，与命令/配置题走同一格分路径。
	// "datastructure" 不是语言栈：题目用 Python 实现，按 python 引擎真实执行。
	runStack := ex.TechStack
	if runStack == "datastructure" {
		runStack = "python"
	}
	executable := (runStack == "python" || runStack == "go") && ex.Subcategory != "FastAPI"
	if executable {
		run, runErr := s.engine.Run(r.Context(), runStack, rebuilt, time.Duration(ex.TimeLimitSecs)*time.Second)
		result.Output = run.Stdout
		result.ExecutionTimeMS = run.TimeMS
		fillResults, score := judge.GradeFill(ex, points, req.FillAnswers, run.Stdout, ex.ExpectedOutput, hintUsage)
		result.FillResults = fillResults
		result.Score = score
		result.Grade = judge.GradeForScore(score)
		if runErr != nil {
			result.Success = false
			result.ErrorMessage = engine.FormatError(runErr, run.Stderr)
			result.IsCorrect = false
			result.WeakAnalysis = judge.WeakAnalysis(fillResults, hintUsage)
			result.Logs = append(result.Logs, "❌ 执行失败", result.ErrorMessage)
			s.recordAttempt(ex, req, &result, hintUsage, level, s.userID(r))
			writeJSON(w, http.StatusOK, result)
			return
		}
		result.Success = true
		expected := engine.EffectiveOutput(run)
		result.IsCorrect = judge.OutputsMatch(run.Stdout, ex.ExpectedOutput)
		result.WeakAnalysis = judge.WeakAnalysis(fillResults, hintUsage)
		result.Logs = append(result.Logs, "✅ 执行成功")
		if result.IsCorrect {
			result.Logs = append(result.Logs, "🎉 恭喜，输出与预期一致！")
		} else if expected != "" {
			result.Logs = append(result.Logs, "⚠️ 输出与预期不一致，请检查训练点")
		}
	} else {
		// command / config / non-python: answer-match judging only
		result.Success = true
		result.Output = "✔ 校验模式：仅比对答案（该环境暂不实际执行）"
		fillResults, score := judge.GradeFill(ex, points, req.FillAnswers, "", "", hintUsage)
		result.FillResults = fillResults
		result.Score = score
		result.Grade = judge.GradeForScore(score)
		result.IsCorrect = score >= 75
		result.WeakAnalysis = judge.WeakAnalysis(fillResults, hintUsage)
		result.Logs = append(result.Logs, "✅ 答案校验完成")
		if result.IsCorrect {
			result.Logs = append(result.Logs, "🎉 全部训练点正确！")
		} else {
			result.Logs = append(result.Logs, "⚠️ 存在错误训练点，请根据解析修正")
		}
	}

	s.recordAttempt(ex, req, &result, hintUsage, level, s.userID(r))
	writeJSON(w, http.StatusOK, result)
}

// recordAttempt persists the attempt and updates trainee progress.
func (s *Server) recordAttempt(ex model.Exercise, req executeRequest, result *model.ExecuteResult, hintUsage map[string]int, level, userID string) {
	execID := fmt.Sprintf("exec-%d", time.Now().UnixNano())
	result.ExecutionID = execID
	rec := model.AttemptRecord{
		ID:              execID,
		UserID:          userID,
		ExerciseID:      ex.ID,
		VariantID:       req.VariantID,
		WeightLevel:     level,
		UserAnswers:     req.FillAnswers,
		ExecutionOutput: result.Output,
		IsCorrect:       result.IsCorrect,
		Score:           result.Score,
		Grade:           result.Grade,
		ErrorMessage:    result.ErrorMessage,
		ExecutionTimeMS: result.ExecutionTimeMS,
		HintUsage:       hintUsage,
		CreatedAt:       time.Now().UTC(),
	}
	hintTotal := 0
	for _, lvl := range hintUsage {
		hintTotal += lvl
	}
	if err := s.store.SaveAttempt(rec); err != nil {
		log.Printf("save attempt: %v", err)
	}
	if err := s.store.UpdateProgress(ex, result.Score, result.Grade, hintTotal, result.IsCorrect, userID); err != nil {
		log.Printf("update progress: %v", err)
	}
}

func (s *Server) getProgress(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.GetProgress(s.userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	recent, _ := s.store.RecentAttempts(s.userID(r), 10)
	writeJSON(w, http.StatusOK, map[string]any{"progress": p, "recent_attempts": recent})
}

func (s *Server) getWeakPoints(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.GetProgress(s.userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	// resolve weak exercise ids to titles
	items := []map[string]any{}
	for id, count := range p.WeakPoints {
		if strings.HasSuffix(id, "#hint") {
			continue
		}
		ex, err := s.store.GetExercise(id)
		title := id
		if err == nil {
			title = ex.Title
		}
		items = append(items, map[string]any{"exercise_id": id, "title": title, "fail_count": count})
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) listEnvironments(w http.ResponseWriter, _ *http.Request) {
	items, err := s.store.ListEnvironments(true)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// --- 教练端环境管理（设计 §14） ---

func (s *Server) adminListEnvironments(w http.ResponseWriter, _ *http.Request) {
	items, err := s.store.ListEnvironments(false)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) adminCreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var e model.Environment
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if e.Name == "" || e.Type == "" {
		writeErr(w, http.StatusBadRequest, "name and type are required")
		return
	}
	e.CreatedAt = time.Now().UTC()
	if err := s.store.UpsertEnvironment(e); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) adminUpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	var e model.Environment
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	existing, err := s.store.GetEnvironment(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "environment not found")
		return
	}
	e.ID = existing.ID
	e.CreatedAt = existing.CreatedAt
	if err := s.store.UpsertEnvironment(e); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) adminDeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteEnvironment(r.PathValue("id")); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// adminTestEnvironment performs a lightweight connectivity/health check.
// Local engines verify the interpreter exists; others are marked healthy by
// configuration (真实 Docker/K8s 探活为后续 Phase).
func (s *Server) adminTestEnvironment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	env, err := s.store.GetEnvironment(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "environment not found")
		return
	}
	status := "healthy"
	detail := "配置校验通过"
	switch env.Type {
	case "local":
		bin := "python3"
		if v, ok := env.Config["interpreter"].(string); ok && v != "" {
			bin = v
		}
		if _, err := exec.LookPath(bin); err != nil {
			status = "unhealthy"
			detail = "未找到解释器: " + bin
		} else {
			detail = "解释器可用: " + bin
		}
	}
	if err := s.store.SetEnvironmentHealth(id, status); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "health_status": status, "detail": detail})
}

func (s *Server) listFavorites(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Favorites(s.userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) toggleFavorite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExerciseID string `json:"exercise_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ExerciseID == "" {
		writeErr(w, http.StatusBadRequest, "exercise_id is required")
		return
	}
	if _, err := s.store.GetExercise(req.ExerciseID); err != nil {
		writeErr(w, http.StatusNotFound, "exercise not found")
		return
	}
	favorited, err := s.store.ToggleFavorite(s.userID(r), req.ExerciseID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"exercise_id": req.ExerciseID, "favorited": favorited})
}

// --- 教练端 ---

func (s *Server) createExercise(w http.ResponseWriter, r *http.Request) {
	var ex model.Exercise
	if err := json.NewDecoder(r.Body).Decode(&ex); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if ex.ID == "" || ex.Title == "" || ex.FullCode == "" {
		writeErr(w, http.StatusBadRequest, "id, title and full_code are required")
		return
	}
	if ex.CreatedAt.IsZero() {
		ex.CreatedAt = time.Now().UTC()
		ex.UpdatedAt = time.Now().UTC()
	}
	if err := s.store.UpsertExercise(ex); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

func (s *Server) updateExercise(w http.ResponseWriter, r *http.Request) {
	var ex model.Exercise
	if err := json.NewDecoder(r.Body).Decode(&ex); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	existing, err := s.store.GetExercise(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "exercise not found")
		return
	}
	if ex.ID == "" {
		ex.ID = existing.ID
	}
	ex.CreatedAt = existing.CreatedAt
	ex.UpdatedAt = time.Now().UTC()
	if err := s.store.UpsertExercise(ex); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

func (s *Server) deleteExercise(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteExercise(r.PathValue("id")); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) importExercises(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	var list []model.Exercise
	if err := json.Unmarshal(body, &list); err != nil {
		writeErr(w, http.StatusBadRequest, "expected JSON array of exercises: "+err.Error())
		return
	}
	count := 0
	for _, ex := range list {
		if ex.ID == "" || ex.Title == "" || ex.FullCode == "" {
			continue
		}
		if ex.CreatedAt.IsZero() {
			ex.CreatedAt = time.Now().UTC()
			ex.UpdatedAt = time.Now().UTC()
		}
		if err := s.store.UpsertExercise(ex); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		count++
	}
	writeJSON(w, http.StatusOK, map[string]any{"imported": count, "total": len(list)})
}

// exportExercises downloads the whole question bank as a JSON array.
func (s *Server) exportExercises(w http.ResponseWriter, _ *http.Request) {
	exs, err := s.store.ListExercises("")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="codegym-exercises-%s.json"`, time.Now().Format("20060102-150405")))
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(exs)
}

func (s *Server) listCategories(w http.ResponseWriter, _ *http.Request) {
	items, err := s.store.ListCategories()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TechStack   string `json:"tech_stack"`
		Subcategory string `json:"subcategory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TechStack == "" || req.Subcategory == "" {
		writeErr(w, http.StatusBadRequest, "tech_stack and subcategory are required")
		return
	}
	if err := s.store.CreateCategory(req.TechStack, req.Subcategory); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tech_stack": req.TechStack, "subcategory": req.Subcategory})
}

func (s *Server) renameCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TechStack      string `json:"tech_stack"`
		Subcategory    string `json:"subcategory"`
		NewSubcategory string `json:"new_subcategory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TechStack == "" || req.Subcategory == "" || req.NewSubcategory == "" {
		writeErr(w, http.StatusBadRequest, "tech_stack, subcategory and new_subcategory are required")
		return
	}
	if err := s.store.RenameCategory(req.TechStack, req.Subcategory, req.NewSubcategory); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) deleteCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TechStack   string `json:"tech_stack"`
		Subcategory string `json:"subcategory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TechStack == "" || req.Subcategory == "" {
		writeErr(w, http.StatusBadRequest, "tech_stack and subcategory are required")
		return
	}
	if err := s.store.DeleteCategory(req.TechStack, req.Subcategory); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) aiGenerate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TechStack       string `json:"tech_stack"`
		Weight          string `json:"weight"`
		KnowledgePoints string `json:"knowledge_points"`
		Count           int    `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if req.TechStack == "" {
		writeErr(w, http.StatusBadRequest, "tech_stack is required")
		return
	}
	if req.Weight == "" {
		req.Weight = model.DefaultWeightLevel
	}
	count := req.Count
	if count <= 0 {
		count = 3
	}
	if count > 5 {
		count = 5
	}
	results, err := s.generateExercises(r.Context(), req.TechStack, req.Weight, req.KnowledgePoints, count)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	ok := 0
	for _, res := range results {
		if res.OK {
			ok++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"total": len(results), "ok": ok, "items": results})
}

// --- LLM 设置（页面可配置，设计 §9） ---

func (s *Server) currentLLMSettings() (string, string) {
	ep, ok := s.store.GetSetting("llm.endpoint")
	if !ok {
		ep = s.llm.Endpoint
	}
	md, ok := s.store.GetSetting("llm.model")
	if !ok {
		md = s.llm.Model
	}
	return ep, md
}

func (s *Server) getLLMSettings(w http.ResponseWriter, r *http.Request) {
	ep, md := s.currentLLMSettings()
	models, err := s.llm.Models()
	if err != nil {
		models = nil
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"endpoint": ep,
		"model":    md,
		"models":   models,
		"online":   err == nil,
	})
}

func (s *Server) setLLMSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
		Model    string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if req.Endpoint != "" {
		if err := s.store.SetSetting("llm.endpoint", req.Endpoint); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if req.Model != "" {
		if err := s.store.SetSetting("llm.model", req.Model); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	s.llm.SetConfig(req.Endpoint, req.Model)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) testLLMSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
		Model    string `json:"model"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	ep, md := s.currentLLMSettings()
	if req.Endpoint != "" {
		ep = req.Endpoint
	}
	if req.Model != "" {
		md = req.Model
	}
	probe := llm.New(ep, md)
	models, err := probe.Models()
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "models": models, "endpoint": ep})
}

func (s *Server) adminStats(w http.ResponseWriter, _ *http.Request) {
	stats, err := s.store.Stats()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	progress, err := s.store.AllProgress()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	var totalAttempts int
	for _, p := range progress {
		totalAttempts += p.TotalAttempts
	}
	stats["avg_attempts_per_user"] = totalAttempts
	writeJSON(w, http.StatusOK, stats)
}

// categoryHint2Map gives a meaningful Level-2 teaching hint per fill-point
// category, covering the categories used across all banks.
var categoryHint2Map = map[string]string{
	// Go 并发
	"goroutine":        "📖 用 `go` 关键字并发执行一个函数调用",
	"waitgroup":        "📖 `WaitGroup` 三件套：`Add` 计入、`Done` 减计、`Wait` 等待归零",
	"mutex":            "📖 修改共享变量前要 `Lock()`，结束后 `Unlock()`",
	"rwmutex":          "📖 读多写少用 `RWMutex`：读加 `RLock()`，写加 `Lock()`",
	"map_write":        "📖 并发写 map 必须放在 `Lock()`/`Unlock()` 之间：`m[key] = value`",
	"critical_section": "📖 该行位于临界区内：写清受锁保护的对共享变量的实际操作", "once": "📖 `sync.Once` 保证初始化只执行一次，即使并发调用",
	"atomic":            "📖 `sync/atomic` 提供无锁原子操作（Add/Load/Store/CAS）",
	"cond":              "📖 条件不满足时 `Wait()` 释放锁等待，条件满足后 `Broadcast()` 唤醒",
	"pool":              "📖 `sync.Pool` 缓存复用对象：`Get` 取、`Put` 还，降低分配压力",
	"channel":           "📖 用 `ch <- v` 发送、`<-ch` 接收；`v, ok := <-ch` 的 ok 为 false 表示通道已关闭",
	"channel_direction": "📖 `chan<-` 只写、`<-chan` 只读，用于约束函数对通道的使用",
	"channel_pattern":   "📖 常见的通道模式：生成器/WorkerPool/流水线，用通道在协程间传数据",
	"select":            "📖 `select` 同时监听多个通道，任一就绪即执行对应分支",
	"context":           "📖 context 通过 `Done()` 通知取消/超时，向下游传递控制信号",
	"timeout":           "📖 用 `time.After` 配合 `select` 做超时等待",
	"ticker":            "📖 `time.Ticker` 周期向 `.C` 发送时间，做限流/心跳/定时任务",
	"case":              "📖 分支值要与被判断对象类型一致，注意匹配的写法",
	"default":           "📖 默认值/兜底：switch 无匹配分支或函数参数未传时用的默认",
	// Python 协程
	"coroutine":   "📖 `async def` 定义协程函数：调用只得到协程对象，须由事件循环驱动执行",
	"event_loop":  "📖 `asyncio.run(入口协程)` 创建事件循环、运行至完成再关闭；进程内不要嵌套调用",
	"await_kw":    "📖 `await 可等待对象`：挂起当前协程，等其完成后取回返回值",
	"task":        "📖 `asyncio.create_task(协程)` 把协程包装成 Task 尽快调度，`await task` 收割结果",
	"gather":      "📖 `await asyncio.gather(c1, c2…)` 并发运行多个协程，按传入顺序返回结果列表",
	"async_sleep": "📖 `asyncio.sleep` 只挂起当前协程并把控制权交还事件循环，不阻塞线程",
	"wait_for":    "📖 `asyncio.wait_for(协程, timeout=秒)` 超时取消并抛出 `asyncio.TimeoutError`",
	"async_iter":  "📖 异步生成器（`async def + yield`）要用 `async for` 驱动，每次 yield 处可挂起",
	"async_queue": "📖 asyncio.Queue 的 put/get 都可等待：队列满/空时挂起，是协程间安全的数据通道",
	"semaphore":   "📖 `async with Semaphore(n)` 限制同时进入的协程数量，配额用尽则挂起等待",
	// Python 多线程
	"thread":          "📖 `threading.Thread(target=函数, args=(…))` 创建线程：`start()` 启动（勿直接调 run）",
	"thread_args":     "📖 线程入口函数的参数用 `args=元组` 按位置传入，单元素也要尾逗号",
	"thread_join":     "📖 `join()` 阻塞调用方直到目标线程结束，是保证先后顺序的基本手段",
	"thread_state":    "📖 `is_alive()` 查询运行状态；`daemon=True` 的线程随主进程退出被终止、不阻塞退出",
	"shared_state":    "📖 线程共享同一份内存：单一内置操作（append 等）受 GIL 保护原子，复合读改写必须加锁",
	"lock_py":         "📖 threading.Lock 保护临界区：优先 `with lock:`；显式 acquire 后务必在 finally 中 release",
	"rlock":           "📖 RLock 可重入：同一线程可多次 acquire（内部计数），须等量 release 才放行他人",
	"semaphore_py":    "📖 Semaphore(n) 发放 n 个许可限制并发进入数；释放后许可归还，等待者依序获得，人人都能完成",
	"event_py":        "📖 Event 是一次性广播开关：`wait()` 阻塞等待，`set()` 放行所有等待者（clear 复位）",
	"barrier":         "📖 Barrier(n) 让 n 个线程在集结点互等，到齐一起放行；action 回调由最后到达者执行恰好一次",
	"thread_local":    "📖 threading.local 的属性按线程隔离：每个线程读写的是自己的独立副本",
	"queue_py":        "📖 queue.Queue 线程安全 FIFO：put/get 内置加锁阻塞，是线程间传值的首选容器",
	"task_done_queue": "📖 消费端每处理完一项调 `task_done()`；主线程 `q.join()` 等未决计数归零（≠队列为空）",
	"queue_sentinel":  "📖 用哨兵值（如 None）通知消费线程退出循环；多消费者时每人发一个",
	"timer_thread":    "📖 threading.Timer(秒, 函数) 是 Thread 子类：到点执行一次，到点前可 cancel()",
	"executor_thread": "📖 ThreadPoolExecutor 复用线程池：map 按提交顺序返回结果；submit 返回 Future 用 .result() 取值",
	// BigQuery
	"bq_table":       "📖 引用表用完整路径「数据集.表」，含连字符时须反引号包裹",
	"bq_select":      "📖 SELECT 列清单：需要哪列写哪列；`* EXCEPT (列)` 排除少数列；DISTINCT 去重",
	"bq_where":       "📖 WHERE 过滤行：= 等值、BETWEEN 闭区间、IN 列表、IS NULL 判空",
	"bq_aggregate":   "📖 聚合函数：COUNT(*)/SUM/AVG/MIN/MAX 忽略 NULL；COUNTIF(条件) 做条件计数",
	"bq_groupby":     "📖 分组：非聚合列必须进 GROUP BY；组级过滤用 HAVING 不能用 WHERE；支持按位置 GROUP BY 1",
	"bq_cte":         "📖 CTE：`WITH 名 AS (子查询) SELECT … FROM 名`，让复杂查询分层可读",
	"bq_union":       "📖 UNION 纵向合并结果集：ALL 保留重复，DISTINCT 去重；两边列数与类型须兼容",
	"bq_string":      "📖 字符串函数：CONCAT 拼接、SUBSTR 截取（1 起始）、UPPER/LOWER、LENGTH；正则用 REGEXP_CONTAINS",
	"bq_array":       "📖 数组：ARRAY_AGG 组内聚合成数组、ARRAY_LENGTH 求长度、UNNEST 展开为行",
	"bq_unnest":      "📖 `表, UNNEST(数组列) AS 元素` 把数组拆成多行（逗号为隐式交叉连接）",
	"bq_date":        "📖 日期：EXTRACT(PART FROM 时间) 取部件、DATE_DIFF(a,b,PART) 差、DATE_ADD/DATE_SUB 加减 INTERVAL、FORMAT_DATE 格式化",
	"bq_window":      "📖 窗口函数：ROW_NUMBER/RANK/SUM OVER (PARTITION BY 分组 ORDER BY 排序)，逐行计算不折叠行",
	"bq_if_case":     "📖 条件表达式：IF(条件,真,假) 二分支；CASE WHEN 条件 THEN 值 … ELSE 兜底 END 多分支",
	"bq_coalesce":    "📖 COALESCE(列, 默认) 返回第一个非 NULL 值，处理空值兜底",
	"bq_math":        "📖 数学：ROUND 四舍五入、MOD 取余、CAST 类型转换、SAFE_DIVIDE 除零返回 NULL",
	"bq_json":        "📖 JSON：JSON 类型列直接 `列.字段` 导航；字符串列用 JSON_VALUE(json, \"$.路径\")",
	"bq_struct":      "📖 STRUCT 复合类型：`STRUCT(列…)` 构造、`结果.字段` 导航；REPEATED STRUCT 配 UNNEST 逐行取字段",
	"bq_wildcard":    "📖 通配表：表名带 `*` 跨多张分片表查询，伪列 `_TABLE_SUFFIX` 存放后缀并可在 WHERE 裁剪",
	"bq_info_schema": "📖 元数据：查 `数据集.INFORMATION_SCHEMA.TABLES/COLUMNS` 获取行数、大小、列结构",
	"bq_qualify":     "📖 `QUALIFY 窗口条件` 在 GROUP BY 之后直接过滤窗口函数结果，组内 Top-N 免套 CTE",
	"bq_script":      "📖 脚本：`DECLARE 名 类型 DEFAULT 初值;` 声明、`SET 名 = 值;` 赋值，可配合 FOR/WHILE 控制流",
	"bq_sample":      "📖 采样：`TABLESAMPLE SYSTEM (n PERCENT)` 随机采样，探查大表可大幅降本",
	"bq_timetravel":  "📖 时间旅行：`FOR SYSTEM_TIME AS OF 时间戳` 读取历史快照（过去 7 天），用于恢复/审计/对比",
	// Terraform
	"terra_cli":         "📖 CLI 命令：init 初始化、plan 干跑、apply 落地、destroy 销毁、validate/fmt 校验格式、show/state/workspace 管理",
	"terra_resource":    "📖 `resource \"类型\" \"名称\" { 属性 = 值 }` 定义托管资源；「类型.名称.属性」可跨资源引用",
	"terra_provider":    "📖 `provider \"云\" { … }` 声明对接的云厂商与区域；required_version 约束 CLI 版本",
	"terra_variable":    "📖 `variable \"名\" { type=… default=… sensitive=… }` 输入参数，用 `var.名` 引用",
	"terra_output":      "📖 `output \"名\" { value = 表达式 }` 暴露资源属性，跨模块/给外部使用",
	"terra_local":       "📖 `locals { 名 = 值 }` 定义仅本配置内复用的中间值，用 `local.名` 引用",
	"terra_data":        "📖 `data \"类型\" \"名称\" { … }` 只读引用外部已有资源，用 `data.类型.名称.属性`",
	"terra_module":      "📖 `module \"名\" { source=地址 version=… }` 复用封装好的资源集合（注册表/Git/本地目录）",
	"terra_depends":     "📖 `depends_on = [资源地址]` 显式声明无引用的先后依赖",
	"terra_meta":        "📖 count 按数量复制（count.index 区分）；for_each 按集合遍历（each.key/value）；lifecycle 控制生命周期",
	"terra_state":       "📖 state 记录托管资源实际状态：state list/show 查看，import 纳入外部资源；生产用远端 backend 共享",
	"terra_backend":     "📖 `terraform { backend \"s3\" { … } }` 把 state 存远端共享并加锁，团队协作不覆盖",
	"terra_provisioner": "📖 `provisioner \"local-exec\"/\"remote-exec\"` 在创建时执行脚本；非幂等，尽量少用",
	"terra_varfile":     "📖 `-var-file=xxx.tfvars` 用文件注入变量；`*.auto.tfvars` 自动加载",
	"terra_func":        "📖 内置函数：join/lookup/element/toset；`列表[*].属性` splat 展开取集合属性",
	"terra_import":      "📖 `terraform import 资源地址 外部ID` 把手工资源纳入 state（需先写对应配置）",
	// Ansible
	"ans_cli":         "📖 CLI：ansible 跑 ad-hoc（-m 模块 -a 参数）、ansible-playbook 跑剧本、--syntax-check/--check 干跑、vault/galaxy/inventory 管理",
	"ans_inventory":   "📖 清单 INI：`[组名]` 分组、`[组:vars]` 组变量、`[组:children]` 嵌套组",
	"ans_play":        "📖 play 结构：`hosts:` 目标 + `tasks:` 任务（每任务 name + 模块）；一个 playbook 可有多个 play",
	"ans_privilege":   "📖 `become: true` 提权（sudo/su）、`remote_user:` 登录身份、`connection: local` 本机执行",
	"ans_variable":    "📖 变量：vars 段定义、`{{ 变量 }}` Jinja2 插值、register 捕获任务结果、vars_files 加载外部文件",
	"ans_condition":   "📖 `when: 表达式` 按条件执行；用事实变量/注册结果，支持 and/or/not",
	"ans_loop":        "📖 `loop:` 遍历列表/字典，块内用 `{{ item }}`；旧式 with_items 等价",
	"ans_handler":     "📖 `notify: 名` 在任务变更时触发；`handlers:` 段定义，play 末尾只跑一次，天然幂等",
	"ans_template":    "📖 `template:` 用 Jinja2 渲染配置（src 模板 → dest 目标），常配 notify 重启服务",
	"ans_module":      "📖 常用模块：copy 传文件、template 渲染、service 管服务（state/enabled）、apt/yum 管包、file 管文件态",
	"ans_facts":       "📖 `gather_facts` 采集主机事实（ansible_ 前缀）；ad-hoc 可用 `-m setup` dump",
	"ans_role":        "📖 `roles:` 引用角色（tasks/handlers/templates/vars 目录）；ansible-galaxy init 生成骨架",
	"ans_tags":        "📖 任务 `tags:` 标记，`--tags/--skip-tags` 选择性执行；always 标签必跑",
	"ans_idempotency": "📖 幂等 = 重复执行结果一致：优先用声明式模块而非 shell，描述目标态而非操作步骤",
	// AWS
	"aws_cli":         "📖 `aws 服务 子命令 --参数`：子命令是动词（create/describe/delete…），资源 id/arn 用 --xxx-id/--xxx-arn 指定",
	"aws_ec2":         "📖 EC2：run-instances 创建、describe/stop/start/terminate 管理；--image-id 选镜像、--instance-type 选规格",
	"aws_ec2_storage": "📖 EBS：create-volume 建卷、attach-volume 挂载、create-snapshot 做快照、create-image 制 AMI；快照是卷的备份",
	"aws_ec2_net":     "📖 网络：authorize-security-group-ingress/egress 控制出入站；弹性 IP（allocate/associate）绑定公网地址",
	"aws_iam":         "📖 IAM：create-user/list-users 管用户、attach-*-policy 授权、create-access-key 发密钥、sts get-caller-identity 验身份",
	"aws_iam_policy":  "📖 策略 JSON：Effect（Allow/Deny）+ Action（服务:操作）+ Resource（ARN）+ Condition（约束）；最小权限只给所需",
	"aws_eks":         "📖 EKS：eksctl/aws eks 管集群与节点组；update-kubeconfig 连集群；kubectl 操作工作负载",
	"aws_eks_iam":     "📖 EKS↔IAM：集群 OIDC 提供方 + ServiceAccount 注解绑定 IAM 角色，Pod 免密钥拿权限",
	"aws_s3":          "📖 S3：s3api 管桶（create-bucket/put-*/head-object），s3 命令传文件（cp/sync/rm/mv），presign 生成临时访问 URL",
	"aws_s3_policy":   "📖 桶策略 JSON：Principal 主体 + Action(s3:*) + Resource(arn:aws:s3:::桶/前缀) 控制桶级访问",
	"aws_s3_storage":  "📖 存储级别：Standard/IA/Glacier/Deep Archive；生命周期按规则降级/删除；SSE-S3/SSE-KMS 加密",
	"aws_vpc":         "📖 VPC：create-vpc/subnet/igw/route-table/nat-gateway；路由表把子网接到网关；安全组（实例级）+ NACL（子网级）",
	"aws_rds":         "📖 RDS：create-db-instance 建库、create-db-snapshot 备份、restore 恢复、read-replica 读写分离、multi-az 高可用",
	// Agentic AI（Python）
	"agent_message": "📖 消息由 role 与 content 组成：system 设定行为，user 输入，assistant 回复，tool 回传工具结果",
	"agent_params":  "📖 调用参数：temperature 控制随机性，max_tokens 限制长度，token 可按字符数÷4 粗估",
	"agent_tool":    "📖 工具用 JSON Schema 声明名称/参数/必填项；模型发出调用后由代码执行，结果以 tool 角色回传",
	"agent_loop":    "📖 Agent 循环：模型决策 → 执行工具 → 观察结果回填消息，直到无工具调用或达到最大步数",
	"agent_memory":  "📖 上下文有限：滑动窗口/摘要压缩/关键词或向量检索/切块，都是为了塞下最相关的信息",
	"agent_output":  "📖 结构化输出：让模型输出 JSON，解析失败就截取/修复/重试，字段要做必填与类型校验",
	"agent_multi":   "📖 多智能体：按角色分工，通过 handoff 交接或按意图路由分发，最后聚合结果",
	"agent_ops":     "📖 生产要点：指数退避重试、按 token 单价核算成本、结构化日志便于回放轨迹",
	// Go 类型系统
	"go_types_basic":   "📖 Go 静态类型：var 有零值约定，:= 推导类型，type 定义新类型（与别名 = 不同）",
	"go_types_convert": "📖 不同类型不能直接运算：T(x) 显式转换；string 与 []byte/rune 互转注意字节与字符",
	"go_types_assert":  "📖 类型断言 x.(T) 带 ok 不 panic；type switch 一次断言多种类型",
	"go_struct":        "📖 结构体：值接收者操作副本，指针接收者可修改原值；嵌入字段被提升；全可比较则可作 map 键",
	"go_interface":     "📖 接口隐式实现；接口值=(类型,值)，带类型的 nil 不等于 nil 接口；any 是空接口",
	"go_generics":      "📖 泛型：[T 约束] 声明类型参数，约束是类型集合，~T 表示底层类型",
	"go_pointer":       "📖 值传参是副本，指针传参共享；map/slice 底层共享，无需取地址即可修改",
	// 数据结构系列（Python）
	"ds_array":      "📖 数组：连续内存 O(1) 随机访问；双指针/滑动窗口/前缀和/差分是四大套路",
	"ds_linkedlist": "📖 链表：节点+指针；虚拟头节点、快慢指针、就地反转是高频套路",
	"ds_stackqueue": "📖 栈 LIFO 后进先出，队列 FIFO 先进先出；单调栈/单调队列维护有序性",
	"ds_hash":       "📖 哈希表用哈希函数定位桶；冲突用链地址或开放寻址；查询均摊 O(1)",
	"ds_tree":       "📖 二叉树递归套路；BST 中序有序；堆是完全二叉树；Trie 按字符分叉",
	"ds_graph":      "📖 图=邻接表；BFS 求无权最短路，DFS 走到底，拓扑排序处理依赖，并查集管连通",
	"bq_schema":     "📖 表优化：PARTITION BY 按时间分区裁剪扫描，CLUSTER BY 加速等值过滤；过滤条件要直接写在分区键上",
	// Go / Python 语法
	"syntax":          "📖 先确认语句的正确语法（关键字、括号、缩进/换行）",
	"variables":       "📖 变量声明与赋值：注意类型与命名",
	"number":          "📖 数值要算对，注意整数/浮点与运算优先级",
	"constant":        "📖 常量在编译期固定，运行期不可改",
	"if":              "📖 判断条件要覆盖到正确分支，注意比较与取模",
	"condition":       "📖 写出使分支成立的条件表达式，注意比较/取模运算",
	"for":             "📖 for 的三段式：初始、条件、步进，缺一不可",
	"while":           "📖 循环条件要能最终变假，否则死循环",
	"loop":            "📖 注意循环边界（含首不含尾）与步进",
	"iteration":       "📖 遍历序列时注意当前元素与下标的关系",
	"iterator":        "📖 明确在遍历哪个可迭代对象",
	"range":           "📖 `range` 同时产出下标与值，用 `_` 忽略其一",
	"break":           "📖 `break` 立即跳出整个循环",
	"continue":        "📖 `continue` 跳过本次迭代剩下的语句",
	"list":            "📖 列表按下标访问，下标从 0 开始",
	"slice":           "📖 切片 `[start:end]` 含首不含尾；`append` 追加、`len` 长度",
	"tuple":           "📖 元组不可变，按下标访问或解包",
	"unpack":          "📖 变量数量与序列元素数一致即可解包",
	"dict":            "📖 字典 `d[键]` 取值、`d[键]=值` 写入；`get` 可给默认值",
	"set":             "📖 集合自动去重，`len` 得元素个数",
	"string":          "📖 字符串常用方法：upper/lower/strip/split/join/replace/find",
	"method":          "📖 方法绑定在类型上，注意接收者（值/指针）",
	"builtins":        "📖 内置函数如 len/sum/max/min/sorted 直接可用",
	"function":        "📖 函数 `def/func` 定义，`return` 返回值，注意参数与默认值",
	"variadic":        "📖 变长参数收集成切片/元组再处理",
	"recursion":       "📖 递归 = 基线条件 + 递推式，缺一不可",
	"lambda":          "📖 `lambda` 定义匿名函数，结果就是表达式值",
	"comprehension":   "📖 `[表达式 for x in 序列 if 条件]` 一行生成列表",
	"format":          "📖 格式化输出：f-string 的 `{}` 或格式占位符",
	"conversion":      "📖 类型转换用 `T(值)` 显式转换（int/str/float）",
	"exception":       "📖 想清楚会抛什么异常，用对应类型捕获并处理",
	"module":          "📖 先 `import` 引入包，再用 `包.函数` 调用",
	"class":           "📖 类里 `__init__` 初始化、`self` 指实例，方法靠实例调用",
	"closure":         "📖 闭包捕获外层变量，可带状态地反复调用",
	"expression":      "📖 想出结果表达式，可用运算符/函数组合得到目标值",
	"statement":       "📖 这里是一条语句，想清楚它的作用再补全",
	"keyword":         "📖 想清楚这个关键字的语义（go/defer/select/close…）",
	"argument":        "📖 参数取值要与函数定义匹配",
	"type_annotation": "📖 类型注解要写对：int/str/bool/自定义类型",
	"init":            "📖 初始化逻辑：声明变量并赋初值",
	"generics":        "📖 泛型用 `[T 约束]` 声明，约束限定可用的类型",
	"interface":       "📖 类型实现接口的全部方法即满足接口",
	"struct":          "📖 结构体用 `Type{字段: 值}` 初始化，`.` 访问字段",
	"pointer":         "📖 `&` 取地址、`*` 解引用；指针参数可修改调用方变量",
	"recursion+":      "📖 先找基线，再写递推",
	// K8s / OpenShift / Docker
	"command_flag":      "📖 命令的选项/参数，注意短参数与长参数往往等价",
	"command_verb":      "📖 想清楚这条命令的动词（动词决定操作）",
	"resource":          "📖 资源类型有全称和缩写（如 deployments 与 deploy）",
	"secret":            "📖 Secret 的 `data` 值要 base64 编码，`stringData` 可写明文",
	"secret_type":       "📖 Secret 的 type：Opaque / kubernetes.io/tls 等",
	"secret_data":       "📖 `data` 值必须 base64；`stringData` 自动编码",
	"configmap":         "📖 ConfigMap 存非敏感配置，Pod 用 env 或 volume 引用",
	"config":            "📖 找出该配置对应的键，填上正确的取值（注意数字/字符串引号、单位）",
	"docker_env":        "📖 Docker 环境变量：Dockerfile 用 `ENV 键=值`；compose 的 `environment:` 下用 `键: 值` 逐行设置（字符串可加引号）",
	"env":               "📖 env 注入环境变量：直接 value，或 valueFrom 引用 ConfigMap/Secret",
	"volume":            "📖 卷先声明来源（emptyDir/PVC/ConfigMap…），再用 volumeMounts 挂载",
	"mount_path":        "📖 挂载到容器内的绝对路径",
	"host_path":         "📖 hostPath 直连节点上的目录，注意只读使用",
	"probe":             "📖 探针三选一：httpGet / tcpSocket / exec；配好端口与周期",
	"probe_port":        "📖 探测端口要对齐容器监听端口",
	"storage":           "📖 存储容量用二进制单位（Mi/Gi），访问模式决定读写法",
	"access_mode":       "📖 访问模式：ReadWriteOnce / ReadOnlyMany / ReadWriteMany",
	"pvc":               "📖 PVC 声明存储需求并绑定 PV，Pod 用 claimName 引用",
	"pvc_name":          "📖 引用的 PVC 名称要与已创建的 PVC 一致",
	"api_version":       "📖 资源属于哪个 API 组、kind 决定了 `apiVersion`",
	"service_type":      "📖 Service 类型：ClusterIP / NodePort / LoadBalancer",
	"labels":            "📖 标签是键值对，选择器按标签匹配资源",
	"label":             "📖 标签是键值对，`key=value` 或 `key: value` 形式",
	"annotation":        "📖 注解是非标识性元数据，键值形式，不用于选择",
	"selector":          "📖 selector 按标签选择后端 Pod，键值要与 Pod 标签一致",
	"ports":             "📖 `\"主机端口:容器端口\"` 发布端口",
	"port":              "📖 端口要与你声明的目标端口一致",
	"replicas":          "📖 期望副本数决定同时运行的实例数",
	"image":             "📖 镜像名 + 标签（如 nginx:1.25）",
	"host":              "📖 对外的域名/主机名",
	"path":              "📖 URL 路径，注意路径类型 Prefix/Exact",
	"uri":               "📖 源码/镜像仓库地址",
	"termination":       "📖 TLS 终止方式：edge / passthrough / reencrypt",
	"strategy":          "📖 构建策略决定如何构建（Source/Docker/Custom）",
	"build_strategy":    "📖 构建策略：Source 从源码构建、Docker 按 Dockerfile",
	"reclaim_policy":    "📖 PV 释放后回收策略：Retain / Delete / Recycle",
	"provisioner":       "📖 存储插件（供应器），决定 PV 如何动态创建",
	"tls":               "📖 TLS 配置：termination 决定在哪里终止 HTTPS",
	"trigger":           "📖 触发器让 DC 在配置/镜像变化时自动部署",
	"depends_on":        "📖 声明服务依赖，控制启动顺序",
	"restart":           "📖 重启策略：always / unless-stopped / on-failure",
	"restart_policy":    "📖 Job/CronJob 中 restartPolicy 只能 Never 或 OnFailure",
	"serviceaccount":    "📖 Pod 用 serviceAccountName 指定身份访问 API",
	"init_container":    "📖 init 容器在主容器前依序执行，用于初始化",
	"security":          "📖 安全上下文：runAsUser/runAsNonRoot/只读根文件系统",
	"lifecycle":         "📖 生命周期钩子：preStop 在终止前执行",
	"healthcheck":       "📖 Dockerfile 用 HEALTHCHECK 声明容器健康检查",
	"entrypoint":        "📖 ENTRYPOINT 固定入口命令，不被 run 参数覆盖",
	"cmd":               "📖 明确这条指令的作用，补上正确的指令名或参数",
	"workdir":           "📖 WORKDIR 设定工作目录，后续指令相对它执行",
	"user":              "📖 USER 切换运行用户（通常非 root）",
	"expose_port":       "📖 EXPOSE 声明容器监听端口",
	"run_cmd":           "📖 RUN 在构建阶段执行的命令",
	"base_image":        "📖 FROM 指定基础镜像名及标签",
	"git_uri":           "📖 构建源码的 Git 仓库地址",
	"source":            "📖 构建来源类型：Git / Binary / Image",
	"cron":              "📖 cron 表达式：分 时 日 月 周",
	"dns":               "📖 dnsPolicy 控制 Pod 的 DNS 解析策略",
	"limit_type":        "📖 LimitRange 作用对象：Container / Pod",
	"utilization":       "📖 HPA 目标 CPU 利用率百分比",
	"volumes":           "📖 volumes 先挂载 `卷:容器路径`，再在顶层声明命名卷",
	"volumes_decl":      "📖 顶层 volumes 里声明命名卷",
	"image_pull_policy": "📖 拉取策略：Always / Never / IfNotPresent",
	"add":               "📖 把元素追加/累加到目标（append / +=）",
	"signature":         "📖 函数签名：参数类型与返回值类型要写对",
	"server":            "📖 服务器/API 端点地址",
	"service":           "📖 要暴露或引用的 Service 名称",
	"copy":              "📖 复制内容：copy(dst, src) 或切片拷贝",
	"arg":               "📖 参数的取值与传递要匹配函数定义",
	"error":             "📖 错误处理：检查 err，返回 nil 表示成功",
	"type":              "📖 类型声明：`type 名 底层类型`",
	"pattern":           "📖 代码模式的关键写法，想清楚模式步骤再补全",
	"project":           "📖 项目名称要唯一且符合命名规范",
	"metadata":          "📖 metadata 里的名称/标签要合理且唯一",
	"annotations":       "📖 注解是非标识性元数据，键值形式",
	"taint_effect":      "📖 污点效果：NoSchedule / PreferNoSchedule / NoExecute",
}

// categoryHint2 returns a meaningful Level-2 teaching hint for a fill point
// category (instead of a generic "look at the context" placeholder).
func categoryHint2(category string) string {
	if h, ok := categoryHint2Map[category]; ok {
		return h
	}
	return "📖 想清楚这里的「" + category + "」在代码中起什么作用，再补全"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
