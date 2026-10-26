# WKY-07 — MCP server (Streamable HTTP)

## Goal
An MCP endpoint in the same binary at `/mcp` (Streamable HTTP, official Go SDK) whose tools an AI client can call correctly from their descriptions and schemas alone.

## Scope
- `github.com/modelcontextprotocol/go-sdk` (latest stable; read its README/examples for the current Streamable HTTP handler API before coding). DECISIONS entry with the version and reason.
- `internal/mcpserver`, source = `mcp`, Bearer middleware from `internal/httpx`.
- Tools (names final unless a DECISIONS entry changes them):
  - `get_plan(from, to)` — sessions with items, state, notation.
  - `save_plan(from, to, sessions[], idempotency_key?)` — declarative range replace; description must state: past days and completed sessions are never changed and come back in `skipped`; sending the same request twice is safe; omitted open future sessions in the range are deleted; `makeup_of` links a makeup session to a missed one.
  - `complete_session(session_id, done=true|false)` — today/past only.
  - `get_topic_groups()` / `save_topic_groups(groups[])`, `save_phases(phases[])` (or one `save_outline`; pick and record).
  - `get_progress()` — progress summary.
- Input schemas: required fields, enums (`book`: M1 M2 M3 G1 G2; `kind`: aksam uzun_blok deneme), formats (`date`, `HH:MM` pattern), minimum/maximum, examples in descriptions; structured output schemas where the SDK supports them. Descriptions in English, state the timezone (Europe/Istanbul) and that only book/chapter/page references are stored, never topic text or questions.
- Errors from the service are returned as tool errors with the same code/field paths as REST.
- Non-goals: MCP resources/prompts, OAuth.

## Acceptance criteria
- [ ] In-process test with the SDK client over Streamable HTTP: list tools, call `save_plan` with the sample week twice (second all unchanged), `get_plan`, `complete_session`, `get_progress`.
- [ ] 401 without token on `/mcp`.
- [ ] Description review: a reader who sees only `tools/list` output can build a valid `save_plan` call (paste the output into the Progress section once, shortened).
- [ ] No business logic in tool handlers.

## Checks
- `go test ./internal/mcpserver/...`
- `go vet ./...`

## Progress
Not started.
