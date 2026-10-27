package config

import (
	"os"
	"strings"
	"testing"
)

func TestDotEnvLoadAndEnvironmentPrecedence(t *testing.T) {
	clearEnv(t)
	t.Chdir(t.TempDir())
	for _, key := range []string{"WORKEY_DEV", "WORKEY_ADDR", "WORKEY_PASSWORD_HASH", "WORKEY_DB_PATH", "WORKEY_COOKIE_SECURE"} {
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	const content = "\ufeff# local defaults\r\nexport WORKEY_DEV=1\r\nWORKEY_ADDR=127.0.0.1:9090 # comment\r\nWORKEY_DB_PATH='data/study plan.db'\r\nWORKEY_PASSWORD_HASH=\"pbkdf2-sha256$600000$salt$hash#literal\"\r\nWORKEY_COOKIE_SECURE=false\r\n"
	if err := os.WriteFile(".env", []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Dev || cfg.CookieSecure || cfg.Addr != "127.0.0.1:9090" || cfg.DBPath != "data/study plan.db" || cfg.PasswordHash != "pbkdf2-sha256$600000$salt$hash#literal" {
		t.Fatal(".env values were not loaded literally")
	}
	if _, exists := os.LookupEnv("WORKEY_PASSWORD_HASH"); exists {
		t.Fatal("Load changed the process environment")
	}
	t.Setenv("WORKEY_ADDR", "127.0.0.1:8088")
	t.Setenv("WORKEY_PASSWORD_HASH", "")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "127.0.0.1:8088" || cfg.PasswordHash != "" {
		t.Fatal("explicit environment, including empty values, must win")
	}
}

func TestDotEnvMalformedDoesNotLeak(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, content := range []string{"SECRET-VALUE", "1KEY=SECRET-VALUE", "KEY='SECRET-VALUE", `KEY="SECRET-VALUE" garbage`, "KEY=" + strings.Repeat("SECRET-VALUE", 10000)} {
		if err := os.WriteFile(".env", []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := Load()
		if err == nil {
			t.Fatal("malformed .env accepted")
		}
		if strings.Contains(err.Error(), "SECRET-VALUE") {
			t.Fatal("error leaked a value")
		}
	}
}
