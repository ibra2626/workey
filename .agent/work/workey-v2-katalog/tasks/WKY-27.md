# WKY-27 — Web: Bugün/Hafta/İlerleme birim görünümü ve işaretleme

## Goal
Surface units in the study screens: the session detail shows which units
(Antrenman, Test, Ders…) to work, each can be marked "bitti" with optional D/Y/B,
and İlerleme reflects unit-level progress — all without disturbing the one-tap
session "tamamlandı" moment.

## Scope
- **Bugün / Hafta (`internal/web/views/today.templ`, `week.templ`):** render each
  session's items as catalog refs (`M1·4, Antrenman 5–7, Test 1–2`) via the
  formatter. Under the session, list its units with a per-unit "bitti" control and
  optional D/Y/B inputs, posting to `set_unit_result` (source `web`, htmx-first,
  CSRF, no-JS fallback). The session's own completion control stays separate and
  primary (independent per WKY-23).
- **İlerleme (`progress.templ`):** add unit-level done and D/Y/B aggregates
  alongside the existing minutes/weeks figures, honest denominators from saved data
  only (no hard-coded totals). Keep the single-column layout and DESIGN.md calm
  tone; no gamification.
- **`present.go`:** helpers for unit rows, done state, D/Y/B display; tokens only.
- Reuse the mark-done control shared with WKY-26 if one was introduced there.
- Non-goals: catalog editing (WKY-26).

## Acceptance criteria
- [ ] `go test ./internal/web/...`: Bugün/Hafta show unit refs and controls;
  marking a unit done does NOT complete the session and vice versa; İlerleme shows
  unit aggregates from seeded data; no-JS path + CSRF work.
- [ ] Visual check (built-in browser) 390×844 and 1280×800: no overflow, Turkish
  glyphs, tokens only, ≥44px targets, contrast; completion animation/feedback
  unchanged from v1.
- [ ] templ + Tailwind regenerated; builds, vet, gofmt clean.

## Checks
- `go tool templ generate` then Tailwind rebuild
- `go test ./internal/web/...`
- `go vet ./...`, `gofmt -l .`, builds (normal + `CGO_ENABLED=0`)

## Notes
- Depends on WKY-23 (unit result path) and WKY-26 (shared control + Kaynaklar).

## Progress

### DONE — 2026-11-04 (web unit view + marking on Bugün/Hafta/İlerleme)

**Plan read model (new, independent of session completion):**
- `plan.UnitRow` (`units.go`): unit id, exact per-unit notation, label/number,
  `IsVideo`, and current global done + D/Y/B. `SessionView.UnitRows []UnitRow`
  with `json:"-"` (web-only; API/MCP session views never serialise it).
- `Service.LoadUnitRows(ctx, []*SessionView)`: one tx — loads the catalog, maps
  each session item to its covered units (de-duped per session), then one
  `UnitProgressByUnits` read fills done/D-Y-B. Covered units = the item's unit
  range, or every unit of a section item without an explicit range; a section-
  less item covers none. Helpers `catalog.coveredUnits` / `catalog.sourceKind`
  (`resolve.go`); `unitRow` renders the single-unit notation via `FormatRef`.

**Web (`internal/web`):**
- Routes `POST /oturum/{id}/birim/{unitID}/sonuc` (Bugün) and
  `POST /hafta/oturum/{id}/birim/{unitID}/sonuc` (Hafta), both auth+CSRF.
  `applyUnitResult` parses `done` (button value) + optional D/Y/B (reusing
  `formOptInt`) and calls `SetUnitResult(SourceWeb, …)`; `todayUnitResult`
  swaps `#bugun-main` (htmx) / 303 → `/`; `weekUnitResult` reads the viewed week
  from `tarih` (FormValue — it rides the action query) and swaps `#hafta-main`
  / 303 → `/hafta?tarih=`. `buildTodayPage`/`buildWeekPage` now attach unit rows
  via `daySessionPtrs`/`weekSessionPtrs` so every render (initial + swap) carries
  them. `screenProgress` computes `unitAggregate` (total catalog units as the
  honest denominator; done + D/Y/B summed from `SourceProgress`).

**Views:**
- `views/units.templ`: `unitMarking`/`unitRow`/`dybField` — one form per unit,
  two `name="done"` submit buttons (Bitti/Kaydet + Geri al), D/Y/B omitted for
  video, done check + prefilled stored result; 44px (h-11) targets, soft tokens.
  `present.go` `unitResultForm(screen,…)` picks the Bugün/Hafta route + swap
  target. Wired into `today.templ` (tonight's `todaySession` + the Bugünün planı
  list) and `week.templ` (mobile list row + a colspan row under the desktop
  table row). İlerleme panel "Birim ilerlemesi" in `progress.templ`
  (`UnitAggregate` + `DonePct`/`Answered`/`HasUnits`), shown only when the
  catalog has units; the session completion control stays separate and primary.

**Tests (`internal/web/units_test.go`):** Bugün renders the unit list, exact
notation, Bitti control, D/Y/B inputs and the post route; marking a unit done
does NOT complete the session and completing the session does NOT mark units;
no-JS 303 + 403-without-CSRF on both Bugün and Hafta; İlerleme shows the
aggregate (2/3 birim, D/Y/B, Çözülen 10).

### Checks (2026-11-04, Go 1.27.3)
- `go tool templ generate` + Tailwind rebuilt (`web/static/app.css`).
- `go test ./internal/config/… ./internal/clock/… ./internal/store/…
  ./internal/plan/… ./internal/web/… ./internal/api/… ./internal/apidoc/…
  ./internal/mcpserver/… ./cmd/…` → all ok.
- `go build ./...` + `CGO_ENABLED=0 go build ./...` + `go vet ./...` green.
- gofmt: new/edited files clean on LF (working-tree CRLF = autocrlf artifact,
  as in WKY-26). NOT run: govulncheck (no dependency change). No commit/push.
- Visual check (built-in browser) at 390×844 and 1280×800: no horizontal
  overflow, Turkish glyphs correct, soft tokens, 44px controls; htmx Bitti swap
  marks the unit done and leaves the session uncompleted (Seans tile 0/1);
  Hafta table shows a colspan unit row under the session; İlerleme panel reads
  "Birim ilerlemesi · 2/3 birim · D/Y/B · Çözülen 10".
