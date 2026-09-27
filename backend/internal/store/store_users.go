package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"codegym/backend/internal/model"
)

// ErrNotFound is returned when a user or session does not exist.
var ErrNotFound = errors.New("not found")

const sessionTTL = 30 * 24 * time.Hour

// --- users ---

// CreateUser inserts a new user. Username is unique (case-insensitive).
func (s *Store) CreateUser(u model.User, passwordHash string) (model.User, error) {
	now := time.Now().UTC()
	if u.ID == "" || u.Username == "" || passwordHash == "" {
		return model.User{}, errors.New("user id, username and password hash are required")
	}
	if u.DisplayName == "" {
		u.DisplayName = u.Username
	}
	sel, _ := json.Marshal(u.SelectedBanks)
	_, err := s.db.Exec(`
		INSERT INTO users (id, username, password_hash, display_name, disabled, selected_banks, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?)`,
		u.ID, u.Username, passwordHash, u.DisplayName, boolInt(u.Disabled), string(sel),
		now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		return model.User{}, err
	}
	u.CreatedAt, u.UpdatedAt = now, now
	return u, nil
}

func scanUser(r rowScanner) (model.User, string, error) {
	var (
		u            model.User
		hash, sel    string
		disabled     int
		created, upd string
	)
	err := r.Scan(&u.ID, &u.Username, &hash, &u.DisplayName, &disabled, &sel, &created, &upd)
	if err != nil {
		return u, "", err
	}
	u.Disabled = disabled != 0
	json.Unmarshal([]byte(sel), &u.SelectedBanks)
	if u.SelectedBanks == nil {
		u.SelectedBanks = []string{}
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, created)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, upd)
	return u, hash, nil
}

const userColumns = `id, username, password_hash, display_name, disabled, selected_banks, created_at, updated_at`

// GetUserByUsername looks up a user by username (case-insensitive).
func (s *Store) GetUserByUsername(username string) (model.User, string, error) {
	row := s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE username=? COLLATE NOCASE`, username)
	u, hash, err := scanUser(row)
	if err == sql.ErrNoRows {
		return model.User{}, "", ErrNotFound
	}
	return u, hash, err
}

// GetUserByID looks up a user by id.
func (s *Store) GetUserByID(id string) (model.User, error) {
	row := s.db.QueryRow(`SELECT `+userColumns+` FROM users WHERE id=?`, id)
	u, _, err := scanUser(row)
	if err == sql.ErrNoRows {
		return model.User{}, ErrNotFound
	}
	return u, err
}

// ListUsers returns all users, newest first.
func (s *Store) ListUsers() ([]model.User, error) {
	rows, err := s.db.Query(`SELECT ` + userColumns + ` FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.User
	for rows.Next() {
		u, _, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// UpdateUser persists display name, disabled flag, password hash (when
// non-empty) and bank selection for a user.
func (s *Store) UpdateUser(u model.User, newPasswordHash string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	sel, _ := json.Marshal(u.SelectedBanks)
	q := `UPDATE users SET display_name=?, disabled=?, selected_banks=?, updated_at=?`
	args := []any{u.DisplayName, boolInt(u.Disabled), string(sel), now}
	if newPasswordHash != "" {
		q += `, password_hash=?`
		args = append(args, newPasswordHash)
	}
	q += ` WHERE id=?`
	args = append(args, u.ID)
	res, err := s.db.Exec(q, args...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// SetUserBanks stores the selected "tech::sub" keys for a user.
func (s *Store) SetUserBanks(userID string, banks []string) error {
	if banks == nil {
		banks = []string{}
	}
	sel, _ := json.Marshal(banks)
	res, err := s.db.Exec(`UPDATE users SET selected_banks=?, updated_at=? WHERE id=?`,
		string(sel), time.Now().UTC().Format(time.RFC3339), userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteUser removes a user and every piece of their data.
func (s *Store) DeleteUser(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range []string{
		`DELETE FROM sessions WHERE user_id=?`,
		`DELETE FROM user_progress WHERE user_id=?`,
		`DELETE FROM favorites WHERE user_id=?`,
		`DELETE FROM attempt_records WHERE user_id=?`,
		`DELETE FROM users WHERE id=?`,
	} {
		if _, err := tx.Exec(q, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// --- sessions ---

// CreateSession issues a session token for a user.
func (s *Store) CreateSession(userID, token string) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(`INSERT INTO sessions (token, user_id, created_at, expires_at) VALUES (?,?,?,?)`,
		token, userID, now.Format(time.RFC3339), now.Add(sessionTTL).Format(time.RFC3339))
	return err
}

// SessionUser returns the user id for a valid (unexpired) token.
func (s *Store) SessionUser(token string) (string, bool) {
	if token == "" {
		return "", false
	}
	var userID, expires string
	err := s.db.QueryRow(`SELECT user_id, expires_at FROM sessions WHERE token=?`, token).Scan(&userID, &expires)
	if err != nil {
		return "", false
	}
	exp, err := time.Parse(time.RFC3339, expires)
	if err != nil || time.Now().UTC().After(exp) {
		s.db.Exec(`DELETE FROM sessions WHERE token=?`, token)
		return "", false
	}
	return userID, true
}

// DeleteSession revokes one token.
func (s *Store) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token=?`, token)
	return err
}

// DeleteUserSessions revokes every token of a user (e.g. after a password reset).
func (s *Store) DeleteUserSessions(userID string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE user_id=?`, userID)
	return err
}

// --- banks ---

// ListBankInfo returns every selectable bank (tech + subcategory) with counts.
func (s *Store) ListBankInfo() ([]model.BankInfo, error) {
	rows, err := s.db.Query(`
		SELECT tech_stack, subcategory, COUNT(*) FROM exercises
		WHERE enabled=1 GROUP BY tech_stack, subcategory
		ORDER BY tech_stack, subcategory`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.BankInfo
	for rows.Next() {
		var b model.BankInfo
		if err := rows.Scan(&b.TechStack, &b.Subcategory, &b.ExerciseCount); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// BankKey joins tech and subcategory into the stored selection key.
func BankKey(tech, sub string) string { return tech + "::" + sub }

// ParseBankKey splits a selection key; ok is false when malformed.
func ParseBankKey(key string) (tech, sub string, ok bool) {
	parts := strings.SplitN(key, "::", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// SaveBankSource records a dynamically loaded bank (upload or directory scan).
func (s *Store) SaveBankSource(name, source string, exerciseCount int, content string) error {
	_, err := s.db.Exec(`
		INSERT INTO bank_sources (name, source, exercise_count, content, updated_at)
		VALUES (?,?,?,?,?)
		ON CONFLICT(name, source) DO UPDATE SET
			exercise_count=excluded.exercise_count, content=excluded.content, updated_at=excluded.updated_at`,
		name, source, exerciseCount, content, time.Now().UTC().Format(time.RFC3339))
	return err
}

// ListBankSources returns all dynamically loaded banks, newest first.
func (s *Store) ListBankSources() ([]model.BankSource, error) {
	rows, err := s.db.Query(`SELECT name, source, exercise_count, updated_at FROM bank_sources ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.BankSource
	for rows.Next() {
		var b model.BankSource
		var upd string
		if err := rows.Scan(&b.Name, &b.Source, &b.ExerciseCount, &upd); err != nil {
			return nil, err
		}
		b.UpdatedAt, _ = time.Parse(time.RFC3339, upd)
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBankSource returns one bank source record by name and source kind.
func (s *Store) GetBankSource(name, source string) (model.BankSource, error) {
	var b model.BankSource
	var upd string
	err := s.db.QueryRow(`SELECT name, source, exercise_count, updated_at FROM bank_sources WHERE name=? AND source=?`,
		name, source).Scan(&b.Name, &b.Source, &b.ExerciseCount, &upd)
	if err == sql.ErrNoRows {
		return b, ErrNotFound
	}
	if err != nil {
		return b, err
	}
	b.UpdatedAt, _ = time.Parse(time.RFC3339, upd)
	return b, nil
}
