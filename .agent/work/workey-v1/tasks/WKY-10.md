# WKY-10 — Hafta screen (mobile + desktop)

## Goal
The week view with volume bars and the week list/table, matching the Hafta artboards.

## Scope
- `internal/web/views/week.templ` + handler on `Week()`; route `/hafta` with `?tarih=YYYY-MM-DD` for other weeks (previous/next week links as plain text links, no secondary button).
- Mobile: header "Hafta N · 26 Eki–1 Kas", phase name, Haftalık hacim X / Y dk, "X / Y seans · N taşındı", volume bars (done ink, moved hatch + strong-hairline ring, today sand + ochre ring, planned line, rest 3px stub), week list rows per DESIGN.md (today row Today Sand with Moss Ink secondary text; moved row "cumartesiye taşındı"; rest "Dinlenme").
- Desktop: 260px figure column + chart, full-width table (Gün, Seans, Sayfa, Süre, Durum) with status words.
- Row action for today and past rows: mark done / undo via htmx (CSRF, source `web`), swapping the row and the summary figures; future rows have no action.
- Remove the "Seri" figure from the desktop header (DECISIONS).

## Acceptance criteria
- [ ] Render tests for every state (done, moved, missed without link, today, planned, rest) and the moved-to weekday text.
- [ ] Handler test: undo on a past row; future row action → 4xx.
- [ ] Screenshots 390 and 1280 compared with the mockup; fixes or reasons in Progress.

## Checks
- `go tool templ generate && go build ./...`
- `go test ./internal/web/...`
- Preview screenshots at 390 and 1280.

## Progress
Not started.
