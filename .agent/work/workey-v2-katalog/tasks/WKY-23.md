# WKY-23 — Birim ilerlemesi ve D/Y/B servis yolu

## Goal
A business-logic path to mark a unit "bitti" (done) with optional D/Y/B
(doğru/yanlış/boş), fully independent of session completion. Writing it is
idempotent and audited; reading it feeds per-source/section progress.

## Scope
- **`internal/plan`** (new `units.go` or extend `complete.go`):
  - `SetUnitResult(ctx, source, req)` where `req = { unit_id, done bool,
    correct *int, wrong *int, blank *int }`:
    - `done=true` upserts a `unit_progress` row (sets `done_at` if newly done,
      keeps the first `done_at` on repeat), stores optional D/Y/B, records
      `completed_source`.
    - `done=false` clears the row (undo), audited.
    - D/Y/B are optional and only meaningful for non-video units; a video unit
      leaves them null. Validate non-negative when present; no CHECK that they sum
      to anything (the user reports what they want).
    - Idempotent: same payload twice → same state, no second audit row.
  - **No coupling to session completion.** Session `complete`/`undo` (v1
    `complete.go`) is untouched; marking all units in a session done does not
    complete the session, and completing a session does not touch units.
  - Read helpers: unit-progress lookup for a set of units; per-source and
    per-section D/Y/B + done-count aggregates for the İlerleme/Kaynaklar pages
    (WKY-26/27).
- **`internal/store`:** `unit_progress` upsert/delete/read-by-units and the
  aggregate queries (counts and D/Y/B sums grouped by source/section).
- Audit: every set/undo writes an `audit_log` row (entity `unit`, before/after
  JSON, source).
- Non-goals: transport (WKY-24/25 expose this), web controls (WKY-27).

## Acceptance criteria
- [ ] `go test ./internal/plan/...`: set done (no D/Y/B); set done with D/Y/B;
  repeat (idempotent, no new audit, `done_at` stable); undo clears row; negative
  D/Y/B rejected; completing all units of a session leaves the session's
  `completed_at` NULL (independence); completing a session leaves units untouched.
- [ ] Aggregate read returns correct done-count and D/Y/B sums per source/section
  over a seeded catalog.
- [ ] Builds (normal + `CGO_ENABLED=0`), vet, gofmt clean.

## Checks
- `go test ./internal/plan/... ./internal/store/...`
- `go vet ./...`, `gofmt -l .`

## Notes
- DECISIONS.md entry: unit completion independent of session completion (user
  decision 2026-11-02) and the reasoning — the one-tap session "tamamlandı" moment
  in PRODUCT.md must stay primary; units are a finer, optional overlay.

## Progress
Done 2026-11-02.
- Store (`internal/store/sources.go`): added `UnitByID` (unit joined to its
  section ordinal + source id/key/name/kind/location_unit), `UnitProgressByUnits`
  (read-by-unit-set, empty input → empty map, no query), and aggregate queries
  `UnitProgressTotalsBySource`/`UnitProgressTotalsBySection` (shared
  `unitProgressTotals(groupCol)`, `COUNT` + `COALESCE(SUM(...),0)` for D/Y/B) with
  a `UnitProgressTotals` struct. The WKY-21 `unit_progress` CRUD was reused as-is.
- Plan (`internal/plan/units.go`, new): `SetUnitResult(source, UnitResultRequest)`
  — done upserts (sticky first `done_at`, optional D/Y/B, writing source), undo
  clears; idempotent no-op writes no audit; one `audit_log` row per real change
  (entity `unit`, actions `complete`/`undo` reused, before/after JSON, NULL hash).
  `checkNonNeg` rejects negative D/Y/B; a `video` source rejects any D/Y/B but can
  still be marked done (null). Returns `UnitResultView` with notation via
  `FormatRef`. Read helpers `UnitResults(unitIDs)`, `SourceProgress()`,
  `SectionProgress()` mapping store aggregates to plan types `UnitResult` /
  `ProgressTotals`.
- Tests (`internal/plan/units_test.go`): done without/with D/Y/B; idempotent
  (stable done_at, 1 audit row); D/Y/B change keeps done_at + audits; undo clears +
  second undo no-op; negative rejected; missing unit not_found; video rejects D/Y/B
  but marks done; two-direction independence (units done ⇏ session complete;
  session complete ⇏ units touched); source/section aggregates + read-by-units.
  New helpers `sectionUnitIDs`, `sectionID`, `seedVideoUnit`, `unitAuditCount`.

### Checks (2026-11-02, Go 1.27.3)
- `go test ./internal/plan/... ./internal/store/...` ok.
- `go vet ./internal/plan/... ./internal/store/...` clean.
- `go build` + `CGO_ENABLED=0 go build` of plan+store ok.
- `gofmt -l` on the three changed files prints nothing.
- NOT run: whole-tree build/vet (web/transport red by design until WKY-28),
  govulncheck. No commit/push.
