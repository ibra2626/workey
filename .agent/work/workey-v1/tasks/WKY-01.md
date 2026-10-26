# WKY-01 — Go skeleton, config, HTTP server

## Goal
A buildable single binary that loads config from the environment and serves `/healthz`, with graceful shutdown. Foundation for all later tasks.

## Scope
- Precondition: Go toolchain is go1.27.3 (latest stable as of 2026-10-26; installed was 1.27.0). Re-check `https://go.dev/VERSION?m=text` at session start; if a newer 1.27.x exists use it. If the system Go is still older, ask the user to upgrade (e.g. the go.dev MSI or `winget upgrade GoLang.Go`) or confirm that `GOTOOLCHAIN=auto` may download it.
- `go.mod` (module path: ask the user, default `github.com/ibra2626/workey`) with `go 1.27.0` and `toolchain go1.27.3`.
- `cmd/workey/main.go`: subcommand `serve` (default); `hash-password` stub left for WKY-08.
- `internal/config`: `WORKEY_ADDR` (default `127.0.0.1:8080`), `WORKEY_DB_PATH`, `WORKEY_API_TOKEN`, `WORKEY_PASSWORD_HASH`, `WORKEY_SESSION_KEY`, `WORKEY_EXAM_DATE` (default `2027-07-18`), `WORKEY_COOKIE_SECURE` (default true). Validation errors name the variable, never print its value. Missing token/hash/key → refuse to start unless `WORKEY_DEV=1`.
- `internal/clock`: `Clock` interface (`Now()`), Istanbul location loaded once (import `time/tzdata` so the binary works without system tzdata), `Today()` returning a civil date, week start (Monday), fake clock for tests.
- `net/http` server with `ReadHeaderTimeout`, `signal.NotifyContext` graceful shutdown, `GET /healthz`.
- `.env.example` (placeholders only); `.gitignore` adds `.env`, `/bin/`, `*.db`, `*.db-*`.
- `CLAUDE.md` "Checks" section updated with the commands below.
- Non-goals: DB, templates, auth logic.

## Acceptance criteria
- [x] `go build ./...` and `go vet ./...` clean; `gofmt -l .` empty.
- [x] Config tests: defaults, missing required vars, dev mode; no secret values in error strings.
- [x] Clock tests: 2026-10-25 21:30 UTC is 2026-10-26 in Istanbul (UTC+3, no DST); week start of 2026-10-28 is 2026-10-26.
- [x] `/healthz` test via `httptest` returns 200.
- [x] `go version` reports go1.27.3 (or the newer patch found at session start).
- [x] DECISIONS entry: module path, Go version/toolchain line, `time/tzdata` import.

## Checks
- `go build ./...`
- `go vet ./...`
- `go test ./internal/config/... ./internal/clock/... ./cmd/...`
- `gofmt -l .`

## Progress
Done (2026-10-26). System Go upgraded to go1.27.3 by the user; latest stable confirmed go1.27.3 at session start. Module path `github.com/ibra2626/workey` (user-confirmed).

Created:
- `go.mod` (`go 1.27.0` + `toolchain go1.27.3`), `.gitignore` (+ `.env`, `/bin/`, `*.db`, `*.db-*`), `.env.example` (placeholders only).
- `internal/config` — env parsing + validation; errors name the variable, never the value; required token/hash/key unless `WORKEY_DEV=1`. Tests: defaults, missing-required, dev mode, no-value-leak, invalid exam date.
- `internal/clock` — `Clock`/`System`/`Fake`, civil `Date`, `Today`, `WeekStart` (Monday); Europe/Istanbul loaded once with embedded `time/tzdata`. Tests cover the two acceptance cases plus Monday/Sunday edges and `Date.String`.
- `cmd/workey/main.go` — `serve` (default) with `ReadHeaderTimeout`, `signal.NotifyContext` graceful shutdown, `GET /healthz`; `hash-password` stub for WKY-08. `httptest` test on `/healthz`.
- `CLAUDE.md` Checks section filled; DECISIONS entry added.

Checks (all clean): `go build ./...`, `go vet ./...`, `go test ./internal/config/... ./internal/clock/... ./cmd/...` (3 ok), `gofmt -l .` empty, `go version` → go1.27.3. No commit made (awaiting explicit request).
