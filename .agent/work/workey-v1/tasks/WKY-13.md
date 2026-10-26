# WKY-13 — End-to-end: Claude Code MCP writes a week

## Goal
Prove the v1 loop: Claude Code connects over MCP, saves a one-week plan, the plan appears on Bugün and Hafta, completion shows on İlerleme.

## Scope
- Run the app locally (preview `workey-dev`, dev DB, dev token from a local `.env` that is never printed).
- Register with `claude mcp add --transport http workey http://127.0.0.1:8080/mcp --header "Authorization: Bearer <token>"` (user runs it if it needs the secret in plain text; scope `local`).
- In a Claude Code session the user (or this session with the user's OK) asks: save the outline and one week's plan for 2026-10-26 from the tool descriptions only. Record any call the AI got wrong and fix descriptions/schemas in `internal/mcpserver`.
- Verify: Bugün shows today's session; Hafta shows all 7 days; mark done on the web → İlerleme figures change; `get_progress` over MCP agrees; repeat `save_plan` → no duplicates, audit shows `mcp` and `web` sources.
- Write `docs/mcp.md`: connection command, tool list, example prompts.
- Non-goals: visual polish (WKY-14).

## Acceptance criteria
- [ ] Transcript summary in Progress: tools called, results, any description fixes.
- [ ] Audit query output (no secrets) showing sources and timestamps.
- [ ] Screenshots of Bugün, Hafta, İlerleme after the run (mobile).

## Checks
- `go test ./internal/mcpserver/... ./internal/plan/...`
- Manual run as above.

## Progress
Not started.
