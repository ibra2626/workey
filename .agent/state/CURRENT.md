# CURRENT — WKY-08 done, web auth live

Updated 2026-10-26.

Active task: .agent/work/workey-v1/tasks/WKY-08.md (done; not committed)

- Feature: `workey-v1` — Go + templ + htmx + Tailwind single binary; REST `/api/v1`, MCP `/mcp` and web share the one business-logic layer `internal/plan`; SQLite (modernc). Local Windows only; WKY-12 deploy remains blocked by the user.
- Plan: `.agent/work/workey-v1/TASKS.md`; shared rules: `.agent/work/workey-v1/README.md`. Product: `PRODUCT.md`; visuals: `DESIGN.md`; decisions: `.agent/state/DECISIONS.md`.
- WKY-01..07 committed; HEAD `97ca6d2` (WKY-07). WKY-08 done; this working tree is the WKY-08 change set, not yet committed. No push performed.

## WKY-08 result (web auth: Giriş, session cookie, CSRF)

Single-user login protecting all web pages; REST/MCP stay Bearer-only.

- `internal/web/password.go`: stdlib `crypto/pbkdf2` SHA-256, 600000 iter, 16-byte salt, 32-byte key; encoding `pbkdf2-sha256$iter$b64salt$b64hash`; `HashPassword`/`VerifyPassword` (constant-time).
- `internal/web/auth.go`: `Auth` + `NewAuth(passwordHash, sessionKey, cookieSecure, dev)`. Signed session cookie `b64url(sid16||expiry8).b64url(HMAC)`, `HttpOnly`/`Secure`/`SameSite=Lax`/`Path=/`, 30-day. CSRF token = `b64url(HMAC(key,"csrf"||sid))`; `requireAuth` (redirect unauth GET → `/giris` 303, else 401; stores token in ctx), `requireCSRF` (non-GET needs `X-CSRF-Token`). `throttle` (backoff after 3 fails, 30s cap). Dev-open: `WORKEY_DEV=1` + no hash bypasses login.
- `internal/web/web.go`: `Auth.Handler()` builds a ServeMux wrapped in `http.CrossOriginProtection`; public `GET/POST /giris`, `GET /static/`; protected `GET /{$}|/hafta|/ilerleme`, `POST /cikis`. Mounted at `/` in `cmd/workey` (specific `/healthz`,`/api/v1/`,`/mcp` still win). `newMux` now takes a third `webHandler`.
- `internal/web/views/login.templ`: self-contained Giriş page (Trail Ground, white panel, Workey wordmark, Parola field, dark ink pill "Giriş yap", plain error, no shadows, centered). `layout.templ` Layout gained a `csrf` param → `hx-headers` on `<body>`; `Placeholder` threads it.
- `cmd/workey/main.go`: `hash-password` subcommand reads stdin (no console echo on Windows via `golang.org/x/sys/windows`; `password_windows.go` + `password_other.go` fallback), prints only the hash. Builds `web.NewAuth` from config and mounts its handler.
- Config already had `PasswordHash`/`SessionKey`/`CookieSecure`/`Dev` (WKY-06); unchanged.

## Verification (Go 1.27.3, 2026-10-26, all passed)

- `go test ./internal/web/... ./cmd/...` — pass. `web_test.go`: wrong/right password, cookie flags, expired + tampered cookie rejected, POST /cikis without CSRF → 403 (with token → 303 + cookie cleared), unauth page → 303 /giris, login public/Turkish, static public, dev-open bypass. `password_test.go`: round trip, fresh salt, malformed-hash errors.
- `go vet ./...` 0; `gofmt -l .` empty; `go build ./...` 0; `CGO_ENABLED=0 go build ./...` 0. `go mod tidy` promoted `golang.org/x/sys` to a direct dependency (go.mod/go.sum changed).
- Live at `http://127.0.0.1:8080` with real-auth env (not dev-open): Giriş verified at 390 and 1280 (on-brand, Turkish ş/İ, sentence case, dark pill, no shadows). Correct password → Bugün; `document.cookie` empty (HttpOnly); `hx-headers` carries CSRF; `POST /cikis` → Giriş. Secure cookie accepted over plain http on 127.0.0.1.

## Next step

WKY-09: Bugün screen (mobile + desktop), depends on WKY-05, WKY-06, WKY-08. Start a new session: `WKY-09 görevine başla`; read the index and that task. Screens render behind `requireAuth` and read the CSRF token from the layout; use dev-open (`WORKEY_DEV=1`) for preview.

WKY-08 paths (not yet committed): `internal/web/` (password.go, auth.go, web.go, *_test.go, views/login.templ + generated, layout.templ/pages.templ/views.go + generated), `cmd/workey/` (main.go, main_test.go, password_windows.go, password_other.go), `go.mod`, `go.sum`, `web/static/app.css`, `.agent/` records. Commit/push/PR require an explicit user request.
