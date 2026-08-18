package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"codegym/backend/internal/model"
)

const aiGenPrompt = `你是资深 {tech_stack} 工程师与教学设计师。请生成 {count} 道 {weight} 级别的 {tech_stack} 编程训练题，仅输出一个 JSON 数组，不要任何多余文字、不要 markdown 代码块。

每道题的 JSON 结构：
{
  "id": "唯一的题号，如 ai-{tech}-001",
  "title": "简短标题",
  "description": "一句话题目描述",
  "exercise_type": "fill",
  "tech_stack": "{tech_stack}",
  "category": "知识点分类",
  "subcategory": "基础",
  "weight": "{weight}",
  "full_code": "完整可运行代码：{tech_stack} 代码，且直接运行会 print 出 expected_output；训练点位置保留原文，不要占位符",
  "expected_output": "运行 full_code 打印出的完整输出",
  "fill_points": [
    {"id": "p1", "text": "要被学员填写的原文片段", "hint": "一句方向性提示", "category": "expression 或 condition 或 statement 等", "min_weight": "easy"}
  ],
  "spotter_hints": {"level_1": "轻提示", "level_2": "中提示", "level_3": "接近答案", "level_4": "正确答案"},
  "explanation": "解题思路解析"
}

要求：
1. 训练点 2-4 个，都是 full_code 中真实存在的原文片段（填空时学员补全该片段），覆盖核心知识点。
2. full_code 必须是完整程序：{tech_stack} 的完整可运行代码，运行输出恰好等于 expected_output。
3. 每个训练点的 hint 要有教学价值；spotter_hints.level_4 给出正确答案。
4. 知识点侧重：{knowledge_points}
5. 代码保持在 20 行以内。`

// aiRawFill mirrors one generated fill point before span computation.
type aiRawFill struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	Hint       string `json:"hint"`
	Category   string `json:"category"`
	MinWeight  string `json:"min_weight"`
}

type aiRawExercise struct {
	ID             string          `json:"id"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	ExerciseType   any             `json:"exercise_type"`
	TechStack      any             `json:"tech_stack"`
	Category       any             `json:"category"`
	Subcategory    any             `json:"subcategory"`
	Weight         any             `json:"weight"`
	FullCode       string          `json:"full_code"`
	ExpectedOutput string          `json:"expected_output"`
	FillPoints     []aiRawFill     `json:"fill_points"`
	SpotterHints   json.RawMessage `json:"spotter_hints"`
	Explanation    any             `json:"explanation"`
}

// buildAIPrompt constructs the prompt for a generation request.
func buildAIPrompt(tech, weight, knowledge string, count int) string {
	p := aiGenPrompt
	p = strings.ReplaceAll(p, "{tech_stack}", tech)
	p = strings.ReplaceAll(p, "{weight}", weight)
	p = strings.ReplaceAll(p, "{knowledge_points}", knowledge)
	p = strings.ReplaceAll(p, "{count}", fmt.Sprintf("%d", count))
	return p
}

// parseAIExercises turns the raw LLM JSON into usable exercises. Span offsets
// are computed by locating each fill's text inside full_code; entries that
// fail validation are returned with an error for the caller to report.
func parseAIExercises(raw string) ([]model.Exercise, []aiGenItemResult, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var items []aiRawExercise
	if err := json.Unmarshal([]byte(cleaned), &items); err != nil {
		// tolerate a single object
		var one aiRawExercise
		if err2 := json.Unmarshal([]byte(cleaned), &one); err2 != nil {
			return nil, nil, fmt.Errorf("模型返回的不是可解析的 JSON：%v", err)
		}
		items = []aiRawExercise{one}
	}

	var exs []model.Exercise
	var results []aiGenItemResult
	for _, it := range items {
		res := aiGenItemResult{ID: it.ID, Title: it.Title}
		ex, err := buildAIExercise(it)
		if err != nil {
			res.OK = false
			res.Error = err.Error()
			results = append(results, res)
			continue
		}
		exs = append(exs, ex)
		res.OK = true
		results = append(results, res)
	}
	return exs, results, nil
}

// buildAIExercise normalizes one raw item into a model.Exercise.
func buildAIExercise(it aiRawExercise) (model.Exercise, error) {
	id := strings.TrimSpace(it.ID)
	title := strings.TrimSpace(it.Title)
	code := it.FullCode
	if id == "" || title == "" || code == "" {
		return model.Exercise{}, fmt.Errorf("缺少 id/title/full_code")
	}

	ex := model.Exercise{
		ID:             id,
		Title:          title,
		Description:    strings.TrimSpace(it.Description),
		ExerciseType:   asString(it.ExerciseType, model.ExerciseTypeFill),
		TechStack:      asString(it.TechStack, "python"),
		Category:       asString(it.Category, ""),
		Subcategory:    asString(it.Subcategory, "基础"),
		Weight:         asString(it.Weight, model.DefaultWeightLevel),
		FullCode:       code,
		ExpectedOutput: strings.TrimSpace(it.ExpectedOutput),
		TimeLimitSecs:  30,
		Enabled:        true,
		CreatedBy:      "ai",
		FillPoints: model.FillPointsConfig{
			MinFills: 1,
			MaxFills: len(it.FillPoints),
		},
	}

	// spotter hints (model may return strings or arrays; normalize to strings)
	ex.SpotterHints = normalizeSpotterHints(it.SpotterHints)

	for _, fp := range it.FillPoints {
		text := strings.TrimSpace(fp.Text)
		if text == "" {
			continue
		}
		start := strings.Index(code, text)
		if start < 0 {
			continue // cannot locate the fill text -> skip this point
		}
		fillID := fp.ID
		if fillID == "" {
			fillID = fmt.Sprintf("p%d", len(ex.FillPoints.Points)+1)
		}
		ex.FillPoints.Points = append(ex.FillPoints.Points, model.FillPoint{
			ID:           fillID,
			Location:     model.FillLocation{Type: "span", SpanStart: start, SpanEnd: start + len(text)},
			OriginalText: text,
			Category:     firstNonEmpty(fp.Category, "expression"),
			MinWeight:    firstNonEmpty(fp.MinWeight, "easy"),
			WeightScore:  1,
			Hint:         strings.TrimSpace(fp.Hint),
			FillType:     "text",
		})
	}
	if len(ex.FillPoints.Points) == 0 {
		return model.Exercise{}, fmt.Errorf("没有能在代码中定位到的训练点")
	}
	return ex, nil
}

// normalizeSpotterHints coerces the LLM's hints (string or array) into strings.
func normalizeSpotterHints(raw json.RawMessage) model.SpotterHints {
	h := model.SpotterHints{}
	if len(raw) == 0 {
		return h
	}
	// try object of strings
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return h
	}
	h.Level1 = asString(obj["level_1"], "")
	h.Level2 = asString(obj["level_2"], "")
	h.Level3 = asString(obj["level_3"], "")
	h.Level4 = asString(obj["level_4"], "")
	return h
}

// asString converts a JSON value (string or array of strings) to a string.
func asString(v any, fallback string) string {
	switch t := v.(type) {
	case string:
		if strings.TrimSpace(t) != "" {
			return strings.TrimSpace(t)
		}
	case []any:
		var parts []string
		for _, e := range t {
			if s, ok := e.(string); ok && strings.TrimSpace(s) != "" {
				parts = append(parts, strings.TrimSpace(s))
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, ",")
		}
	case []string:
		if len(t) > 0 {
			return strings.Join(t, ",")
		}
	}
	return fallback
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// aiGenItemResult reports per-exercise generation outcome to the UI.
type aiGenItemResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// generateExercises drives the whole AI generation flow.
func (s *Server) generateExercises(ctx context.Context, tech, weight, knowledge string, count int) ([]aiGenItemResult, error) {
	prompt := buildAIPrompt(tech, weight, knowledge, count)
	raw, err := s.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}
	exs, results, err := parseAIExercises(raw)
	if err != nil {
		return nil, err
	}
	for _, ex := range exs {
		if err := s.store.UpsertExercise(ex); err != nil {
			results = append(results, aiGenItemResult{ID: ex.ID, Title: ex.Title, OK: false, Error: "保存失败: " + err.Error()})
			continue
		}
	}
	return results, nil
}
