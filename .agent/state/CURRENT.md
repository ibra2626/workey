# CURRENT — WKY-27 done (Bugün/Hafta/İlerleme unit view + marking); WKY-28 next

Updated 2026-11-04.
Active task: .agent/work/workey-v2-katalog/tasks/WKY-27.md (DONE)
Next: WKY-28 — Uçtan uca doğrulama ve v2 kapısı (end-to-end gate; depends on
WKY-24, 25, 27 — all now done).

## Feature: workey-v2-katalog (WKY-21..28)
Generalize workey from a fixed DGS/5-book tracker to a general source catalog
(Kaynak → Bölüm → Birim + separate unit progress + D/Y/B). Decisions in
`.agent/state/DECISIONS.md` (2026-11-02). Shared rules:
`.agent/work/workey-v2-katalog/README.md`. No data migration; schema rebuilt
clean (WKY-21). Service path (`SaveSource`/`Sources`, `SetUnitResult`,
`SourceProgress`/`SectionProgress`) since WKY-22/23; MCP WKY-24; REST WKY-25;
web Kaynaklar WKY-26; web unit view + marking WKY-27.

## WKY-27 result — 2026-11-04 (web unit view + marking)
Full detail in the task's Progress section. Summary:

**Plan read model:** `plan.UnitRow` + `SessionView.UnitRows []UnitRow`
(`json:"-"`, web-only — API/MCP unaffected). `Service.LoadUnitRows(ctx,
[]*SessionView)` resolves each session's covered units (unit range, or all of a
section item's units; section-less item → none) with exact per-unit notation and
current global done/D-Y-B, in one tx (catalog + one `UnitProgressByUnits`).
Helpers `catalog.coveredUnits`/`catalog.sourceKind` (`resolve.go`).

**Web:** routes `POST /oturum/{id}/birim/{unitID}/sonuc` and
`POST /hafta/oturum/{id}/birim/{unitID}/sonuc` (auth+CSRF). `applyUnitResult`
parses `done` (button value) + optional D/Y/B and calls `SetUnitResult(SourceWeb)`
— wholly independent of session completion. Bugün swaps `#bugun-main` / 303 → `/`;
Hafta reads the viewed week from `tarih` (FormValue, action query) and swaps
`#hafta-main` / 303 → `/hafta?tarih=`. `buildTodayPage`/`buildWeekPage` attach
rows for every render via `daySessionPtrs`/`weekSessionPtrs`. `screenProgress`
computes `unitAggregate` (total catalog units = honest denominator; done + D/Y/B
summed from `SourceProgress`).

**Views:** new `views/units.templ` (`unitMarking`/`unitRow`/`dybField`): one form
per unit, two `name="done"` submit buttons (Bitti/Kaydet + Geri al), D/Y/B omitted
for video, done check + prefilled result, 44px targets, soft tokens. Wired into
`today.templ` (todaySession + Bugünün planı list) and `week.templ` (mobile list
row + desktop colspan row). İlerleme panel "Birim ilerlemesi" in `progress.templ`
(`UnitAggregate`). Session completion control stays separate and primary.

**Tests:** `internal/web/units_test.go` — Bugün/Hafta render units + controls;
unit-done ≠ session-done (both directions); no-JS 303 + 403-without-CSRF; İlerleme
aggregates (2/3 birim, D/Y/B, Çözülen 10).

### Checks (2026-11-04, Go 1.27.3)
- templ generate + Tailwind rebuilt (`web/static/app.css`).
- `go test` across config/clock/store/plan/web/web·views/api/apidoc/mcpserver/cmd
  → all ok. `go build ./...` + `CGO_ENABLED=0 go build ./...` + `go vet ./...`
  green. gofmt: new/edited files clean on LF (working-tree CRLF = autocrlf
  artifact). NOT run: govulncheck (no dep change). No commit/push.
- Visual check (built-in browser) 390×844 + 1280×800: no overflow, Turkish
  glyphs, 44px controls; htmx Bitti marks a unit done and leaves the session
  uncompleted (Seans 0/1); Hafta colspan unit row; İlerleme "Birim ilerlemesi ·
  2/3 birim · Çözülen 10".

## Whole-tree build GREEN (since WKY-26)
`go build ./...` / `CGO_ENABLED=0` / `go vet ./...` pass across the module.
WKY-21..27 done; only WKY-28 (end-to-end gate) remains.

## Prior state (unchanged)
SF-01..06 study flow, study room, Plan calendar shipped on main (merged PR #1).
PRODUCT.md / DESIGN.md / `.impeccable/design.json` are v1 truth; the catalog
feature supersedes the v1 item model only. WKY-12 deploy held; WKY-20 200% zoom
pending. Branch: feat/v2-katalog. No commits for WKY-22..27 yet.
