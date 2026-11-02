# CURRENT — WKY-25 done (REST catalog endpoints); WKY-26 next

Updated 2026-11-02.
Active task: .agent/work/workey-v2-katalog/tasks/WKY-25.md (DONE)
Next: WKY-26 — Web: Kaynaklar düzeltme sayfası ve D/Y/B raporu (`internal/web`,
`internal/web/views`). WKY-26 also restores that layer's compile against the
catalog `plan.Item`/notation and regenerates the verbatim-synced usage docs
(study-contract / planning-inputs) + the in-app API usage guide text.

## Feature: workey-v2-katalog (WKY-21..28)
Generalize workey from a fixed DGS/5-book tracker to a general source catalog
(Kaynak → Bölüm → Birim) with separate unit progress + D/Y/B. Decisions (WKY-21
..24) in `.agent/state/DECISIONS.md` (2026-11-02). Shared rules:
`.agent/work/workey-v2-katalog/README.md`. No data migration; schema rebuilt
clean (WKY-21). Service path (`SaveSource`/`Sources`, `SetUnitResult`) in place
since WKY-22/23; MCP exposes it (WKY-24); REST exposes it now (WKY-25).

## WKY-25 result — 2026-11-02 (REST catalog endpoints + plan item contract)
`internal/api` + `internal/apidoc` build/vet/test green. Full detail: the task's
Progress section. Summary:
- **`internal/api/api.go`:** new thin routes `GET /api/v1/sources`
  (`svc.Sources`), `POST /api/v1/sources` (`svc.SaveSource`, Idempotency-Key
  header), `POST /api/v1/units/{id}/result` (`svc.SetUnitResult`, unit id from
  path, body `{done, correct?, wrong?, blank?}`). `sessionID` → shared
  `pathID(w,r,label)`. plan/outline PUT already carried the catalog shapes.
  Chose POST-on-collection over PUT /{key} so handlers stay parse+call.
- **`internal/apidoc`:** Operations + schemas `Catalog`/`SaveSource`/
  `SaveSourceResult`/`UnitResultView` + body-only `UnitResult`; `requestSchema`
  on the two new inputs; numeric constraints (ordinal/number/unit_id ≥1,
  correct/wrong/blank ≥0, done_at date-time). `exampleM1Source()` (stable ids)
  backs Catalog + SaveSourceResult examples; `ExampleSource`/`ExampleUnitResult`
  request examples. Fixed the `Session` example `section_name` → "Sayı
  basamakları" (resolved from the catalog, not the topic group).
- **Tests:** seeds POST sources first in api_test/study_contract; new
  `TestCatalogRoundTrip` (write→read→unit result `M1·3, Antrenman 1`→re-write
  keeps unit id→no-op; id 0=422, unknown id=404) and `TestPlanUnknownSource`
  (422 validation_failed, field `sessions[0].items[0].source_key`,
  code `unknown_source`). `TestOpenAPIContract` seeds the catalog + SaveSource/
  Catalog example drift.
- **docs/api.md:** agent-guide fields (source_kinds/location_units), plan item
  example (source/section/loc), outline+progress `question_*`, three new route
  sections. Verbatim-synced study-contract/planning-inputs blocks left for
  WKY-26/27.

**Interpretation:** the WKY-25 criterion "unknown source → 400" is implemented as
the service's typed `422 validation_failed` (unknown_source ≠ not_found),
surfaced unchanged by `writePlanError` — same as MCP; a 400 would need API-layer
re-checking the task forbids. In-app API guide text deferred to WKY-26/27 (lives
in the still-red `internal/web/views/usage_content.go`).

### Checks (2026-11-02, Go 1.27.3)
- `go test -count=1 ./internal/api/... ./internal/apidoc/...` ok.
- `go build` + `CGO_ENABLED=0 go build` + `go vet` of api+apidoc ok.
- `go test ./internal/mcpserver/... ./internal/plan/...` green (unaffected).
- gofmt: changed files clean on LF (working-tree CRLF = autocrlf artifact).
  NOT run: govulncheck (no dep change); whole-tree `go build ./...` (still red by
  design in `internal/web/views` → web, cmd until WKY-28). No commit/push.

## Whole-tree build is RED by design (user split, 2026-11-02)
Changing `plan.Item` broke the v1-concept layers; each downstream task restores
its own layer. Done: WKY-24 (`internal/mcpserver`, `internal/apidoc`), WKY-25
(`internal/api`). Remaining red: `internal/web`, `internal/web/views`
(present.go, study.go, *.templ; usage_content.go; the verbatim-synced usage
docs) → WKY-26/27. `go build ./...` / `go vet ./...` go green again at WKY-28.
Details: WKY-22.md "Scope split" and TASKS.md.

## Prior state (unchanged)
SF-01..06 study flow, study room, Plan calendar shipped on main (merged PR #1).
PRODUCT.md / DESIGN.md / `.impeccable/design.json` are v1 truth; the catalog
feature supersedes the v1 item model only. WKY-12 deploy held; WKY-20 200% zoom
pending. Branch: feat/v2-katalog. No commits for WKY-22/23/24/25 yet.
