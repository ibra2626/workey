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
(planned)
