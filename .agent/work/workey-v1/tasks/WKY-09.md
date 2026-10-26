# WKY-09 — Bugün screen (mobile + desktop)

## Goal
The 20:30 screen: tonight's assignment and one-tap "Tamamlandı", matching the Bugün artboards of `design/directions/04-antrenman-gunlugu.html`.

## Scope
- `internal/web/views/today.templ` + handler on `Day()` read model; route `/` (and `/bugun`).
- Mobile 390 first: header line (date · "Hafta N · Faz N"), first panel per the First Panel Rule (book ref, chapter title, book line with pages, session profile with axis and legend, "Tamamlandı" pill — no scrolling at 390×844), stat tiles: Sınava N gün, Bu hafta X / Y dakika, **Seans X / Y** (replaces Seri, see DECISIONS), generated note line.
- Desktop 1280: two-column grid, labelled session-profile blocks, side panels Bu hafta (small bars), Yarın, Not (generated), full-width phase panel with week blocks. Remove the "H4 · deneme" annotation unless a `deneme` session exists that week.
- htmx: POST complete/undo (source `web`, CSRF header) swaps the first panel; after completion show the done state and a quiet "Geri al" text link. Works without JS as a normal form POST + redirect.
- Session profile proportions come from item splits when present; otherwise default 20/50/20 for `aksam`; `uzun_blok` and `deneme` get a simple single-block profile (design it within DESIGN.md, note it).
- Empty states: no plan yet ("Plan henüz yazılmadı." + what the AI should call), rest day ("Bugün dinlenme."), before start, after exam. No fabricated figures.
- Seed for preview: load `testdata/sample-week.json` via `PUT /api/v1/plan` against a dev DB; screens show it only because it was saved, never hard-coded.

## Acceptance criteria
- [ ] templ render tests: notation exact, Turkish dates, Seans tile present, no "Seri", `aria-current` on Bugün.
- [ ] Handler test: complete via htmx returns the swapped fragment; without CSRF → 403.
- [ ] Screenshots 390×844 and 1280 compared with the mockup; differences fixed or listed with reason in Progress.

## Checks
- `go tool templ generate && go build ./...`
- `go test ./internal/web/...`
- Preview: 390×844 and 1280 screenshots, console free of errors.

## Progress
Done (2026-10-26, not committed separately yet — committed with this task).

- `internal/web/views/today.templ`: `Today(TodayPage)` + the swappable `TodayPanel` (id `bugun-panel`). Mobile-first first panel (book ref, chapter, book line + pages, interval profile with mobile legend / desktop in-block labels + 3-col detail, "Tamamlandı" pill); mobile stat tiles **Sınava / Bu hafta / Seans** (no Seri) and the generated note; desktop two-column grid with Bu hafta mini chart, Yarın, Not side panels and the full-width current-phase week blocks. Empty ("Plan henüz yazılmadı."), rest ("Bugün dinlenme.") and after-exam states.
- `internal/web/views/present.go`: presentation helpers only (profile blocks, volume bars, 38-week strip cells, status words, Turkish dative weekday, book names, axis ticks). No business rule — the service already derived state/notation.
- `internal/web/views/layout.templ`: `Layout` now takes `LayoutProps` and the rail foot shows "Sınava N gün". `views.go` gained `styleAttr` for the dynamic chart geometry (token colours via CSS vars, never hex). `input.css` gained the `hatch` utility and a `prefers-reduced-motion` guard.
- Handlers in `internal/web/web.go`: `screenToday`, `completeSession`/`undoSession` at `POST /oturum/{id}/{tamamla|geri-al}` (source `web`, behind requireAuth+requireCSRF). htmx (HX-Request) swaps `#bugun-panel`; a plain form POST (hidden `_csrf` field, now accepted by `requireCSRF`) redirects to `/`. `Auth` carries the `*plan.Service` + exam date (`NewAuth` signature changed; `cmd/workey` and the web tests updated).
- Tests (`today_test.go`): notation (`M1·3`, `s. 41–58`, `Antr. 1–3`), Turkish date, Seans tile present, no "Seri", `aria-current`, empty state, htmx complete returns the panel fragment with "Geri al", complete without CSRF → 403.
- Checks green (Go 1.27.3): `go tool templ generate`, Tailwind rebuild, `go build ./...`, `go vet ./...`, `gofmt -l .` empty, `go test ./internal/web/...` pass. Verified live (dev-open, seeded via `PUT /api/v1/{outline,plan}`) at 390 and 1280: on-brand, Turkish glyphs correct, first panel fits 390×844, htmx completion swaps to the done state with "Geri al". Console clean.
- Note: completing on Bugün swaps only the first panel (per scope); the Bu hafta figures refresh on the next full load, not via OOB swap.
