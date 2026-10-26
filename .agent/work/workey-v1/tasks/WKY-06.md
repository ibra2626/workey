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
- [ ] httptest suite: 401 without/with wrong token; full round trip with `testdata/sample-week.json`; PUT twice → second response reports all unchanged; idempotency key replay; validation error shape with field paths; complete future → 409/422 with code.
- [ ] Handlers contain no business rules (review: every branch maps to a service call or an error mapping).
- [ ] `docs/api.md` matches the implemented routes.

## Checks
- `go test ./internal/httpx/... ./internal/api/...`
- `go vet ./...`

## Progress
Not started.
