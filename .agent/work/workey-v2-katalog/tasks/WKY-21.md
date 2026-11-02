# WKY-21 — Katalog şeması ve store katmanı

## Goal
Replace the DGS-specific item storage with a general source catalog:
`sources` → `sections` → `units`, plus a separate `unit_progress` table so a
source re-write never wipes study progress. No data migration (plan not started;
old DB disposable).

## Scope
- **Rewrite `internal/store/migrations/0001_init.sql`** (clean, no second
  migration — nothing committed to preserve). New tables:
  - `sources`: `id`, `key` TEXT UNIQUE, `name`, `kind` TEXT CHECK in
    (`book`,`video`,`course`,`other`), `location_unit` TEXT CHECK in
    (`page`,`minute`,`lesson`,`other`), `ordinal`, `created_at`, `updated_at`.
  - `sections`: `id`, `source_id` FK ON DELETE CASCADE, `ordinal`, `name`,
    `loc_from` INTEGER NULL, `loc_to` INTEGER NULL, CHECK range order,
    UNIQUE(`source_id`,`ordinal`).
  - `units`: `id`, `section_id` FK ON DELETE CASCADE, `ordinal`, `label` TEXT
    (source-supplied word, no CHECK list), `number` INTEGER NULL,
    `loc_from`/`loc_to` INTEGER NULL, CHECK range order,
    UNIQUE(`section_id`,`ordinal`).
  - `unit_progress`: `unit_id` PK FK ON DELETE CASCADE, `done_at` TEXT NULL,
    `correct`/`wrong`/`blank` INTEGER NULL, `completed_source` CHECK in
    (`mcp`,`api`,`web`). One row per completed unit; absence = not done.
  - Replace `session_items` columns: drop `book`/`chapter_*`/`page_*`/`konu_*`/
    `soru_*`/`antrenman_*`. New shape: `id`, `session_id` FK, `ordinal`,
    `source_id` FK, `section_id` FK NULL, `unit_from`/`unit_to` INTEGER NULL,
    `loc_from`/`loc_to` INTEGER NULL, CHECK range orders,
    UNIQUE(`session_id`,`ordinal`). (A session item points at a source + section
    + a unit/location range; multiple items render joined.)
  - Keep `phases`, `sessions`, `audit_log`, `idempotency_keys`. Rename
    `topic_groups.dgs_question_min/max` → `question_min/max` (generic).
  - Relax `sessions.kind`: drop the `aksam/uzun_blok/deneme` CHECK; `kind` becomes
    an open TEXT label (default/empty allowed). (Presentation still recognises the
    known values; the store no longer constrains them.)
- **`internal/store`**: new `Source`/`Section`/`Unit`/`UnitProgress` Go types and
  CRUD in a new `sources.go` (list/get-by-key/insert/update-by-key/delete, plus
  section and unit replace-by-source helpers); `UnitProgress` upsert/read; update
  `queries.go` session-item read/write for the new columns; keep `WithTx`.
  Update `TopicGroup` field names to `QuestionMin/Max`.
- Non-goals: domain validation, notation, any transport. Store holds no business
  rules (CHECK + FK only).

## Acceptance criteria
- [ ] `0001_init.sql` applies cleanly on a fresh DB; `go test ./internal/store/...`
  green with table tests for each new table (insert/list/update-by-key/delete,
  cascade delete source→sections→units→progress, UNIQUE/CHECK violations).
- [ ] `unit_progress` survives a source section/unit *content* re-write only where
  the unit id is kept; document the expectation in the test (progress keyed by
  unit id).
- [ ] Session-item read/write round-trips the new reference columns.
- [ ] `CGO_ENABLED=0 go build ./...` passes (pure-Go SQLite).

## Checks
- `go test ./internal/store/...`
- `go vet ./...`
- `gofmt -l .` (empty)
- `go build ./...` and `CGO_ENABLED=0 go build ./...`

## Notes
- Because `session_items` columns change, `internal/plan` will not compile until
  WKY-22. Keep this task's store API shaped for what WKY-22 needs (item refs by
  source id + section id + unit range). It is acceptable for `internal/plan` and
  above to be temporarily broken at the end of WKY-21; say so in Progress.
- DECISIONS.md entry: the catalog table design and the "progress separate from
  catalog" choice (alternative: progress columns on `units`, rejected — save_source
  would wipe them).

## Progress
Done (2026-11-02).
- **`0001_init.sql` rewritten** to the catalog schema: `sources`, `sections`
  (FK CASCADE, UNIQUE(source,ordinal)), `units` (free-TEXT `label`, FK CASCADE,
  UNIQUE(section,ordinal)), `unit_progress` (PK=unit_id FK CASCADE,
  `completed_source` NOT NULL, D/Y/B each CHECK ≥ 0). `session_items` reshaped to
  `source_id`+`section_id`+`unit_from/to`+`loc_from/to` (range CHECKs,
  UNIQUE(session,ordinal)); `sessions.kind` CHECK dropped (open TEXT);
  `topic_groups.dgs_question_*` → `question_*`. `phases`, `sessions`,
  `audit_log`, `idempotency_keys` kept; `0002_study_flow.sql` untouched.
- **`internal/store/sources.go`** (new): `Source`/`Section`/`Unit`/`UnitProgress`
  types and CRUD — source list/get-by-key/insert/update-by-key/delete; section &
  unit insert/update(by id)/delete + `ReplaceSections`/`ReplaceUnits`;
  `SectionsBySource`/`UnitsBySection`/`UnitsBySource`; `UpsertUnitProgress`,
  `UnitProgressByUnit`, `UnitProgressBySource`, `DeleteUnitProgress`.
- **`queries.go`**: `SessionItem` reshaped, all read/write/scan sites updated;
  `Session.Kind` comment relaxed. **`catalog.go`**: `TopicGroup.Question{Min,Max}`.
- **Tests**: `store_test.go` ported to the new item/source shapes (open-kind now
  accepted; range CHECK still enforced on items); new `sources_test.go` with
  table tests — source CRUD/list, source→section→unit→progress cascade,
  UNIQUE/CHECK/FK violations, progress upsert + read-by-source, and the
  **keep-unit-id progress-survival** case (UpdateUnit keeps progress;
  ReplaceUnits drops it, both asserted). `study_test.go` legacy item fixed.

### Checks run
- `go test ./internal/store/...` → ok.
- `go vet ./internal/store/...` → clean.
- `go build ./internal/store/...` and `CGO_ENABLED=0 go build ./internal/store/...`
  → ok.
- `gofmt`: new/edited Go is gofmt-clean on LF-normalized content; `gofmt -l`
  flags only the pre-existing CRLF working-tree artifact (core.autocrlf=true;
  git stores LF, as for untouched `study.go`).

### Known/expected breakage (hand-off to WKY-22)
- `go build ./...` fails in `internal/plan` and above (REST/MCP/web): the
  `SessionItem` columns and `TopicGroup.Question*` rename. This is the intended
  WKY-21 boundary — WKY-22 rebinds the plan domain to the catalog (item refs by
  source id + section id + unit range; new notation). The store API is shaped for
  that: both id-preserving update helpers and replace helpers exist.
- Not run (out of scope / would not compile): full-tree build, `go test ./...`,
  `govulncheck`, templ/Tailwind. No commit/push.
