package web

import (
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/apidoc"
	"github.com/ibra2626/workey/internal/web/views"
)

func TestUsagePagesAndDownload(t *testing.T) {
	a := newTestAuth(t)
	cookie := login(t, a)
	paths := []string{"/kullanim", "/kullanim/mcp", "/kullanim/api", "/kullanim/api/openapi.json", "/kullanim/mcp?view=markdown", "/kullanim/api?view=markdown", "/kullanim/api?view=unknown"}
	for _, client := range []string{"codex-desktop", "codex-cli", "claude-code"} {
		for _, view := range []string{"", "markdown"} {
			paths = append(paths, "/kullanim/mcp?client="+client+"&view="+view)
		}
	}
	for _, path := range paths {
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
				contract := httptest.NewRecorder()
				apidoc.ServeHTTP(contract, httptest.NewRequest("GET", "/api/v1/openapi.json", nil))
				if res.Header.Get("Cache-Control") != "no-store" || string(b) != contract.Body.String() {
					t.Fatal("download must be uncached and match the current contract")
				}
				return
			}
			body := string(b)
			if res.Header.Get("Cache-Control") != "no-store" {
				t.Fatal("usage page must not be cached")
			}
			for _, want := range []string{"MCP Kullanımı", "API Kullanımı", "grid-cols-5", "2026-10-28", "2027-07-18", `href="/kullanim"`, `aria-current="page"`} {
				if !strings.Contains(strings.ReplaceAll(body, `\-`, `-`), want) {
					t.Errorf("missing %q", want)
				}
			}
			if path == "/kullanim/mcp" && (!strings.Contains(body, "bearer_token_env_var") || !strings.Contains(body, "get_agent_guide")) {
				t.Fatal("missing MCP setup")
			}
		})
	}
}

func TestUsageMarkdownViews(t *testing.T) {
	a := newTestAuth(t)
	cookie := login(t, a)
	target := regexp.MustCompile(`(?s)<textarea[^>]*id="usage-markdown"[^>]*>(.*?)</textarea>`)
	for _, key := range []string{"mcp", "api"} {
		want := views.NewUsageDocument(key, a.svc.AgentGuide(a.examDate)).Markdown()
		for _, view := range []string{"", "?view=markdown", "?view=unknown"} {
			r := httptest.NewRequest("GET", "/kullanim/"+key+view, nil)
			r.AddCookie(cookie)
			res := do(t, a, r)
			b, _ := io.ReadAll(res.Body)
			res.Body.Close()
			body := string(b)
			if res.StatusCode != 200 || res.Header.Get("Content-Type") != "text/html; charset=utf-8" || res.Header.Get("Cache-Control") != "no-store" {
				t.Fatalf("invalid response: %v", res)
			}
			match := target.FindStringSubmatch(body)
			if len(match) != 2 || html.UnescapeString(match[1]) != want {
				t.Fatalf("%s%s: copy target differs from full Markdown", key, view)
			}
			panelHidden := regexp.MustCompile(`<div[^>]*id="usage-markdown-panel"[^>]*hidden`).MatchString(body)
			if panelHidden == (view == "?view=markdown") {
				t.Fatalf("%s: wrong raw view visibility", view)
			}
			for _, text := range []string{`readonly`, `aria-live="polite"`, `data-copy-target="usage-markdown"`, `aria-label="Rehber görünümü"`, `/static/copy.js`, `for="usage-markdown"`} {
				if !strings.Contains(body, text) {
					t.Errorf("missing %s", text)
				}
			}
			// API disclosures render only in the guide; unknown views use that default.
			if key == "api" && strings.Contains(body, "<details") == (view == "?view=markdown") {
				t.Fatal("wrong formatted guide visibility")
			}
		}
	}
}

func TestUsageMCPClientSelection(t *testing.T) {
	a := newTestAuth(t)
	cookie := login(t, a)
	raw := regexp.MustCompile(`(?s)<textarea[^>]*id="usage-markdown"[^>]*>(.*?)</textarea>`)
	wantMarkdown := views.NewUsageDocument("mcp", a.svc.AgentGuide(a.examDate)).Markdown()
	for _, value := range []string{"", "invalid", "codex-desktop", "codex-cli", "claude-code", "%3Cscript%3E"} {
		selected := "codex-desktop"
		if value == "codex-cli" || value == "claude-code" {
			selected = value
		}
		for _, view := range []string{"", "markdown", "unknown"} {
			t.Run(value+"/"+view, func(t *testing.T) {
				r := httptest.NewRequest("GET", "/kullanim/mcp?client="+value+"&view="+view, nil)
				r.AddCookie(cookie)
				res := do(t, a, r)
				b, _ := io.ReadAll(res.Body)
				res.Body.Close()
				body := string(b)
				match := raw.FindStringSubmatch(body)
				if res.StatusCode != 200 || len(match) != 2 || html.UnescapeString(match[1]) != wantMarkdown {
					t.Fatal("full Markdown changed with selection")
				}
				body = raw.ReplaceAllString(body, "")
				for _, client := range []string{"codex-desktop", "codex-cli", "claude-code"} {
					has := strings.Contains(body, `data-client="`+client+`"`)
					if has != (view != "markdown" && client == selected) {
						t.Errorf("wrong setup visibility for %s", client)
					}
				}
				for _, link := range []string{"?client=" + selected + `"`, "?client=" + selected + `&amp;view=markdown"`} {
					if !strings.Contains(body, link) {
						t.Errorf("selection lost in view link: %s", link)
					}
				}
				if view != "markdown" {
					for _, want := range []string{"1. Workey", "2. İstemcinizi", "3. Bağlantıyı", "4. Sınava kadar çalışma planımı oluştur", `data-copy-target="usage-prompt-exam-plan"`, `data-copy-target="usage-prompt-connection"`, "Workey MCP bağlantısını kullanarak ajan rehberini oku ve bağlantıyı doğrula"} {
						if !strings.Contains(body, want) {
							t.Errorf("missing %q", want)
						}
					}
					if regexp.MustCompile(`<details[^>]*\bopen\b`).MatchString(body) {
						t.Fatal("technical details must start closed")
					}
				}
				if len(res.Cookies()) != 0 {
					t.Fatal("client selection must not set a preference cookie")
				}
			})
		}
	}
}
