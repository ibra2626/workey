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

	"github.com/ibra2626/workey/internal/config"
	web "github.com/ibra2626/workey/internal/web"
)

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

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           newMux(),
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
// testable with httptest.
func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	web.AddRoutes(mux)
	return mux
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
