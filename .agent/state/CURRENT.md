# CURRENT — WKY-07 done, MCP server live at /mcp

Updated 2026-10-26.

Active task: .agent/work/workey-v1/tasks/WKY-07.md (done; not committed)

- Feature: `workey-v1` — Go + templ + htmx + Tailwind single binary; REST `/api/v1`, MCP `/mcp` and web share the one business-logic layer `internal/plan`; SQLite (modernc). Local Windows only; WKY-12 deploy remains blocked by the user.
- Plan: `.agent/work/workey-v1/TASKS.md`; shared rules: `.agent/work/workey-v1/README.md`. Product: `PRODUCT.md`; visuals: `DESIGN.md`; decisions: `.agent/state/DECISIONS.md`.
- WKY-01..06 committed; HEAD `a2ff624` (WKY-06). WKY-07 done; this working tree is the WKY-07 change set, about to be committed. No push performed.

## WKY-07 result

MCP server in `internal/mcpserver` on the official Go SDK `github.com/modelcontextprotocol/go-sdk` v1.8.0, mounted at `/mcp` (Streamable HTTP, stateful default), wrapped with `httpx.SecurityHeaders` + `httpx.BearerAuth` using the same `WORKEY_API_TOKEN` as REST (one user, one token; token is a `New` argument so it can split later).

- `internal/mcpserver/mcpserver.go`: `New(svc, examDate, token) (http.Handler, error)`; registers six typed tools via `mcp.AddTool`; `mapError` renders plan errors as the REST JSON envelope (same `code`/`fields` paths) so they return as tool errors with `IsError` (handler returns a Go error → SDK packs it, output not marshalled).
- `internal/mcpserver/schema.go`: schemas inferred with `jsonschema.For` then enriched — `book` enum M1/M2/M3/G1/G2, `kind` enum aksam/uzun_blok/deneme, `YYYY-MM-DD` and `HH:MM` patterns, min/max bounds, English descriptions (Europe/Istanbul, references-only). `complete_session.done` optional, schema default `true`.
- Tools (mirror the service 1:1): `get_plan(from,to)`, `save_plan(SavePlanRequest)`, `complete_session(session_id, done=true)`, `get_outline()`, `save_outline(SaveOutlineRequest)`, `get_progress()`. Inputs reuse `plan` request types for the writes (zero mapping) + small local structs for the two simple reads; outputs are the service view/result types (structured output schemas). Chose `get_outline`/`save_outline` over the task's 3-tool split (recorded).
- `cmd/workey/main.go`: builds the MCP handler and mounts `/mcp`; `newMux(apiHandler, mcpHandler)`. `cmd/workey/main_test.go` updated for the second argument.

## Verification

Go 1.27.3; all passed on 2026-10-26:

- `go test ./internal/mcpserver/... ./internal/api/... ./internal/plan/... ./cmd/...` — pass. mcpserver tests use an in-process SDK client over Streamable HTTP: list tools, save_outline+save_plan, save_plan again (all unchanged), get_plan, complete_session (today), get_progress, default-done, tool-error code (`future`), and 401 without a token.
- `go vet ./...` — exit 0.
- `gofmt -l .` — empty, exit 0.
- `go build ./...` — exit 0.
- `CGO_ENABLED=0 go build ./...` — exit 0.
- No templ/Tailwind regeneration: no UI source changed. `go.mod`/`go.sum` grew (go-sdk + jsonschema-go v0.4.3 + oauth2/time/jwt/segmentio transitive deps); `go mod tidy` run.

## Next step

WKY-08: Web auth (Giriş screen, session cookie, CSRF), depends on WKY-02, WKY-03. Start a new session with `WKY-08 görevine devam et`; read the index and that task. Note WKY-09/10/11 (screens) and WKY-13 (end-to-end Claude Code MCP writes a week) depend on this MCP work.

WKY-07 paths (committed with this change set): `internal/mcpserver/`, `cmd/workey/main.go`+`main_test.go`, `go.mod`, `go.sum`, and the `.agent/` records. Push and PR still require an explicit user request.
