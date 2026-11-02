# CURRENT — WKY-26 done (Kaynaklar screen + web layer restored); WKY-27 next

Updated 2026-11-02.
Active task: .agent/work/workey-v2-katalog/tasks/WKY-26.md (DONE)
Next: WKY-27 — Web: Bugün/Hafta/İlerleme birim görünümü ve işaretleme
(unit-level done + D/Y/B marking from the day/week/progress screens).

## Feature: workey-v2-katalog (WKY-21..28)
Generalize workey from a fixed DGS/5-book tracker to a general source catalog
(Kaynak → Bölüm → Birim + separate unit progress + D/Y/B). Decisions in
`.agent/state/DECISIONS.md` (2026-11-02 entries). Shared rules:
`.agent/work/workey-v2-katalog/README.md`. No data migration; schema rebuilt
clean (WKY-21). Service path (`SaveSource`/`Sources`, `SetUnitResult`,
`SourceProgress`/`SectionProgress`) since WKY-22/23; MCP WKY-24; REST WKY-25.

## WKY-26 result — 2026-11-02 (web layer restored to the catalog + Kaynaklar screen)
`internal/web` + `internal/web/views` build/vet/test green; whole tree green.
Full detail in the task's Progress section. Summary:

**Compile restoration (no `plan`/`apidoc` changes):**
- `views/present.go`: v1 item helpers rewritten against `ItemView`, names kept so
  templ churn stayed small — `chapterRef`=`itemHead` (source·section),
  `chapterTitle`=section name (→ source name), `chapterBook`=source name,
  `headPages`/`pagesPlain`=bare loc span, `dgsLabel`→`QuestionMin/Max`. New
  `itemRef` (via `plan.FormatRef`), `itemNotations`/`itemNotation` (split
  `sv.Notation` on `" · "` — the formatter's own join — to recover per-item
  **exact** unit-prefixed notation, since `ItemView` has no location unit),
  `itemLabel(sv,i)`. Dropped `konuPages`/`soruPages`/`bookName`/`plan.Ref`.
- `views/study.go`: `sourceRanges(sv,i)`, `studyFirstStep`, `studyTopic`,
  `studyTitle` rebound to `ItemView`/`BookmarkView`; `plan.FormatItem` gone.
- templ: `today.templ` hero shows `itemNotation(sv,0)`; multi-item list +
  `study.templ` use `itemLabel`/`sourceRanges(sv,i)`. `*_templ.go` regenerated.
- konu/soru/antrenman page-split presentation removed (no catalog backing); the
  session profile keeps its 20/50/20 **time** partition with empty zone pages.

**Kaynaklar screen (new top-level, `/kaynaklar`):**
- Nav: `views.go` `navItems` + `i-sources` sprite; mobile tab bar now
  `grid-cols-6`. Routes in `web.go`: `GET /kaynaklar`, `POST /kaynaklar/{key}`.
- `internal/web/sources.go`: `screenSources`/`buildSourcesPage`/`saveSource`.
  `parseSaveSource` rebuilds the whole `SaveSourceRequest` from hidden ordinal
  lists (`section_ordinals`, `s<o>_unit_ordinals`) + field scheme (`s<o>_*`,
  `s<o>u<u>_*`), so reconcile-by-ordinal **preserves `unit_progress`**. htmx
  swaps `#kaynaklar-main`; no-JS POST (hidden `_csrf`) redirects to `/kaynaklar`.
- `views/sources.templ` + `sources_present.go`: `SourcesPage`, in-place edit
  form, per-source/section report (`progressFigures`); **video** = done-only.
- Scope = in-place editing only; add/remove/reorder deferred (user-chosen).

**Docs:** `docs/api.md`/`docs/mcp.md` study-contract blocks re-synced to
`StudyContractMarkdown()`; `usage_content.go` REST "Plan isteği" item description
rewritten from v1 book/chapter/page to the catalog source_key/section/unit/loc.

**Tests:** new `internal/web/sources_test.go` (render + figures match aggregates;
edit preserves unit progress; no-JS 303 vs 403 without CSRF; video done-only).
Updated v1 web tests to catalog output; `seedWeek` now seeds `sample-sources.json`
before `SavePlan`; `usage_test.go` tab bar `grid-cols-6`.

### Checks (2026-11-02, Go 1.27.3)
- templ generate + Tailwind rebuilt (`web/static/app.css`).
- `go test ./internal/config/... ./internal/clock/... ./internal/store/...
  ./internal/plan/... ./internal/web/... ./internal/web/views/... ./internal/api/...
  ./internal/apidoc/... ./internal/mcpserver/... ./cmd/...` → all ok.
- `go build ./...` + `CGO_ENABLED=0 go build ./...` + `go vet ./...` green
  (whole tree compiles again — WKY-28 gate effectively reached for the web layer).
- gofmt: edited/new files clean on LF (working-tree CRLF = autocrlf artifact).
- Visual check (built-in browser) at 1280×800 and 390×844: no horizontal overflow
  (scrollW==clientW), Turkish glyphs correct, 44px inputs, 6-item nav fits,
  figures match aggregates, htmx save swaps with a "kaydedildi" banner and keeps
  progress. NOT run: govulncheck (no dep change). No commit/push.

## Whole-tree build is GREEN again (2026-11-02)
WKY-22 left `internal/web`/`internal/web/views` un-compiling by design; WKY-26
restored them. `go build ./...` / `go vet ./...` pass across the module. WKY-27
(day/week/progress unit marking) and WKY-28 (end-to-end gate) remain.

## Prior state (unchanged)
SF-01..06 study flow, study room, Plan calendar shipped on main (merged PR #1).
PRODUCT.md / DESIGN.md / `.impeccable/design.json` are v1 truth; the catalog
feature supersedes the v1 item model only. WKY-12 deploy held; WKY-20 200% zoom
pending. Branch: feat/v2-katalog. No commits for WKY-22..26 yet.
