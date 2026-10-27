package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUsagePagesAndDownload(t *testing.T) {
	a := newTestAuth(t)
	cookie := login(t, a)
	for _, path := range []string{"/kullanim", "/kullanim/mcp", "/kullanim/api", "/kullanim/api/openapi.json"} {
		t.Run(path, func(t *testing.T) {
			r := httptest.NewRequest("GET", path, nil)
			unauth := do(t, a, r)
			unauth.Body.Close()
			if unauth.StatusCode != http.StatusSeeOther {
				t.Fatalf("unprotected: %d", unauth.StatusCode)
			}
			r.AddCookie(cookie)
			res := do(t, a, r)
			defer res.Body.Close()
			b, _ := io.ReadAll(res.Body)
			if res.StatusCode != 200 {
				t.Fatalf("%d: %s", res.StatusCode, b)
			}
			if strings.HasSuffix(path, ".json") {
				if !strings.Contains(res.Header.Get("Content-Disposition"), "attachment") || !strings.Contains(string(b), `"openapi": "3.1.0"`) {
					t.Fatal("invalid download")
				}
				return
			}
			body := string(b)
			for _, want := range []string{"MCP Kullanımı", "API Kullanımı", "grid-cols-4", "2026-10-28", `href="/kullanim"`, `aria-current="page"`} {
				if !strings.Contains(body, want) {
					t.Errorf("missing %q", want)
				}
			}
			if path == "/kullanim/mcp" && (!strings.Contains(body, "bearer_token_env_var") || !strings.Contains(body, "get_agent_guide")) {
				t.Fatal("missing MCP setup")
			}
		})
	}
}
