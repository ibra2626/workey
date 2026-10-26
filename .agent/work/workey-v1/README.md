# workey-v1 — shared rules

Read once per session together with your task file. Product truth: `PRODUCT.md`; visuals: `DESIGN.md`; reference: `design/directions/04-antrenman-gunlugu.html`; v1 decisions: `.agent/state/DECISIONS.md` (2026-10-26 entries override DESIGN.md where they differ: no Seri tile, no deneme panels).

## Invariants (every task)
- One business-logic layer: `internal/plan` (service). Web, REST and MCP only parse input, call the service, render output. No rule is implemented twice.
- Plan save is idempotent: the same request twice leaves the same state and adds no duplicate rows or audit entries.
- Completed sessions and sessions dated before today are frozen for plan writes.
- Every change (plan write, outline write, complete, undo) writes an audit row: time (UTC), source `mcp` | `api` | `web`, action, session id, before/after JSON.
- "Today" and week boundaries (Monday–Sunday) use Europe/Istanbul via an injectable clock; tests never use the wall clock.
- Plan items store only book, chapter number, chapter title, page ranges, antrenman range. No topic text, no question content, no free-text notes.
- Notation is produced by one formatter: `M2·17, s. 210–225` (middle dot U+00B7, en dash U+2013, decimal comma).
- Secrets only from environment; never logged; `.env` ignored, `.env.example` committed.

## Intended layout (adjust only with a DECISIONS entry)
- `cmd/workey/` — main, subcommands (`serve`, `hash-password`).
- `internal/config/` — env parsing.
- `internal/clock/` — Istanbul clock interface + fake.
- `internal/store/` — SQLite (modernc.org/sqlite), embedded SQL migrations.
- `internal/plan/` — domain types, validation, notation, service.
- `internal/httpx/` — shared middleware (bearer auth, JSON errors, security headers).
- `internal/api/` — REST `/api/v1`.
- `internal/mcpserver/` — MCP tools at `/mcp`.
- `internal/web/` — handlers, session auth, CSRF; `internal/web/views/` — templ components.
- `web/static/` — built CSS, `htmx.min.js`, fonts (embedded with `embed`); `web/styles/input.css` — Tailwind source.
- `deploy/` — systemd unit, Nginx example.
- `testdata/` — sample plan fixtures, clearly labelled sample data.

## Conventions
- Go: latest stable is go1.27.3 (go.dev, released 2026-10-19); installed is go1.27.0. `go.mod` uses `go 1.27.0` + `toolchain go1.27.3`; build and govulncheck with 1.27.3. Stdlib `net/http` with method+pattern routing; no web framework.
- Runtime target for v1: local Windows only (WKY-12 deploy is blocked). Keep code OS-neutral (paths via `filepath`, no Unix-only syscalls) so a Linux server stays possible.
- Each new module gets a DECISIONS.md entry (why, alternatives) in the task that adds it.
- Tools not installed yet: templ (add via `go get -tool`), Tailwind standalone CLI, govulncheck (`go run golang.org/x/vuln/cmd/govulncheck@latest` or tool directive). Downloading binaries, fonts or htmx needs the user's explicit OK in that session (state file, source, size).
- Keep `CLAUDE.md` "Checks" current when a check is added.
