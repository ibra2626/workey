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
Not started.
