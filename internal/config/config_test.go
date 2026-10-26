package config

import (
	"strings"
	"testing"
)

// clearEnv unsets every variable Load reads so each test starts from a known
// baseline regardless of the ambient environment.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"WORKEY_ADDR", "WORKEY_DB_PATH", "WORKEY_API_TOKEN",
		"WORKEY_PASSWORD_HASH", "WORKEY_SESSION_KEY", "WORKEY_EXAM_DATE",
		"WORKEY_COOKIE_SECURE", "WORKEY_DEV",
	} {
		t.Setenv(k, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("WORKEY_DEV", "1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Addr != "127.0.0.1:8080" {
		t.Errorf("Addr = %q, want 127.0.0.1:8080", cfg.Addr)
	}
	if !cfg.CookieSecure {
		t.Error("CookieSecure = false, want true by default")
	}
	if got := cfg.ExamDate.Format("2006-01-02"); got != "2027-07-18" {
		t.Errorf("ExamDate = %q, want 2027-07-18", got)
	}
	if !cfg.Dev {
		t.Error("Dev = false, want true")
	}
}

func TestLoadMissingRequired(t *testing.T) {
	clearEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing required vars")
	}
	for _, name := range []string{"WORKEY_API_TOKEN", "WORKEY_PASSWORD_HASH", "WORKEY_SESSION_KEY"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not name %s", err.Error(), name)
		}
	}
}

func TestLoadDevModeSkipsRequired(t *testing.T) {
	clearEnv(t)
	t.Setenv("WORKEY_DEV", "1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil in dev mode", err)
	}
	if !cfg.Dev {
		t.Error("Dev = false, want true")
	}
}

func TestLoadDoesNotLeakValues(t *testing.T) {
	clearEnv(t)
	const secretLike = "s3cr3t-should-not-appear"

	// An invalid boolean value must be rejected without echoing the value.
	t.Setenv("WORKEY_DEV", "1")
	t.Setenv("WORKEY_COOKIE_SECURE", secretLike)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for invalid WORKEY_COOKIE_SECURE")
	}
	if !strings.Contains(err.Error(), "WORKEY_COOKIE_SECURE") {
		t.Errorf("error %q does not name the variable", err.Error())
	}
	if strings.Contains(err.Error(), secretLike) {
		t.Errorf("error %q leaked the variable value", err.Error())
	}
}

func TestLoadInvalidExamDate(t *testing.T) {
	clearEnv(t)
	t.Setenv("WORKEY_DEV", "1")
	t.Setenv("WORKEY_EXAM_DATE", "18-07-2027")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for invalid WORKEY_EXAM_DATE")
	}
	if !strings.Contains(err.Error(), "WORKEY_EXAM_DATE") {
		t.Errorf("error %q does not name the variable", err.Error())
	}
	if strings.Contains(err.Error(), "18-07-2027") {
		t.Errorf("error %q leaked the variable value", err.Error())
	}
}
