package web

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"sync"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
)

// Cookie and header names, and the session lifetime.
const (
	sessionCookie = "workey_session"
	csrfHeader    = "X-CSRF-Token"
	sessionTTL    = 30 * 24 * time.Hour
	sidLen        = 16
)

// devSID is the fixed session id used in dev-open mode (WORKEY_DEV=1 with no
// password hash), so CSRF tokens are stable without a real login.
var devSID = []byte("workey-dev-sid--") // 16 bytes

// csrfCtxKey is the private type for the request-context CSRF token.
type csrfCtxKey struct{}

// Auth holds the web authentication state: the stored password hash, the HMAC
// key for signing session cookies and deriving CSRF tokens, and cookie flags.
// devOpen bypasses login for local development (no password configured).
type Auth struct {
	svc          *plan.Service
	examDate     clock.Date
	passwordHash string
	key          []byte
	cookieSecure bool
	devOpen      bool
	throttle     *throttle
}

// NewAuth builds an Auth from configuration and the plan service the screens
// render. In dev mode with no password hash (devOpen), login is bypassed and an
// ephemeral key is generated when none is set, so the server still starts for
// UI work without secrets configured.
func NewAuth(svc *plan.Service, examDate clock.Date, passwordHash, sessionKey string, cookieSecure, dev bool) (*Auth, error) {
	a := &Auth{
		svc:          svc,
		examDate:     examDate,
		passwordHash: passwordHash,
		cookieSecure: cookieSecure,
		devOpen:      dev && passwordHash == "",
		throttle:     newThrottle(),
	}
	if sessionKey != "" {
		a.key = []byte(sessionKey)
	} else {
		a.key = make([]byte, 32)
		if _, err := rand.Read(a.key); err != nil {
			return nil, err
		}
	}
	return a, nil
}

// issueSession writes a signed session cookie that expires after sessionTTL.
func (a *Auth) issueSession(w http.ResponseWriter) {
	sid := make([]byte, sidLen)
	_, _ = rand.Read(sid)
	expiry := time.Now().Add(sessionTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    a.encodeSession(sid, expiry),
		Path:     "/",
		Expires:  expiry,
		MaxAge:   int(sessionTTL / time.Second),
		HttpOnly: true,
		Secure:   a.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSession expires the session cookie (logout).
func (a *Auth) clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// encodeSession returns base64url(sid||expiry) + "." + base64url(mac). The MAC
// covers both the session id and the expiry, so neither can be altered.
func (a *Auth) encodeSession(sid []byte, expiry time.Time) string {
	payload := make([]byte, sidLen+8)
	copy(payload, sid)
	binary.BigEndian.PutUint64(payload[sidLen:], uint64(expiry.Unix()))
	mac := a.mac(payload)
	enc := base64.RawURLEncoding
	return enc.EncodeToString(payload) + "." + enc.EncodeToString(mac)
}

// decodeSession verifies the cookie value and returns the session id when the
// signature is valid and the session has not expired.
func (a *Auth) decodeSession(value string) (sid []byte, ok bool) {
	enc := base64.RawURLEncoding
	dot := -1
	for i := 0; i < len(value); i++ {
		if value[i] == '.' {
			dot = i
			break
		}
	}
	if dot < 0 {
		return nil, false
	}
	payload, err := enc.DecodeString(value[:dot])
	if err != nil || len(payload) != sidLen+8 {
		return nil, false
	}
	gotMAC, err := enc.DecodeString(value[dot+1:])
	if err != nil {
		return nil, false
	}
	if subtle.ConstantTimeCompare(gotMAC, a.mac(payload)) != 1 {
		return nil, false
	}
	expiry := int64(binary.BigEndian.Uint64(payload[sidLen:]))
	if time.Now().Unix() >= expiry {
		return nil, false
	}
	return payload[:sidLen], true
}

// mac returns HMAC-SHA256(key, msg).
func (a *Auth) mac(msg []byte) []byte {
	h := hmac.New(sha256.New, a.key)
	h.Write(msg)
	return h.Sum(nil)
}

// csrfToken derives the per-session CSRF token from the session id. It is a
// keyed HMAC, so a client cannot forge one without a valid session, and it is
// stable for the life of the session (no server-side storage needed).
func (a *Auth) csrfToken(sid []byte) string {
	h := hmac.New(sha256.New, a.key)
	h.Write([]byte("csrf"))
	h.Write(sid)
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// csrfTokenFromRequest returns the CSRF token stored in the request context by
// requireAuth, or "" when the request did not pass through it.
func csrfTokenFromRequest(r *http.Request) string {
	tok, _ := r.Context().Value(csrfCtxKey{}).(string)
	return tok
}

// withCSRF returns r with the session's CSRF token in its context.
func withCSRF(r *http.Request, token string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), csrfCtxKey{}, token))
}

// requireAuth wraps a protected handler: a request with a valid session (or any
// request in dev-open mode) proceeds with its CSRF token in context; otherwise
// a page GET is redirected to /giris and anything else gets 401.
func (a *Auth) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.devOpen {
			next.ServeHTTP(w, withCSRF(r, a.csrfToken(devSID)))
			return
		}
		c, err := r.Cookie(sessionCookie)
		if err == nil {
			if sid, ok := a.decodeSession(c.Value); ok {
				next.ServeHTTP(w, withCSRF(r, a.csrfToken(sid)))
				return
			}
		}
		if r.Method == http.MethodGet {
			http.Redirect(w, r, "/giris", http.StatusSeeOther)
			return
		}
		http.Error(w, "oturum gerekli", http.StatusUnauthorized)
	})
}

// requireCSRF wraps a state-changing handler and rejects it unless the request
// carries the session's CSRF token in the X-CSRF-Token header. Safe methods
// pass through unchecked. Must sit inside requireAuth so the session is known.
func (a *Auth) requireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		want := csrfTokenFromRequest(r)
		got := r.Header.Get(csrfHeader)
		if got == "" {
			// No-JS fallback: a plain form POST carries the token in a hidden
			// field instead of the htmx header.
			got = r.PostFormValue("_csrf")
		}
		if want == "" || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			http.Error(w, "csrf doğrulaması başarısız", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// throttle applies a small in-memory backoff after repeated failed logins. For
// a single-user local app one global counter is enough; it is not a defence
// against a distributed attacker, only against rapid guessing.
type throttle struct {
	mu       sync.Mutex
	fails    int
	nextTime time.Time
	now      func() time.Time
}

func newThrottle() *throttle {
	return &throttle{now: time.Now}
}

// retryAfter reports how long the caller must wait before another attempt is
// allowed, or zero when an attempt may proceed now.
func (t *throttle) retryAfter() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	if d := t.nextTime.Sub(t.now()); d > 0 {
		return d
	}
	return 0
}

// fail records a failed attempt and arms the backoff: after 3 failures each
// further attempt waits, growing to a 30s cap.
func (t *throttle) fail() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fails++
	if t.fails < 3 {
		return
	}
	backoff := time.Duration(t.fails-2) * time.Second
	if backoff > 30*time.Second {
		backoff = 30 * time.Second
	}
	t.nextTime = t.now().Add(backoff)
}

// success clears the backoff after a correct password.
func (t *throttle) success() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fails = 0
	t.nextTime = time.Time{}
}
