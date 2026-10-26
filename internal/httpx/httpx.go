// Package httpx holds transport-level HTTP helpers shared by the REST API
// (internal/api) and the MCP server (internal/mcpserver): a constant-time
// Bearer-token guard, a JSON error envelope, a JSON writer, a request-body size
// limit and security headers. It contains no business rules.
package httpx

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

// FieldDetail is one field-level problem inside an error envelope: a path into
// the request, a stable machine code and a human message. It mirrors
// plan.FieldError without importing the plan package.
type FieldDetail struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error is the body of a failed response: a machine code, a human message and
// optional per-field details.
type Error struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Fields  []FieldDetail `json:"fields,omitempty"`
}

// envelope wraps an Error so every failure serialises as {"error": {...}}.
type envelope struct {
	Error Error `json:"error"`
}

// WriteJSON writes v as JSON with the given status and a UTF-8 content type.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes an Error as the {"error": ...} envelope with the status.
func WriteError(w http.ResponseWriter, status int, e Error) {
	WriteJSON(w, status, envelope{Error: e})
}

// BearerAuth guards next with a Bearer token check. The expected token and the
// presented one are compared as SHA-256 digests with crypto/subtle so neither
// the comparison time nor the token length leaks. An empty expected token, a
// missing or malformed Authorization header, or a mismatch all fail closed with
// 401 and a WWW-Authenticate challenge.
func BearerAuth(expected string, next http.Handler) http.Handler {
	expectedSum := sha256.Sum256([]byte(expected))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		presentedSum := sha256.Sum256([]byte(token))
		ok := expected != "" && token != "" &&
			subtle.ConstantTimeCompare(expectedSum[:], presentedSum[:]) == 1
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			WriteError(w, http.StatusUnauthorized, Error{
				Code:    "unauthorized",
				Message: "a valid Bearer token is required",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// bearerToken extracts the token from an "Authorization: Bearer <token>"
// header, with a case-insensitive scheme. It returns "" when the header is
// missing or not a Bearer credential.
func bearerToken(header string) string {
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

// LimitBody caps each request body at maxBytes before calling next, so a large
// upload cannot exhaust memory. Reads past the limit fail and the handler's
// decode returns an error it maps to 400.
func LimitBody(maxBytes int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders sets conservative headers for a JSON API: no MIME sniffing,
// no referrer leakage and no shared caching of responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
