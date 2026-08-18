package judge

import (
	"math"
	"strings"

	"codegym/backend/internal/model"
)

// hintPenalty mirrors design section 7.2: hint level -> score penalty.
var hintPenalty = map[int]float64{0: 0, 1: 0, 2: -5, 3: -15, 4: -30}

// normalize strips all whitespace so code-equivalent answers (e.g. `x%2==0`
// vs `x % 2 == 0`) compare equal.
// OutputsMatch compares the full program stdout against the expected output,
// ignoring whitespace so formatting differences don't fail a correct answer.
func OutputsMatch(stdout, expected string) bool {
	if expected == "" {
		return false
	}
	return normalize(stdout) == normalize(expected)
}

func normalize(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, s)
}

// IsCorrect compares a user answer with the expected text (精确匹配).
func IsCorrect(user, expected string) bool {
	return normalize(user) == normalize(expected)
}

// IsAcceptedAnswer checks the user answer against the expected text and any
// explicitly acceptable alternatives (e.g. `-A` vs `--all-namespaces`).
func IsAcceptedAnswer(user string, p model.FillPoint) bool {
	if IsCorrect(user, p.OriginalText) {
		return true
	}
	for _, alt := range p.AcceptableAnswers {
		if IsCorrect(user, alt) {
			return true
		}
	}
	return false
}

// IsPartiallyCorrect is a lightweight semantic fallback (大小写/空白无关或包含关系).
func IsPartiallyCorrect(user, expected string) bool {
	if user == "" {
		return false
	}
	u := strings.ToLower(normalize(user))
	e := strings.ToLower(normalize(expected))
	return strings.Contains(e, u) || strings.Contains(u, e) || strings.Contains(u, strings.Trim(e, "()[]{}"))
}

// GradeForScore maps a 0-100 score to S/A/B/C/D (design section 7.3).
func GradeForScore(score float64) string {
	switch {
	case score >= 90:
		return "S"
	case score >= 75:
		return "A"
	case score >= 60:
		return "B"
	case score >= 40:
		return "C"
	default:
		return "D"
	}
}

// GradeLabel returns the label text used by the frontend grade badge.
func GradeLabel(grade string) string {
	return map[string]string{
		"S": "🎉 完美！完全掌握",
		"A": "✅ 优秀，继续加油",
		"B": "📖 良好，建议复习薄弱点",
		"C": "📝 需要加强训练",
		"D": "🔄 建议重新学习基础知识",
	}[grade]
}

// GradeFill judges the whole exercise after execution per design section 7.
// hintUsage maps fill point id -> max hint level used.
func GradeFill(ex model.Exercise, points []model.FillPoint, answers map[string]string, output, expected string, hintUsage map[string]int) ([]model.FillResult, float64) {
	var results []model.FillResult
	totalWeight, weightedSum := 0.0, 0.0

	for _, p := range points {
		user := strings.TrimSpace(answers["FILL_"+p.ID])
		exp := p.OriginalText
		fr := model.FillResult{
			ID:         "FILL_" + p.ID,
			PointID:    p.ID,
			UserAnswer: user,
			Expected:   exp,
		}
		base := 0.0
		switch {
		case IsAcceptedAnswer(user, p):
			base = 100
			fr.IsCorrect = true
		case IsPartiallyCorrect(user, exp):
			base = 50
		}

		// hint usage deducts from this point (设计 6.4/7.2)
		score := base + hintPenalty[hintUsage[p.ID]]
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		fr.Score = score
		results = append(results, fr)

		w := p.WeightScore
		if w <= 0 {
			w = 1
		}
		totalWeight += w
		weightedSum += score * w
	}

	total := 0.0
	if totalWeight > 0 {
		total = math.Round(weightedSum/totalWeight*100) / 100
	}
	return results, total
}

// GradeBugfix judges a bugfix exercise per design section 7.4.
// `output` is the full multi-line stdout produced by running the user's code
// with the exercise's validation harness.
func GradeBugfix(ex model.Exercise, userCode string, output string, hintUsage map[string]int) (float64, int, []model.FillResult) {
	var results []model.FillResult

	lines := nonEmptyLines(output)
	// 1. test pass rate -> 70 base points
	passed := 0
	for _, tc := range ex.TestCases {
		for _, line := range lines {
			if normalize(tc.Output) == normalize(line) {
				passed++
				break
			}
		}
	}
	total := len(ex.TestCases)
	testScore := 0.0
	if total > 0 {
		testScore = float64(passed) / float64(total) * 70
	}
	results = append(results, model.FillResult{
		ID: "FILL_bug", PointID: "bug", UserAnswer: userCode,
		Expected: ex.ExpectedOutput, IsCorrect: passed == total, Score: testScore,
	})

	// 2. bug location accuracy -> +20
	locationScore := 0.0
	if normalize(userCode) == normalize(ex.FullCode) || strings.Contains(normalize(userCode), normalize(ex.FullCode)) {
		locationScore = 20
	} else if ex.BugLocation != "" && !strings.Contains(userCode, ex.BugLocation) {
		locationScore = 20
	}
	if locationScore == 20 {
		// report the fixed line that should replace the buggy one, if derivable
		fixed := fixSnippet(ex)
		fixExp := "已定位到 bug 并修复"
		if fixed != "" {
			fixExp = fixExp + "：" + fixed
		}
		results = append(results, model.FillResult{
			ID: "FILL_location", PointID: "bug_location",
			UserAnswer: userCode, Expected: fixExp,
			IsCorrect: true, Score: locationScore,
		})
	} else {
		results = append(results, model.FillResult{
			ID: "FILL_location", PointID: "bug_location",
			UserAnswer: userCode, Expected: "未定位/未修复出错的代码行",
			IsCorrect: false, Score: 0,
		})
	}

	// 3. code quality -> +10 when all tests pass
	qualityScore := 0.0
	if testScore == 70 && total > 0 {
		qualityScore = 10
	}
	results = append(results, model.FillResult{
		ID: "FILL_quality", PointID: "quality",
		UserAnswer: userCode, Expected: "通过全部测试后给质量分",
		IsCorrect: qualityScore == 10, Score: qualityScore,
	})

	// 4. hint penalty (values are negative, so adding subtracts)
	penalty := hintPenalty[maxHintLevel(hintUsage)]

	score := testScore + locationScore + qualityScore + penalty
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return math.Round(score*100) / 100, passed, results
}

// fixSnippet tries to surface the corrected line from the exercise explanation
// (our explanations mention the fix, e.g. "加 (y%100 != 0 || y%400 == 0)").
func fixSnippet(ex model.Exercise) string {
	// pick a concise line reference from the explanation between backticks if any
	// fall back to empty (frontend shows the whole answer code separately)
	if ex.Explanation == "" {
		return ""
	}
	// avoid dumping the whole explanation; just a trimmed first clause
	cut := ex.Explanation
	if i := strings.Index(cut, "。"); i > 0 && i < 60 {
		cut = cut[:i]
	}
	if len(cut) > 48 {
		cut = cut[:48] + "…"
	}
	return cut
}

// WeakAnalysis builds a short review recommendation for the result screen.
func WeakAnalysis(results []model.FillResult, hintUsage map[string]int) string {
	var parts []string
	for _, r := range results {
		if !r.IsCorrect {
			parts = append(parts, "「"+r.PointID+"」填成 `"+truncate(r.UserAnswer, 40)+"`，与预期不符")
		}
	}
	if maxHintLevel(hintUsage) >= 3 {
		parts = append(parts, "本组使用了较高级辅助，建议回到该知识点的入门负重重新练习")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "；")
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

func truncate(s string, n int) string {
	r := []rune(strings.Join(strings.Fields(s), " "))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "…"
}

func maxHintLevel(usage map[string]int) int {
	m := 0
	for _, lvl := range usage {
		if lvl > m {
			m = lvl
		}
	}
	return m
}
