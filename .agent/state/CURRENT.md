# CURRENT — WKY-17 complete

Updated 2026-10-27.
Active task: none.
Next task: `.agent/work/workey-v1/tasks/WKY-18.md`.
WKY-12 Linux deployment remains held by the user's explicit instruction.

## Objective and result

WKY-17 simplifies /kullanim/mcp into four steps: prepare Workey, connect the
selected client, test the connection, create a plan. Step four is a placeholder
for WKY-18. Codex Desktop is the default; client=codex-desktop|codex-cli|claude-code
is normalized on the server, unknown values use Desktop. View links preserve
selection; no persistent preference. Only the selected setup appears in HTML;
full Markdown includes every client and technical detail independently of selection.

Token source, equality with server WORKEY_API_TOKEN, process inheritance,
custom ports and Desktop restart are explained beside setup. Technical material
is in closed details. The connection prompt has its own readonly copy target
and live status using unchanged WKY-16 copy.js. No JS still supports navigation
and manual text selection. API guide content is unchanged.

## Active paths

- internal/web/usage.go and usage_test.go
- internal/web/views/usage.templ and generated usage_templ.go
- internal/web/views/usage_document.go, usage_content.go, usage_document_test.go
- .agent/work/workey-v1/TASKS.md and tasks/WKY-17.md

## Verification

Passed: go tool templ generate; Tailwind rebuild (no generated CSS change);
go test ./internal/web/...; go vet ./internal/web/...; temporary cmd/workey build;
gofmt; git diff --check. Go 1.27.3.
Headless Edge at 390x844 and 1280x844: all three clients, keyboard/touch,
selection across views, full Markdown parity, separate prompt copy with clipboard
readback, closed details, no horizontal overflow, JS-disabled navigation/manual
selection. Screenshots visually inspected. Disposable dev DB used; real .env,
tokens and database untouched. Test servers stopped. Temporary browser script,
binary and screenshots are ignored under bin/wky17-preview.
Official configuration source checked: https://learn.chatgpt.com/docs/extend/mcp?surface=cli
Existing connection commands preserved; Desktop now includes the Codex add command.
Full suite, CGO-disabled build and vulnerability scan not rerun for this focused change.

## Handoff

WKY-01..11,13..17 complete; WKY-18..19 pending. WKY-12 blocked.
WKY-18 adds the exam plan prompt to step four. WKY-19 owns combined acceptance.
User authorized commit and push after completion. Release includes WKY-17
implementation and task records, plus existing WKY-18/19 definitions referenced
by the task index. No deployment requested.
