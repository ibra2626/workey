// Command workey runs the study-plan server.
//
// Usage:
//
//	workey [serve]        start the HTTP server (default)
//	workey hash-password  print a bcrypt hash of a password (added in WKY-08)
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
		// Implemented in WKY-08 (single-user web auth).
		fmt.Fprintln(os.Stderr, "workey: hash-password is not implemented yet")
		os.Exit(1)
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

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           newMux(apiHandler, mcpHandler),
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
// testable with httptest. apiHandler and mcpHandler may be nil (e.g. in tests
// that exercise only the always-on routes).
func newMux(apiHandler, mcpHandler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	if apiHandler != nil {
		mux.Handle("/api/v1/", apiHandler)
	}
	if mcpHandler != nil {
		mux.Handle("/mcp", mcpHandler)
	}
	web.AddRoutes(mux)
	return mux
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
