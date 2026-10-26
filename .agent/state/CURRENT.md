# CURRENT — WKY-10 done, Hafta screen live

Updated 2026-10-26.

Active task: .agent/work/workey-v1/tasks/WKY-11.md (WKY-09, WKY-10 done, committed)

- Feature: `workey-v1` — Go + templ + htmx + Tailwind single binary; REST `/api/v1`, MCP `/mcp` and web share the one business-logic layer `internal/plan`; SQLite (modernc). Local Windows only; WKY-12 deploy remains **blocked** by the user (do not start until lifted).
- Plan: `.agent/work/workey-v1/TASKS.md`; shared rules: `.agent/work/workey-v1/README.md`. Product: `PRODUCT.md`; visuals: `DESIGN.md`; decisions: `.agent/state/DECISIONS.md`.
- WKY-01..08 committed (HEAD before this session `0e1b0ee` WKY-08). WKY-09 committed this session.

## This session

Running all remaining tasks in order (user: "kalan taskları sırası ile bana sormadan yap, her task sonrası anlamlı mesaj ile commit at"). Order: WKY-09 ✓, WKY-10, WKY-11, WKY-12 (blocked → skip), WKY-13, WKY-14.

## WKY-09 result (Bugün screen)

- `internal/web/views/today.templ`: `Today(TodayPage)` + swappable `TodayPanel` (`#bugun-panel`). Mobile first panel (ref, chapter, book line + pages, interval profile with legend, Tamamlandı pill), stat tiles Sınava/Bu hafta/**Seans** (no Seri), note; desktop 2-col grid + Bu hafta mini / Yarın / Not + full-width phase blocks. Empty/rest/after-exam states.
- `internal/web/views/present.go`: presentation-only helpers (profile, volume bars, 38-week strip, status words, book names, axis ticks, Turkish dative weekdays). Reused by WKY-10/11.
- `internal/web/views/layout.templ`: `Layout(LayoutProps)` with rail foot "Sınava N gün". `views.go` `styleAttr`; `input.css` `hatch` utility + reduced-motion guard.
- `internal/web/{auth,web}.go`: `Auth` now holds `*plan.Service` + exam date; `NewAuth(svc, examDate, hash, key, secure, dev)`. Handlers `screenToday`, `completeSession`/`undoSession` at `POST /oturum/{id}/{tamamla|geri-al}` (source web). htmx swaps the panel; plain POST (hidden `_csrf`, accepted by `requireCSRF`) redirects. `cmd/workey` + web tests updated.

## Verification (Go 1.27.3, all green)

`templ generate`, Tailwind rebuild, `go build ./...`, `go vet ./...`, `gofmt -l .` empty, `go test ./internal/web/...` pass. Live (dev-open, seeded via `PUT /api/v1/{outline,plan}` with sample week) at 390 and 1280: on-brand, Turkish glyphs, first panel fits 390×844, htmx completion → done state + "Geri al", console clean.

## Next step

WKY-11: İlerleme screen (progress.templ + handler on Progress()); route /ilerleme. Reuse present.go progressStrip, percentClass. Remove deneme/seviye panels; rebalance columns.

Dev preview: `tmp-dev.db` seeded on 127.0.0.1:8099 (WORKEY_DEV=1, token `devtoken`); real system "today" is 2026-10-26 (Monday).
