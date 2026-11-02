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
- [ ] Whole project builds again (normal + `CGO_ENABLED=0`); `go vet`, `gofmt`
  clean. (This is the task that makes the tree compile after WKY-21.)

## Checks
- `go test ./internal/plan/... ./internal/store/...`
- `go vet ./...`, `gofmt -l .`
- `go build ./...` and `CGO_ENABLED=0 go build ./...`

## Notes
- Update `testdata/sample-*.json` fixtures to the catalog shape (a small sample
  source with a couple of sections and labelled units), clearly marked sample data.
- DECISIONS.md entry: items reference the source by `key` (stable, user-facing)
  resolved to ids at write time (alternative: items carry raw ids — rejected, the
  agent writes by human-readable key and ids are internal).

## Progress
(planned)
