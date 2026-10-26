# Architectural Decisions

Record accepted decisions with date, context, decision, alternatives, and consequences. Do not record proposals as accepted decisions.

## 2026-10-26 — Stack: Go + templ + htmx + Tailwind standalone CLI
- Context: greenfield single-user web app; user is a web developer.
- Decision: server-rendered Go with templ, htmx for interactivity, Tailwind via standalone CLI, no JS framework (user's choice during `/impeccable init`).
- Alternatives: static HTML/CSS, Next.js/React, Astro (offered, not chosen).
- Consequences: Tailwind opacity modifiers on CSS-variable colors do not compile; use explicit soft tokens (see DESIGN.md Soft Token Rule).

## 2026-10-26 — Visual direction locked: "Antrenman günlüğü" (Yön 04)
- Context: six directions explored (notebook, Swiss grid, night focus, training log, two hybrids).
- Decision: user chose Yön 04; recorded in `DESIGN.md` and `.impeccable/design.json`; reference artifact `design/directions/04-antrenman-gunlugu.html`. Other directions and review screenshots deleted.
- Alternatives: 01–03, 05, 06 (deleted, not recoverable — never committed).
- Consequences: teal z1 retuned #5b9c93 → #66a69d so ink labels pass 4.5:1; focus rings use Forest Ink.

## 2026-10-26 — v1 runs locally on Windows; deploy task on hold
- Context: the brief asked for systemd + Nginx deploy files; the user will run v1 only locally on Windows for now.
- Decision (user): WKY-12 (Linux build, systemd, Nginx) is blocked until the user lifts it; a Linux server may come later. The v1 gate does not include deploy.
- Alternatives: write the deploy files now without using them.
- Consequences: default bind stays `127.0.0.1:8080`; code stays OS-neutral. Phone access is not possible with this setup unless the app is later exposed on the LAN or deployed; that needs a separate decision (TLS, Secure cookies).

## 2026-10-26 — Go toolchain: go1.27.3
- Context: installed Go is 1.27.0; go.dev reports go1.27.3 as latest stable (2026-10-19). Patch releases carry security fixes that govulncheck would report against stdlib.
- Decision: target go1.27.3 (`go 1.27.0` + `toolchain go1.27.3` in go.mod); upgrade the system Go before or at WKY-01.
- Alternatives: stay on 1.27.0 and accept stdlib govulncheck findings.

## 2026-10-26 — WKY-01: module path, tzdata embed, Go upgrade done
- Context: scaffolding the Go skeleton; needed a module path, a timezone source that does not depend on the host, and the pinned toolchain in place.
- Decision: module path `github.com/ibra2626/workey` (user-confirmed, matches git user ibra2626). `internal/clock` imports `_ "time/tzdata"` so Europe/Istanbul loads from the embedded tzdata and the single binary works on a machine without a system tz database. System Go upgraded to go1.27.3 (user did it this session); `go.mod` is `go 1.27.0` + `toolchain go1.27.3`.
- Alternatives: bare module name `workey` (not future-proof for a remote repo); rely on host tzdata (breaks on minimal systems); keep system Go at 1.27.0 and let GOTOOLCHAIN=auto download (user preferred upgrading outright).
- Consequences: all internal packages import under `github.com/ibra2626/workey/...`. tzdata adds a small amount to the binary; acceptable for a self-contained local app.

## 2026-10-26 — v1 scope: no streak, no deneme tracking
- Context: the mockup shows a "Seri" tile, a "Seviye testi 8,5 / 25" panel, a "Denemeler · net" chart and an "H4 · deneme" marker; the v1 brief puts question/deneme tracking and streak-style gamification out of scope.
- Decision (user, during `/task-plan workey-v1`): remove the Seri tile, the seviye testi panel and the Denemeler chart. The third mobile Bugün stat tile becomes "Seans 1 / 5" (done / planned sessions this week). A deneme day may exist only as a planned session of kind `deneme`; no results are stored.
- Alternatives: remove everything with two stat tiles; keep Seri as a pure measurement.
- Consequences: DESIGN.md mentions of Seri and Denemeler are superseded for v1; screen tasks follow this record.

## 2026-10-26 — Missed sessions: past is frozen, makeup link
- Context: auto-rescheduling is out of scope, but the design shows "cumartesiye taşındı".
- Decision (user): a session whose date (Europe/Istanbul) has passed without completion shows as "yapılamadı". Plan writes never touch days before today or completed sessions. When the AI adds a makeup session it may set `makeup_of` to the missed session's id; the missed row then reads "<gün>e taşındı" with the hatched bar.
- Alternatives: "yapılamadı" only, no link; let the AI overwrite past uncompleted sessions.
- Consequences: save-plan responses report skipped (frozen) sessions; the moved state is derived from the link, never set directly.

## 2026-10-26 — Completion: undo and marking from Hafta
- Decision (user): after "Tamamlandı" the panel shows a quiet "Geri al" text link. Hafta rows for today and past days can be marked done or undone. Future days cannot be marked. Every change is audited with its source (web/api/mcp).
- Alternatives: today only with undo; no undo.

## 2026-10-26 — WKY-02: frontend toolchain and base layout

- Context: WKY-02 set up the server-rendered UI toolchain (templ, Tailwind, self-hosted fonts, htmx) and the base layout, with everything embedded so the app is a single binary and no request leaves localhost. User approved the three downloads in-session (source + size).
- Decisions:
  - **templ `v0.3.1100`**, added as a Go tool (`tool github.com/a-h/templ/cmd/templ` in `go.mod`); run with `go tool templ generate`. Generated `*_templ.go` **are committed** (so `go build ./...` and a fresh clone build without first running templ; `scripts/build.sh` still regenerates them). Runtime lib `github.com/a-h/templ` is a direct dependency.
  - **Tailwind CSS standalone CLI `v4.3.3`**, binary at `./bin/tailwindcss.exe` (`/bin/` is gitignored — the binary is a ~110 MB dev tool, re-downloadable from `tailwindlabs/tailwindcss` releases). v4 CSS-first config: tokens live in `web/styles/input.css` `@theme`, sources via `@source "../../internal/web/views"`; output minified to `web/static/app.css`. Soft tints are explicit tokens (`--color-z1-soft`, `--color-z3-soft`), never opacity modifiers (Soft Token Rule).
  - **htmx `2.0.11`** (npm stable `latest`; GitHub's "latest" release is a v4.0.0 pre-release — not used), saved to `web/static/htmx.min.js`, loaded with `defer`. Only client-side JS; served locally.
  - **Fonts self-hosted** via Fontsource (jsDelivr): Barlow 400/500/600/700 and Barlow Condensed 500/600/700, each in `latin` + `latin-ext` woff2 (14 files) under `web/static/fonts/` with per-subset `unicode-range` (latin-ext carries ş ğ İ; latin carries ı ç ö ü) and `font-display: swap`. `OFL.txt` sits alongside. No Google Fonts / CDN link anywhere.
  - **Assets embedded** with `embed.FS` in the top-level `web` package (`web/assets.go`); served under `GET /static/` with `Cache-Control: public, max-age=31536000, immutable` and a `?v=<sha256-12>` cache-busting query derived from `app.css`.
  - **Responsive breakpoint = Tailwind `md` (768px)**: mobile bottom tab bar below `md`, desktop 232px rail at `md`+. DESIGN.md left this switch-over "not yet decided"; this is the v1 choice.
  - **Packages:** top-level `web` (embedded assets) and `internal/web` (handlers) are both `package web`; importers alias the former as `assets`. templ components live in `internal/web/views` (`package views`).
  - **Dev launch:** `.claude/launch.json` `workey-dev` runs `go run ./cmd/workey serve` with `env WORKEY_DEV=1` on port 8080 (a `bash -lc` launcher resolved to WSL bash and failed on this machine).
- Alternatives: gitignore generated templ files and generate in CI (rejected — breaks `go build` from a clean checkout for a local-first app); Tailwind via npm/PostCSS (rejected — the stack decision is the standalone CLI, no Node); Google Fonts CDN (rejected — must work offline, no external request); htmx 4.0.0 pre-release (rejected — task wants 2.x and npm stable is 2.0.11).
- Consequences: editing a `.templ` requires `go tool templ generate`; changing classes/tokens requires a Tailwind rebuild before `go build` (app.css is embedded). The Tailwind binary must be re-downloaded on a fresh machine (not in git). CLAUDE.md "Checks" updated accordingly.

## 2026-10-26 — Session data model: 1..n items, no free-text note
- Decision (user): a session has date, start time, duration (minutes), kind (`aksam`, `uzun_blok`, `deneme`), optional topic group, optional `makeup_of`, and 1..n items. An item has book (M1, M2, M3, G1, G2), chapter number, chapter title, page range, and optional splits: konu pages, soru pages, antrenman range. No free-text note field; the desktop "Not" panel is generated from state (missed/moved sessions).
- Alternatives: same plus a 140-character AI note; one item per session.
- Consequences: the app stores references only (book, chapter, pages), never topic explanations or question content.

## 2026-10-26 — WKY-03: SQLite store on modernc.org/sqlite, embedded migrations
- Context: v1 needs a persistence layer. The stack is a single self-contained binary for local Windows (Linux later), so a CGO SQLite driver (mattn/go-sqlite3) would force a C toolchain and break `CGO_ENABLED=0` cross-builds.
- Decisions:
  - **Driver `modernc.org/sqlite` v1.61.0** (pure Go, no CGO): registered under database/sql name `sqlite`; builds with `CGO_ENABLED=0`, keeping the single-binary promise.
  - **Connection:** opened via DSN pragmas `busy_timeout(5000)`, `foreign_keys(1)`, `journal_mode(WAL)` so every connection carries them. `SetMaxOpenConns(1)` — the single-writer pattern: for a single-user local app one connection serialises writes and (with WAL + busy timeout) avoids "database is locked" without a separate reader pool.
  - **Migrations:** ordered `internal/store/migrations/*.sql` embedded with `//go:embed`; applied in filename order, each in its own tx, tracked by filename in a `schema_migrations` table. `Migrate` is idempotent (skips recorded versions). v1 ships one file, `0001_init.sql`.
  - **Schema:** `phases`, `topic_groups`, `sessions`, `session_items`, `audit_log`, `idempotency_keys`. Dates/times/timestamps stored as TEXT (civil strings / ISO-8601 UTC); the store keeps no Go date types. CHECK constraints enforce enums (`kind`, `book`, audit/completed `source`) and range ordering (`page_from <= page_to`, and the optional konu/soru/antrenman/phase/dgs ranges); `session_items.session_id` is `ON DELETE CASCADE`, `sessions.makeup_of` is a self-FK `ON DELETE SET NULL`; `UNIQUE(date, start_time)` and `UNIQUE(session_id, ordinal)`.
  - **API:** `Store` wraps one `*sql.DB`; `WithTx(ctx, func(dbtx) error)` runs a transaction; query helpers take a `dbtx` (satisfied by `*sql.DB` and `*sql.Tx`) so they run directly or in a tx. No business rules in this layer (no validation, freezing or notation — those live in `internal/plan`).
- Alternatives: `mattn/go-sqlite3` (rejected — needs CGO/C toolchain, no `CGO_ENABLED=0` build); a golang-migrate-style tool (rejected — an embedded loop is enough for a local app and adds no dependency); storing dates as Go `time`/`clock.Date` in the store (rejected — keeps the store a thin, type-agnostic persistence layer).
- Consequences: new migrations are added as higher-numbered `.sql` files and never edited once applied. `go.sum` grows (modernc + its deps). `CGO_ENABLED=0 go build ./...` is now a meaningful check (added to the task checklist).
