package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestMux() *http.ServeMux {
	mux := http.NewServeMux()
	AddRoutes(mux)
	return mux
}

func getBody(t *testing.T, path string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	newTestMux().ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status = %d, want %d", path, res.StatusCode, http.StatusOK)
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

func TestLayoutIsTurkish(t *testing.T) {
	body := getBody(t, "/")
	if !strings.Contains(body, `<html lang="tr">`) {
		t.Errorf(`response does not declare lang="tr"`)
	}
}

func TestActiveTabHasAriaCurrent(t *testing.T) {
	// The active section is marked with aria-current="page" and no other link
	// is. The marker appears twice per page: once in the desktop rail, once in
	// the mobile tab bar.
	cases := []struct {
		path  string
		label string
	}{
		{"/", "Bugün"},
		{"/hafta", "Hafta"},
		{"/ilerleme", "İlerleme"},
	}
	for _, c := range cases {
		body := getBody(t, c.path)
		if got := strings.Count(body, `aria-current="page"`); got != 2 {
			t.Errorf("GET %s: aria-current count = %d, want 2 (rail + tab bar)", c.path, got)
		}
		// The marked link must be the one whose label matches the section.
		marker := `aria-current="page"`
		idx := strings.Index(body, marker)
		if idx < 0 {
			t.Fatalf("GET %s: no aria-current link", c.path)
		}
		// The label text follows shortly after the opening anchor tag.
		after := body[idx:]
		end := strings.Index(after, "</a>")
		if end < 0 || !strings.Contains(after[:end], c.label) {
			t.Errorf("GET %s: active link is not %q", c.path, c.label)
		}
	}
}

func TestStaticAssetServedWithCache(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
	rec := httptest.NewRecorder()
	newTestMux().ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /static/app.css: status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if cc := res.Header.Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("Cache-Control = %q, want immutable", cc)
	}
}
