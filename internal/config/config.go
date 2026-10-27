// Package config loads and validates workey's runtime configuration from the
// environment and an optional .env in the working directory. Environment
// variables take precedence. Error messages never include configuration values,
// so secrets are not leaked to logs.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds validated runtime configuration.
type Config struct {
	Addr         string    // WORKEY_ADDR
	DBPath       string    // WORKEY_DB_PATH
	APIToken     string    // WORKEY_API_TOKEN
	PasswordHash string    // WORKEY_PASSWORD_HASH
	SessionKey   string    // WORKEY_SESSION_KEY
	ExamDate     time.Time // WORKEY_EXAM_DATE (civil date, Istanbul)
	CookieSecure bool      // WORKEY_COOKIE_SECURE
	Dev          bool      // WORKEY_DEV=1
}

const (
	defaultAddr     = "127.0.0.1:8080"
	defaultExamDate = "2027-07-18"
	examDateLayout  = "2006-01-02"
)

// Load reads .env defaults and environment overrides, then validates them. When
// WORKEY_DEV is not "1", the API token, password hash and session key are
// required and Load refuses to start without them.
func Load() (*Config, error) {
	fileEnv, err := readDotEnv(".env")
	if err != nil {
		return nil, err
	}
	get := func(key string) string {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
		return fileEnv[key]
	}
	getDefault := func(key, fallback string) string {
		if value := get(key); value != "" {
			return value
		}
		return fallback
	}
	cfg := &Config{
		Addr:         getDefault("WORKEY_ADDR", defaultAddr),
		DBPath:       get("WORKEY_DB_PATH"),
		APIToken:     get("WORKEY_API_TOKEN"),
		PasswordHash: get("WORKEY_PASSWORD_HASH"),
		SessionKey:   get("WORKEY_SESSION_KEY"),
		Dev:          get("WORKEY_DEV") == "1",
	}

	secure, err := parseBoolDefault("WORKEY_COOKIE_SECURE", get("WORKEY_COOKIE_SECURE"), true)
	if err != nil {
		return nil, err
	}
	cfg.CookieSecure = secure

	examDate, err := time.ParseInLocation(examDateLayout, getDefault("WORKEY_EXAM_DATE", defaultExamDate), time.UTC)
	if err != nil {
		return nil, fmt.Errorf("WORKEY_EXAM_DATE: must be a date in YYYY-MM-DD form")
	}
	cfg.ExamDate = examDate

	if !cfg.Dev {
		var missing []string
		if cfg.APIToken == "" {
			missing = append(missing, "WORKEY_API_TOKEN")
		}
		if cfg.PasswordHash == "" {
			missing = append(missing, "WORKEY_PASSWORD_HASH")
		}
		if cfg.SessionKey == "" {
			missing = append(missing, "WORKEY_SESSION_KEY")
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("missing required variables (set WORKEY_DEV=1 for local development): %s", strings.Join(missing, ", "))
		}
	}

	return cfg, nil
}

func parseBoolDefault(key, v string, def bool) (bool, error) {
	if v == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s: must be a boolean (true or false)", key)
	}
	return b, nil
}
