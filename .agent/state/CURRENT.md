# CURRENT — WKY-06 done, protected REST API live

Updated 2026-10-26.

Active task: .agent/work/workey-v1/tasks/WKY-06.md (done; not committed)

- Feature: `workey-v1` — Go + templ + htmx + Tailwind single binary; REST `/api/v1` and MCP `/mcp` share `internal/plan`; SQLite (modernc). Local Windows only; WKY-12 deploy remains blocked by the user.
- Plan: `.agent/work/workey-v1/TASKS.md`; shared rules: `.agent/work/workey-v1/README.md`. Product: `PRODUCT.md`; visuals: `DESIGN.md`; decisions: `.agent/state/DECISIONS.md`.
- WKY-01..05 committed; HEAD `e5fd26e` (WKY-05). WKY-06 done, uncommitted; TASKS updated. No commit/push performed.

## WKY-06 result

- `internal/httpx/{httpx.go,httpx_test.go}`: reusable Bearer auth (SHA-256 digests + constant-time comparison; empty configured token fails closed), 401 challenge, JSON error envelope, request body cap and security headers.
- `internal/api/{api.go,api_test.go}`: nine routes — GET/PUT plan, GET/PUT outline, GET today/week/progress, POST/DELETE session completion. Input parsing and service error mapping only; write source `api`. Optional `Idempotency-Key` header overrides the body key.
- Errors: malformed/over-limit JSON 400; validation 422 (including field code `future`); all-not-found 404; conflicting idempotency key 409; unexpected errors generic 500. API cap 1 MiB; no-store/nosniff/no-referrer headers.
- `internal/plan/read.go`: `Range`, `Outline`, `Today`, `TodayDate`, outline read types, exported `ParseDate`. State/notation/calendar stay in the service.
- `cmd/workey/main.go`: opens/migrates SQLite at startup, default path `workey.db`, creates service/API and mounts `/api/v1/`. `main_test.go` adjusted for the mux argument.
- `docs/api.md`: auth, dates/times, routes, examples, idempotency and errors. Acceptance criteria checked; handler boundaries and documented routes reviewed.

## Verification

Go 1.27.3; all passed on 2026-10-26:

- `go test ./internal/httpx/... ./internal/api/... ./internal/plan/... ./cmd/...` — all four packages pass (fixture round trip, auth, unchanged repeat, key replay/conflict, field errors, completion/undo/future/not-found).
- `go vet ./...` — exit 0.
- `gofmt -l .` — empty, exit 0.
- `go build ./...` — exit 0.
- `CGO_ENABLED=0 go build ./...` — exit 0; PowerShell used `$env:CGO_ENABLED = '0'; go build ./...`.
- Both builds emitted a non-fatal module stat-cache write permission warning under the user Go cache; compilation succeeded.
- No templ/Tailwind regeneration: no UI source changed.

## Next step

WKY-07: MCP server (Streamable HTTP), depending on WKY-06. Start a new session with `WKY-07 görevine devam et`; read the index and that task. Reuse `internal/httpx` and `plan.Service`; preserve the single business-logic layer and injectable Istanbul clock.

Pending WKY-06 paths: `cmd/workey/main.go`, `cmd/workey/main_test.go`, `internal/plan/read.go`, new `internal/api/`, new `internal/httpx/`, new `docs/api.md`, and four updated `.agent/` records. Release actions require an explicit user request.
