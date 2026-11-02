# WKY-24 — MCP: save_source/get_sources, bookEnum kaldırma, ajan rehberi

## Goal
Give the AI agent MCP tools to write and read the source catalog, update the
existing plan tools to the catalog shape, remove the fixed book enum, and rewrite
the agent guide so a client can build a valid catalog + plan from `tools/list`
and `get_agent_guide` alone.

## Scope
- **`internal/mcpserver`:**
  - New tool `save_source`: writes one source with its sections and units
    (declarative replace by source `key`; a section/unit set present replaces that
    source's structure). Schema from `plan.SaveSourceRequest`, enriched with
    descriptions: `kind` enum (book/video/course/other), `location_unit` enum
    (page/minute/lesson/other), unit `label` free text, optional numbers/ranges.
    **save_source must not touch `unit_progress`** (re-extracting a source keeps
    the user's done/D-Y-B where unit ids are preserved).
  - New tool `get_sources`: returns the catalog (sources → sections → units) for
    the agent to reference and correct.
  - New tool `set_unit_result` (wraps WKY-23 `SetUnitResult`): mark a unit done
    with optional D/Y/B, or undo.
  - **Remove `bookEnum`** and the book/chapter/konu/soru/antrenman item fields from
    `schema.go`. Rewrite `savePlanSchema`/`enrichItemSchema` to the new item shape
    (`source_key`, `section_ordinal`, `unit_from`/`unit_to`, `loc_from`/`loc_to`).
    Rename the outline `dgs_question_*` schema to `question_*`.
  - `kind` on a session: enum relaxed to a free label with the known values given
    as examples (not an enum constraint).
- **`get_agent_guide`** (in `internal/plan/guide.go`): rewrite the flow — the agent
  (1) extracts a source's structure from a PDF/TOC/video list/Udemy syllabus,
  (2) writes it with `save_source`, (3) reads it back with `get_sources` and the
  user corrects it in the web Kaynaklar page, (4) builds the plan referencing
  sources/sections/units, (5) optionally records unit results. Drop all DGS/
  five-book wording and the fixed-book assumption. Keep the "references only, no
  topic/question content" rule.
- Update `docs/mcp.md` and the in-app MCP usage guide text to the new tools/flow.

## Acceptance criteria
- [ ] `go test ./internal/mcpserver/...`: `tools/list` advertises save_source,
  get_sources, set_unit_result, save_plan (new item schema), get_plan, save_outline
  (question_*), complete_session, get_agent_guide; no `book` enum anywhere.
- [ ] A save_source → get_sources round-trip over the MCP handler returns the
  written catalog; a following save_source that keeps unit ids leaves prior
  `unit_progress` intact.
- [ ] save_plan rejects an item whose `source_key` is unknown, with the typed
  validation error surfaced as an MCP error.
- [ ] `get_agent_guide` text contains no fixed-book/DGS-only wording; docs/mcp.md
  matches the implemented schema.
- [ ] Builds (normal + `CGO_ENABLED=0`), vet, gofmt clean.

## Checks
- `go test ./internal/mcpserver/... ./internal/plan/...`
- `go vet ./...`, `gofmt -l .`

## Notes
- Depends on WKY-22 (item/catalog domain) and WKY-23 (unit result path).
- DECISIONS.md entry: MCP catalog tools + removal of the fixed book enum; the agent
  owns structure extraction, the user owns correction via the web page.

## Progress
(planned)
