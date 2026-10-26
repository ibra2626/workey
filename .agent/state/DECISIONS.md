# Architectural Decisions

Record accepted decisions with date, context, decision, alternatives, and consequences. Do not record proposals as accepted decisions.

## 2026-10-26 — WKY-07: MCP server (Streamable HTTP) in `internal/mcpserver`
- Context: expose the plan service to an AI client as MCP tools in the same binary, preserving the one business-logic layer and reusing `internal/httpx`.
- Decisions:
  - **SDK `github.com/modelcontextprotocol/go-sdk` v1.8.0** (latest stable; v1.1.0 was the first v1 GA). Mounted at `/mcp` via `mcp.NewStreamableHTTPHandler(getServer, nil)` (stateful default, SSE), wrapped with `httpx.SecurityHeaders(httpx.BearerAuth(token, …))`. It pulls in `github.com/google/jsonschema-go` v0.4.3 (the SDK's schema inference/validation library, 2020-12 draft).
  - **The `/mcp` Bearer token is `WORKEY_API_TOKEN`** — the same secret as `/api/v1`. A single-user local app has no reason to issue two tokens; config stays unchanged. A separate `WORKEY_MCP_TOKEN` can be split out later without touching the handler (it already takes the token as an argument).
  - **Six tools, mirroring the service 1:1:** `get_plan(from,to)`, `save_plan(SavePlanRequest)`, `complete_session(session_id, done=true)`, `get_outline()`, `save_outline(SaveOutlineRequest)`, `get_progress()`. Chose **`get_outline`/`save_outline`** over the task's alternative split (`get_topic_groups`/`save_topic_groups`/`save_phases`) because the service exposes outline reads/writes as one unit (`Outline`/`SaveOutline`) and REST already does the same — fewer tools, no logic split. `complete_session` addresses the session by the `id` returned in `get_plan` (as REST does), while `save_plan`'s `makeup_of` stays a `(date, start_time)` slot.
  - **Typed `mcp.AddTool` with enriched schemas.** Inputs reuse the `plan` request types directly (zero mapping) for `save_plan`/`save_outline`, and small local structs for `get_plan`/`complete_session`; schemas are inferred with `jsonschema.For` then enriched in `schema.go` with book/kind enums, `YYYY-MM-DD` / `HH:MM` patterns, min/max bounds and English descriptions stating Europe/Istanbul and "references only, never topic text". `complete_session.done` is optional with schema default `true`. Outputs are the service view/result types, so the SDK emits structured output schemas.
  - **Tool errors mirror REST.** A `plan` service error is returned from the handler as a Go error whose message is the REST error envelope JSON (`{code,message,fields}` with the same machine codes and field paths); the SDK packs it into the result with `IsError=true` and never marshals the output. Argument-parse failures (bad date in `get_plan`) produce the same `validation_failed` shape.
- Alternatives: raw `Server.AddTool` with hand-written schemas and manual validation (rejected — typed `AddTool` gives inference + input/output validation for free); duplicate MCP input structs mapped to `plan` types (rejected for `save_plan`/`save_outline` — reusing the domain request types avoids a second definition and keeps the wire shape identical to REST; descriptions are set programmatically to keep the domain layer free of AI-facing text); a second `WORKEY_MCP_TOKEN` (rejected — one user, one token); SSE transport or the older HTTP+SSE handler (rejected — Streamable HTTP is the current spec and what the task asked for).
- Consequences: `go.mod`/`go.sum` grow (go-sdk + jsonschema-go + oauth2/time/jwt/segmentio transitive deps). `cmd/workey` mounts `/mcp` alongside `/api/v1/`; `newMux` now takes both handlers. Focused tests (`internal/mcpserver` in-process SDK client over Streamable HTTP: list, save×2, read, complete, progress, 401, tool-error code), `go vet`, `gofmt`, and normal + `CGO_ENABLED=0` builds pass with Go 1.27.3. MCP resources/prompts and OAuth remain out of scope.

## 2026-10-26 — WKY-06: REST transport and shared HTTP helpers
- Context: expose the plan service through `/api/v1`, preserving one business-logic layer and reusable transport helpers for WKY-07.
- Decision: stdlib method/path routing in `internal/api`; handlers parse, call the service with source `api`, render JSON. `internal/httpx` owns Bearer auth (SHA-256 digests + `crypto/subtle`, empty token fails closed, 401 challenge), JSON errors, body limits and security headers. API cap: 1 MiB. Validation maps to 422, all-not-found to 404, idempotency conflict to 409, unexpected failures to generic 500. `Idempotency-Key` header overrides the body key.
- Decision: `Range`, `Outline`, `Today`, `TodayDate` and exported `ParseDate` live in `internal/plan` for reusable reads and injected-clock access. `serve` opens/migrates SQLite before HTTP, defaults an unset DB path to `workey.db`, mounts the protected API at `/api/v1/`.
- Alternatives: persistence/state rules in handlers (duplicates domain logic); a web framework (stdlib suffices); raw token comparison (digest comparison fixes comparison length); mandatory DB path for every launch (default supports local startup).
- Consequences: MCP can reuse helpers and service reads; web cookie auth remains WKY-08. `docs/api.md` records wire behavior. Focused tests, vet, formatting and normal/CGO-disabled builds pass with Go 1.27.3.

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

## 2026-10-26 — WKY-04: plan write path (`internal/plan`), store composability
- Context: `internal/plan` is the single business-logic layer for plan writes (validation, freezing, idempotency, audit, notation). It composes `internal/store` query helpers inside one transaction.
- Decisions:
  - **`store.WithTx` callback now takes `*sql.Tx`** instead of the unexported `dbtx` interface. External packages cannot name `dbtx`, so they could not build a `func(tx dbtx) error`; `*sql.Tx` satisfies the (still unexported) `dbtx` that every query helper accepts, so plan can compose helpers in a transaction. `store_test.go` updated to match. The query helpers keep taking `dbtx` so `*sql.DB` still works for read-only paths.
  - **`makeup_of` is referenced by `(date, start_time)` (`SlotRef`), not by raw id.** The write API is declarative and matches sessions by that slot; the service resolves the slot to the missed session's stored id and requires it to be past and uncompleted. The external caller (AI via MCP/API) never sees internal ids. (DECISIONS 2026-10-26 "Missed sessions" said the AI "may set `makeup_of` to the missed session's id"; the id is the stored form, the slot is the write-API form that resolves to it.)
  - **Range replace is declarative and frozen-aware.** For dates in `[from,to]` that are today-or-later (Europe/Istanbul via the injectable clock), the open sessions become exactly the request's set, matched by slot: matched → update-in-place (id kept) or unchanged, absent → deleted. Past-dated and completed sessions are frozen (never updated/deleted); an incoming session targeting one is returned in `skipped[]` with a machine reason, not applied.
  - **Idempotency has two layers.** Without a key, re-applying the same payload yields all-unchanged and zero audit rows (content equality check per session). With `idempotency_key`, the canonical request hash + stored response are kept in `idempotency_keys`; a repeat replays the stored response, and the same key with a different payload returns `ErrIdempotencyConflict`.
  - **Item replacement, not item diffing.** On update, a session's items are deleted and re-inserted with ordinals 0..n; simpler than per-item diff and correct under the `UNIQUE(session_id, ordinal)` constraint.
  - **Typed errors.** `ValidationError` carries `FieldError{path, code, message}` (stable machine codes) so REST/MCP render without re-checking; store-dependent checks (unknown topic-group key, invalid makeup target, referenced-topic-group delete) produce the same typed errors. Structural validation (`validate.go`) is store-free.
  - **Outline (`SaveOutline`)** replaces phases and topic groups by key (create/update/unchanged/delete), idempotent and audited; deleting a topic group still referenced by sessions fails with code `referenced` and rolls the whole transaction back.
  - **New store helpers** (persistence only, no rules): `catalog.go` adds `Phase`/`TopicGroup` with list/insert/update-by-key/delete-by-key, `TopicGroupByKey`, `CountSessionsByTopicGroup`; `queries.go` adds `UpdateSession`, `DeleteSession`, `DeleteSessionItems`, `SessionByDateAndTime`.
- Alternatives: export `dbtx` as `DBTX` (rejected — passing the concrete `*sql.Tx` is simpler and leaks nothing); reference makeup by id in the write API (rejected — exposes internal ids to the AI/API); per-item diff on update (rejected — more code for no user-visible gain); a separate idempotency table scope per source (rejected — the key is caller-chosen and global is enough for one user).
- Consequences: all plan write rules live here and nowhere else (REST/MCP/web will only parse+call+render). The `testdata/` fixtures (`sample-week.json`, `sample-outline.json`) are labelled sample data for tests and demos, not real plan content.

## 2026-10-26 — WKY-05: plan read models and completion (`internal/plan`)
- Context: the read side (Bugün/Hafta/İlerleme view models) and complete/undo, computed once in the service so web/REST/MCP only render.
- Decisions:
  - **Three read methods** `Day(ctx, date)`, `Week(ctx, date)`, `Progress(ctx, examDate)` return fully-prepared view models (`DayView`/`WeekView`/`ProgressView` with `SessionView`). State is always derived against the clock's "today", so any date can be read.
  - **Derived `SessionState`** `planned|today|done|missed|moved` with precedence done > (past → moved if referenced by a makeup, else missed) > today > planned. "moved" is derived from the makeup link (`ListMakeupLinks`), never stored; the moved row carries `MovedTo{Date, Weekday}`.
  - **`examDate` is a `Progress` parameter**, not a `New` field — the handler passes it from `WORKEY_EXAM_DATE`; the write-path `New(st, clk)` signature is unchanged. `WeeksTotal = 38` is a package constant (PRODUCT.md), plan start = the first phase's Monday (fallback: earliest session's Monday); both avoid new env vars.
  - **Completed share per topic group = done minutes / planned minutes** within the group (bounded 0..1, `Percent` rounded). The mockup's "%13" is illustrative sample data and is not reproduced; the figure chosen is defensible and consistent with weekly "hacim" (minutes).
  - **Completion `Complete`/`Undo`** allowed for today-or-past only (future → `ValidationError` code `future`); idempotent (already-in-target-state → no write, no audit row); audited as action `complete`/`undo` with before/after JSON and a NULL request hash (completions have no request payload to hash). Both return the updated `SessionView`.
  - **Turkish calendar lives in the plan layer** (`dates.go`), not templates: name tables + `FormatDayHeading`/`FormatShortDate`/`FormatWeekRange` (en dash U+2013) + `FormatHoursTR` (decimal comma). Lowercase weekday forms are written out rather than produced by case conversion, so the Turkish i/İ ↔ ı/I rule never applies by accident.
  - **Civil-date arithmetic is zone-free** (`toDate`/`addDays`/`civilDays` on UTC midnights). Safe because Turkey has had no DST since 2016; UTC avoids any DST discontinuity regardless.
  - **New store helpers** (persistence only): `SessionByID`, `AllSessions`, `SetSessionCompletion`, `ListMakeupLinks` (+ `MakeupLink`), private `sessionItems`.
- Alternatives: add `examDate`/plan-start to `New` or to `config` as new env vars (rejected — more wiring/churn for a value only Progress needs; 38 and the first phase already pin the window); match the mockup's %13 exactly (rejected — sample/illustrative, meaning undefined); store a `moved` flag or a `completed` boolean separately (rejected — derive from the makeup link and `completed_at`); Turkish `strings.ToLower` for weekday notes (rejected — locale-unaware, risks i/İ bugs).
- Consequences: screens (WKY-09/10/11) consume these view models directly and format nothing date/notation-related themselves. REST/MCP (WKY-06/07) expose completion via `Complete`/`Undo` and the read models as JSON.
