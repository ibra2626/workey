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
- [ ] Recorded end-to-end run: two sources (book + video) defined, a plan written
  referencing both, units completed with/without D/Y/B, screens render correctly;
  session and unit completion shown independent.
- [ ] No fixed-book / `dgs_question_*` identifiers left in code; `get_agent_guide`,
  docs/mcp.md, docs/api.md all describe the catalog flow.
- [ ] Full focused checks + govulncheck pass (or findings reported); evidence
  written to a docs/ verification note like WKY-14/19.
- [ ] State docs updated.

## Checks
- `scripts/build.sh` (templ → Tailwind → build)
- `go test ./internal/... ./cmd/...`
- `go vet ./...`, `gofmt -l .`, `CGO_ENABLED=0 go build ./...`
- `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...`

## Notes
- Depends on WKY-24, WKY-25, WKY-27. Commit/push/PR only on explicit user request.

## Progress
(planned)
