# CURRENT — WKY-22 done (plan domain on catalog); WKY-23 next

Updated 2026-11-02.
Active task: .agent/work/workey-v2-katalog/tasks/WKY-22.md (DONE)
Next: WKY-23 — unit progress + D/Y/B service path (independent of session
completion). Then WKY-24/25 (transport) and WKY-26/27 (web), which also restore
their layer's build (see "Whole-tree build is red" below).

## Feature: workey-v2-katalog (WKY-21..28)
Generalize workey from a fixed DGS/5-book tracker to a general source catalog
(Kaynak → Bölüm → Birim) with separate unit progress + D/Y/B. Direction + WKY-21
+ WKY-22 decisions: `.agent/state/DECISIONS.md` (three 2026-11-02 entries). Shared
rules: `.agent/work/workey-v2-katalog/README.md`. No data migration; schema rebuilt
clean (WKY-21).

## WKY-22 result — 2026-11-02 (plan domain bound to the catalog)
Rebound `internal/plan` to the WKY-21 catalog store. Key shapes:
- `plan.Item = {SourceKey string, SectionOrdinal *int, UnitFrom/UnitTo *int,
  LocFrom/LocTo *int}`. UnitFrom/To are 1-based unit ordinals within the section;
  Loc* are raw numbers in the source's location unit. Removed `validBooks`; `kind`
  is now a free label (no enum gate). `TopicGroupInput` + read models renamed
  `DGSQuestion*`/`DGSMin/Max` → `Question*`/`QuestionMin/Max`.
- New `internal/plan/resolve.go`: per-tx `catalog` snapshot (`loadCatalog`).
  `resolve`→`ItemRef` (notation), `itemView`→`ItemView` (read), `resolveWrite`→
  `store.SessionItem` with typed errors `unknown_source`/`unknown_section`/
  `unit_out_of_range` (codes added in errors.go).
- `notation.go` rewritten: `FormatRef`/`FormatRefs(ItemRef)` render e.g.
  `M1·4, Antrenman 5–7, Test 1–2` — source key·section ordinal, unit label groups
  (same-label runs collapse to a number range via each unit's `number`), then the
  location range with a unit prefix (page→"s. ", minute→"dk ", lesson→"ders ",
  other→none). U+00B7 / U+2013 preserved; multi-item join " · ".
- `service.go` SavePlan: resolves/validates items in-tx before any write; replace,
  idempotency, frozen-past, audit semantics unchanged. `itemContentEqual` and
  `snapshotJSON` on the new fields. `validate.go`: structural item checks (source
  key required; unit range needs a section; ordered ranges; needs a unit or loc
  range). `read.go`: joins the catalog; `SessionView.Items` is now `[]ItemView`;
  added `ItemView`. `outline.go`: Question* store fields. `study_view.go`: bookmark
  resolved via catalog (`BookmarkView.Item` is `ItemView`, Notation via FormatRef).
- `study.go` `validateBookmark`: `ResumePage` is now a location value validated
  against the item's loc range; the v1 konu/soru sub-range narrowing is gone; the
  stage gate stays (konu/soru/yanlis_defteri for akşam, calisma otherwise). Fuller
  stage semantics revisited in WKY-27.
- `plan.go` also defines the catalog write-path inputs for WKY-24/25:
  `SaveSourceRequest{Source SourceInput, Sections []SectionInput}` (each section
  carries `[]UnitInput`); unused this task.
- Tests: `testdata/sample-*.json` moved to the catalog shape. Plan tests reworked
  — `seedCatalog`/`sourceID` helpers seed sources M1 (sections 1–4, Antrenman units
  1–3 under section 3) and G1 via the store; `loadWeek(t, svc, st)` now seeds first;
  new validation cases; notation tests use resolved `ItemRef` with Turkish labels.

### Checks (2026-11-02, Go 1.27.3)
- `go test ./internal/plan/... ./internal/store/...` ok.
- `go vet ./internal/plan/... ./internal/store/...` clean.
- `go build` + `CGO_ENABLED=0 go build` of plan+store ok.
- gofmt-clean on LF-normalized content; `gofmt -l` still flags pre-existing CRLF
  working-tree artifacts (core.autocrlf=true; git stores LF) — environmental only.
- NOT run: whole-tree build/vet (web/transport deferred — red by design),
  govulncheck. No commit/push.

## Whole-tree build is RED by design (user split, 2026-11-02)
Changing `plan.Item` breaks the v1-concept layers; the user chose to split rather
than gut WKY-24..27's work. Each now also restores its layer's compile against the
catalog Item/notation:
- WKY-24 → `internal/mcpserver`, `internal/apidoc`
- WKY-25 → `internal/api`
- WKY-26/27 → `internal/web`, `internal/web/views` (present.go, study.go, *.templ;
  the akşam konu/soru/yanlış-defteri profile and page/stage bookmark rendering are
  redefined there)
`go build ./...` / `go vet ./...` go green again at WKY-28. Details: WKY-22.md
"Scope split" and TASKS.md.

## Prior state (unchanged)
SF-01..06 study flow, study room, Plan calendar shipped on main (merged PR #1).
PRODUCT.md / DESIGN.md / `.impeccable/design.json` are v1 truth; the catalog feature
supersedes the v1 item model only. WKY-12 deploy held; WKY-20 200% zoom pending.
Branch: feat/v2-katalog. No commits for WKY-22 yet.
