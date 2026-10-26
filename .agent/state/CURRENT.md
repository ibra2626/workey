# CURRENT — WKY-01 done, skeleton builds

Updated 2026-10-26.

Active task: .agent/work/workey-v1/tasks/WKY-02.md (next; not started)

- Feature: `workey-v1` — Go + templ + htmx + Tailwind single binary; Giriş / Bugün / Hafta / İlerleme; REST `/api/v1` and MCP `/mcp` over one `internal/plan` service; SQLite (modernc). Local Windows only; WKY-12 (deploy) blocked by the user.
- Plan: `.agent/work/workey-v1/TASKS.md`; shared invariants + layout in `.agent/work/workey-v1/README.md` (read with each task).

## WKY-01 result (done)
- System Go upgraded to go1.27.3 (user). Module path `github.com/ibra2626/workey` (user-confirmed).
- Created: `go.mod` (`go 1.27.0` + `toolchain go1.27.3`), `.gitignore` (+ `.env`, `/bin/`, `*.db`, `*.db-*`), `.env.example` (placeholders).
  - `internal/config` — env parse/validate; errors name the variable, never the value; token/hash/key required unless `WORKEY_DEV=1`.
  - `internal/clock` — `Clock`/`System`/`Fake`, civil `Date`, `Today`, `WeekStart` (Monday); Europe/Istanbul via embedded `time/tzdata`.
  - `cmd/workey/main.go` — `serve` default (ReadHeaderTimeout, signal.NotifyContext graceful shutdown, `GET /healthz`); `hash-password` stub for WKY-08.
  - `CLAUDE.md` Checks section filled; DECISIONS entry added (module path, tzdata, Go upgrade).
- Checks all clean: `go build ./...`, `go vet ./...`, `go test ./internal/config/... ./internal/clock/... ./cmd/...` (3 pkg ok), `gofmt -l .` empty, `go version` → go1.27.3.
- Nothing committed (commit only on explicit request). Tools still not installed: templ, Tailwind standalone CLI, govulncheck.

## Next step
New session → `WKY-02 görevine başla` (frontend toolchain: templ, Tailwind, fonts, htmx, base layout). WKY-02 will need the user's OK to download Tailwind CLI, htmx and fonts (source, size).
