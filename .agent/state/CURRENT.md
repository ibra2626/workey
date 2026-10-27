# CURRENT — WKY-18 complete

Updated 2026-10-27.
Active task: none.
Next task: `.agent/work/workey-v1/tasks/WKY-19.md`.
WKY-12 Linux deployment remains held by the user's explicit instruction.

## Objective and result

WKY-18 adds the exam-plan prompt to step four of /kullanim/mcp. The shared
UsagePrompt block supplies the readonly display, dedicated Promptu kopyala
button target and full Markdown. The longer prompt has eighteen visible rows.
Visible exam date comes from AgentGuide; execution instructions refresh today,
timezone and exam_date through get_agent_guide. User fields cover level/goal,
availability, duration, rest and verified book/chapter/page references, with
an editable four-evening/one-weekend suggestion.

The prompt requires initial reads, clarification before writes, full outline
preservation, bounded weekly writes with immediate rereads, explicit open-session
replacement semantics, protected past/completed records and makeup links,
result inspection, reread verification, stopping on failures/blocking skips,
and a Turkish report of completed/remaining work. It forbids completion calls.
The page itself only copies text; no plan-writing behavior was added.

## Active paths

- internal/web/views/usage_content.go, usage_document.go
- internal/web/views/usage.templ and generated usage_templ.go
- internal/web/views/usage_document_test.go, internal/web/usage_test.go
- .agent/work/workey-v1/TASKS.md and tasks/WKY-18.md

## Verification

Passed with Go 1.27.3: go tool templ generate; Tailwind rebuild (no CSS diff);
go test ./internal/web/...; go vet ./internal/web/...;
go build -o bin/wky18-preview/workey.exe ./cmd/workey; gofmt; git diff --check.
Build emitted a stat-cache access warning but exited 0.
Tests cover fake Istanbul midnight/alternate exam date, critical instructions,
and exact rendered/copy/Markdown prompt parity for all three clients.
Headless Edge at 390x844 and 1280x844 passed all clients, keyboard/touch,
clipboard readback (Windows CRLF normalized), Markdown parity, no horizontal
overflow, closed details and JS-disabled fallback. Screenshots visually checked.
Sandbox browser failed to launch; browser and isolated server subsequently ran
together outside the sandbox successfully. Both test servers stopped. Temporary
checks, binaries, DBs and screenshots are ignored under bin/wky18-preview.

Existing-plan, missing-reference, partial-last-week, skipped and interrupted-write
scenarios reviewed against prompt instructions. No real plan writes or real
.env/token/DB access. Full suite, CGO-disabled build and vulnerability scan not
rerun for this focused guide change.

## Handoff

WKY-01..11,13..18 complete; WKY-19 pending; WKY-12 blocked.
WKY-19 owns combined acceptance and documentation. User requested a local commit
of WKY-18 after verification. No push or deployment requested.
