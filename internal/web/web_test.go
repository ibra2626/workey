package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/store"
)

// testPassword and its hash are used across the auth tests. The hash is
// recomputed at test time so the test never hard-codes a stored value.
const testPassword = "doğru-parola-42"

// testExamDate anchors the screens' exam countdown in the web tests.
var testExamDate = clock.Date{Year: 2027, Month: 7, Day: 18}

// newService opens a migrated store on a temp database and wraps it in a plan
// service whose clock is fixed, so the protected screens have real read models
// to render.
func newService(t *testing.T, now time.Time) *plan.Service {
	t.Helper()
	path := filepath.Join(t.TempDir(), "workey.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return plan.New(st, clock.Fake{Current: now})
}

func newTestAuth(t *testing.T) *Auth {
	t.Helper()
	hash, err := HashPassword(testPassword)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	svc := newService(t, time.Date(2026, 10, 28, 9, 0, 0, 0, clock.Location))
	a, err := NewAuth(svc, testExamDate, hash, "test-session-key-0123456789abcdef", true, false)
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	return a
}

// do runs one request against the auth handler and returns the response.
func do(t *testing.T, a *Auth, req *http.Request) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	a.Handler().ServeHTTP(rec, req)
	return rec.Result()
}

// login performs a successful login and returns the session cookie.
func login(t *testing.T, a *Auth) *http.Cookie {
	t.Helper()
	form := strings.NewReader("parola=" + testPassword)
	req := httptest.NewRequest(http.MethodPost, "/giris", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := do(t, a, req)
	defer res.Body.Close()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("login status = %d, want %d", res.StatusCode, http.StatusSeeOther)
	}
	for _, c := range res.Cookies() {
		if c.Name == sessionCookie {
			return c
		}
	}
	t.Fatal("login set no session cookie")
	return nil
}

func TestLoginPageIsPublicAndTurkish(t *testing.T) {
	a := newTestAuth(t)
	res := do(t, a, httptest.NewRequest(http.MethodGet, "/giris", nil))
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /giris status = %d, want 200", res.StatusCode)
	}
	b, _ := io.ReadAll(res.Body)
	body := string(b)
	if !strings.Contains(body, `<html lang="tr">`) {
		t.Error(`login page is not lang="tr"`)
	}
	if !strings.Contains(body, "Giriş yap") {
		t.Error("login page missing the Giriş yap action")
	}
}

func TestUnauthenticatedPageRedirectsToGiris(t *testing.T) {
	a := newTestAuth(t)
	for _, path := range []string{"/", "/hafta", "/ilerleme"} {
		res := do(t, a, httptest.NewRequest(http.MethodGet, path, nil))
		res.Body.Close()
		if res.StatusCode != http.StatusSeeOther {
			t.Errorf("GET %s status = %d, want 303", path, res.StatusCode)
		}
		if loc := res.Header.Get("Location"); loc != "/giris" {
			t.Errorf("GET %s redirect = %q, want /giris", path, loc)
		}
	}
}

func TestWrongPasswordRejected(t *testing.T) {
	a := newTestAuth(t)
	form := strings.NewReader("parola=yanlış")
	req := httptest.NewRequest(http.MethodPost, "/giris", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := do(t, a, req)
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password status = %d, want 401", res.StatusCode)
	}
	if len(res.Cookies()) != 0 {
		t.Error("wrong password set a cookie")
	}
	b, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(b), "Parola hatalı") {
		t.Error("wrong password did not show the error sentence")
	}
}

func TestRightPasswordGrantsAccess(t *testing.T) {
	a := newTestAuth(t)
	cookie := login(t, a)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	res := do(t, a, req)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("authenticated GET / status = %d, want 200", res.StatusCode)
	}
	b, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(b), "hx-headers") || !strings.Contains(string(b), csrfHeader) {
		t.Error("authenticated page did not inject the CSRF token via hx-headers")
	}
}

func TestSessionCookieFlags(t *testing.T) {
	a := newTestAuth(t)
	c := login(t, a)
	if !c.HttpOnly {
		t.Error("session cookie is not HttpOnly")
	}
	if !c.Secure {
		t.Error("session cookie is not Secure")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("session cookie SameSite = %v, want Lax", c.SameSite)
	}
	if c.Path != "/" {
		t.Errorf("session cookie Path = %q, want /", c.Path)
	}
	if c.MaxAge <= 0 {
		t.Errorf("session cookie MaxAge = %d, want > 0", c.MaxAge)
	}
}

func TestExpiredCookieRejected(t *testing.T) {
	a := newTestAuth(t)
	sid := []byte("0123456789abcdef")
	value := a.encodeSession(sid, time.Now().Add(-time.Minute))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: value})
	res := do(t, a, req)
	res.Body.Close()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("expired cookie status = %d, want redirect 303", res.StatusCode)
	}
}

func TestTamperedCookieRejected(t *testing.T) {
	a := newTestAuth(t)
	c := login(t, a)
	// Flip the last character of the signed value.
	tampered := c.Value[:len(c.Value)-1]
	if strings.HasSuffix(c.Value, "A") {
		tampered += "B"
	} else {
		tampered += "A"
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: tampered})
	res := do(t, a, req)
	res.Body.Close()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("tampered cookie status = %d, want redirect 303", res.StatusCode)
	}
}

func TestLogoutRequiresCSRFToken(t *testing.T) {
	a := newTestAuth(t)
	cookie := login(t, a)

	// Without the CSRF header: 403.
	req := httptest.NewRequest(http.MethodPost, "/cikis", nil)
	req.AddCookie(cookie)
	res := do(t, a, req)
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("POST /cikis without CSRF status = %d, want 403", res.StatusCode)
	}

	// With the correct token: logout clears the cookie.
	sid, ok := a.decodeSession(cookie.Value)
	if !ok {
		t.Fatal("could not decode the session cookie")
	}
	req = httptest.NewRequest(http.MethodPost, "/cikis", nil)
	req.AddCookie(cookie)
	req.Header.Set(csrfHeader, a.csrfToken(sid))
	res = do(t, a, req)
	res.Body.Close()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("POST /cikis with CSRF status = %d, want 303", res.StatusCode)
	}
	var cleared bool
	for _, c := range res.Cookies() {
		if c.Name == sessionCookie && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("logout did not clear the session cookie")
	}
}

func TestStaticAssetIsPublicWithCache(t *testing.T) {
	a := newTestAuth(t)
	res := do(t, a, httptest.NewRequest(http.MethodGet, "/static/app.css", nil))
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /static/app.css status = %d, want 200", res.StatusCode)
	}
	if cc := res.Header.Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("Cache-Control = %q, want immutable", cc)
	}
}

func TestDevOpenBypassesLogin(t *testing.T) {
	svc := newService(t, time.Date(2026, 10, 28, 9, 0, 0, 0, clock.Location))
	a, err := NewAuth(svc, testExamDate, "", "", false, true) // dev, no password hash
	if err != nil {
		t.Fatalf("NewAuth dev: %v", err)
	}
	if !a.devOpen {
		t.Fatal("expected devOpen in dev mode without a password hash")
	}
	res := do(t, a, httptest.NewRequest(http.MethodGet, "/", nil))
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("dev-open GET / status = %d, want 200 (no redirect)", res.StatusCode)
	}
}
