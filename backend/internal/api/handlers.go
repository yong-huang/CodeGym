package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"codegym/backend/internal/engine"
	"codegym/backend/internal/judge"
	"codegym/backend/internal/model"
	"codegym/backend/internal/store"
	"codegym/backend/internal/variant"
)

// Server wires the store and engine into HTTP handlers.
type Server struct {
	store  *store.Store
	engine *engine.Engine
	now    func() time.Time
}

// New creates an API server.
func New(st *store.Store, eng *engine.Engine) *Server {
	return &Server{store: st, engine: eng, now: time.Now}
}

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

	mux.HandleFunc("POST /api/v1/admin/exercises", s.createExercise)
	mux.HandleFunc("PUT /api/v1/admin/exercises/{id}", s.updateExercise)
	mux.HandleFunc("DELETE /api/v1/admin/exercises/{id}", s.deleteExercise)
	mux.HandleFunc("POST /api/v1/admin/import", s.importExercises)
	mux.HandleFunc("GET /api/v1/admin/export", s.exportExercises)
	mux.HandleFunc("GET /api/v1/admin/categories", s.listCategories)
	mux.HandleFunc("POST /api/v1/admin/categories", s.createCategory)
	mux.HandleFunc("PUT /api/v1/admin/categories", s.renameCategory)
	mux.HandleFunc("DELETE /api/v1/admin/categories", s.deleteCategory)
	mux.HandleFunc("GET /api/v1/admin/stats", s.adminStats)

	return s.withCommon(mux)
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
	ExerciseID     string                   `json:"exercise_id"`
	Title          string                   `json:"title"`
	Description    string                   `json:"description"`
	ExerciseType   string                   `json:"exercise_type"`
	TechStack      string                   `json:"tech_stack"`
	Category       string                   `json:"category"`
	WeightLevel    string                   `json:"weight_level"`
	VariantID      string                   `json:"variant_id"`
	MarkedCode     string                   `json:"marked_code"`
	CodeForEdit    string                   `json:"code_for_edit,omitempty"`
	BugDescription string                   `json:"bug_description,omitempty"`
	FillPoints     []variantFillPoint       `json:"fill_points"`
	SpotterHints   model.SpotterHints       `json:"spotter_hints"`
	TimeLimitSecs  int                      `json:"time_limit_seconds"`
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
			HintLevel2: "📖 思路：从「" + p.Category + "」出发，看看前后文再补全",
			HintLevel3: "👀 接近答案：" + truncate(p.OriginalText, 20),
			HintLevel4: "📝 正确答案：" + p.OriginalText,
		})
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
	ExerciseID   string            `json:"exercise_id"`
	Environment  string            `json:"environment_id"`
	FillAnswers  map[string]string `json:"fill_answers"`
	VariantID    string            `json:"variant_id"`
	WeightLevel  string            `json:"weight_level"`
	HintUsage    map[string]int    `json:"hint_usage"`
	UserCode     string            `json:"user_code,omitempty"` // bugfix direct-edit mode
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

	executable := ex.TechStack == "python" || ex.TechStack == "go"
	if executable {
		run, runErr := s.engine.Run(r.Context(), ex.TechStack, rebuilt, time.Duration(ex.TimeLimitSecs)*time.Second)
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
	writeJSON(w, http.StatusOK, []map[string]any{
		{"name": "python-3.10", "type": "local", "config": map[string]any{"interpreter": "python3", "timeout": "30s"}, "health_status": "healthy", "enabled": true},
		{"name": "go-1.26", "type": "local", "config": map[string]any{"toolchain": "go", "timeout": "30s"}, "health_status": "healthy", "enabled": true},
		{"name": "answer-match", "type": "local", "config": map[string]any{"description": "command/config/non-python answer matching"}, "health_status": "healthy", "enabled": true},
	})
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
