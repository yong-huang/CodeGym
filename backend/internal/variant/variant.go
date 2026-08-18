package variant

import (
	"crypto/rand"
	"encoding/hex"
	mathrand "math/rand/v2"
	"sort"

	"codegym/backend/internal/model"
)

// weightRank maps a weight level name to an integer for eligibility filtering.
var weightRank = map[string]int{
	"beginner": 1,
	"easy":     2,
	"medium":   3,
	"hard":     4,
	"expert":   5,
}

// NewVariantID returns a short random id used to correlate an attempt with its variant.
func NewVariantID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// SelectPoints picks the fill points for a given weight level following design
// section 5.4:
//  1. read weight config -> min/max fills for the level
//  2. filter eligible points (level rank >= point min_weight rank, or always_fill)
//  3. randomly choose a count in [min, max] and pick that many points,
//     skipping spans that overlap an already selected point.
func SelectPoints(fillPoints model.FillPointsConfig, level string) []model.FillPoint {
	cfg, ok := model.WeightLevels[level]
	if !ok {
		cfg = model.WeightLevels[model.DefaultWeightLevel]
	}

	eligible := EligiblePoints(fillPoints, level)
	if len(eligible) == 0 {
		return nil
	}

	minFills := fillPoints.MinFills
	if cfg.MinFills > 0 {
		minFills = cfg.MinFills
	}
	maxFills := fillPoints.MaxFills
	if cfg.MaxFills > 0 {
		maxFills = cfg.MaxFills
	}
	minFills = clamp(minFills, 1, len(eligible))
	maxFills = clamp(maxFills, minFills, len(eligible))

	count := minFills
	if maxFills > minFills {
		count = mathrand.IntN(maxFills-minFills+1) + minFills
	}

	// shuffle and take the first `count`, skipping overlapping spans
	mathrand.Shuffle(len(eligible), func(i, j int) { eligible[i], eligible[j] = eligible[j], eligible[i] })
	var out []model.FillPoint
	for _, p := range eligible {
		if len(out) >= count {
			break
		}
		if overlapsSelected(out, p) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// EligiblePoints returns the fill points allowed at the given weight level.
func EligiblePoints(fillPoints model.FillPointsConfig, level string) []model.FillPoint {
	if _, ok := model.WeightLevels[level]; !ok {
		level = model.DefaultWeightLevel
	}
	rank := weightRank[level]
	eligible := make([]model.FillPoint, 0, len(fillPoints.Points))
	for _, p := range fillPoints.Points {
		if p.AlwaysFill {
			eligible = append(eligible, p)
			continue
		}
		pRank, ok := weightRank[p.MinWeight]
		if !ok {
			pRank = rank // unknown min_weight -> assume eligible at any level
		}
		if pRank <= rank {
			eligible = append(eligible, p)
		}
	}
	return eligible
}

// FilterForAnswers returns the subset of eligible points the user actually
// answered, so grading is consistent with the variant they saw.
func FilterForAnswers(fillPoints model.FillPointsConfig, level string, answers map[string]string) []model.FillPoint {
	eligible := EligiblePoints(fillPoints, level)
	var out []model.FillPoint
	for _, p := range eligible {
		if _, ok := answers["FILL_"+p.ID]; ok {
			out = append(out, p)
		}
	}
	return out
}

// overlapsSelected reports whether p's span overlaps any already selected point.
func overlapsSelected(selected []model.FillPoint, p model.FillPoint) bool {
	for _, s := range selected {
		if p.Location.SpanStart < s.Location.SpanEnd && s.Location.SpanStart < p.Location.SpanEnd {
			return true
		}
	}
	return false
}

// ApplyAnswers merges user answers into a filled copy of the code.
// Each selected fill point's span is replaced by the user's answer
// (or its original text when missing/blank).
func ApplyAnswers(code string, points []model.FillPoint, answers map[string]string) string {
	// Apply replacements right-to-left by span so earlier offsets stay valid.
	ordered := make([]model.FillPoint, len(points))
	copy(ordered, points)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Location.SpanStart < ordered[j].Location.SpanStart })

	type repl struct{ start, end int; text string }
	repls := make([]repl, 0, len(ordered))
	for _, p := range ordered {
		txt, ok := answers["FILL_"+p.ID]
		if !ok || txt == "" {
			txt = p.OriginalText
		}
		repls = append(repls, repl{start: p.Location.SpanStart, end: p.Location.SpanEnd, text: txt})
	}
	for i := len(repls) - 1; i >= 0; i-- {
		r := repls[i]
		if r.start < 0 || r.end > len(code) || r.start > r.end {
			continue
		}
		code = code[:r.start] + r.text + code[r.end:]
	}
	return code
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
