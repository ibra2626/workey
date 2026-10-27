# workey

## Project rules

- **State of the repo:** no application code and no commits yet (verified 2026-10-26). Product truth is in `PRODUCT.md`, the locked visual system in `DESIGN.md` (+ `.impeccable/design.json`); the only UI artifact is `design/directions/04-antrenman-gunlugu.html`.
- **Stack (user-decided, not yet scaffolded):** Go + templ + htmx, Tailwind CSS via the standalone CLI, no JS framework. Server-rendered HTML. The plan is written by an AI assistant through an MCP/API interface; the app does not generate study content.
- **UI language Turkish:** ş ğ ı İ ç ö ü must render; never rely on CSS uppercase for Turkish; notation `M2·17, s. 210–225`, decimal comma.
- **Design:** follow `DESIGN.md`; tints are explicit soft tokens (no `/35` opacity on CSS-variable colors).
- **Checks (run only what the change touches):**
  - `go tool templ generate` (regenerate `*_templ.go` after editing any `.templ`; the generated files are committed)
  - `./bin/tailwindcss.exe -i web/styles/input.css -o web/static/app.css --minify` (rebuild CSS after a class or token change; `web/static/app.css` is embedded so it must exist before `go build`)
  - `go build ./...` (and `CGO_ENABLED=0 go build ./...` — the SQLite driver `modernc.org/sqlite` is pure Go, so the single binary must build with CGO off)
  - `go vet ./...`
  - `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...` (v1 gate; report findings explicitly; installed/verified in WKY-14)
  - `go test ./internal/config/... ./internal/clock/... ./internal/store/... ./internal/web/... ./cmd/...` (or the focused package)
  - `gofmt -l .` (must print nothing; generated `*_templ.go` are already gofmt-clean)
  - `scripts/build.sh` runs templ generate → Tailwind → `go build` in order.
  - Toolchain: `go.mod` pins `toolchain go1.27.3`; build/test with 1.27.3. templ (`go tool templ`) and the Tailwind standalone CLI (`./bin/tailwindcss.exe`, gitignored) are installed; `govulncheck` lands in a later task.
- **Commits:** no convention established yet; commit/push/PR only on explicit request.

## Agent state

The global claude-kit rules (`~/.claude/CLAUDE.md`) apply. In this project:

- `.agent/state/CURRENT.md` is the working snapshot (at most 8,000 characters, enforced by `.claude/hooks/guard.mjs`); it names the active task as `Active task: .agent/work/<feature>/tasks/<ID>.md`. Accepted decisions go to `.agent/state/DECISIONS.md` with date, context, decision, alternatives and consequences.
- Long work lives in `.agent/work/<feature>/`: `TASKS.md` is a one-line-per-task index, `tasks/<ID>.md` holds each task. Read the index, then only the task you need.
- Without the global kit: work solo, run only focused checks, and commit/push/PR only on explicit request with exact paths and no force push.
