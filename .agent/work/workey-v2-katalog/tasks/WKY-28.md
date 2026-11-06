# WKY-28 — Uçtan uca doğrulama ve v2 kapısı

## Goal
Prove the generalization end to end with two different source kinds and close the
v2 gate: a user can define and plan with their own sources via the agent, with no
code change and no DGS/book assumption left anywhere.

## Scope
- **End-to-end via MCP (as in WKY-13):** drive the live `/mcp` endpoint to
  (1) `save_source` a book-kind source with sections and Antrenman/Test units,
  (2) `save_source` a video-kind source (location_unit `minute` or `lesson`,
  units labelled "Ders"), (3) `get_sources` and confirm structure,
  (4) `save_plan` a week referencing both sources, (5) `set_unit_result` for a few
  units (with D/Y/B on the book, none on the video), (6) read Bugün/Hafta/İlerleme
  and confirm unit refs, independent session vs unit completion, and aggregates.
- **Grep/gate for leftover DGS specifics:** confirm no `bookEnum`, no
  `M1/M2/M3/G1/G2` CHECK or enum, no `dgs_question_*` identifiers remain in code
  (strings in historical docs are fine). The fixed exam date / 38-week framing is
  config/plan-content, not a schema constraint — note where it lives.
- **v1 gate re-run:** `go tool templ generate`, Tailwind rebuild, full focused test
  set across all touched packages, `go vet`, `gofmt -l .`, normal + `CGO_ENABLED=0`
  builds, and `govulncheck ./...`; report findings explicitly.
- Update `.agent/state/CURRENT.md`, `ARCHITECTURE.md` and `PROJECT.md` to the v2
  model; mark WKY-21..28 statuses in TASKS.md.

## Acceptance criteria
- [x] Recorded end-to-end run: two sources (book + video) defined, a plan written
  referencing both, units completed with/without D/Y/B, screens render correctly;
  session and unit completion shown independent.
- [x] No fixed-book / `dgs_question_*` identifiers left in code; `get_agent_guide`,
  docs/mcp.md, docs/api.md all describe the catalog flow.
- [x] Full focused checks + govulncheck pass (or findings reported); evidence
  written to a docs/ verification note like WKY-14/19.
- [x] State docs updated.

## Checks
- `scripts/build.sh` (templ → Tailwind → build)
- `go test ./internal/... ./cmd/...`
- `go vet ./...`, `gofmt -l .`, `CGO_ENABLED=0 go build ./...`
- `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...`

## Notes
- Depends on WKY-24, WKY-25, WKY-27. Commit/push/PR only on explicit user request.

## Progress

### DONE — 2026-11-04 (end-to-end gate closed)

**End-to-end run (recorded, re-runnable).** New `TestCatalogEndToEnd` in
`internal/mcpserver/webworkflow_test.go` drives the live `/mcp` endpoint (official
Go MCP SDK client, real localhost HTTP) + the web UI over one store, fixed clock
2026-10-26 09:00 İstanbul:
`save_source` book (GR, page, Antrenman/Test) → `save_source` video (VD, lesson,
Ders) → `get_sources` (structure + kinds + labels confirmed, ids grabbed) →
`save_plan` a week referencing both (2 created) → `get_plan` notation carries
`GR·1 … Antrenman` and `VD·1 … Ders` → `set_unit_result` book Antrenman 1 with
D/Y/B (8/2/0), video Ders 1 with none (returned view has null D/Y/B) → web Bugün
shows `GR·1, Antrenman 1`/`GR·1, Test 1` + `name="correct"`, Hafta shows
`VD·1, Ders 1`; `get_progress` SessionsDone=0 with 2 units done; web-form (CSRF)
completion of today's session → SessionsDone=1 while `unit_progress` stays 2
rows; İlerleme shows "Birim ilerlemesi" / "6 birim" / "Çözülen". Proves unit and
session completion independent, two kinds, no code change.

**Re-enabled by default.** Removed the `//go:build webworkflow` tag so
`TestMCPWebWorkflow` + `TestCatalogEndToEnd` run in the normal test set now that
WKY-26/27 rebuilt the web views; updated the `newMCP`/`webworkflow` doc comments.
`TestVisualReview` stays opt-in via `WORKEY_VISUAL_REVIEW=1`.

**Grep/gate.** No `bookEnum` / `dgs_question_*` / `dgsQuestion*` in `*.go` (only
in `.agent/` history). `M1/M2/G1/G2` in code are user-given source `key` examples
and the generic `kind="book"` label, not a fixed enum. Schema `0001_init.sql` is
clean v2 (free `key`, generic `kind`/`location_unit` CHECKs, free `label`,
`question_*`, no `sessions.kind` CHECK). Guide content + docs/mcp.md + docs/api.md
describe the catalog flow. Fixed exam date / 38-week framing deliberately kept —
it lives in `WORKEY_EXAM_DATE` config and `phases` date spans (plan content), not
the schema.

**One leftover fixed.** `get_agent_guide` **tool description** still said
"supported books/session kinds" → "supported source kinds, location units and
session kinds" (no test asserted the string).

**v1 gate (Go 1.27.3).** templ generate (0 updates), Tailwind rebuilt,
`go build ./...`, `CGO_ENABLED=0 go build ./...`, `go vet ./...` all green;
focused test set across config/clock/store/plan/web/web·views/api/apidoc/
mcpserver/cmd → all ok; `gofmt -l .` clean on LF (CRLF list = autocrlf artifact,
verified each file gofmt-clean after stripping CR); `govulncheck@v1.9.0 ./...` →
No vulnerabilities found.

**Docs.** Evidence in `docs/v2-gate-verification.md`. State docs updated:
`.agent/state/PROJECT.md`, `.agent/state/ARCHITECTURE.md`, this file + TASKS.md
statuses (WKY-21 & WKY-28 → done), CURRENT.md.

No commit/push/PR (awaiting explicit request). WKY-12 deploy hold stands.
