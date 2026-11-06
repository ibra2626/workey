# WKY-22 — Plan domain modelini kataloğa bağla

## Goal
Rebind `internal/plan` to the catalog: a session `Item` references a source +
section + unit/location range instead of a fixed book + pages. Validation checks
references against the stored catalog; the notation formatter renders the new
refs. Keep SavePlan's declarative-replace, idempotency, frozen-past and audit
semantics unchanged.

## Scope
- **`plan.go` domain types:**
  - Remove `validBooks`. New `Item`: `SourceKey string`, `SectionOrdinal *int`,
    `UnitFrom`/`UnitTo *int`, `LocFrom`/`LocTo *int`. (Items name the source by its
    stable `key`; the service resolves key → id and section ordinal → id.)
  - `Source`/`Section`/`Unit` input types for the catalog write path live here too
    (used by WKY-24/25) — define them now: `SaveSourceRequest{ Source, Sections[] }`
    where each section carries its units.
  - Relax `validKinds`: `kind` is a free label; keep the known constants as display
    hints only, not a validation gate (a blank or unknown kind is allowed).
  - `TopicGroupInput.DGSQuestion*` → `Question*`.
- **`validate.go`:** structural checks (ranges ordered, ordinals ≥ 1, source key
  non-empty). Add catalog-reference validation in the service (needs the store):
  unknown `source_key`, unknown section ordinal for that source, or a unit range
  outside the section's units → typed `ValidationError` with field path + machine
  code (`unknown_source`, `unknown_section`, `unit_out_of_range`).
- **`notation.go`:** render `M1·4, Antrenman 5–7, Test 1–2` from catalog refs
  (source key · section ordinal, then unit label + number range, and/or location
  range). Multiple items join with ` · `. Preserve U+00B7 / U+2013 / decimal comma;
  keep the ASCII-hyphen-rejected test discipline. The formatter takes resolved
  catalog data (label, numbers), not raw ids.
- **`service.go` SavePlan:** resolve each item's `source_key`/section ordinal
  against the catalog inside the tx; persist resolved ids to `session_items`;
  everything else (range replace, skip frozen, idempotency, audit before/after)
  unchanged. A plan referencing a missing source fails validation before any write.
- **`read.go`:** read models now join `session_items` → sources/sections/units so
  Bugün/Hafta render unit refs; expose the data the formatter needs.
- Non-goals: unit completion/D-Y-B (WKY-23), transport (WKY-24/25), web (WKY-26/27).

## Acceptance criteria
- [ ] `go test ./internal/plan/...` green: SavePlan create/idempotent/change-one/
  past-skip/completed-untouched/shrink-delete still pass with catalog-referencing
  fixtures; add unknown-source / unknown-section / unit-out-of-range validation
  cases.
- [ ] Notation tests with Turkish labels and exact runes; single values collapse
  ranges; multi-item join.
- [ ] Read model returns source key, section, and unit label/number for an item.
- [ ] `internal/plan` + `internal/store` build (normal + `CGO_ENABLED=0`), `go vet`
  and `gofmt` clean for those packages. (Scope split, user 2026-11-02: the
  **whole-tree** build stays red until the downstream layers are rebound — see
  Notes. WKY-22 only guarantees the plan/store packages.)

## Checks
- `go test ./internal/plan/... ./internal/store/...`
- `go vet ./...`, `gofmt -l .`
- `go build ./...` and `CGO_ENABLED=0 go build ./...`

## Scope split (user decision, 2026-11-02)
Changing `plan.Item` to the catalog shape breaks every layer that renders the v1
book/chapter/page/konu/soru/antrenman concepts — those concepts no longer exist,
so a "shallow compile fix" would mean gutting real logic that belongs to the
downstream tasks. The user chose to **split**: WKY-22 delivers the plan core
(and keeps the whole `internal/plan` package, study flow included, compiling with
its tests green); the downstream compile-and-rebind is carried by the tasks that
already own those layers. Until they land, `go build ./...` is intentionally red.
- `internal/mcpserver`, `internal/apidoc` → **WKY-24** (also restores MCP build).
- `internal/api` → **WKY-25** (also restores REST build).
- `internal/web`, `internal/web/views` (present.go / study.go / *.templ) → **WKY-26/27**
  (also restores web build). The akşam konu/soru/yanlış-defteri study profile and
  the page/stage bookmark rendering are redefined there against the catalog model.
- In `internal/plan` the session-study flow is adapted minimally now: `ResumePage`
  is kept but validated against the item's **location** range (konu/soru sub-ranges
  are gone); its richer stage semantics are revisited in WKY-27.
- Sample fixtures `testdata/sample-*.json` move to the catalog shape here, so the
  deferred api/mcp/web tests that still read them stay red until their task.

## Notes
- Update `testdata/sample-*.json` fixtures to the catalog shape (a small sample
  source with a couple of sections and labelled units), clearly marked sample data.
- DECISIONS.md entry: items reference the source by `key` (stable, user-facing)
  resolved to ids at write time (alternative: items carry raw ids — rejected, the
  agent writes by human-readable key and ids are internal).

## Progress
DONE (plan core) — 2026-11-02.
- `plan.Item` → catalog shape (`source_key`, `section_ordinal`, `unit_from/to`,
  `loc_from/to`); removed `validBooks`; `kind` is a free label (no enum gate);
  `TopicGroupInput`/read models `DGSQuestion*`→`Question*`. Catalog write-path
  input types (`SaveSourceRequest`/`SourceInput`/`SectionInput`/`UnitInput`) defined.
- New `resolve.go`: per-tx `catalog` snapshot; resolves stored items → `ItemRef`
  (notation) / `ItemView` (read), and write items → `store.SessionItem` with
  `unknown_source`/`unknown_section`/`unit_out_of_range` typed errors.
- `notation.go` rewritten: `FormatRef`/`FormatRefs` render `M1·4, Antrenman 5–7,
  Test 1–2` from resolved data; U+00B7/U+2013 kept; loc prefix per location unit.
- `validate.go` structural item checks; `service.go` SavePlan resolves/validates
  items in-tx (replace/idempotency/frozen/audit unchanged), `itemContentEqual` and
  `snapshotJSON` on new fields; `read.go` joins catalog, `SessionView.Items`→`[]ItemView`;
  `outline.go` Question*; `study.go` bookmark validated against loc range;
  `study_view.go` bookmark resolved via catalog.
- `testdata/sample-*.json` → catalog shape; plan tests reworked (catalog seed via
  store, new validation cases, notation Turkish-label tests).
- Checks (Go 1.27.3): `go test ./internal/plan/... ./internal/store/...` ok;
  `go vet` clean; build normal + `CGO_ENABLED=0` ok; gofmt-clean on LF content
  (`gofmt -l` still flags pre-existing CRLF working-tree artifacts). NOT run:
  whole-tree build/vet (web/transport deferred — red by design), govulncheck. No commit.
- Boundary moved (user split): web/transport compile-fix → WKY-24/25/26/27 (see
  Scope split above and TASKS.md).
