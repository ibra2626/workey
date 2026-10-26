// Package web holds the server-rendered HTTP handlers. Handlers only parse the
// request, call the plan service and render templ views; no business rule lives
// here (see .agent/work/workey-v1/README.md invariants).
package web

import (
	"net/http"

	"github.com/ibra2626/workey/internal/web/views"
	assets "github.com/ibra2626/workey/web"
)

// AddRoutes registers the web UI and static-asset routes on mux.
func AddRoutes(mux *http.ServeMux) {
	mux.Handle("GET /static/", assets.StaticHandler())
	mux.HandleFunc("GET /{$}", page("bugun", "Workey · Bugün", "Bugün"))
	mux.HandleFunc("GET /hafta", page("hafta", "Workey · Hafta", "Hafta"))
	mux.HandleFunc("GET /ilerleme", page("ilerleme", "Workey · İlerleme", "İlerleme"))
}

// page renders the placeholder view for one section. Replaced by real screens
// in WKY-03 onwards.
func page(active, title, heading string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = views.Placeholder(active, title, heading).Render(r.Context(), w)
	}
}
