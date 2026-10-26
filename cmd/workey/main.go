// Command workey runs the study-plan server.
//
// Usage:
//
//	workey [serve]        start the HTTP server (default)
//	workey hash-password  read a password from stdin and print its PBKDF2 hash
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ibra2626/workey/internal/api"
	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/config"
	"github.com/ibra2626/workey/internal/mcpserver"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/store"
	web "github.com/ibra2626/workey/internal/web"
)

// defaultDBPath is used when WORKEY_DB_PATH is unset, so a local run works out
// of the box.
const defaultDBPath = "workey.db"

func main() {
	cmd := "serve"
	args := os.Args[1:]
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "serve":
		if err := runServe(); err != nil {
			fmt.Fprintln(os.Stderr, "workey:", err)
			os.Exit(1)
		}
	case "hash-password":
		if err := runHashPassword(); err != nil {
			fmt.Fprintln(os.Stderr, "workey:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "workey: unknown command %q (use \"serve\" or \"hash-password\")\n", cmd)
		os.Exit(2)
	}
}

func runServe() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = defaultDBPath
	}
	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	if err := st.Migrate(context.Background()); err != nil {
		return err
	}

	svc := plan.New(st, clock.System{})
	examDate := clock.Date{Year: cfg.ExamDate.Year(), Month: cfg.ExamDate.Month(), Day: cfg.ExamDate.Day()}
	apiHandler := api.New(svc, examDate, cfg.APIToken)
	mcpHandler, err := mcpserver.New(svc, examDate, cfg.APIToken)
	if err != nil {
		return err
	}
	auth, err := web.NewAuth(svc, examDate, cfg.PasswordHash, cfg.SessionKey, cfg.CookieSecure, cfg.Dev)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           newMux(apiHandler, mcpHandler, auth.Handler()),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		fmt.Printf("workey: listening on %s\n", cfg.Addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		stop() // restore default signal handling for a second Ctrl-C
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

// newMux builds the HTTP router. Kept separate from runServe so handlers are
// testable with httptest. Any handler may be nil (e.g. in tests that exercise
// only the always-on routes). The web handler is mounted at "/"; the more
// specific /healthz, /api/v1/ and /mcp patterns still take precedence.
func newMux(apiHandler, mcpHandler, webHandler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	if apiHandler != nil {
		mux.Handle("/api/v1/", apiHandler)
	}
	if mcpHandler != nil {
		mux.Handle("/mcp", mcpHandler)
	}
	if webHandler != nil {
		mux.Handle("/", webHandler)
	}
	return mux
}

// runHashPassword reads a password from stdin (without terminal echo when stdin
// is a terminal) and prints only its PBKDF2 hash, for WORKEY_PASSWORD_HASH.
func runHashPassword() error {
	password, err := readPassword()
	if err != nil {
		return err
	}
	if password == "" {
		return errors.New("empty password")
	}
	hash, err := web.HashPassword(password)
	if err != nil {
		return err
	}
	fmt.Println(hash)
	return nil
}

// readPassword reads one password. On a console it disables echo and prompts on
// stderr (readConsolePassword); otherwise, or when the platform cannot suppress
// echo, it reads a single line from stdin (e.g. a pipe).
func readPassword() (string, error) {
	if pw, ok, err := readConsolePassword(); ok {
		return pw, err
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
