# CURRENT — WKY-23 done (unit progress + D/Y/B service path); WKY-24 next

Updated 2026-11-02.
Active task: .agent/work/workey-v2-katalog/tasks/WKY-23.md (DONE)
Next: WKY-24 — MCP tools (`save_source`/`get_sources`/`set_unit_result`) +
`internal/mcpserver`/`internal/apidoc`, which also restores that layer's build
against the catalog Item/notation (see "Whole-tree build is red by design").

## Feature: workey-v2-katalog (WKY-21..28)
Generalize workey from a fixed DGS/5-book tracker to a general source catalog
(Kaynak → Bölüm → Birim) with separate unit progress + D/Y/B. Direction + WKY-21
+ WKY-22 + WKY-23 decisions: `.agent/state/DECISIONS.md` (four 2026-11-02
entries). Shared rules: `.agent/work/workey-v2-katalog/README.md`. No data
migration; schema rebuilt clean (WKY-21).

## WKY-23 result — 2026-11-02 (unit progress + D/Y/B service path)
A business-logic path to mark a unit "bitti" with optional D/Y/B, independent of
session completion. Details: DECISIONS.md "WKY-23" + the task's Progress section.
- **Store** (`internal/store/sources.go`): added `UnitByID` (unit joined to its
  section ordinal + source id/key/name/kind/location_unit), `UnitProgressByUnits`
  (read-by-unit-set; empty input → empty map), and aggregates
  `UnitProgressTotalsBySource`/`...BySection` (shared `unitProgressTotals`,
  `COUNT` + `COALESCE(SUM(...),0)`) with a `UnitProgressTotals` struct. WKY-21's
  `unit_progress` CRUD (`UpsertUnitProgress`/`DeleteUnitProgress`/
  `UnitProgressByUnit`/`...BySource`) reused as-is.
- **Plan** (`internal/plan/units.go`, new): `SetUnitResult(source,
  UnitResultRequest{UnitID, Done, Correct/Wrong/Blank *int})` — done upserts
  (sticky first `done_at`, optional D/Y/B, writing source), undo clears.
  Idempotent no-op (compare done_at + D/Y/B + completed_source) writes nothing +
  no audit; a real change writes one `audit_log` row (entity `unit`, actions
  `complete`/`undo` reused from complete.go, before/after JSON, NULL request
  hash). `checkNonNeg` rejects negative D/Y/B; a `video` source rejects any D/Y/B
  but can still be marked done (null). Returns `UnitResultView` (identity +
  `FormatRef` notation + result). Reads `UnitResults(unitIDs)`,
  `SourceProgress()`, `SectionProgress()` map store aggregates → plan types
  `UnitResult` / `ProgressTotals` (web never sees store types).
- **Tests** (`internal/plan/units_test.go`): done ±D/Y/B; idempotent (stable
  done_at, 1 audit row); D/Y/B change keeps done_at + audits; undo clears + second
  undo no-op; negative rejected; missing unit; video rejects D/Y/B yet marks done;
  two-direction independence; source/section aggregates + read-by-units. Helpers
  `sectionUnitIDs`/`sectionID`/`seedVideoUnit`/`unitAuditCount` (reuse
  `seedCatalog`, `ptr` from notation_test.go).

### Checks (2026-11-02, Go 1.27.3)
- `go test ./internal/plan/... ./internal/store/...` ok.
- `go vet ./internal/plan/... ./internal/store/...` clean.
- `go build` + `CGO_ENABLED=0 go build` of plan+store ok.
- `gofmt -l` on the three changed files prints nothing (whole-tree `gofmt -l .`
  still flags pre-existing CRLF working-tree artifacts; environmental only).
- NOT run: whole-tree build/vet (web/transport red by design), govulncheck. No
  commit/push.

## WKY-22 result — 2026-11-02 (plan domain bound to the catalog)
Rebound `internal/plan` to the WKY-21 catalog store. `plan.Item = {SourceKey,
SectionOrdinal *int, UnitFrom/UnitTo *int, LocFrom/LocTo *int}`; `kind` a free
label; `Question*` rename. `resolve.go` per-tx `catalog` snapshot resolves refs
(read `ItemView`, write `store.SessionItem`, typed errors
`unknown_source`/`unknown_section`/`unit_out_of_range`). `notation.go` renders
`M1·4, Antrenman 5–7, Test 1–2` (U+00B7 / U+2013, join " · "). SavePlan validates
items in-tx before any write. Catalog write-path inputs (`SaveSourceRequest`)
defined for WKY-24/25, unused there. Full detail: DECISIONS.md "WKY-22".

## Whole-tree build is RED by design (user split, 2026-11-02)
Changing `plan.Item` breaks the v1-concept layers; the user chose to split rather
than gut WKY-24..27's work. Each also restores its layer's compile against the
catalog Item/notation:
- WKY-24 → `internal/mcpserver`, `internal/apidoc`
- WKY-25 → `internal/api`
- WKY-26/27 → `internal/web`, `internal/web/views` (present.go, study.go, *.templ;
  the akşam konu/soru/yanlış-defteri profile and page/stage bookmark rendering are
  redefined there)
`go build ./...` / `go vet ./...` go green again at WKY-28. Details: WKY-22.md
"Scope split" and TASKS.md. The catalog unit-write path (SetUnitResult) and its
reads are now in place for WKY-24 (MCP `set_unit_result`) and WKY-27 (web
controls) to expose.

## Prior state (unchanged)
SF-01..06 study flow, study room, Plan calendar shipped on main (merged PR #1).
PRODUCT.md / DESIGN.md / `.impeccable/design.json` are v1 truth; the catalog
feature supersedes the v1 item model only. WKY-12 deploy held; WKY-20 200% zoom
pending. Branch: feat/v2-katalog. No commits for WKY-22/23 yet.
