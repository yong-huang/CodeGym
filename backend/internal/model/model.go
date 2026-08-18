package model

import "time"

// WeightLevel represents a fill density level per design section 5.
type WeightLevel struct {
	Label       string   `json:"label"`
	Description string   `json:"description"`
	MinFills    int      `json:"min_fills"`
	MaxFills    int      `json:"max_fills"`
	FillTypes   []string `json:"fill_types"`
}

// ExerciseType per design section 4.
const (
	ExerciseTypeFill     = "fill"
	ExerciseTypeBugfix   = "bugfix"
	ExerciseTypeCommand  = "command"
	ExerciseTypeConfig   = "config"
)

// Weight levels per design section 5.2.
var WeightLevels = map[string]WeightLevel{
	"beginner": {
		Label:       "🌱 入门",
		Description: "适合第一次接触这个知识点",
		MinFills:    1,
		MaxFills:    2,
		FillTypes:   []string{"text", "select"},
	},
	"easy": {
		Label:       "🌿 简单",
		Description: "已经理解了基本概念",
		MinFills:    2,
		MaxFills:    3,
		FillTypes:   []string{"text", "select"},
	},
	"medium": {
		Label:       "🌳 中等",
		Description: "需要掌握完整实现",
		MinFills:    3,
		MaxFills:    4,
		FillTypes:   []string{"text", "code"},
	},
	"hard": {
		Label:       "🔥 困难",
		Description: "挑战深度理解",
		MinFills:    4,
		MaxFills:    6,
		FillTypes:   []string{"text", "code", "code_block"},
	},
	"expert": {
		Label:       "💎 专家",
		Description: "接近独立编写水平",
		MinFills:    6,
		MaxFills:    10,
		FillTypes:   []string{"code_block"},
	},
}

// DefaultWeightLevel is used when no level is specified.
const DefaultWeightLevel = "medium"

// FillLocation locates a fill point inside full_code (design section 11.2).
type FillLocation struct {
	Type         string `json:"type"`
	Line         int    `json:"line"`
	StartColumn  int    `json:"start_column,omitempty"`
	EndColumn    int    `json:"end_column,omitempty"`
	SpanStart    int    `json:"span_start"`
	SpanEnd      int    `json:"span_end"`
	IsBlankBlock bool   `json:"is_blank_block,omitempty"`
}

// FillPoint describes a single trainee action (设计: 训练点).
type FillPoint struct {
	ID                string       `json:"id"`
	Location          FillLocation `json:"location"`
	OriginalText      string       `json:"original_text"`
	AcceptableAnswers []string     `json:"acceptable_answers,omitempty"`
	Category          string       `json:"category"` // expression / condition / statement / keyword / config / command
	MinWeight         string       `json:"min_weight"`
	WeightScore       float64      `json:"weight_score"`
	AlwaysFill        bool         `json:"always_fill"`
	Hint              string       `json:"hint"`
	FillType          string       `json:"fill_type"` // text / select / code / code_block
	Options           []string     `json:"options,omitempty"`
}

// FillPointsConfig holds the points list plus density bounds.
type FillPointsConfig struct {
	Points   []FillPoint `json:"points"`
	MinFills int         `json:"min_fills"`
	MaxFills int         `json:"max_fills"`
}

// TestCase is a single input/output pair for bugfix grading.
type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// SpotterHints is the 4-level progressive hint system (设计 section 6).
type SpotterHints struct {
	Level1 string `json:"level_1"`
	Level2 string `json:"level_2"`
	Level3 string `json:"level_3"`
	Level4 string `json:"level_4"`
}

// Exercise mirrors the exercises table (design section 11.1).
type Exercise struct {
	ID             string           `json:"id"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	ExerciseType   string           `json:"exercise_type"`
	TechStack      string           `json:"tech_stack"`
	Weight         string           `json:"weight"`
	Category       string           `json:"category"`
	Subcategory    string           `json:"subcategory"` // 目录树第二层：基础/高级/实战/web…
	FullCode       string           `json:"full_code"`
	BuggyCode      string           `json:"buggy_code,omitempty"`
	BugDescription string           `json:"bug_description,omitempty"`
	BugLocation    string           `json:"bug_location,omitempty"`
	FillPoints     FillPointsConfig `json:"fill_points"`
	ExpectedOutput string           `json:"expected_output"`
	ValidationCode string           `json:"validation_code,omitempty"`
	TestCases      []TestCase       `json:"test_cases,omitempty"`
	SpotterHints   SpotterHints     `json:"spotter_hints"`
	TimeLimitSecs  int              `json:"time_limit_seconds"`
	Explanation    string           `json:"explanation,omitempty"`
	Enabled        bool             `json:"enabled"`
	CreatedBy      string           `json:"created_by,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// FavoriteItem is a favorited exercise summary (设计: 收藏).
type FavoriteItem struct {
	ExerciseID   string    `json:"exercise_id"`
	Title        string    `json:"title"`
	TechStack    string    `json:"tech_stack"`
	ExerciseType string    `json:"exercise_type"`
	FavoritedAt  time.Time `json:"favorited_at"`
}

// Category is a managed subcategory under a tech stack (目录树第二层).
type Category struct {
	TechStack     string `json:"tech_stack"`
	Subcategory   string `json:"subcategory"`
	ExerciseCount int    `json:"exercise_count"`
}

// Environment mirrors the environments table (设计 §8.3 / §11.1).
type Environment struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Type          string         `json:"type"` // local / docker / kubernetes / pyodide
	Config        map[string]any `json:"config"`
	HealthStatus  string         `json:"health_status"` // unknown / healthy / unhealthy
	MaxConcurrent int            `json:"max_concurrent"`
	Enabled       bool           `json:"enabled"`
	CreatedAt     time.Time      `json:"created_at"`
}

// AttemptRecord mirrors the attempt_records table.
type AttemptRecord struct {
	ID              string            `json:"id"`
	UserID          string            `json:"user_id"`
	ExerciseID      string            `json:"exercise_id"`
	VariantID       string            `json:"variant_id"`
	WeightLevel     string            `json:"weight_level"`
	FilledPoints    []FillPoint       `json:"filled_points"`
	UserAnswers     map[string]string `json:"user_answers"`
	ExecutionOutput string            `json:"execution_output"`
	IsCorrect       bool              `json:"is_correct"`
	Score           float64           `json:"score"`
	Grade           string            `json:"grade"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	ExecutionTimeMS int64             `json:"execution_time_ms"`
	HintUsage       map[string]int    `json:"hint_usage"`
	CreatedAt       time.Time         `json:"created_at"`
}

// UserProgress mirrors the user_progress table.
type UserProgress struct {
	UserID             string            `json:"user_id"`
	TotalAttempts      int               `json:"total_attempts"`
	CorrectAttempts    int               `json:"correct_attempts"`
	CompletedExercises map[string]bool   `json:"completed_exercises"`
	WeakPoints         map[string]int    `json:"weak_points"`
	TechStackScores    map[string]float64 `json:"tech_stack_scores"`
	TotalHintUsed      int               `json:"total_hint_used"`
	CurrentStreak      int               `json:"current_streak"`
	LongestStreak      int               `json:"longest_streak"`
	LastAttemptAt      *time.Time        `json:"last_attempt_at,omitempty"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// FillResult is one graded fill point (design section 12.3).
type FillResult struct {
	ID         string  `json:"id"`
	PointID    string  `json:"point_id"`
	UserAnswer string  `json:"user_answer"`
	Expected   string  `json:"expected"`
	IsCorrect  bool    `json:"is_correct"`
	Score      float64 `json:"score"`
}

// ExecuteResult is the response of POST /api/v1/execute.
type ExecuteResult struct {
	ExecutionID     string       `json:"execution_id"`
	Success         bool         `json:"success"`
	Output          string       `json:"output"`
	ExpectedOutput  string       `json:"expected_output"`
	IsCorrect       bool         `json:"is_correct"`
	Score           float64      `json:"score"`
	Grade           string       `json:"grade"`
	FillResults     []FillResult `json:"fill_results"`
	ExecutionTimeMS int64        `json:"execution_time_ms"`
	Logs            []string     `json:"logs"`
	WeakAnalysis    string       `json:"weak_analysis,omitempty"`
	ErrorMessage    string       `json:"error_message,omitempty"`
	CorrectAnswer   string       `json:"correct_answer,omitempty"` // 修复后的代码 / 正确答案
	Explanation     string       `json:"explanation,omitempty"`    // 解析
}
