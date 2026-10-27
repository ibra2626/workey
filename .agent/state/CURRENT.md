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
