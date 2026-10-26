# WKY-06 — REST API /api/v1 with Bearer auth

## Goal
A JSON API exposing the same operations as MCP, thin over `internal/plan`, protected by a constant-time Bearer token check.

## Scope
- `internal/httpx`: Bearer middleware (`WORKEY_API_TOKEN`, `crypto/subtle.ConstantTimeCompare` on SHA-256 digests so length does not leak; `WWW-Authenticate` on 401), JSON error envelope `{error:{code,message,fields[]}}`, body size limit, security headers. Reused by WKY-07.
- `internal/api` routes (Go 1.22+ method patterns), source = `api`:
  - `GET /api/v1/plan?from=YYYY-MM-DD&to=YYYY-MM-DD` — sessions with items, state, notation.
  - `PUT /api/v1/plan` — SavePlan body; optional `Idempotency-Key` header.
  - `POST /api/v1/sessions/{id}/complete`, `DELETE /api/v1/sessions/{id}/complete` (undo).
  - `GET /api/v1/outline`, `PUT /api/v1/outline` — phases and topic groups.
  - `GET /api/v1/progress` — progress summary.
  - `GET /api/v1/today`, `GET /api/v1/week?date=` — the read models the screens use.
- Dates are civil dates (`YYYY-MM-DD`), times `HH:MM`, all Europe/Istanbul; documented in `docs/api.md` with request/response examples taken from tests.
- Non-goals: web cookie auth, MCP.

## Acceptance criteria
- [x] httptest suite: 401 without/with wrong token; full round trip with `testdata/sample-week.json`; PUT twice → second response reports all unchanged; idempotency key replay; validation error shape with field paths; complete future → 409/422 with code.
- [x] Handlers contain no business rules (review: every branch maps to a service call or an error mapping).
- [x] `docs/api.md` matches the implemented routes.

## Checks
- `go test ./internal/httpx/... ./internal/api/...`
- `go vet ./...`

## Progress
Done 2026-10-26; not committed.

- `internal/httpx`: reusable Bearer auth using SHA-256 digests and constant-time comparison, fail-closed empty token, 401 challenge, JSON errors, body limit and security headers.
- `internal/api`: all nine method/path routes; handlers parse, call `plan.Service` with source `api`, render. Validation maps to 422 (future completion has field code `future`), all-not-found to 404, idempotency conflict to 409, unexpected failures to generic 500. Body cap: 1 MiB.
- `internal/plan/read.go`: added `Range`, `Outline`, `Today`, `TodayDate`, outline read models and exported `ParseDate`; state, notation and calendar logic stay in the service.
- `cmd/workey`: startup opens/migrates SQLite (default `workey.db`) and mounts the protected API. `docs/api.md` checked against routes, models and fixture tests.
- Checks passed with Go 1.27.3: `go test ./internal/httpx/... ./internal/api/... ./internal/plan/... ./cmd/...`, `go vet ./...`, `gofmt -l .` (empty), `go build ./...`, `CGO_ENABLED=0 go build ./...` (PowerShell environment assignment). Both builds exited 0 despite a non-fatal Go module stat-cache write permission warning.
- No templ/Tailwind regeneration needed: no UI sources changed. CURRENT, DECISIONS and TASKS updated. Next: WKY-07; no commit/push performed.
