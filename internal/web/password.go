package web

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Password hashing uses PBKDF2-HMAC-SHA256 from the standard library
// (crypto/pbkdf2, Go 1.24+), so single-user auth needs no external dependency.
// The encoded form is pbkdf2-sha256$<iter>$<b64 salt>$<b64 hash>, self-describing
// so the iteration count can be raised later without breaking existing hashes.
const (
	pbkdf2Prefix  = "pbkdf2-sha256"
	pbkdf2Iter    = 600000 // OWASP 2023 floor for PBKDF2-HMAC-SHA256
	pbkdf2KeyLen  = 32
	pbkdf2SaltLen = 16
)

// errMalformedHash reports a stored hash that is not in the expected encoding.
var errMalformedHash = errors.New("web: malformed password hash")

// HashPassword derives a PBKDF2-SHA256 hash of password with a fresh random salt
// and returns the self-describing encoded form.
func HashPassword(password string) (string, error) {
	salt := make([]byte, pbkdf2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("web: generate salt: %w", err)
	}
	dk, err := pbkdf2.Key(sha256.New, password, salt, pbkdf2Iter, pbkdf2KeyLen)
	if err != nil {
		return "", fmt.Errorf("web: derive key: %w", err)
	}
	return fmt.Sprintf("%s$%d$%s$%s",
		pbkdf2Prefix,
		pbkdf2Iter,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk),
	), nil
}

// VerifyPassword reports whether password matches the encoded hash. The final
// comparison is constant-time; a malformed encoding returns an error without
// leaking which part failed.
func VerifyPassword(encoded, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != pbkdf2Prefix {
		return false, errMalformedHash
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter < 1 {
		return false, errMalformedHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false, errMalformedHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, errMalformedHash
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, iter, len(want))
	if err != nil {
		return false, fmt.Errorf("web: derive key: %w", err)
	}
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
