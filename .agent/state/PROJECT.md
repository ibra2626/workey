# Project map

Bootstrapped 2026-10-26; updated 2026-11-04 to the v2 source-catalog model
(WKY-21..28). Filled from verified repository evidence.

- Product: a general **source-catalog study tracker**. Any user installs it
  locally and studies for any exam with their own books / videos / courses. The
  app is the single source of truth; an AI assistant extracts each source's
  structure and writes the plan through the MCP/API interface (the app never
  generates study content). Truth: `PRODUCT.md`; the v2 direction is the
  `2026-11-02 — v2 kaynak kataloğu` entry in `.agent/state/DECISIONS.md`.
- Concept model: **Source** (kaynak: user-given `key`, name, `kind`
  book/video/course/other, `location_unit` page/minute/lesson/other) → **Section**
  (bölüm) → **Unit** (birim: label is the source's own word — Antrenman/Test/
  Föy/Ders — never hard-coded). Unit progress (done + optional D/Y/B) is stored
  separately from the catalog and is independent of session completion.
- Stack (implemented): Go 1.27.3 (toolchain pinned in `go.mod`) + templ
  (`go tool templ`, generated `*_templ.go` committed) + htmx, SQLite via
  `modernc.org/sqlite` (pure Go; single binary builds with `CGO_ENABLED=0`),
  Tailwind standalone CLI (`./bin/tailwindcss.exe`, gitignored). No JS framework;
  server-rendered HTML with form fallbacks and small local enhancement scripts.
- Interfaces: web UI (Bugün `/`, Hafta `/hafta`, İlerleme `/ilerleme`, Kaynaklar
  correction page, usage pages), REST `/api/v1`, MCP `/mcp`. All three call one
  business layer (`internal/plan`); no rule is implemented twice.
- Main directories:
  - `PRODUCT.md`, `DESIGN.md` (+ `.impeccable/design.json`) — product truth and
    locked design system. `design/directions/04-antrenman-gunlugu.html` — original
    exploration mockup.
  - `cmd/workey/` — entry point (HTTP server, mux). `internal/` — see
    `ARCHITECTURE.md`.
  - `docs/` — `mcp.md`, `api.md`, `study-flow-acceptance.md`,
    `usage-verification.md`, `v2-gate-verification.md` (WKY-28 end-to-end gate).
  - `.claude/launch.json` — preview server.
- Focused verification: `scripts/build.sh` (templ → Tailwind → build); the test
  set and checks in `CLAUDE.md` ## Checks; `govulncheck@v1.9.0 ./...`.
- Constraints: Turkish UI with correct Turkish casing (never CSS uppercase);
  notation `M2·17, s. 210–225` (middle dot U+00B7, en dash U+2013, decimal comma);
  plan content arrives via MCP/API only; mobile-first. The fixed exam date
  (`WORKEY_EXAM_DATE`, default 2027-07-18) and 38-week framing live in plan
  content (phase dates) and per-install config, not in the schema. Secrets only
  from the environment.
