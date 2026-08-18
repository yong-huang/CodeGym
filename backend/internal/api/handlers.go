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
			HintLevel2: categoryHint2(p.Category),
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

// categoryHint2Map gives a meaningful Level-2 teaching hint per fill-point
// category, covering the categories used across all banks.
var categoryHint2Map = map[string]string{
	// Go 并发
	"goroutine":      "📖 用 `go` 关键字并发执行一个函数调用",
	"waitgroup":      "📖 `WaitGroup` 三件套：`Add` 计入、`Done` 减计、`Wait` 等待归零",
	"mutex":          "📖 修改共享变量前要 `Lock()`，结束后 `Unlock()`",
	"rwmutex":        "📖 读多写少用 `RWMutex`：读加 `RLock()`，写加 `Lock()`",
	"once":           "📖 `sync.Once` 保证初始化只执行一次，即使并发调用",
	"atomic":         "📖 `sync/atomic` 提供无锁原子操作（Add/Load/Store/CAS）",
	"cond":           "📖 条件不满足时 `Wait()` 释放锁等待，条件满足后 `Broadcast()` 唤醒",
	"pool":           "📖 `sync.Pool` 缓存复用对象：`Get` 取、`Put` 还，降低分配压力",
	"channel":        "📖 用 `ch <- v` 发送、`<-ch` 接收；带缓冲容量可避免阻塞",
	"channel_direction": "📖 `chan<-` 只写、`<-chan` 只读，用于约束函数对通道的使用",
	"channel_pattern": "📖 常见的通道模式：生成器/WorkerPool/流水线，用通道在协程间传数据",
	"select":         "📖 `select` 同时监听多个通道，任一就绪即执行对应分支",
	"context":        "📖 context 通过 `Done()` 通知取消/超时，向下游传递控制信号",
	"timeout":        "📖 用 `time.After` 配合 `select` 做超时等待",
	"ticker":         "📖 `time.Ticker` 周期向 `.C` 发送时间，做限流/心跳/定时任务",
	"case":           "📖 分支值要与被判断对象类型一致，注意匹配的写法",
	"default":        "📖 默认值/兜底：switch 无匹配分支或函数参数未传时用的默认",
	// Go / Python 语法
	"syntax":        "📖 先确认语句的正确语法（关键字、括号、缩进/换行）",
	"variables":     "📖 变量声明与赋值：注意类型与命名",
	"number":        "📖 数值要算对，注意整数/浮点与运算优先级",
	"constant":      "📖 常量在编译期固定，运行期不可改",
	"if":            "📖 判断条件要覆盖到正确分支，注意比较与取模",
	"condition":     "📖 写出使分支成立的条件表达式，注意比较/取模运算",
	"for":           "📖 for 的三段式：初始、条件、步进，缺一不可",
	"while":         "📖 循环条件要能最终变假，否则死循环",
	"loop":          "📖 注意循环边界（含首不含尾）与步进",
	"iteration":     "📖 遍历序列时注意当前元素与下标的关系",
	"iterator":      "📖 明确在遍历哪个可迭代对象",
	"range":         "📖 `range` 同时产出下标与值，用 `_` 忽略其一",
	"break":         "📖 `break` 立即跳出整个循环",
	"continue":      "📖 `continue` 跳过本次迭代剩下的语句",
	"list":          "📖 列表按下标访问，下标从 0 开始",
	"slice":         "📖 切片 `[start:end]` 含首不含尾；`append` 追加、`len` 长度",
	"tuple":         "📖 元组不可变，按下标访问或解包",
	"unpack":        "📖 变量数量与序列元素数一致即可解包",
	"dict":          "📖 字典 `d[键]` 取值、`d[键]=值` 写入；`get` 可给默认值",
	"set":           "📖 集合自动去重，`len` 得元素个数",
	"string":        "📖 字符串常用方法：upper/lower/strip/split/join/replace/find",
	"method":        "📖 方法绑定在类型上，注意接收者（值/指针）",
	"builtins":      "📖 内置函数如 len/sum/max/min/sorted 直接可用",
	"function":      "📖 函数 `def/func` 定义，`return` 返回值，注意参数与默认值",
	"variadic":      "📖 变长参数收集成切片/元组再处理",
	"recursion":     "📖 递归 = 基线条件 + 递推式，缺一不可",
	"lambda":        "📖 `lambda` 定义匿名函数，结果就是表达式值",
	"comprehension": "📖 `[表达式 for x in 序列 if 条件]` 一行生成列表",
	"format":        "📖 格式化输出：f-string 的 `{}` 或格式占位符",
	"conversion":    "📖 类型转换用 `T(值)` 显式转换（int/str/float）",
	"exception":     "📖 想清楚会抛什么异常，用对应类型捕获并处理",
	"module":        "📖 先 `import` 引入包，再用 `包.函数` 调用",
	"class":         "📖 类里 `__init__` 初始化、`self` 指实例，方法靠实例调用",
	"closure":       "📖 闭包捕获外层变量，可带状态地反复调用",
	"expression":    "📖 想出结果表达式，可用运算符/函数组合得到目标值",
	"statement":     "📖 这里是一条语句，想清楚它的作用再补全",
	"keyword":       "📖 想清楚这个关键字的语义（go/defer/select/close…）",
	"argument":      "📖 参数取值要与函数定义匹配",
	"type_annotation": "📖 类型注解要写对：int/str/bool/自定义类型",
	"init":          "📖 初始化逻辑：声明变量并赋初值",
	"generics":      "📖 泛型用 `[T 约束]` 声明，约束限定可用的类型",
	"interface":     "📖 类型实现接口的全部方法即满足接口",
	"struct":        "📖 结构体用 `Type{字段: 值}` 初始化，`.` 访问字段",
	"pointer":       "📖 `&` 取地址、`*` 解引用；指针参数可修改调用方变量",
	"recursion+":    "📖 先找基线，再写递推",
	// K8s / OpenShift / Docker
	"command_flag":  "📖 命令的选项/参数，注意短参数与长参数往往等价",
	"command_verb":  "📖 想清楚这条命令的动词（动词决定操作）",
	"resource":      "📖 资源类型有全称和缩写（如 deployments 与 deploy）",
	"secret":        "📖 Secret 的 `data` 值要 base64 编码，`stringData` 可写明文",
	"secret_type":   "📖 Secret 的 type：Opaque / kubernetes.io/tls 等",
	"secret_data":   "📖 `data` 值必须 base64；`stringData` 自动编码",
	"configmap":     "📖 ConfigMap 存非敏感配置，Pod 用 env 或 volume 引用",
	"config":        "📖 找出该配置对应的键，填上正确的取值（注意数字/字符串引号、单位）",
	"env":           "📖 env 注入环境变量：直接 value，或 valueFrom 引用 ConfigMap/Secret",
	"volume":        "📖 卷先声明来源（emptyDir/PVC/ConfigMap…），再用 volumeMounts 挂载",
	"mount_path":    "📖 挂载到容器内的绝对路径",
	"host_path":     "📖 hostPath 直连节点上的目录，注意只读使用",
	"probe":         "📖 探针三选一：httpGet / tcpSocket / exec；配好端口与周期",
	"probe_port":    "📖 探测端口要对齐容器监听端口",
	"storage":       "📖 存储容量用二进制单位（Mi/Gi），访问模式决定读写法",
	"access_mode":   "📖 访问模式：ReadWriteOnce / ReadOnlyMany / ReadWriteMany",
	"pvc":           "📖 PVC 声明存储需求并绑定 PV，Pod 用 claimName 引用",
	"pvc_name":      "📖 引用的 PVC 名称要与已创建的 PVC 一致",
	"api_version":   "📖 资源属于哪个 API 组、kind 决定了 `apiVersion`",
	"service_type":  "📖 Service 类型：ClusterIP / NodePort / LoadBalancer",
	"labels":        "📖 标签是键值对，选择器按标签匹配资源",
	"label":         "📖 标签是键值对，`key=value` 或 `key: value` 形式",
	"annotation":    "📖 注解是非标识性元数据，键值形式，不用于选择",
	"selector":      "📖 selector 按标签选择后端 Pod，键值要与 Pod 标签一致",
	"ports":         "📖 `\"主机端口:容器端口\"` 发布端口",
	"port":          "📖 端口要与你声明的目标端口一致",
	"replicas":      "📖 期望副本数决定同时运行的实例数",
	"image":         "📖 镜像名 + 标签（如 nginx:1.25）",
	"host":          "📖 对外的域名/主机名",
	"path":          "📖 URL 路径，注意路径类型 Prefix/Exact",
	"uri":           "📖 源码/镜像仓库地址",
	"termination":   "📖 TLS 终止方式：edge / passthrough / reencrypt",
	"strategy":      "📖 构建策略决定如何构建（Source/Docker/Custom）",
	"build_strategy": "📖 构建策略：Source 从源码构建、Docker 按 Dockerfile",
	"reclaim_policy":"📖 PV 释放后回收策略：Retain / Delete / Recycle",
	"provisioner":   "📖 存储插件（供应器），决定 PV 如何动态创建",
	"tls":           "📖 TLS 配置：termination 决定在哪里终止 HTTPS",
	"trigger":       "📖 触发器让 DC 在配置/镜像变化时自动部署",
	"depends_on":    "📖 声明服务依赖，控制启动顺序",
	"restart":       "📖 重启策略：always / unless-stopped / on-failure",
	"restart_policy": "📖 Job/CronJob 中 restartPolicy 只能 Never 或 OnFailure",
	"serviceaccount": "📖 Pod 用 serviceAccountName 指定身份访问 API",
	"init_container": "📖 init 容器在主容器前依序执行，用于初始化",
	"security":      "📖 安全上下文：runAsUser/runAsNonRoot/只读根文件系统",
	"lifecycle":     "📖 生命周期钩子：preStop 在终止前执行",
	"healthcheck":   "📖 Dockerfile 用 HEALTHCHECK 声明容器健康检查",
	"entrypoint":    "📖 ENTRYPOINT 固定入口命令，不被 run 参数覆盖",
	"cmd":           "📖 明确这条指令的作用，补上正确的指令名或参数",
	"workdir":       "📖 WORKDIR 设定工作目录，后续指令相对它执行",
	"user":          "📖 USER 切换运行用户（通常非 root）",
	"expose_port":   "📖 EXPOSE 声明容器监听端口",
	"run_cmd":       "📖 RUN 在构建阶段执行的命令",
	"base_image":    "📖 FROM 指定基础镜像名及标签",
	"git_uri":       "📖 构建源码的 Git 仓库地址",
	"source":        "📖 构建来源类型：Git / Binary / Image",
	"cron":          "📖 cron 表达式：分 时 日 月 周",
	"dns":           "📖 dnsPolicy 控制 Pod 的 DNS 解析策略",
	"limit_type":    "📖 LimitRange 作用对象：Container / Pod",
	"utilization":   "📖 HPA 目标 CPU 利用率百分比",
	"volumes":       "📖 volumes 先挂载 `卷:容器路径`，再在顶层声明命名卷",
	"volumes_decl":  "📖 顶层 volumes 里声明命名卷",
	"image_pull_policy": "📖 拉取策略：Always / Never / IfNotPresent",
	"add":           "📖 把元素追加/累加到目标（append / +=）",
	"signature":     "📖 函数签名：参数类型与返回值类型要写对",
	"server":        "📖 服务器/API 端点地址",
	"service":       "📖 要暴露或引用的 Service 名称",
	"copy":          "📖 复制内容：copy(dst, src) 或切片拷贝",
	"arg":           "📖 参数的取值与传递要匹配函数定义",
	"error":         "📖 错误处理：检查 err，返回 nil 表示成功",
	"type":          "📖 类型声明：`type 名 底层类型`",
	"pattern":       "📖 代码模式的关键写法，想清楚模式步骤再补全",
	"project":       "📖 项目名称要唯一且符合命名规范",
	"metadata":      "📖 metadata 里的名称/标签要合理且唯一",
	"annotations":   "📖 注解是非标识性元数据，键值形式",
	"taint_effect":  "📖 污点效果：NoSchedule / PreferNoSchedule / NoExecute",
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
