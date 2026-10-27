# CURRENT — local Windows v1 checks complete

Updated 2026-10-27.

Active task: none. WKY-01..11 were already complete. WKY-13/14 completed this
session. WKY-12 Linux deploy remains blocked by the user's explicit hold.

## Results

- WKY-13: real Streamable HTTP integration test: MCP writes outline/plan, web
  completes/undoes with CSRF, MCP reads shared progress; repeat requests add no
  duplicate sessions/audit. 5 sessions/540 min; completion = 1/90 min/1,5 hours.
  docs/mcp.md contains connection, semantics, prompts and audit evidence.
- Initial MCP verification used the official Go SDK, not a Claude Code model.
  Subsequent real Codex CLI acceptance passed using the agent guide and tools.
- WKY-14: 27 screenshots at 390×844 and 1280×800, including all screens and key
  states. docs/review/README.md contains findings and limits. Temporary images,
  raw accessibility output and tmp-dev/tmp-mcp databases were cleaned before commit.
- Fixed stale Today completion/undo totals by swapping TodayMain; list every
  multi-source reference; mobile inset 12px; semantic headings; plain empty copy.
  DESIGN.md aligned with existing v1 decisions. Generated templ/CSS rebuilt.
- Existing real/local DB files untouched. All review data disposable and synthetic.
- Added authenticated MCP/API usage pages, mobile Usage navigation, shared agent
  guide, tool annotations, OpenAPI 3.1 contract and contract/example tests.
  docs/usage-verification.md records real CLI acceptance and Desktop test limits.

## Verification

Go 1.27.3: templ generate, Tailwind, go test ./..., go vet ./..., normal and
CGO_ENABLED=0 builds, gofmt and diff whitespace checks. govulncheck v1.9.0:
No vulnerabilities found. Existing progress_test.go formatted. Mobile text
>=12px, computed plain-text contrast >=4.60:1, no horizontal overflow. Login/nav
keyboard focus checked; reduced-motion source checked; no screen-reader or
motion-emulation certification.

## Remaining / handoff

Local Windows implementation and technical v1 checks complete with the client
substitution above. No unblocked implementation task remains. WKY-12 must stay
held until the user lifts it. Optional external acceptance: Claude Code model
session using tools/list alone; separate Codex Desktop model acceptance also
remains unverified. User authorized committing the completed changes after test
artifact cleanup. No push/deploy. Temporary review servers are stopped after
capture; reproduction is in docs/mcp.md.
