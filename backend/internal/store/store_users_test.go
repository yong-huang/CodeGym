package store

import (
	"path/filepath"
	"testing"

	"codegym/backend/internal/model"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestUserCRUDAndSessions(t *testing.T) {
	st := newTestStore(t)

	u, err := st.CreateUser(model.User{ID: "u_1", Username: "Alice", DisplayName: "爱丽丝"}, "hash1")
	if err != nil {
		t.Fatal(err)
	}
	if u.DisplayName != "爱丽丝" {
		t.Fatalf("display name = %q", u.DisplayName)
	}

	// case-insensitive lookup
	got, hash, err := st.GetUserByUsername("alice")
	if err != nil || got.ID != "u_1" || hash != "hash1" {
		t.Fatalf("lookup alice: %+v hash=%q err=%v", got, hash, err)
	}
	if _, _, err := st.GetUserByUsername("nobody"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// duplicate username rejected
	if _, err := st.CreateUser(model.User{ID: "u_2", Username: "ALICE"}, "hash2"); err == nil {
		t.Fatal("duplicate username should fail")
	}

	// sessions
	if err := st.CreateSession("u_1", "tok-1"); err != nil {
		t.Fatal(err)
	}
	if uid, ok := st.SessionUser("tok-1"); !ok || uid != "u_1" {
		t.Fatalf("session lookup = %q,%v", uid, ok)
	}
	if _, ok := st.SessionUser("bad-token"); ok {
		t.Fatal("bad token accepted")
	}
	if err := st.DeleteSession("tok-1"); err != nil {
		t.Fatal(err)
	}
	if _, ok := st.SessionUser("tok-1"); ok {
		t.Fatal("revoked token accepted")
	}

	// bank selection
	if err := st.SetUserBanks("u_1", []string{"python::基础", "testing::pytest"}); err != nil {
		t.Fatal(err)
	}
	got, err = st.GetUserByID("u_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.SelectedBanks) != 2 || got.SelectedBanks[1] != "testing::pytest" {
		t.Fatalf("selected banks = %v", got.SelectedBanks)
	}

	// disable + password reset
	got.Disabled = true
	if err := st.UpdateUser(got, "hash-new"); err != nil {
		t.Fatal(err)
	}
	_, hash, _ = st.GetUserByUsername("alice")
	if hash != "hash-new" {
		t.Fatalf("password hash = %q, want hash-new", hash)
	}

	// delete cascades
	if err := st.CreateSession("u_1", "tok-2"); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteUser("u_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.GetUserByID("u_1"); err != ErrNotFound {
		t.Fatalf("user still present after delete: %v", err)
	}
	if _, ok := st.SessionUser("tok-2"); ok {
		t.Fatal("session survived user delete")
	}
}

func TestListUsersAndBankInfo(t *testing.T) {
	st := newTestStore(t)
	if _, err := st.CreateUser(model.User{ID: "u_1", Username: "a"}, "h"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateUser(model.User{ID: "u_2", Username: "b"}, "h"); err != nil {
		t.Fatal(err)
	}
	users, err := st.ListUsers()
	if err != nil || len(users) != 2 {
		t.Fatalf("users = %d (err %v)", len(users), err)
	}

	for _, ex := range []model.Exercise{
		{ID: "e1", Title: "t", FullCode: "x", TechStack: "python", Subcategory: "基础", Weight: "easy", ExerciseType: "fill", Enabled: true},
		{ID: "e2", Title: "t", FullCode: "x", TechStack: "python", Subcategory: "基础", Weight: "easy", ExerciseType: "fill", Enabled: true},
		{ID: "e3", Title: "t", FullCode: "x", TechStack: "testing", Subcategory: "pytest", Weight: "easy", ExerciseType: "fill", Enabled: true},
	} {
		if err := st.UpsertExercise(ex); err != nil {
			t.Fatal(err)
		}
	}
	infos, err := st.ListBankInfo()
	if err != nil || len(infos) != 2 {
		t.Fatalf("bank infos = %+v (err %v)", infos, err)
	}
	if infos[0].TechStack != "python" || infos[0].ExerciseCount != 2 {
		t.Errorf("first bank = %+v, want python/2", infos[0])
	}
	if k := BankKey("python", "基础"); k != "python::基础" {
		t.Errorf("BankKey = %q", k)
	}
	if tech, sub, ok := ParseBankKey("testing::pytest"); !ok || tech != "testing" || sub != "pytest" {
		t.Errorf("ParseBankKey = %q,%q,%v", tech, sub, ok)
	}
	if _, _, ok := ParseBankKey("broken"); ok {
		t.Error("broken key should not parse")
	}

	// bank sources
	if err := st.SaveBankSource("mybank", "upload", 3, "[]"); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveBankSource("mybank", "upload", 5, "[]"); err != nil {
		t.Fatal(err)
	}
	src, err := st.GetBankSource("mybank", "upload")
	if err != nil || src.ExerciseCount != 5 {
		t.Fatalf("source = %+v (err %v)", src, err)
	}
	list, err := st.ListBankSources()
	if err != nil || len(list) != 1 {
		t.Fatalf("sources = %d (err %v)", len(list), err)
	}
}
