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
- [x] In-process test with the SDK client over Streamable HTTP: list tools, call `save_plan` with the sample week twice (second all unchanged), `get_plan`, `complete_session`, `get_progress`. — `TestListTools`, `TestSavePlanIdempotentThenReads`.
- [x] 401 without token on `/mcp`. — `TestUnauthorized` (SDK client with no Bearer header fails to connect).
- [x] Description review: a reader who sees only `tools/list` output can build a valid `save_plan` call. — `TestSavePlanSchemaSelfDescribing`; shortened output below.
- [x] No business logic in tool handlers. — handlers only parse args, call `plan` with source `mcp`, and map errors; all rules stay in `internal/plan`.

## Checks
- `go test ./internal/mcpserver/...`
- `go vet ./...`

## Progress

Done 2026-10-26. MCP server in `internal/mcpserver` on the official Go SDK v1.8.0, mounted at `/mcp` (Streamable HTTP) behind `httpx.SecurityHeaders` + `httpx.BearerAuth` using `WORKEY_API_TOKEN`. See the DECISIONS entry (SDK version/reason, tool naming, token reuse).

Files:
- `internal/mcpserver/mcpserver.go` — `New(svc, examDate, token)` → `http.Handler`; registers six tools; `mapError` renders plan errors as the REST JSON envelope (same codes/field paths) so they surface as tool errors with `IsError`.
- `internal/mcpserver/schema.go` — schema builders: enums (`book` M1/M2/M3/G1/G2, `kind` aksam/uzun_blok/deneme), date/time patterns, bounds, English descriptions (Europe/Istanbul, references-only); `complete_session.done` optional, default true.
- `internal/mcpserver/mcpserver_test.go` — in-process SDK client over Streamable HTTP.
- `cmd/workey/main.go` — builds the MCP handler and mounts `/mcp`; `newMux(api, mcp)`.

Tools: `get_plan`, `save_plan`, `complete_session`, `get_outline`, `save_outline`, `get_progress`.

Verification (Go 1.27.3, all pass): `go test ./internal/mcpserver/... ./internal/api/... ./internal/plan/... ./cmd/...`; `go vet ./...`; `gofmt -l .` empty; `go build ./...`; `CGO_ENABLED=0 go build ./...`.

### `save_plan` from `tools/list` (shortened)

Description:
> Declaratively replace the open sessions for every date in [from, to] (Europe/Istanbul). For each date in range that is today or later, the open (uncompleted) sessions are made to match exactly the given set, keyed by date + start_time: a matching session is updated in place, a new one is created, and an open future session omitted from the request is deleted. Past days and completed sessions are frozen — … returned in 'skipped' … . The call is safe to repeat … 'unchanged' (pass idempotency_key to also replay …). Link a makeup session … with 'makeup_of' (the missed session's date + start_time). Only book/chapter/page references are stored, never topic text or question content.

Input schema (trimmed):
```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["from", "to", "sessions"],
  "properties": {
    "from": { "type": "string", "pattern": "^\\d{4}-\\d{2}-\\d{2}$", "description": "Inclusive start date 'YYYY-MM-DD' (Europe/Istanbul). Example: 2026-10-26." },
    "to":   { "type": "string", "pattern": "^\\d{4}-\\d{2}-\\d{2}$", "description": "Inclusive end date 'YYYY-MM-DD'." },
    "idempotency_key": { "type": "string", "description": "Optional; repeat replays the first result." },
    "sessions": {
      "type": "array",
      "description": "The complete set of open sessions for every date in [from, to]; omitted open future sessions are deleted.",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["date", "start_time", "duration_min", "kind", "items"],
        "properties": {
          "date":        { "type": "string", "pattern": "^\\d{4}-\\d{2}-\\d{2}$" },
          "start_time":  { "type": "string", "pattern": "^([01]\\d|2[0-3]):[0-5]\\d$" },
          "duration_min":{ "type": "integer", "minimum": 1, "maximum": 1440 },
          "kind":        { "type": "string", "enum": ["aksam", "uzun_blok", "deneme"] },
          "topic_group": { "type": "string", "description": "Optional topic-group key from save_outline." },
          "makeup_of":   { "type": "object", "properties": { "date": {"pattern":"…"}, "start_time": {"pattern":"…"} }, "description": "Link to the missed session (past, uncompleted)." },
          "items": {
            "type": "array", "minItems": 1,
            "items": {
              "type": "object",
              "additionalProperties": false,
              "required": ["book", "chapter_no", "chapter_title", "page_from", "page_to"],
              "properties": {
                "book":          { "type": "string", "enum": ["M1","M2","M3","G1","G2"] },
                "chapter_no":    { "type": "integer", "minimum": 1 },
                "chapter_title": { "type": "string" },
                "page_from":     { "type": "integer", "minimum": 1 },
                "page_to":       { "type": "integer", "minimum": 1 },
                "konu_from":     { "type": ["null","integer"], "minimum": 1 },
                "konu_to":       { "type": ["null","integer"], "minimum": 1 },
                "soru_from":     { "type": ["null","integer"], "minimum": 1 },
                "soru_to":       { "type": ["null","integer"], "minimum": 1 },
                "antrenman_from":{ "type": ["null","integer"], "minimum": 1 },
                "antrenman_to":  { "type": ["null","integer"], "minimum": 1 }
              }
            }
          }
        }
      }
    }
  }
}
```
