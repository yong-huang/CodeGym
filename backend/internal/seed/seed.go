package seed

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codegym/backend/internal/model"
)

// seedPoint declares a fill point by searching for its text in full_code,
// so span offsets are always correct even after the code text is edited.
type seedPoint struct {
	id                string
	text              string
	category          string
	minWeight         string
	weightScore       float64
	hint              string
	fillType          string
	options           []string
	acceptableAnswers []string
	alwaysFill        bool
	occurrence        int   // 0-based occurrence of `text` to use
	orderGroup        string // 顺序可互换分组（同组判题时双射匹配）
}

type seedExercise struct {
	id           string
	title        string
	description  string
	exerciseType string
	techStack    string
	weight       string
	category     string
	subcategory  string
	fullCode     string
	buggyCode    string
	bugDesc      string
	bugLoc       string
	expected     string
	validation   string
	testCases    []model.TestCase
	points       []seedPoint
	hints        model.SpotterHints
	explanation  string
}

// bankPointJSON is the on-disk JSON schema of one fill point. Locations are
// NOT stored: spans are recomputed from `text` in the code on every load, so
// the JSON stays hand-editable.
type bankPointJSON struct {
	ID                string   `json:"id"`
	Text              string   `json:"text"`
	Category          string   `json:"category"`
	MinWeight         string   `json:"min_weight"`
	WeightScore       float64  `json:"weight_score"`
	Hint              string   `json:"hint"`
	FillType          string   `json:"fill_type"`
	Options           []string `json:"options,omitempty"`
	AcceptableAnswers []string `json:"acceptable_answers,omitempty"`
	AlwaysFill        bool     `json:"always_fill,omitempty"`
	Occurrence        int      `json:"occurrence,omitempty"`
	OrderGroup        string   `json:"order_group,omitempty"`
}

// bankSeedJSON is the on-disk JSON schema of one question.
type bankSeedJSON struct {
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	ExerciseType string           `json:"exercise_type"`
	TechStack    string           `json:"tech_stack"`
	Weight       string           `json:"weight"`
	Category     string           `json:"category"`
	Subcategory  string           `json:"subcategory,omitempty"`
	FullCode     string           `json:"full_code"`
	BuggyCode    string           `json:"buggy_code,omitempty"`
	BugDesc      string           `json:"bug_description,omitempty"`
	BugLoc       string           `json:"bug_location,omitempty"`
	Expected     string           `json:"expected"`
	Validation   string           `json:"validation_code,omitempty"`
	TestCases    []model.TestCase `json:"test_cases,omitempty"`
	Points       []bankPointJSON  `json:"points"`
	Hints        model.SpotterHints `json:"hints"`
	Explanation  string           `json:"explanation,omitempty"`
}

func (j bankPointJSON) toSeed() seedPoint {
	return seedPoint{
		id: j.ID, text: j.Text, category: j.Category, minWeight: j.MinWeight,
		weightScore: j.WeightScore, hint: j.Hint, fillType: j.FillType,
		options: j.Options, acceptableAnswers: j.AcceptableAnswers,
		alwaysFill: j.AlwaysFill, occurrence: j.Occurrence, orderGroup: j.OrderGroup,
	}
}

func (p seedPoint) toBankJSON() bankPointJSON {
	return bankPointJSON{
		ID: p.id, Text: p.text, Category: p.category, MinWeight: p.minWeight,
		WeightScore: p.weightScore, Hint: p.hint, FillType: p.fillType,
		Options: p.options, AcceptableAnswers: p.acceptableAnswers,
		AlwaysFill: p.alwaysFill, Occurrence: p.occurrence, OrderGroup: p.orderGroup,
	}
}

func (j bankSeedJSON) toSeed() seedExercise {
	points := make([]seedPoint, 0, len(j.Points))
	for _, p := range j.Points {
		points = append(points, p.toSeed())
	}
	return seedExercise{
		id: j.ID, title: j.Title, description: j.Description,
		exerciseType: j.ExerciseType, techStack: j.TechStack,
		weight: j.Weight, category: j.Category, subcategory: j.Subcategory,
		fullCode: j.FullCode, buggyCode: j.BuggyCode,
		bugDesc: j.BugDesc, bugLoc: j.BugLoc,
		expected: j.Expected, validation: j.Validation,
		testCases: j.TestCases, points: points,
		hints: j.Hints, explanation: j.Explanation,
	}
}

func (s seedExercise) toBankJSON() bankSeedJSON {
	points := make([]bankPointJSON, 0, len(s.points))
	for _, p := range s.points {
		points = append(points, p.toBankJSON())
	}
	return bankSeedJSON{
		ID: s.id, Title: s.title, Description: s.description,
		ExerciseType: s.exerciseType, TechStack: s.techStack,
		Weight: s.weight, Category: s.category, Subcategory: s.subcategory,
		FullCode: s.fullCode, BuggyCode: s.buggyCode,
		BugDesc: s.bugDesc, BugLoc: s.bugLoc,
		Expected: s.expected, Validation: s.validation,
		TestCases: s.testCases, Points: points,
		Hints: s.hints, Explanation: s.explanation,
	}
}

func buildExercise(s seedExercise) (model.Exercise, error) {
	ex := model.Exercise{
		ID:             s.id,
		Title:          s.title,
		Description:    s.description,
		ExerciseType:   s.exerciseType,
		TechStack:      s.techStack,
		Weight:         s.weight,
		Category:       s.category,
		Subcategory:    s.subcategory,
		FullCode:       s.fullCode,
		BuggyCode:      s.buggyCode,
		BugDescription: s.bugDesc,
		BugLocation:    s.bugLoc,
		ExpectedOutput: s.expected,
		ValidationCode: s.validation,
		TestCases:      s.testCases,
		SpotterHints:   s.hints,
		TimeLimitSecs:  30,
		Explanation:    s.explanation,
		Enabled:        true,
		CreatedBy:      "seed",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		FillPoints: model.FillPointsConfig{
			MinFills: 1,
			MaxFills: len(s.points),
		},
	}

	code := s.fullCode
	if s.exerciseType == model.ExerciseTypeBugfix {
		code = s.buggyCode
	}
	for _, p := range s.points {
		start, ok := occurrenceIndex(code, p.text, p.occurrence)
		if !ok {
			return ex, fmt.Errorf("seed %s: text %q occurrence %d not found in code", s.id, p.text, p.occurrence)
		}
		line := strings.Count(code[:start], "\n") + 1
		ex.FillPoints.Points = append(ex.FillPoints.Points, model.FillPoint{
			ID:                p.id,
			Location:          model.FillLocation{Type: "span", Line: line, SpanStart: start, SpanEnd: start + len(p.text)},
			OriginalText:      p.text,
			AcceptableAnswers: p.acceptableAnswers,
			Category:          p.category,
			MinWeight:         p.minWeight,
			WeightScore:       p.weightScore,
			AlwaysFill:        p.alwaysFill,
			Hint:              p.hint,
			FillType:          p.fillType,
			Options:           p.options,
			OrderGroup:        p.orderGroup,
		})
	}
	// fill points are fully baked into the bank JSON (including auto ones),
	// so no extra enrichment happens at build time.
	ensureBeginnerPoint(ex.FillPoints.Points)
	return ex, nil
}

// ensureBeginnerPoint guarantees every exercise has at least one training
// point fillable at the beginner level. Otherwise low weight levels would have
// no blank to fill (a broken/empty exercise).
func ensureBeginnerPoint(points []model.FillPoint) {
	ranks := map[string]int{"beginner": 1, "easy": 2, "medium": 3, "hard": 4, "expert": 5}
	hasFillableAtBeginner := false
	lowest := 99
	lowIdx := -1
	for i, p := range points {
		if p.MinWeight == "" || ranks[p.MinWeight] <= 1 {
			hasFillableAtBeginner = true
		}
		if r := ranks[p.MinWeight]; r > 0 && r < lowest {
			lowest, lowIdx = r, i
		}
	}
	if !hasFillableAtBeginner && lowIdx >= 0 {
		points[lowIdx].MinWeight = "beginner"
	}
}

// occurrenceIndex returns the start of the nth occurrence of sub in s.
func occurrenceIndex(s, sub string, n int) (int, bool) {
	idx := -1
	for i := 0; i <= n; i++ {
		next := strings.Index(s[idx+1:], sub)
		if next < 0 {
			return 0, false
		}
		idx += next + 1
	}
	return idx, true
}

// CategoryPair is a (tech_stack, subcategory) default managed category.
type CategoryPair struct {
	Tech  string
	Sub   string
}

// CategoryPlan returns the default set of managed subcategories. Empty ones
// (高级/实战/web) are just manageable containers until questions are added.
func CategoryPlan() []CategoryPair {
	return []CategoryPair{
		{Tech: "python", Sub: "基础"},
		{Tech: "python", Sub: "协程"},
		{Tech: "python", Sub: "多线程"},
		{Tech: "python", Sub: "装饰器"},
		{Tech: "python", Sub: "迭代器"},
		{Tech: "python", Sub: "描述符"},
		{Tech: "python", Sub: "上下文管理器"},
		{Tech: "python", Sub: "元类与类机制"},
		{Tech: "python", Sub: "魔术方法"},
		{Tech: "python", Sub: "Agent 开发"},
		{Tech: "datastructure", Sub: "数组"},
		{Tech: "datastructure", Sub: "链表"},
		{Tech: "datastructure", Sub: "栈与队列"},
		{Tech: "datastructure", Sub: "哈希表"},
		{Tech: "datastructure", Sub: "树"},
		{Tech: "datastructure", Sub: "图"},
		{Tech: "python", Sub: "FastAPI"},
		{Tech: "python", Sub: "高级"},
		{Tech: "python", Sub: "实战"},
		{Tech: "python", Sub: "web"},
		{Tech: "javascript", Sub: "基础"},
		{Tech: "go", Sub: "基础"},
		{Tech: "go", Sub: "类型系统"},
		{Tech: "go", Sub: "通信机制"},
		{Tech: "go", Sub: "同步与原子"},
		{Tech: "go", Sub: "并发模式"},
		{Tech: "go", Sub: "高级"},
		{Tech: "go", Sub: "实战"},
		{Tech: "go", Sub: "web"},
		{Tech: "kubernetes", Sub: "基础"},
		{Tech: "kubernetes", Sub: "资源配置"},
		{Tech: "kubernetes", Sub: "调度"},
		{Tech: "kubernetes", Sub: "高级"},
		{Tech: "kubernetes", Sub: "实战"},
		{Tech: "kubernetes", Sub: "web"},
		{Tech: "cka", Sub: "集群架构"},
		{Tech: "cka", Sub: "工作负载"},
		{Tech: "cka", Sub: "服务网络"},
		{Tech: "cka", Sub: "存储"},
		{Tech: "cka", Sub: "故障排查"},
		{Tech: "cka", Sub: "高级"},
		{Tech: "cka", Sub: "实战"},
		{Tech: "cka", Sub: "web"},
		{Tech: "openshift", Sub: "基础"},
		{Tech: "openshift", Sub: "对比"},
		{Tech: "openshift", Sub: "高级"},
		{Tech: "openshift", Sub: "实战"},
		{Tech: "openshift", Sub: "web"},
		{Tech: "sql", Sub: "基础"},
		{Tech: "sql", Sub: "高级"},
		{Tech: "sql", Sub: "实战"},
		{Tech: "sql", Sub: "web"},
		{Tech: "bigquery", Sub: "SQL 基础"},
		{Tech: "bigquery", Sub: "BigQuery 特有"},
		{Tech: "bigquery", Sub: "高级"},
		{Tech: "bigquery", Sub: "实战"},
		{Tech: "bigquery", Sub: "web"},
		{Tech: "devops", Sub: "terraform"},
		{Tech: "devops", Sub: "ansible"},
		{Tech: "devops", Sub: "docker"},
		{Tech: "devops", Sub: "prometheus"},
		{Tech: "devops", Sub: "高级"},
		{Tech: "devops", Sub: "实战"},
		{Tech: "devops", Sub: "web"},
		{Tech: "aws", Sub: "ec2"},
		{Tech: "aws", Sub: "iam"},
		{Tech: "aws", Sub: "eks"},
		{Tech: "aws", Sub: "s3"},
		{Tech: "aws", Sub: "vpc"},
		{Tech: "aws", Sub: "rds"},
		{Tech: "aws", Sub: "高级"},
		{Tech: "aws", Sub: "实战"},
		{Tech: "aws", Sub: "web"},
		{Tech: "openstack", Sub: "Nova"},
		{Tech: "openstack", Sub: "Neutron"},
		{Tech: "openstack", Sub: "Cinder"},
		{Tech: "openstack", Sub: "Swift"},
		{Tech: "openstack", Sub: "TripleO"},
		{Tech: "openstack", Sub: "Keystone"},
		{Tech: "openstack", Sub: "Glance"},
		{Tech: "openstack", Sub: "运维"},
		{Tech: "openstack", Sub: "升级"},
		{Tech: "openstack", Sub: "Ceph"},
		{Tech: "openstack", Sub: "高级"},
		{Tech: "openstack", Sub: "实战"},
		{Tech: "openstack", Sub: "web"},
		{Tech: "testing", Sub: "pytest"},
		{Tech: "testing", Sub: "unittest"},
		{Tech: "testing", Sub: "selenium"},
		{Tech: "testing", Sub: "playwright"},
		{Tech: "testing", Sub: "cypress"},
		{Tech: "testing", Sub: "locust"},
		{Tech: "testing", Sub: "高级"},
		{Tech: "testing", Sub: "实战"},
		{Tech: "testing", Sub: "web"},
		{Tech: "fastapi", Sub: "基础"},
		{Tech: "fastapi", Sub: "高级"},
		{Tech: "fastapi", Sub: "实战"},
		{Tech: "fastapi", Sub: "web"},
	}
}

// resolveBankDir locates the question-bank JSON directory: an explicit path,
// else $CODEGYM_BANK_DIR, else the nearest ancestor of cwd containing data/bank.
func resolveBankDir(dir string) string {
	if dir != "" {
		return dir
	}
	if env := os.Getenv("CODEGYM_BANK_DIR"); env != "" {
		return env
	}
	if cwd, err := os.Getwd(); err == nil {
		for p := cwd; ; p = filepath.Dir(p) {
			cand := filepath.Join(p, "data", "bank")
			if fi, err := os.Stat(cand); err == nil && fi.IsDir() {
				return cand
			}
			if filepath.Dir(p) == p {
				break
			}
		}
	}
	return filepath.Join("data", "bank")
}

// ResolveBankDir exposes bank directory resolution: explicit path >
// $CODEGYM_BANK_DIR > nearest ancestor of cwd containing data/bank.
func ResolveBankDir(dir string) string { return resolveBankDir(dir) }

// LoadBank reads one question-bank JSON file (raw seed sources) from dir.
func LoadBank(dir, file string) ([]seedExercise, error) {
	if dir == "" {
		dir = resolveBankDir("")
	}
	data, err := os.ReadFile(filepath.Join(dir, file+".json"))
	if err != nil {
		return nil, err
	}
	return parseBank(file+".json", data)
}

// parseBank parses one bank JSON file into raw seed sources.
func parseBank(name string, data []byte) ([]seedExercise, error) {
	var list []bankSeedJSON
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	out := make([]seedExercise, 0, len(list))
	for _, j := range list {
		out = append(out, j.toSeed())
	}
	return out, nil
}

// buildBank loads one bank file and builds every exercise in it.
func buildBank(dir, file string) ([]model.Exercise, error) {
	list, err := LoadBank(dir, file)
	if err != nil {
		return nil, err
	}
	out := make([]model.Exercise, 0, len(list))
	for _, s := range list {
		ex, err := buildExercise(s)
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}

// BuildAll loads every bank JSON file in dir and builds the exercises.
func BuildAll(dir string) ([]model.Exercise, error) {
	if dir == "" {
		dir = resolveBankDir("")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []model.Exercise
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		list, err := LoadBank(dir, strings.TrimSuffix(e.Name(), ".json"))
		if err != nil {
			return nil, err
		}
		for _, s := range list {
			ex, err := buildExercise(s)
			if err != nil {
				return nil, err
			}
			out = append(out, ex)
		}
	}
	return out, nil
}

// All returns every seeded exercise, ready for Upsert. It loads all question
// banks from the data/bank JSON directory (see resolveBankDir).
func All() ([]model.Exercise, error) {
	return BuildAll("")
}

// BuildBank builds every exercise of one bank JSON file in dir.
func BuildBank(dir, file string) ([]model.Exercise, error) {
	return buildBank(dir, file)
}

// BuildBankJSON builds exercises from raw bank JSON bytes (dynamically loaded
// banks: admin upload or directory scan).
func BuildBankJSON(name string, data []byte) ([]model.Exercise, error) {
	list, err := parseBank(name, data)
	if err != nil {
		return nil, err
	}
	out := make([]model.Exercise, 0, len(list))
	for _, s := range list {
		ex, err := buildExercise(s)
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}
