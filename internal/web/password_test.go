package web

import (
	"strings"
	"testing"
)

func TestHashPasswordRoundTrip(t *testing.T) {
	const pw = "çokGizli-şifre_2026"
	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, pbkdf2Prefix+"$") {
		t.Errorf("hash %q does not start with %q", hash, pbkdf2Prefix)
	}
	if strings.Contains(hash, pw) {
		t.Error("hash leaks the plaintext password")
	}

	ok, err := VerifyPassword(hash, pw)
	if err != nil {
		t.Fatalf("VerifyPassword (correct): %v", err)
	}
	if !ok {
		t.Error("correct password did not verify")
	}

	ok, err = VerifyPassword(hash, pw+"x")
	if err != nil {
		t.Fatalf("VerifyPassword (wrong): %v", err)
	}
	if ok {
		t.Error("wrong password verified")
	}
}

func TestHashPasswordUsesFreshSalt(t *testing.T) {
	h1, _ := HashPassword("same")
	h2, _ := HashPassword("same")
	if h1 == h2 {
		t.Error("two hashes of the same password are identical (salt not random)")
	}
}

func TestVerifyPasswordMalformed(t *testing.T) {
	for _, bad := range []string{
		"",
		"not-a-hash",
		"pbkdf2-sha256$abc$c2FsdA$aGFzaA", // non-numeric iterations
		"bcrypt$10$c2FsdA$aGFzaA",         // wrong algorithm
		"pbkdf2-sha256$600000$!!!$aGFzaA", // bad base64 salt
	} {
		if _, err := VerifyPassword(bad, "x"); err == nil {
			t.Errorf("VerifyPassword(%q) error = nil, want malformed-hash error", bad)
		}
	}
}
