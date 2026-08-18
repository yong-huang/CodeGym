package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"codegym/backend/internal/model"

	_ "modernc.org/sqlite"
)

// Store persists exercises, attempts and user progress in SQLite.
type Store struct {
	db *sql.DB
}

// Open opens (and initializes) a SQLite database at the given path.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1) // sqlite single-writer
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS exercises (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			exercise_type TEXT DEFAULT 'fill',
			tech_stack TEXT NOT NULL,
			weight TEXT NOT NULL,
			category TEXT,
			subcategory TEXT DEFAULT '基础',
			full_code TEXT NOT NULL,
			buggy_code TEXT,
			bug_description TEXT,
			bug_location TEXT,
			fill_points TEXT NOT NULL,
			expected_output TEXT,
			validation_code TEXT,
			test_cases TEXT,
			spotter_hints TEXT,
			time_limit_seconds INTEGER DEFAULT 30,
			explanation TEXT,
			enabled INTEGER DEFAULT 1,
			created_by TEXT,
			created_at TEXT,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS attempt_records (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			exercise_id TEXT NOT NULL,
			variant_id TEXT,
			weight_level TEXT,
			filled_points TEXT,
			user_answers TEXT,
			execution_output TEXT,
			is_correct INTEGER,
			score REAL,
			grade TEXT,
			error_message TEXT,
			execution_time_ms INTEGER,
			hint_usage TEXT,
			created_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS user_progress (
			user_id TEXT PRIMARY KEY,
			total_attempts INTEGER DEFAULT 0,
			correct_attempts INTEGER DEFAULT 0,
			completed_exercises TEXT,
			weak_points TEXT,
			tech_stack_scores TEXT,
			total_hint_used INTEGER DEFAULT 0,
			current_streak INTEGER DEFAULT 0,
			longest_streak INTEGER DEFAULT 0,
			last_attempt_at TEXT,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS favorites (
			user_id TEXT NOT NULL,
			exercise_id TEXT NOT NULL,
			created_at TEXT,
			PRIMARY KEY (user_id, exercise_id)
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			tech_stack TEXT NOT NULL,
			subcategory TEXT NOT NULL,
			created_at TEXT,
			PRIMARY KEY (tech_stack, subcategory)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_attempts_exercise ON attempt_records(exercise_id)`,
		`CREATE INDEX IF NOT EXISTS idx_attempts_user ON attempt_records(user_id)`,
	}
	for _, st := range stmts {
		if _, err := s.db.Exec(st); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	// add subcategory column to pre-existing databases
	if err := s.ensureExerciseColumn("subcategory", "TEXT DEFAULT '基础'"); err != nil {
		return fmt.Errorf("migrate subcategory: %w", err)
	}
	return nil
}

func (s *Store) ensureExerciseColumn(name, ddl string) error {
	rows, err := s.db.Query(`PRAGMA table_info(exercises)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notnull, pk int
		var cname, ctype string
		var dflt any
		if err := rows.Scan(&cid, &cname, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if cname == name {
			return nil
		}
	}
	_, err = s.db.Exec(`ALTER TABLE exercises ADD COLUMN ` + name + ` ` + ddl)
	return err
}

// --- exercises ---

// UpsertExercise inserts or replaces an exercise.
func (s *Store) UpsertExercise(ex model.Exercise) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if ex.Subcategory == "" {
		ex.Subcategory = "基础"
	}
	if err := s.ensureCategory(ex.TechStack, ex.Subcategory); err != nil {
		return err
	}
	fp, _ := json.Marshal(ex.FillPoints)
	tc, _ := json.Marshal(ex.TestCases)
	sh, _ := json.Marshal(ex.SpotterHints)
	_, err := s.db.Exec(`
		INSERT INTO exercises (
			id, title, description, exercise_type, tech_stack, weight, category, subcategory,
			full_code, buggy_code, bug_description, bug_location, fill_points,
			expected_output, validation_code, test_cases, spotter_hints,
			time_limit_seconds, explanation, enabled, created_by, created_at, updated_at
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			title=excluded.title, description=excluded.description,
			exercise_type=excluded.exercise_type, tech_stack=excluded.tech_stack,
			weight=excluded.weight, category=excluded.category, subcategory=excluded.subcategory,
			full_code=excluded.full_code, buggy_code=excluded.buggy_code,
			bug_description=excluded.bug_description, bug_location=excluded.bug_location,
			fill_points=excluded.fill_points, expected_output=excluded.expected_output,
			validation_code=excluded.validation_code, test_cases=excluded.test_cases,
			spotter_hints=excluded.spotter_hints, time_limit_seconds=excluded.time_limit_seconds,
			explanation=excluded.explanation, enabled=excluded.enabled,
			created_by=excluded.created_by, updated_at=excluded.updated_at`,
		ex.ID, ex.Title, ex.Description, ex.ExerciseType, ex.TechStack, ex.Weight, ex.Category, ex.Subcategory,
		ex.FullCode, ex.BuggyCode, ex.BugDescription, ex.BugLocation, string(fp),
		ex.ExpectedOutput, ex.ValidationCode, string(tc), string(sh),
		ex.TimeLimitSecs, ex.Explanation, boolInt(ex.Enabled), ex.CreatedBy, now, now,
	)
	return err
}

// ListExercises returns all enabled exercises.
func (s *Store) ListExercises(techStack string) ([]model.Exercise, error) {
	q := `SELECT id, title, description, exercise_type, tech_stack, weight, category, subcategory,
		full_code, buggy_code, bug_description, bug_location, fill_points,
		expected_output, validation_code, test_cases, spotter_hints,
		time_limit_seconds, explanation, enabled, created_by, created_at, updated_at
		FROM exercises WHERE enabled=1`
	args := []any{}
	if techStack != "" {
		q += ` AND tech_stack=?`
		args = append(args, techStack)
	}
	q += ` ORDER BY tech_stack, title`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Exercise
	for rows.Next() {
		ex, err := scanExercise(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, rows.Err()
}

// GetExercise returns an exercise by id.
func (s *Store) GetExercise(id string) (model.Exercise, error) {
	row := s.db.QueryRow(`
		SELECT id, title, description, exercise_type, tech_stack, weight, category, subcategory,
		full_code, buggy_code, bug_description, bug_location, fill_points,
		expected_output, validation_code, test_cases, spotter_hints,
		time_limit_seconds, explanation, enabled, created_by, created_at, updated_at
		FROM exercises WHERE id=?`, id)
	ex, err := scanExercise(row)
	if err != nil {
		return model.Exercise{}, err
	}
	return ex, nil
}

// DeleteExercise removes an exercise by id.
func (s *Store) DeleteExercise(id string) error {
	_, err := s.db.Exec(`DELETE FROM exercises WHERE id=?`, id)
	return err
}

// CountExercises returns the number of exercises (admin stats).
func (s *Store) CountExercises() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM exercises WHERE enabled=1`).Scan(&n)
	return n, err
}

type rowScanner interface{ Scan(dest ...any) error }

func scanExercise(r rowScanner) (model.Exercise, error) {
	var (
		ex             model.Exercise
		fp, tc, sh     string
		enabled        int
		createdAt, upd string
	)
	err := r.Scan(
		&ex.ID, &ex.Title, &ex.Description, &ex.ExerciseType, &ex.TechStack, &ex.Weight, &ex.Category, &ex.Subcategory,
		&ex.FullCode, &ex.BuggyCode, &ex.BugDescription, &ex.BugLocation, &fp,
		&ex.ExpectedOutput, &ex.ValidationCode, &tc, &sh,
		&ex.TimeLimitSecs, &ex.Explanation, &enabled, &ex.CreatedBy, &createdAt, &upd,
	)
	if err != nil {
		return ex, err
	}
	json.Unmarshal([]byte(fp), &ex.FillPoints)
	json.Unmarshal([]byte(tc), &ex.TestCases)
	json.Unmarshal([]byte(sh), &ex.SpotterHints)
	ex.Enabled = enabled != 0
	ex.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	ex.UpdatedAt, _ = time.Parse(time.RFC3339, upd)
	return ex, nil
}

// --- attempt records ---

// SaveAttempt persists an attempt record.
func (s *Store) SaveAttempt(a model.AttemptRecord) error {
	fp, _ := json.Marshal(a.FilledPoints)
	ua, _ := json.Marshal(a.UserAnswers)
	hu, _ := json.Marshal(a.HintUsage)
	_, err := s.db.Exec(`
		INSERT INTO attempt_records (
			id, user_id, exercise_id, variant_id, weight_level, filled_points,
			user_answers, execution_output, is_correct, score, grade,
			error_message, execution_time_ms, hint_usage, created_at
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		a.ID, a.UserID, a.ExerciseID, a.VariantID, a.WeightLevel, string(fp),
		string(ua), a.ExecutionOutput, boolInt(a.IsCorrect), a.Score, a.Grade,
		a.ErrorMessage, a.ExecutionTimeMS, string(hu), a.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// RecentAttempts returns the most recent attempts for a user.
func (s *Store) RecentAttempts(userID string, limit int) ([]model.AttemptRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, exercise_id, variant_id, weight_level, filled_points,
		user_answers, execution_output, is_correct, score, grade,
		error_message, execution_time_ms, hint_usage, created_at
		FROM attempt_records WHERE user_id=? ORDER BY created_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.AttemptRecord
	for rows.Next() {
		var a model.AttemptRecord
		var fp, ua, hu, createdAt string
		var correct int
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.ExerciseID, &a.VariantID, &a.WeightLevel, &fp,
			&ua, &a.ExecutionOutput, &correct, &a.Score, &a.Grade,
			&a.ErrorMessage, &a.ExecutionTimeMS, &hu, &createdAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(fp), &a.FilledPoints)
		json.Unmarshal([]byte(ua), &a.UserAnswers)
		json.Unmarshal([]byte(hu), &a.HintUsage)
		a.IsCorrect = correct != 0
		a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		out = append(out, a)
	}
	return out, rows.Err()
}

// --- user progress ---

// GetProgress returns a user's progress (creating defaults if absent).
func (s *Store) GetProgress(userID string) (model.UserProgress, error) {
	p := model.UserProgress{UserID: userID, CompletedExercises: map[string]bool{}, WeakPoints: map[string]int{}, TechStackScores: map[string]float64{}}
	var ce, wp, ts, lastAt, updated string
	err := s.db.QueryRow(`
		SELECT user_id, total_attempts, correct_attempts, completed_exercises, weak_points,
		tech_stack_scores, total_hint_used, current_streak, longest_streak, last_attempt_at, updated_at
		FROM user_progress WHERE user_id=?`, userID).Scan(
		&p.UserID, &p.TotalAttempts, &p.CorrectAttempts, &ce, &wp, &ts,
		&p.TotalHintUsed, &p.CurrentStreak, &p.LongestStreak, &lastAt, &updated)
	if err == sql.ErrNoRows {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	json.Unmarshal([]byte(ce), &p.CompletedExercises)
	json.Unmarshal([]byte(wp), &p.WeakPoints)
	json.Unmarshal([]byte(ts), &p.TechStackScores)
	if lastAt != "" {
		if t, e := time.Parse(time.RFC3339, lastAt); e == nil {
			p.LastAttemptAt = &t
		}
	}
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return p, nil
}

// UpdateProgress applies a just-completed attempt to the stored progress.
func (s *Store) UpdateProgress(ex model.Exercise, score float64, grade string, hintUsed int, correct bool, userID string) error {
	now := time.Now().UTC()
	p, err := s.GetProgress(userID)
	if err != nil {
		return err
	}

	p.TotalAttempts++
	if correct {
		p.CorrectAttempts++
	}
	p.CompletedExercises[ex.ID] = true
	if !correct {
		p.WeakPoints[ex.ID]++
	}
	p.TechStackScores[ex.TechStack] += score
	p.TotalHintUsed += hintUsed
	if hintUsed > 0 {
		p.WeakPoints[ex.ID+"#hint"]++
	}

	// streak bookkeeping
	if p.LastAttemptAt != nil {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		last := time.Date(p.LastAttemptAt.Year(), p.LastAttemptAt.Month(), p.LastAttemptAt.Day(), 0, 0, 0, 0, time.UTC)
		diff := today.Sub(last)
		if diff <= 24*time.Hour && diff >= 0 {
			p.CurrentStreak++
		} else if diff > 24*time.Hour {
			p.CurrentStreak = 1
		}
		if p.CurrentStreak > p.LongestStreak {
			p.LongestStreak = p.CurrentStreak
		}
	} else {
		p.CurrentStreak = 1
		p.LongestStreak = 1
	}
	p.LastAttemptAt = &now
	p.UpdatedAt = now

	ce, _ := json.Marshal(p.CompletedExercises)
	wp, _ := json.Marshal(p.WeakPoints)
	ts, _ := json.Marshal(p.TechStackScores)
	lastAt := p.LastAttemptAt.Format(time.RFC3339)
	_, err = s.db.Exec(`
		INSERT INTO user_progress (
			user_id, total_attempts, correct_attempts, completed_exercises, weak_points,
			tech_stack_scores, total_hint_used, current_streak, longest_streak, last_attempt_at, updated_at
		) VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(user_id) DO UPDATE SET
			total_attempts=excluded.total_attempts, correct_attempts=excluded.correct_attempts,
			completed_exercises=excluded.completed_exercises, weak_points=excluded.weak_points,
			tech_stack_scores=excluded.tech_stack_scores, total_hint_used=excluded.total_hint_used,
			current_streak=excluded.current_streak, longest_streak=excluded.longest_streak,
			last_attempt_at=excluded.last_attempt_at, updated_at=excluded.updated_at`,
		p.UserID, p.TotalAttempts, p.CorrectAttempts, string(ce), string(wp), string(ts),
		p.TotalHintUsed, p.CurrentStreak, p.LongestStreak, lastAt, p.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// AllProgress returns progress for all users (admin stats).
func (s *Store) AllProgress() ([]model.UserProgress, error) {
	rows, err := s.db.Query(`SELECT user_id FROM user_progress`)
	if err != nil {
		return nil, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.UserProgress, 0, len(ids))
	for _, id := range ids {
		p, err := s.GetProgress(id)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// Stats returns lightweight platform statistics for the admin dashboard.
func (s *Store) Stats() (map[string]any, error) {
	out := map[string]any{}
	exCount, err := s.CountExercises()
	if err != nil {
		return nil, err
	}
	out["exercise_count"] = exCount

	var attempts, users int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM attempt_records`).Scan(&attempts); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM attempt_records`).Scan(&users); err != nil {
		return nil, err
	}
	out["attempt_count"] = attempts
	out["user_count"] = users
	return out, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ValidateWeightLevel returns the level if valid, otherwise the default.
func ValidateWeightLevel(level string) string {
	level = strings.ToLower(strings.TrimSpace(level))
	if _, ok := model.WeightLevels[level]; ok {
		return level
	}
	return model.DefaultWeightLevel
}

// --- categories ---

// ensureCategory creates the (tech_stack, subcategory) row if absent.
func (s *Store) ensureCategory(tech, sub string) error {
	if sub == "" {
		return nil
	}
	_, err := s.db.Exec(`INSERT INTO categories (tech_stack, subcategory, created_at) VALUES (?,?,?)
		ON CONFLICT(tech_stack, subcategory) DO NOTHING`,
		tech, sub, time.Now().UTC().Format(time.RFC3339))
	return err
}

// CreateCategory adds a new subcategory (may be empty initially).
func (s *Store) CreateCategory(tech, sub string) error {
	return s.ensureCategory(tech, sub)
}

// ListCategories returns all managed subcategories with per-category counts.
func (s *Store) ListCategories() ([]model.Category, error) {
	rows, err := s.db.Query(`
		SELECT c.tech_stack, c.subcategory,
			(SELECT COUNT(*) FROM exercises e WHERE e.tech_stack = c.tech_stack AND e.subcategory = c.subcategory AND e.enabled = 1)
		FROM categories c
		ORDER BY c.tech_stack, c.subcategory`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.TechStack, &c.Subcategory, &c.ExerciseCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// RenameCategory renames a subcategory across the categories table and all its
// exercises (moving them along).
func (s *Store) RenameCategory(tech, old, new string) error {
	if new == "" {
		return fmt.Errorf("new subcategory is required")
	}
	if err := s.ensureCategory(tech, new); err != nil {
		return err
	}
	if _, err := s.db.Exec(`UPDATE exercises SET subcategory=? WHERE tech_stack=? AND subcategory=?`, new, tech, old); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM categories WHERE tech_stack=? AND subcategory=?`, tech, old); err != nil {
		return err
	}
	return nil
}

// DeleteCategory removes a subcategory; its exercises move back to "基础".
func (s *Store) DeleteCategory(tech, sub string) error {
	if sub == "基础" {
		return fmt.Errorf("不能删除默认分类「基础」")
	}
	if _, err := s.db.Exec(`UPDATE exercises SET subcategory='基础' WHERE tech_stack=? AND subcategory=?`, tech, sub); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM categories WHERE tech_stack=? AND subcategory=?`, tech, sub)
	return err
}

// --- favorites ---

// ToggleFavorite adds or removes a favorite for a user and returns the new state.
func (s *Store) ToggleFavorite(userID, exerciseID string) (bool, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var exists int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM favorites WHERE user_id=? AND exercise_id=?`, userID, exerciseID).Scan(&exists); err != nil {
		return false, err
	}
	if exists > 0 {
		if _, err := s.db.Exec(`DELETE FROM favorites WHERE user_id=? AND exercise_id=?`, userID, exerciseID); err != nil {
			return false, err
		}
		return false, nil
	}
	if _, err := s.db.Exec(`INSERT INTO favorites (user_id, exercise_id, created_at) VALUES (?,?,?)`, userID, exerciseID, now); err != nil {
		return false, err
	}
	return true, nil
}

// IsFavorite reports whether a user has favorited an exercise.
func (s *Store) IsFavorite(userID, exerciseID string) (bool, error) {
	var exists int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM favorites WHERE user_id=? AND exercise_id=?`, userID, exerciseID).Scan(&exists); err != nil {
		return false, err
	}
	return exists > 0, nil
}

// Favorites returns the favorited exercise ids for a user, newest first.
func (s *Store) Favorites(userID string) ([]model.FavoriteItem, error) {
	rows, err := s.db.Query(`SELECT exercise_id, created_at FROM favorites WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []model.FavoriteItem
	for rows.Next() {
		var id, at string
		if err := rows.Scan(&id, &at); err != nil {
			return nil, err
		}
		item := model.FavoriteItem{ExerciseID: id}
		item.FavoritedAt, _ = time.Parse(time.RFC3339, at)
		ids = append(ids, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// enrich with exercise metadata
	for i := range ids {
		if ex, err := s.GetExercise(ids[i].ExerciseID); err == nil {
			ids[i].Title = ex.Title
			ids[i].TechStack = ex.TechStack
			ids[i].ExerciseType = ex.ExerciseType
		}
	}
	return ids, nil
}
