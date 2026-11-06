# CURRENT — WKY-28 done: v2 end-to-end gate closed (feature complete)

Updated 2026-11-04.
Active task: .agent/work/workey-v2-katalog/tasks/WKY-28.md (DONE)
Feature workey-v2-katalog (WKY-21..28) is complete. No commits for WKY-22..28 yet.

## Feature: workey-v2-katalog (WKY-21..28) — COMPLETE
Generalized workey from a fixed DGS/5-book tracker to a general source catalog
(Kaynak → Bölüm → Birim + separate unit progress + D/Y/B). Decisions in
`.agent/state/DECISIONS.md` (2026-11-02). Shared rules:
`.agent/work/workey-v2-katalog/README.md`. No data migration; schema rebuilt
clean (WKY-21). Service path since WKY-22/23; MCP WKY-24; REST WKY-25; web
Kaynaklar WKY-26; web unit view + marking WKY-27; end-to-end gate WKY-28.

## WKY-28 result — 2026-11-04 (end-to-end gate)
Full detail in the task's Progress section + `docs/v2-gate-verification.md`.

**End-to-end run.** New `TestCatalogEndToEnd` in
`internal/mcpserver/webworkflow_test.go` drives the live `/mcp` endpoint (Go MCP
SDK client, real localhost HTTP) + web over one store, clock 2026-10-26 09:00:
save_source book (GR/page/Antrenman+Test) → save_source video (VD/lesson/Ders) →
get_sources (structure/kinds/labels confirmed) → save_plan referencing both (2
created; get_plan notation has GR·1/Antrenman and VD·1/Ders) → set_unit_result
book Antrenman 1 with D/Y/B 8/2/0, video Ders 1 with none (view null D/Y/B) → web
Bugün shows GR·1 units + name="correct", Hafta shows VD·1, Ders 1; get_progress
SessionsDone=0 with 2 units done; web-form (CSRF) completion → SessionsDone=1
while unit_progress stays 2 rows; İlerleme "Birim ilerlemesi / 6 birim / Çözülen".
Proves two kinds, unit vs session independence, no code change.

**Re-enabled by default.** Removed `//go:build webworkflow` tag (WKY-26/27 rebuilt
the views), so TestMCPWebWorkflow + TestCatalogEndToEnd run in the normal test
set; updated newMCP/webworkflow doc comments. TestVisualReview stays opt-in
(WORKEY_VISUAL_REVIEW=1).

**Grep/gate.** No bookEnum / dgs_question_* / dgsQuestion* in *.go (only .agent/
history). M1/M2/G1/G2 in code = user-given source key examples + generic
kind="book" label, not an enum. Schema 0001_init.sql is clean v2. guide.go + docs
describe the catalog flow. Fixed exam date / 38-week framing kept on purpose: it
lives in WORKEY_EXAM_DATE config + phases date spans (plan content), not schema.

**One leftover fixed.** get_agent_guide tool *description* "supported
books/session kinds" → "supported source kinds, location units and session kinds"
(internal/mcpserver/mcpserver.go; no test asserted it).

### Checks (2026-11-04, Go 1.27.3) — all green
templ generate (0 updates); Tailwind rebuilt (web/static/app.css);
`go build ./...`; `CGO_ENABLED=0 go build ./...`; `go vet ./...`;
focused test set config/clock/store/plan/web/web·views/api/apidoc/mcpserver/cmd
→ all ok; `gofmt -l .` clean on LF (CRLF list = autocrlf artifact, each file
verified gofmt-clean after stripping CR); `govulncheck@v1.9.0 ./...` → No
vulnerabilities found.

## Next
Feature complete. Possible follow-ups (not started, need explicit request):
- Commit/push/PR the WKY-22..28 work (branch feat/v2-katalog; no commits yet).
- WKY-12 deploy (held). WKY-20 real 200% zoom acceptance (pending).

## Prior state (unchanged)
SF-01..06 study flow, study room, Plan calendar shipped on main (merged PR #1).
PRODUCT.md / DESIGN.md / `.impeccable/design.json` are v1 truth; the catalog
feature supersedes the v1 item model only. Branch: feat/v2-katalog.
