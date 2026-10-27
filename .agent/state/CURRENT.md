# CURRENT — WKY-16 complete

Updated 2026-10-27.
Active task: none.
Next task: `.agent/work/workey-v1/tasks/WKY-17.md`.
WKY-12 Linux deployment remains held by the user's explicit instruction.

## Objective and result

WKY-16 adds Rehber / Markdown links and full-Markdown copy controls to
/kullanim/mcp and /kullanim/api. ?view=markdown opens a labelled readonly
textarea inside the authenticated HTML shell. Unknown values use the guide.
The same full UsageDocument.Markdown output is used in both views. Auth,
no-store and OpenAPI download remain covered by tests.

web/static/copy.js is dependency-free and reusable through data-copy-target /
data-copy-status IDs. Success is announced only after Clipboard writeText
resolves. Unsupported/denied clipboard access reveals, focuses and selects
the target with manual instructions. Without JS the copy button stays hidden;
URL navigation and manual selection work. No saved view preference.

## Active paths

- internal/web/usage.go and usage_test.go
- internal/web/views/usage.templ, generated usage_templ.go,
  usage_document_test.go
- web/static/copy.js, web/styles/input.css, generated web/static/app.css
- .agent/work/workey-v1/TASKS.md and tasks/WKY-16.md

## Verification

Passed: go tool templ generate; standalone Tailwind rebuild;
go test ./internal/web/...; go vet ./internal/web/...; gofmt; git diff --check.
Normal and CGO_ENABLED=0 go build ./... passed; a temporary binary was also built for browser tests.
Headless Edge: both guides at 390x844 / 1280x844, keyboard navigation/copy and
focus, actual clipboard write/readback, full text parity, no overflow,
unsupported clipboard and denied access (NotAllowedError) fallback, whole-text
selection and focus, JS-disabled view links/manual selection. Screenshots
visually inspected at both widths. Clipboard API input matches Markdown;
Windows clipboard readback uses CRLF. Tests use disposable dev DBs; existing
real DB and .env untouched. Go tests cover real auth redirects and escaping.
No full suite or vulnerability scan rerun for this focused UI change.

## Handoff

WKY-01..11,13..16 complete; WKY-17..19 pending. WKY-12 remains blocked.
User requested commit and push after completion; no deployment requested. Existing uncommitted WKY-17..19
plan files were preserved. WKY-17 can reuse copy.js with separate readonly text
targets and live status nodes. WKY-19 owns final combined acceptance.

Temporary browser-test servers stopped after verification.
