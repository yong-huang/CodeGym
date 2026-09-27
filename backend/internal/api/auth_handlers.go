package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"codegym/backend/internal/auth"
	"codegym/backend/internal/model"
	"codegym/backend/internal/seed"
	"codegym/backend/internal/store"
)

var (
	usernameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{2,31}$`)
	bankNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)
)

// bearerToken extracts the session token from Authorization or X-Auth-Token.
func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	return strings.TrimSpace(r.Header.Get("X-Auth-Token"))
}

// currentUser resolves the authenticated user from the session token.
func (s *Server) currentUser(r *http.Request) (model.User, bool) {
	token := bearerToken(r)
	if token == "" {
		return model.User{}, false
	}
	uid, ok := s.store.SessionUser(token)
	if !ok {
		return model.User{}, false
	}
	u, err := s.store.GetUserByID(uid)
	if err != nil || u.Disabled {
		return model.User{}, false
	}
	return u, true
}

type credentialsRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func (s *Server) issueSession(u model.User) (authResponse, error) {
	token, err := auth.NewToken()
	if err != nil {
		return authResponse{}, err
	}
	if err := s.store.CreateSession(u.ID, token); err != nil {
		return authResponse{}, err
	}
	return authResponse{Token: token, User: u}, nil
}

// register creates a trainee account and logs it in.
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if !usernameRe.MatchString(req.Username) {
		writeErr(w, http.StatusBadRequest, "用户名需为 3-32 位字母/数字/下划线/连字符，且以字母或数字开头")
		return
	}
	if len(req.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "密码至少 6 位")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	id, err := auth.NewUserID()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	u, err := s.store.CreateUser(model.User{
		ID: id, Username: req.Username, DisplayName: strings.TrimSpace(req.DisplayName),
	}, hash)
	if err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			writeErr(w, http.StatusConflict, "用户名已存在")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp, err := s.issueSession(u)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// login verifies credentials and issues a session token.
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}
	u, hash, err := s.store.GetUserByUsername(strings.TrimSpace(req.Username))
	if err != nil || !auth.CheckPassword(req.Password, hash) {
		writeErr(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	if u.Disabled {
		writeErr(w, http.StatusForbidden, "账号已被禁用，请联系教练")
		return
	}
	resp, err := s.issueSession(u)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// logout revokes the current session token.
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if token := bearerToken(r); token != "" {
		if err := s.store.DeleteSession(token); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// me returns the current user profile.
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	u, ok := s.currentUser(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "未登录或登录已过期")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": u})
}

// updateMe updates display name and/or password of the current user.
func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	u, ok := s.currentUser(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "未登录或登录已过期")
		return
	}
	var req struct {
		DisplayName *string `json:"display_name"`
		OldPassword string  `json:"old_password"`
		NewPassword string  `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}
	if req.DisplayName != nil {
		u.DisplayName = strings.TrimSpace(*req.DisplayName)
	}
	newHash := ""
	if req.NewPassword != "" {
		if len(req.NewPassword) < 6 {
			writeErr(w, http.StatusBadRequest, "新密码至少 6 位")
			return
		}
		_, oldHash, err := s.store.GetUserByUsername(u.Username)
		if err != nil || !auth.CheckPassword(req.OldPassword, oldHash) {
			writeErr(w, http.StatusForbidden, "原密码不正确")
			return
		}
		newHash, err = auth.HashPassword(req.NewPassword)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := s.store.UpdateUser(u, newHash); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if newHash != "" {
		// 改密后其它会话全部失效，当前 token 保留
		_ = s.store.DeleteUserSessions(u.ID)
		if token := bearerToken(r); token != "" {
			_ = s.store.CreateSession(u.ID, token)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": u})
}

// myBanks lists every selectable bank plus the current user's selection.
func (s *Server) myBanks(w http.ResponseWriter, r *http.Request) {
	u, ok := s.currentUser(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "未登录或登录已过期")
		return
	}
	infos, err := s.store.ListBankInfo()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"available": infos, "selected": u.SelectedBanks})
}

// setMyBanks persists the user's selected "tech::sub" bank keys.
func (s *Server) setMyBanks(w http.ResponseWriter, r *http.Request) {
	u, ok := s.currentUser(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "未登录或登录已过期")
		return
	}
	var req struct {
		Selected []string `json:"selected"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}
	infos, err := s.store.ListBankInfo()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	valid := make(map[string]bool, len(infos))
	for _, b := range infos {
		valid[store.BankKey(b.TechStack, b.Subcategory)] = true
	}
	var cleaned []string
	seen := map[string]bool{}
	for _, k := range req.Selected {
		if valid[k] && !seen[k] {
			seen[k] = true
			cleaned = append(cleaned, k)
		}
	}
	if err := s.store.SetUserBanks(u.ID, cleaned); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"selected": cleaned})
}

// --- admin: user management ---

// adminListUsers returns every user with a light progress summary.
func (s *Server) adminListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.ListUsers()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]model.AdminUser, 0, len(users))
	for _, u := range users {
		p, err := s.store.GetProgress(u.ID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		out = append(out, model.AdminUser{
			User:           u,
			TotalAttempts:  p.TotalAttempts,
			CompletedCount: len(p.CompletedExercises),
			CurrentStreak:  p.CurrentStreak,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// adminCreateUser creates an account from the coach side.
func (s *Server) adminCreateUser(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if !usernameRe.MatchString(req.Username) {
		writeErr(w, http.StatusBadRequest, "用户名需为 3-32 位字母/数字/下划线/连字符，且以字母或数字开头")
		return
	}
	if len(req.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "密码至少 6 位")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	id, err := auth.NewUserID()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	u, err := s.store.CreateUser(model.User{
		ID: id, Username: req.Username, DisplayName: strings.TrimSpace(req.DisplayName),
	}, hash)
	if err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			writeErr(w, http.StatusConflict, "用户名已存在")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

// adminUpdateUser updates display name / disabled flag / password.
func (s *Server) adminUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	u, err := s.store.GetUserByID(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "用户不存在")
		return
	}
	var req struct {
		DisplayName *string `json:"display_name"`
		Disabled    *bool   `json:"disabled"`
		Password    string  `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}
	if req.DisplayName != nil {
		u.DisplayName = strings.TrimSpace(*req.DisplayName)
	}
	if req.Disabled != nil {
		u.Disabled = *req.Disabled
	}
	newHash := ""
	if req.Password != "" {
		if len(req.Password) < 6 {
			writeErr(w, http.StatusBadRequest, "密码至少 6 位")
			return
		}
		newHash, err = auth.HashPassword(req.Password)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := s.store.UpdateUser(u, newHash); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if newHash != "" || u.Disabled {
		_ = s.store.DeleteUserSessions(u.ID)
	}
	writeJSON(w, http.StatusOK, u)
}

// adminDeleteUser removes a user and all of their data.
func (s *Server) adminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteUser(id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- admin: dynamic bank loading ---

// adminListBanks lists dynamically loaded banks (uploads + scanned files).
func (s *Server) adminListBanks(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListBankSources()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// adminUploadBank parses an uploaded bank JSON and upserts its exercises
// immediately, so it is practiceable without a restart.
func (s *Server) adminUploadBank(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}
	name := strings.TrimSpace(req.Name)
	if !bankNameRe.MatchString(name) {
		writeErr(w, http.StatusBadRequest, "题库名需为 1-64 位字母/数字/下划线/连字符，且以字母或数字开头")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeErr(w, http.StatusBadRequest, "题库内容为空")
		return
	}
	exs, err := seed.BuildBankJSON(name+".json", []byte(req.Content))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "题库解析失败："+err.Error())
		return
	}
	if len(exs) == 0 {
		writeErr(w, http.StatusBadRequest, "题库不包含任何题目")
		return
	}
	for _, ex := range exs {
		if err := s.store.UpsertExercise(ex); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := s.store.SaveBankSource(name, "upload", len(exs), req.Content); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": name, "imported": len(exs)})
}
