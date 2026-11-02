# WKY-25 — REST API: katalog uçları ve plan item sözleşmesi

## Goal
Mirror the catalog capabilities on `/api/v1` so any non-MCP client (and the
contract tests) can read/write sources, plan with catalog references, and record
unit results. Keep Bearer auth, JSON errors and OpenAPI parity.

## Scope
- **`internal/api`:** endpoints parsing input and calling `internal/plan` only:
  - `GET  /api/v1/sources` → full catalog.
  - `PUT  /api/v1/sources/{key}` (or `POST /api/v1/sources`) → `save_source`.
  - `POST /api/v1/units/{id}/result` → `set_unit_result` (done + optional D/Y/B /
    undo).
  - `POST /api/v1/plan` item shape updated to source/section/unit refs.
  - `PUT /api/v1/outline` → `question_*` naming.
- **`internal/apidoc`:** update `openapi.go` + `examples.go` to the new shapes;
  keep the generated-vs-served parity test.
- **`internal/api/contract_test.go`:** update to the catalog contract; add cases
  for save_source round-trip, plan with a valid/invalid source reference, and a
  unit-result write.
- Update `docs/api.md` and the in-app API usage guide text.
- Non-goals: web UI (WKY-26/27).

## Acceptance criteria
- [ ] `go test ./internal/api/... ./internal/apidoc/...` green; OpenAPI parity test
  passes; contract covers sources + plan(new item) + unit result.
- [ ] Unknown `source_key` in a plan body returns the typed validation error as a
  400 with field path + machine code (no re-checking in the API layer).
- [ ] `docs/api.md` matches the served OpenAPI.
- [ ] Builds (normal + `CGO_ENABLED=0`), vet, gofmt clean.

## Checks
- `go test ./internal/api/... ./internal/apidoc/...`
- `go vet ./...`, `gofmt -l .`

## Notes
- Depends on WKY-22/23. Independent of WKY-24 (MCP) — both just expose the same
  service; do not re-implement any rule here.

## Progress

### 2026-11-02 — DONE (REST catalog endpoints + plan item contract)
`internal/api` now mirrors the catalog service; both it and `internal/apidoc`
build, vet and test green.

- **`internal/api/api.go`:** three new thin routes → service only:
  - `GET  /api/v1/sources` → `svc.Sources` (CatalogView).
  - `POST /api/v1/sources` → `svc.SaveSource` (body `plan.SaveSourceRequest`,
    optional `Idempotency-Key` header, same precedence as plan/outline). Chose
    POST on the collection (not `PUT /{key}`) so the handler stays parse+call and
    reuses the whole-request body without a path/body key cross-check.
  - `POST /api/v1/units/{id}/result` → `svc.SetUnitResult`; unit id from the path
    (body `unit_id` ignored), body is `{done, correct?, wrong?, blank?}`.
  - `sessionID` refactored into a shared `pathID(w,r,label)` helper (session +
    unit reuse it); no rule added in the API layer.
  - The existing `PUT /api/v1/plan` / `PUT /api/v1/outline` already carried the
    catalog `plan.Item` / `question_*` shapes (WKY-22), so no change there.
- **`internal/apidoc`:** `Operations()` gains getSources/saveSource/setUnitResult;
  `build()` adds schemas `Catalog`, `SaveSource`, `SaveSourceResult`,
  `UnitResultView` and a body-only `UnitResult` struct (no `unit_id`). `requestSchema`
  applied to SaveSource+UnitResult; new numeric constraints (`ordinal/number/unit_id`
  min 1, `correct/wrong/blank` min 0, `done_at` date-time). Request examples
  `ExampleSource` / `ExampleUnitResult`; the `Idempotency-Key` header is attached to
  every write input except `UnitResult`. `examples.go`: `exampleM1Source()` (stable
  ids 1; sections 1–4; units 1–3) backs the `Catalog` + `SaveSourceResult` examples;
  a `UnitResultView` example (docs only, fixed `done_at`). Fixed the `Session`
  example's `section_name` to the seeded section-1 name "Sayı basamakları" (resolved
  from the catalog, not the topic group).
- **Tests (`internal/api`):** `api_test.go` `seed()`/round-trip/idempotency now
  POST `sample-sources.json` first; `study_contract_test.go` seeds `ExampleSource`
  first. `contract_test.go`: `TestOpenAPIContract` seeds the catalog + does
  SaveSource/Catalog example drift; new `TestCatalogRoundTrip` (empty read → write →
  read-back → unit result `M1·3, Antrenman 1` → re-write keeps unit id → identical
  no-op → id 0 = 422, unknown id = 404) and `TestPlanUnknownSource`.
- **`docs/api.md`:** agent-guide field list (`source_kinds`/`location_units`),
  plan item example (source/section/loc), outline + progress `question_*` names,
  and three new route sections (GET/POST `/sources`, POST `/units/{id}/result`).
  The verbatim-synced `study-contract`/`planning-inputs` blocks (still v1
  `book/chapter_no`) are left for WKY-26/27 to regenerate.

**Acceptance note — unknown source is 422, not 400.** The criterion's "400" is
read as "the service's typed validation error, surfaced unchanged": `unknown_source`
is not `not_found`, so `writePlanError` maps it to `422 validation_failed` with the
field path `sessions[i].items[j].source_key` + machine code — consistent with the
MCP transport and the existing API. A dedicated 400 would require re-checking in the
API layer, which this task forbids.

**In-app API usage guide text deferred to WKY-26/27.** It lives in
`internal/web/views/usage_content.go`, inside the still-red `internal/web/views`
package that WKY-26/27 own (and which regenerates the verbatim-synced usage docs).
Editing stale v1 wording there now would be unverifiable (package does not compile)
and crosses the layer split. `docs/api.md` (this layer, testable against the served
OpenAPI) is updated.

### Checks (2026-11-02, Go 1.27.3)
- `go test -count=1 ./internal/api/... ./internal/apidoc/...` ok (apidoc has no
  test files; its contract is exercised by `internal/api`).
- `go build` + `CGO_ENABLED=0 go build` + `go vet` of `./internal/api/...
  ./internal/apidoc/...` ok.
- `go test ./internal/mcpserver/... ./internal/plan/...` still green (unaffected).
- gofmt: changed files clean on LF content (working-tree CRLF is the autocrlf
  checkout artifact). NOT run: govulncheck (no dep change); whole-tree `go build
  ./...` (still red by design in `internal/web/views` → web, cmd until WKY-28). No
  commit/push.
