package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ibra2626/workey/internal/httpx"
)

func TestBearerAuth(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot) // a marker the guard let through
	})

	cases := []struct {
		name     string
		expected string
		header   string
		want     int
	}{
		{"valid", "secret", "Bearer secret", http.StatusTeapot},
		{"valid lowercase scheme", "secret", "bearer secret", http.StatusTeapot},
		{"wrong token", "secret", "Bearer nope", http.StatusUnauthorized},
		{"missing header", "secret", "", http.StatusUnauthorized},
		{"not bearer", "secret", "Basic secret", http.StatusUnauthorized},
		{"empty expected fails closed", "", "Bearer ", http.StatusUnauthorized},
		{"empty expected rejects anything", "", "Bearer x", http.StatusUnauthorized},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/x", nil)
			if c.header != "" {
				r.Header.Set("Authorization", c.header)
			}
			rec := httptest.NewRecorder()
			httpx.BearerAuth(c.expected, ok).ServeHTTP(rec, r)
			if rec.Code != c.want {
				t.Errorf("status = %d, want %d", rec.Code, c.want)
			}
			if c.want == http.StatusUnauthorized {
				if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
					t.Errorf("WWW-Authenticate = %q, want Bearer", got)
				}
			}
		})
	}
}
