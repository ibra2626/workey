# WKY-08 — Web auth: Giriş screen, session cookie, CSRF

## Goal
Single-user login protecting all web pages, with secure cookies and CSRF protection on every state-changing htmx request.

## Scope
- Password hash in `WORKEY_PASSWORD_HASH`. Prefer stdlib (`crypto/pbkdf2` SHA-256, high iteration count, encoded `pbkdf2-sha256$iter$salt$hash`) to avoid a dependency; if bcrypt/argon2 via `golang.org/x/crypto` is chosen instead, record why. `workey hash-password` subcommand reads the password from stdin without echo where possible and prints only the hash.
- Local Windows use (WKY-12 blocked): the app is reached at `http://127.0.0.1:8080`; browsers accept `Secure` cookies on localhost, so keep `Secure` on by default and verify login works over plain http on 127.0.0.1 in Chrome/Edge.
- Session: signed (HMAC-SHA256 with `WORKEY_SESSION_KEY`) or server-side session row; cookie `HttpOnly`, `Secure` (from `WORKEY_COOKIE_SECURE`), `SameSite=Lax` or `Strict` (record choice), `Path=/`, 30-day expiry; logout clears it.
- CSRF: per-session token sent by htmx via `hx-headers` on `<body>` and checked on every non-GET web route; additionally `http.CrossOriginProtection` (Go 1.25+) on the web mux. API/MCP use Bearer only and are not cookie-authenticated.
- Login throttling: small in-memory backoff after repeated failures.
- `Giriş` screen (`internal/web/views/login.templ`): no mockup exists; derive from DESIGN.md — Trail Ground, one white panel, "Workey" wordmark, one password field with label, the dark ink pill as the only action ("Giriş yap"), plain error sentence, no shadows. Mobile 390 first, then desktop centered.
- Non-goals: user management, password reset.

## Acceptance criteria
- [ ] Tests: wrong/right password, cookie flags, expired/tampered cookie rejected, POST without CSRF token → 403, unauthenticated page → redirect to `/giris`.
- [ ] Hash command round trip test.
- [ ] Giriş screenshots at 390 and 1280 checked against DESIGN.md (contrast, focus ring, sentence case).

## Checks
- `go test ./internal/web/... ./cmd/...`
- `go vet ./...`
- Preview screenshots at 390 and 1280.

## Progress
Done (not committed) — 2026-10-26.

- Password hashing: stdlib `crypto/pbkdf2` (SHA-256, 600000 iter, 16-byte salt, 32-byte key), encoded `pbkdf2-sha256$iter$b64salt$b64hash` (`internal/web/password.go`). No bcrypt/argon2 dependency. `workey hash-password` reads from stdin with echo disabled on the Windows console via `golang.org/x/sys/windows` (already a verified dep — `golang.org/x/term` could not be verified offline against sum.golang.org), prints only the hash; non-Windows falls back to a plain line read (`cmd/workey/password_{windows,other}.go`).
- Session: signed cookie (HMAC-SHA256 with `WORKEY_SESSION_KEY`), value `b64url(sid16||expiry8).b64url(mac)`, 30-day expiry, `HttpOnly`, `Secure` (from `WORKEY_COOKIE_SECURE`, default true), `SameSite=Lax`, `Path=/`; logout expires it (`internal/web/auth.go`).
- CSRF: per-session token = `b64url(HMAC(key,"csrf"||sid))`, injected on `<body>` via `hx-headers` (layout gained a `csrf` param), checked on every non-GET web route (`requireCSRF`). Plus `http.CrossOriginProtection` wrapping the whole web mux.
- Throttle: global in-memory backoff after 3 failed logins, growing to a 30s cap.
- Routes: web handler now mounts at `/` (specific `/healthz`, `/api/v1/`, `/mcp` still win); `GET/POST /giris`, `POST /cikis`, protected `GET /{$}|/hafta|/ilerleme`. Static and `/giris` are public. Dev-open mode (`WORKEY_DEV=1` + no hash) bypasses login for UI work.
- `Giriş` screen (`internal/web/views/login.templ`): self-contained page, Trail Ground, one white panel, Workey wordmark, "Parola" label, dark ink pill "Giriş yap", plain error sentence, no shadows, centered mobile-first.

## Verification (Go 1.27.3, 2026-10-26)
- `go test ./internal/web/... ./cmd/...` — pass. Covers: wrong/right password, cookie flags, expired + tampered cookie rejected, POST /cikis without CSRF → 403 (and with token → 303 + cookie cleared), unauthenticated page → 303 /giris, login page public/Turkish, static public, dev-open bypass, hash round trip + fresh salt + malformed-hash errors.
- `go vet ./...`, `gofmt -l .` (empty), `go build ./...`, `CGO_ENABLED=0 go build ./...` — all clean. `go mod tidy` promoted `golang.org/x/sys` to a direct dependency.
- Live on `http://127.0.0.1:8080` (real-auth env, not dev-open): Giriş screenshotted at 390 and 1280 (on-brand, Turkish ş/İ correct, sentence case, dark pill). Login with the correct password redirects to Bugün; `document.cookie` empty (HttpOnly) and `hx-headers` carries the CSRF token; logout via `POST /cikis` returns to Giriş. Secure cookie accepted over plain http on 127.0.0.1.
