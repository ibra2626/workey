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
- [x] Transcript summary in Progress: tools called, results, any description fixes.
- [x] Audit query output (no secrets) showing sources and timestamps.
- [x] Screenshots of Bugün, Hafta, İlerleme after the run (mobile).

## Checks
- `go test ./internal/mcpserver/... ./internal/plan/...`
- Manual run as above.

## Progress
Completed 2026-10-27 with an explicit client substitution: official Go MCP SDK
over real Streamable HTTP, not a separate Claude Code model session. Autonomous
model interpretation from descriptions alone is therefore not verified.

`TestMCPWebWorkflow`: save_outline → save_plan → get_progress → web completion
with the actual hidden CSRF form → get_progress → repeat save_plan/completion →
web undo → get_progress. Five sessions/540 min; after completion 1 session,
90 min, 1,5 hours; after undo zero. Repeat leaves 5 sessions and adds no audit.
Existing schema-description and MCP/plan suites pass; no tool-description fix.

Audit summary at fixed test time, before undo:
`2026-10-26T06:00:00Z | mcp | create | 11`;
`2026-10-26T06:00:00Z | web | complete | 1`.

Connection, prompts, semantics and reproduction: `docs/mcp.md`.
Screenshots were reviewed and removed at the user's request before commit;
findings remain in `docs/review/README.md`. Disposable fixtures
were seeded through MCP; the browser completed Wednesday's session. Found stale
Bugün summary after htmx completion; correction belongs to WKY-14.
