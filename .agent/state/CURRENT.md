# CURRENT — WKY-20 implemented; zoom acceptance pending

Updated 2026-10-27.
Active task: .agent/work/workey-v1/tasks/WKY-20.md
Next: verify actual 200% browser zoom; then close WKY-20. WKY-12 Linux deployment remains held by the user's explicit instruction.

## Objective and result

API/MCP usage pages now use Workey's existing palette, Barlow typography,
sibling white panels and spacing. Shared content/Markdown source retained.
Stable section fragments and responsive contents navigation added. Client
selector moved into MCP connection setup; switching preserves #connect.
Long exam prompt is in a closed preview with an external copy button and visible
replacement warning. API first read precedes reference; read/write/re-read flow
and replacement warning grouped. Code blocks have language labels, copy controls
and local scrolling. Copy fallback reveals closed ancestors, selects textarea or
code Range, focuses it and reports failure honestly. All prompt labels now use
İstemi kopyala. No backend/auth/API/MCP contract change.

## Active paths

- internal/web/views/usage.templ and generated usage_templ.go
- internal/web/views/usage_document.go, usage_content.go, usage_document_test.go
- web/styles/input.css, web/static/app.css, web/static/copy.js
- docs/api.md, docs/mcp.md, docs/usage-verification.md (detailed WKY-20 evidence)
- .agent/work/workey-v1/TASKS.md and tasks/WKY-20.md

## Verification

Go 1.27.3: templ and Tailwind generation; focused web/views/apidoc tests;
normal and CGO_ENABLED=0 builds; vet; gofmt (empty); diff whitespace check pass.
Regression covers stable valid fragments/unique IDs, exact code targets, client
selector placement, closed prompt preview and shared prompt/Markdown content.
Existing auth/no-store/OpenAPI parity tests pass. Layout detector: no findings.

Rendered initial/final states compared with Today/Week/Progress using disposable
sample fixtures. IAB matrix: 320x844, 390x844, 1280x800; three MCP clients x two
views plus API views. No page overflow (305/375/1265px document widths). Mobile
end content clears fixed menu. Headings h1/h2/h3; 28px h1, 15px body; visible
interactive guide targets >=44px; measured text contrast minimum 4.60:1.
Keyboard navigation/client/view/disclosures and visible focus verified.
30 clipboard targets matched source exactly after CRLF normalization (two full
Markdowns, six prompts, 22 code snippets). Long prompt copied closed, 3945 chars.
Injected Clipboard denial/absence: focus, full selection, opened prompt/Markdown,
error status and reenabled controls correct. CSP script-src none: copy controls
hidden; navigation/client/view/disclosures and real Ctrl+A/Ctrl+C work.

## Remaining acceptance and limits

Actual 200% browser zoom is NOT verified: IAB Ctrl+plus left innerWidth/DPR
unchanged. Equivalent 640x400 reflow for 1280x800 passed on both guides without
horizontal overflow, but is explicitly not claimed as actual zoom. WKY-20 stays
in progress for this reason; application implementation is complete.
No screen reader, other browser engines, full test suite or vulnerability rescan
in this task. WKY-19 and earlier model-client acceptance limits remain recorded
in docs/usage-verification.md. Impeccable document refresh completed on 2026-10-27;
DESIGN.md and its v2 sidecar now capture current navigation, guide controls, inputs, actual breakpoints/easing and 10 self-contained component snippets. Token/narrative checks and desktop/mobile shadow-DOM previews passed. This separately requested documentation refresh does not close the pending actual-zoom acceptance.

## Handoff

Review screenshots are outside the repo in the chat visualization directory.
Temporary test proxies and fixture servers stopped; task-owned proxy file removed.
No personal .env/token/database/plan touched. User authorized committing and pushing WKY-20 and the design-document refresh on 2026-10-27. No PR or deployment requested.
WKY-01..11,13..19 complete; WKY-12 held; WKY-20 zoom acceptance pending.

## Follow-up — token setup and root README (2026-10-27)

User requested explicit token acquisition instructions and basic project setup.
Added README.md: prerequisites, clone/.env setup, separate random API token and
session key generation, password hashing, run/health check, basic screens,
MCP/API links, env precedence, SQLite backup, build and frontend tooling.
Shared apiTokenSetup content now precedes client connection/first API read in
both formatted and full Markdown guides. Clarified that tokens are operator
created, Read-Host only inputs them, and existing installations reuse the active
server token (process environment can override .env). .env.example and docs
cross-link setup. No actual user environment/credentials/database changed.

Passed focused web/views/apidoc tests. README PowerShell fences parsed; generation
command executed twice with no files/settings written, distinct 32-byte Base64
outputs confirmed without logging the values. README/guide generator text matches.
IAB: 390px MCP and 1280px API token sections render without page overflow;
code copy reports success and copies the 203-character generation command;
manual Markdown Ctrl+A/Ctrl+C matched source. Full clipboard readback initially
returned stale tool state; initializing the test clipboard then copying confirmed
it changed to the code length. Preview fixtures stopped. No templ/CSS changes,
no full install against personal .env, no commit/push for this follow-up.
The earlier actual 200% zoom acceptance remains open.

## Follow-up — supplied logo (2026-10-27)

Integrated user-provided web/workey.png into the desktop rail and login screen
through shared brandLogo markup. The original is preserved; an identical copy
in web/static is embedded and cache-versioned with the application. An SVG
viewport hides outer whitespace without changing image pixels or proportions;
the accessible image name is Workey. Mobile study-page geometry is unchanged.
Passed templ/Tailwind generation, focused web tests, normal and CGO-disabled
builds, focused vet, gofmt and diff whitespace checks. Visually verified login
at 390x844 and desktop rail at 1280x800 using a disposable render-only server;
server stopped and temporary source removed. No commit/push or deployment.
