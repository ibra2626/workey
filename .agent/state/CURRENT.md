# CURRENT — WKY-19 complete

Updated 2026-10-27.
Active task: none.
Next task: none queued; WKY-12 Linux deployment remains held by the user's explicit instruction.

## Objective and result

WKY-19 completes combined acceptance of the MCP/API usage guides after WKY-15..18.
Updated docs/mcp.md and docs/api.md for view/client selection, full Markdown versus
separate prompt copying, clipboard fallback, script-free use and the exam-plan
prompt workflow. Dates/rules still come from AgentGuide and the shared document.
No application behavior or styles changed; no product defect found.

## Active paths

- docs/mcp.md, docs/api.md, docs/usage-verification.md (detailed evidence matrix)
- internal/web/usage_test.go: all client/view auth cases, uncached contract download parity
- internal/web/views/usage_document_test.go: all six prompt targets across all clients
- .agent/work/workey-v1/TASKS.md and tasks/WKY-19.md

## Verification

Passed with Go 1.27.3: templ generate (no generated diff);
go test ./internal/web/... ./internal/apidoc/...; go test ./internal/api/...;
go vet ./...; go build ./...; CGO_ENABLED=0 go build ./...;
gofmt -l . (empty); git diff --check.
Builds emitted module stat-cache permission warnings but exited 0.
No class/token change; embedded CSS remains current without a rebuild.

In-app browser at 390x844 and 1280x800: both pages/views, all three MCP clients,
closed/open disclosures, keyboard view/client selection and disclosures,
readonly selection, visible focus, long content and no page horizontal overflow.
Measured document widths 375/1265 px. MCP full Markdown plus six separate prompts
and API full Markdown matched clipboard readback after CRLF normalization.
Test proxies forced Clipboard rejection/absence: manual selection/focus and
error status were correct, buttons reenabled, no false success.
CSP script-src none blocked all scripts: navigation/disclosures and real
Ctrl+A/Ctrl+C manual copy worked, copy buttons stayed hidden.

Preview used an empty disposable DB, no personal .env/token/plan or permanent
client configuration. Files remain ignored under bin/wky19-preview.
The sandbox network prevented browser access; isolated loopback preview and
proxies ran outside the sandbox for acceptance. All test servers stopped.

Limits: denial was injected by a test script; script-free behavior used CSP,
not permanent browser settings. No screen reader, other browser engines/macOS,
new autonomous model planning, full suite or vulnerability rescan in this task.
Earlier Codex CLI acceptance and Desktop-model acceptance limits still apply.

## Handoff

WKY-01..11,13..19 complete; WKY-12 blocked by user instruction.
No next task is queued. User requested committing and pushing WKY-19 after
acceptance; this snapshot is included in that commit. No deployment requested.
