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
Not started.
