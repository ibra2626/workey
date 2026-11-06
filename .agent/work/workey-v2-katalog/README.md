# workey-v2-katalog — shared rules

Read once per session together with your task file. Product truth: `PRODUCT.md`;
visuals: `DESIGN.md`; v1 decisions: `.agent/state/DECISIONS.md`; the v2 direction
decision is the `2026-11-02 — v2 kaynak kataloğu` entry in DECISIONS.md.

This feature generalizes workey from a fixed DGS/book tracker into a **general
source-catalog study tracker**: any user installs it locally and studies for any
exam with their own books, videos, YouTube playlists or Udemy courses. The AI
agent extracts each source's structure and writes it through MCP; the app stays
the single source of truth.

## Why (from the brain vault, 2026-11-01 decision)
- Users work through *units* ("Antrenman", "Test", "Föy", "Ders") inside a
  source, not just page ranges. v1 only stored page ranges, so these units were
  invisible in the plan detail.
- The system opens to several people. Each picks a different exam and different
  sources. Nothing may remain hard-coded to DGS or the five fixed books.
- Approach **A (source catalog)** was chosen over B (free unit list per session)
  and C (single JSON blob): the catalog keeps source structure reusable, enables
  per-source progress, and survives the first external user without a rewrite.

## Scope decisions settled for this feature (user, 2026-11-02)
- **No data migration.** İbrahim has not started his plan yet; the old 149-session
  DGS plan is disposable. The schema is rebuilt clean — no lossless-migration task,
  no mapping of old `book` codes. A fresh local database is acceptable.
- **Unit completion is independent of session completion.** The one-tap session
  "tamamlandı" moment from PRODUCT.md stays exactly as it is. Units are marked
  "bitti" separately, each with optional D/Y/B (doğru/yanlış/boş). Neither
  triggers the other.
- Core target is the catalog + units + D/Y/B + the agent/web flows. Cheap
  by-products of the clean rebuild are included: the fixed `book` enum is removed,
  `session.kind` is relaxed to an open label, and `topic_groups.dgs_question_*`
  is renamed to a generic `question_*`. The fixed exam date and 38-week framing
  are **not** torn out in this feature — they already live in saved plan content
  (phase dates) and per-install config, so they generalize without schema work.

## New concept model (supersedes the v1 item model)
- **Source (kaynak):** `key` (user-given short code, e.g. "M1"; no fixed list),
  `name`, `kind` (book / video / course / other), `location_unit`
  (page / minute / lesson / other), `ordinal`.
- **Section (bölüm):** belongs to a source; `ordinal`, `name`, optional location
  range in the source's unit.
- **Unit (birim):** ordered study piece inside a section; `label` comes from the
  source ("Antrenman" / "Test" / "Föy" / "Ders"), `number`, optional location
  range. The code never hard-codes the label words.
- **Unit progress:** separate from the catalog definition so re-writing a source
  never wipes progress — `done_at`, optional `correct` / `wrong` / `blank`,
  `completed_source`. Video units simply leave D/Y/B empty.
- **Session item:** references source + section + a unit range (and/or a location
  range); pages are never invented. Render example: `M1·4, Antrenman 5–7, Test 1–2`.

## Invariants (every task) — amended from v1
- One business-logic layer: `internal/plan`. Web, REST and MCP only parse input,
  call the service, render output. No rule is implemented twice.
- Plan/outline/source writes are idempotent and audited (time UTC, source
  `mcp` | `api` | `web`, action, entity id, before/after JSON).
- Completed sessions and sessions dated before today stay frozen for plan writes.
- "Today" / week boundaries use Europe/Istanbul via the injectable clock; tests
  never use the wall clock.
- Items store only references (source, section, unit, numeric ranges) — never
  topic text or question content. (v1 "book, chapter, page" wording is replaced
  by the source-catalog references above.)
- Notation stays produced by one formatter; it now renders source + unit refs
  (middle dot U+00B7, en dash U+2013, decimal comma) instead of the fixed `book`
  codes.
- Writing a source definition (save_source) must not overwrite unit progress.
- Secrets only from environment; never logged.

## Layout additions (adjust only with a DECISIONS entry)
- `internal/store/` — new `sources` / `sections` / `units` / `unit_progress`
  tables and their CRUD; the v1 `session_items` columns are replaced by catalog
  references. Rewrite the embedded migration (no committed data to preserve).
- `internal/plan/` — catalog domain types, catalog-aware validation, the new
  notation, unit-result service path.
- `internal/mcpserver/` — `save_source` / `get_sources` tools; the fixed
  `bookEnum` is removed; `save_plan` / `get_plan` schemas and `get_agent_guide`
  describe the catalog flow.
- `internal/api/` — REST source endpoints and the updated plan-item shape.
- `internal/web/views/` — a "Kaynaklar" correction page; Bugün/Hafta/İlerleme
  show units and let a unit be marked done with optional D/Y/B.

## Conventions
- Same as v1: Go 1.27.3, stdlib `net/http`, templ (`go tool templ`, generated
  `*_templ.go` committed), Tailwind standalone CLI, CSS-variable tokens only.
- Each new module or schema change gets a DECISIONS.md entry (why, alternatives).
- Keep `CLAUDE.md` "Checks" current when a check is added.
- Commit/push/PR only on explicit user request, exact paths, no force push.
