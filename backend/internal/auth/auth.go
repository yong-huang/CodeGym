// Package auth implements password hashing (PBKDF2-HMAC-SHA256) and session
// token generation for the local user system.
package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const (
	// DefaultIterations is the PBKDF2 iteration count for new hashes.
	DefaultIterations = 120_000
	saltLen           = 16
	keyLen            = 32
	// TokenBytes is the entropy of a session token (hex-encoded).
	TokenBytes = 32
)

// HashPassword derives a storable hash of the form
// pbkdf2_sha256$<iterations>$<salt-hex>$<hash-hex>.
func HashPassword(password string) (string, error) {
	return hashPassword(password, DefaultIterations)
}

func hashPassword(password string, iter int) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, iter, keyLen)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", iter, hex.EncodeToString(salt), hex.EncodeToString(key)), nil
}

// CheckPassword reports whether password matches the stored hash.
func CheckPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, iter, len(want))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}

// NewToken returns a random hex session token.
func NewToken() (string, error) {
	b := make([]byte, TokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// NewUserID returns a random hex user id.
func NewUserID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "u_" + hex.EncodeToString(b), nil
}
