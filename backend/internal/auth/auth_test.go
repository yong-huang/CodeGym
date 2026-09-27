package auth

import (
	"strings"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	h, err := HashPassword("s3cret!")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "pbkdf2_sha256$") {
		t.Fatalf("unexpected hash format: %s", h)
	}
	if !CheckPassword("s3cret!", h) {
		t.Error("correct password rejected")
	}
	if CheckPassword("wrong", h) {
		t.Error("wrong password accepted")
	}
	if CheckPassword("s3cret!", "garbage") {
		t.Error("garbage hash accepted")
	}
}

func TestHashIsSalted(t *testing.T) {
	a, _ := HashPassword("same")
	b, _ := HashPassword("same")
	if a == b {
		t.Error("hashes should differ due to random salt")
	}
}

func TestNewTokenAndUserID(t *testing.T) {
	tok, err := NewToken()
	if err != nil || len(tok) != TokenBytes*2 {
		t.Fatalf("token = %q (err %v)", tok, err)
	}
	id, err := NewUserID()
	if err != nil || !strings.HasPrefix(id, "u_") {
		t.Fatalf("user id = %q (err %v)", id, err)
	}
}
