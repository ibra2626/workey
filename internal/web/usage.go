package web

import (
	"github.com/ibra2626/workey/internal/web/views"
	"net/http"
)

func (a *Auth) screenUsage(w http.ResponseWriter, r *http.Request) {
	key := "kullanim"
	if r.URL.Path == "/kullanim/mcp" {
		key = "mcp"
	}
	if r.URL.Path == "/kullanim/api" {
		key = "api"
	}
	doc := views.NewUsageDocument(key, a.svc.AgentGuide(a.examDate))
	if key == "mcp" {
		doc = doc.SelectClient(r.URL.Query().Get("client"))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = views.Usage(key, csrfTokenFromRequest(r), doc, r.URL.Query().Get("view") == "markdown").Render(r.Context(), w)
}
