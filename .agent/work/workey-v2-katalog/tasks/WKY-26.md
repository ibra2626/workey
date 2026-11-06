# WKY-26 — Web: Kaynaklar düzeltme sayfası ve D/Y/B raporu

## Goal
A web page where the user reviews the agent-extracted catalog and corrects it —
the agent can be wrong (e.g. the M1 TOC-vs-page-heading "+/−" mismatch that was
only caught by looking). Also show per-source/section D/Y/B and done-count
reporting. Follows DESIGN.md (soft tokens, Barlow, no gamification).

## Scope
- **`internal/web`:** a `Kaynaklar` route (desktop rail + mobile menu entry) listing
  sources → sections → units from `get_sources`/read models.
- **Correction UI:** edit a source's `name`, `kind`, `location_unit`; edit section
  names and location ranges; edit unit `label`/`number`/location range; reorder
  and add/remove within a source. Writes go through the `plan` service (source
  `web`), audited, and **must not clobber `unit_progress`** for kept unit ids.
  htmx-first with a no-JS form fallback and CSRF, matching the WKY-08/09 pattern.
- **Reporting:** per-source and per-section done-count and D/Y/B sums (from WKY-23
  aggregates), rendered as calm figures (no badges/streaks/confetti). Video units
  show done-only, no D/Y/B.
- **`internal/web/views`:** new templ component(s) + `present.go` helpers; all
  colours are CSS-variable tokens; Turkish casing correct (no CSS uppercase);
  notation via the one formatter.
- Non-goals: marking units done from Bugün/Hafta (WKY-27) — though the shared
  mark-done control may be introduced here if cleaner, call it in Progress.

## Acceptance criteria
- [ ] `go test ./internal/web/...` covers: page renders a seeded catalog; an edit
  updates the source and preserves unit progress; no-JS form path passes CSRF;
  reporting figures match the aggregates.
- [ ] Visual check (built-in browser) at 390×844 and 1280×800: no overflow, correct
  Turkish glyphs (ş ğ ı İ ç ö ü), tokens only, ≥44px touch targets, contrast per
  DESIGN.md.
- [ ] templ + Tailwind regenerated; `web/static/app.css` rebuilt; builds, vet,
  gofmt clean.

## Checks
- `go tool templ generate` then `./bin/tailwindcss.exe -i web/styles/input.css -o web/static/app.css --minify`
- `go test ./internal/web/...`
- `go vet ./...`, `gofmt -l .`, builds (normal + `CGO_ENABLED=0`)

## Notes
- Depends on WKY-22/23. DECISIONS.md entry if the navigation/layout changes beyond
  what DESIGN.md describes (a new top-level screen).

## Progress
Done — 2026-11-02.

**Compile restoration (`internal/web`, `internal/web/views`):**
- `views/present.go`: rewrote the v1 item helpers against `ItemView` keeping their
  names so the templ churn stayed minimal — `chapterRef`=`itemHead` (source·section),
  `chapterTitle`=section name (→ source name), `chapterBook`=source name, `headPages`/
  `pagesPlain`=bare loc span, `dgsLabel`→`QuestionMin/Max`. Dropped `konuPages`/`soruPages`/
  `bookName`/`plan.Ref`. New `itemRef` (via `plan.FormatRef`), `itemNotations`/`itemNotation`
  (split `sv.Notation` on `" · "`) and `itemLabel(sv,i)` for exact per-item notation.
- `views/study.go`: `sourceRanges(sv,i)`, `studyFirstStep`, `studyTopic`, `studyTitle`
  rebound to `ItemView`/`BookmarkView`; `plan.FormatItem` gone.
- templ: `today.templ` hero shows `itemNotation(sv,0)` (full prefixed ref) + source line,
  multi-item list uses `itemLabel(sv,i)`; `study.templ` source/option/help use `itemLabel`/
  `sourceRanges(sv,i)`. Regenerated `*_templ.go`.
- Test fixes: `study_week_test.go` literal → `[]plan.ItemView` (+ local `ptr`);
  `study_completion_test.go` `PageFrom`→`*LoFrom`; `today_test.go` assertions → catalog
  output (no konu/soru, catalog section names); `study_test.go` → `konum aralığında`;
  web test fixture (`today_test.go seedWeek`) now seeds `sample-sources.json` first;
  `usage_test.go` tab bar `grid-cols-6`.

**Kaynaklar screen (new):**
- Nav: `views.go` `navItems` + `i-sources` sprite; mobile tab bar `grid-cols-6`.
- Routes (`web.go`): `GET /kaynaklar`, `POST /kaynaklar/{key}` (requireAuth+requireCSRF).
- `internal/web/sources.go`: `screenSources`/`buildSourcesPage`/`saveSource`; `parseSaveSource`
  rebuilds the whole `SaveSourceRequest` from the hidden ordinal lists (`section_ordinals`,
  `s<o>_unit_ordinals`) + field scheme (`s<o>_*`, `s<o>u<u>_*`), so reconcile-by-ordinal
  preserves `unit_progress`. htmx swaps `#kaynaklar-main`; no-JS redirects to `/kaynaklar`.
- `views/sources.templ` + `sources_present.go`: `SourcesPage`, per-source/section report
  (`progressFigures`; video = done-only), edit form components, token-only, Barlow.
- `internal/web/sources_test.go`: render+report figures match aggregates; edit preserves
  unit progress (id unchanged, `SourceProgress` stable); no-JS `_csrf` 303 vs 403 without;
  video shows no D/Y/B.

**Docs:** `docs/api.md`/`docs/mcp.md` study-contract blocks re-synced to
`StudyContractMarkdown()`; `usage_content.go` REST "Plan isteği" item description rewritten
to the catalog contract (+ "kaynak referanslarıyla").

**Checks (Go 1.27.3):** templ generate + Tailwind rebuilt; `go test` of config/clock/store/
plan/web/web.views/api/mcpserver/cmd green; `go build ./...` + `CGO_ENABLED=0 go build ./...`
+ `go vet ./...` green (whole tree compiles again). `gofmt`: edited/new files clean on LF
(working-tree CRLF = autocrlf artifact, documented). Visual check in the built-in browser at
1280×800 and 390×844: no horizontal overflow, Turkish glyphs correct, 44px inputs, 6-item
nav fits, report figures match aggregates, htmx save swaps with a "kaydedildi" banner and
preserves progress. Not run: govulncheck (no dependency change). No commit/push.

**Deferred (not this task):** add/remove/reorder of sections/units (own follow-up);
Bugün/Hafta/İlerleme unit marking is WKY-27.
