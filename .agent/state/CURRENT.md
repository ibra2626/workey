# CURRENT — WKY-21 done (catalog store); WKY-22 next

Updated 2026-11-02.
Active task: .agent/work/workey-v2-katalog/tasks/WKY-21.md (DONE)
Next: WKY-22 — bind the plan domain to the catalog (Item → source/section/unit,
notation), which also repairs the intended compile breakage below.

## Feature: workey-v2-katalog (WKY-21..28)
Generalize workey from a fixed DGS/5-book tracker to a general source catalog
(Kaynak → Bölüm → Birim) with separate unit progress + D/Y/B. Direction +
WKY-21 implementation decisions: `.agent/state/DECISIONS.md` (both 2026-11-02
entries). Shared rules: `.agent/work/workey-v2-katalog/README.md`. No data
migration (old DGS plan disposable); schema rebuilt clean.

## WKY-21 result — 2026-11-02 (store + schema)
Rewrote `internal/store/migrations/0001_init.sql` to the catalog schema:
- `sources(key UNIQUE, name, kind CHECK book/video/course/other, location_unit
  CHECK page/minute/lesson/other, ordinal, created_at, updated_at)`
- `sections(source_id FK CASCADE, ordinal, name, loc_from/to, CHECK range,
  UNIQUE(source_id,ordinal))`
- `units(section_id FK CASCADE, ordinal, label free TEXT (no enum), number,
  loc_from/to, CHECK range, UNIQUE(section_id,ordinal))`
- `unit_progress(unit_id PK FK CASCADE, done_at, correct/wrong/blank CHECK ≥0,
  completed_source NOT NULL CHECK mcp/api/web)` — presence = done; D/Y/B optional.
- `session_items` reshaped: `source_id` NOT NULL FK, `section_id` nullable FK,
  `unit_from/to`, `loc_from/to` (range CHECKs), UNIQUE(session,ordinal). Only the
  session→item edge cascades; source/section use default NO ACTION.
- `sessions.kind` CHECK dropped (open TEXT); `topic_groups.dgs_question_*` →
  `question_*`. `phases`/`sessions`/`audit_log`/`idempotency_keys` kept.
  `0002_study_flow.sql` untouched (references only `sessions`).

Store Go API:
- New `internal/store/sources.go`: `Source`/`Section`/`Unit`/`UnitProgress` types;
  source list/get-by-key/insert/update-by-key/delete; section & unit
  insert/update(by id)/delete + `ReplaceSections`/`ReplaceUnits`;
  `SectionsBySource`/`UnitsBySection`/`UnitsBySource`; `UpsertUnitProgress`
  (INSERT…ON CONFLICT(unit_id)), `UnitProgressByUnit`, `UnitProgressBySource`
  (joined, keyed by unit id), `DeleteUnitProgress`. `requireOne` turns a
  zero-row UPDATE into wrapped `sql.ErrNoRows`.
- `queries.go`: `SessionItem` reshaped; all read/write/scan sites updated.
- `catalog.go`: `TopicGroup.Question{Min,Max}` (fields + SQL).
- Tests: `store_test.go` ported (open kind accepted; item range CHECK kept);
  new `sources_test.go` (CRUD/list, full cascade, UNIQUE/CHECK/FK violations,
  progress upsert + read-by-source, keep-id progress-survival vs ReplaceUnits
  drop); `study_test.go` legacy item fixed.

Design choice (why id-preserving update helpers exist): progress is keyed by
unit id, so WKY-22 must reconcile sections/units by id (UpdateUnit) to keep
progress; `ReplaceUnits` is the full-rewrite path that intentionally drops it.

### Checks (2026-11-02, Go 1.27.3)
- `go test ./internal/store/...` ok; `go vet ./internal/store/...` clean.
- `go build ./internal/store/...` + `CGO_ENABLED=0 go build ./internal/store/...`
  ok.
- gofmt clean on LF-normalized content; `gofmt -l` flags only pre-existing CRLF
  working-tree artifacts (core.autocrlf=true; git stores LF).
- NOT run: full-tree build/test, govulncheck, templ/Tailwind. No commit/push.

### Expected breakage (WKY-22 repairs)
`go build ./...` fails in `internal/plan` and above — `SessionItem` columns and
`TopicGroup.Question*` rename. Intended WKY-21 boundary (README/task Notes).

## Prior v1 state (unchanged)
SF-01..06 study flow, study room, Plan calendar shipped on main (merged PR #1).
PRODUCT.md / DESIGN.md / `.impeccable/design.json` are v1 truth; the catalog
feature supersedes the v1 item model only. WKY-12 deploy held; WKY-20 200% zoom
acceptance pending.
