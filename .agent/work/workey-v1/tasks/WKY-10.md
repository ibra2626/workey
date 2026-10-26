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
Done (2026-10-26).

- `internal/web/views/week.templ`: `Week(WeekPage)` + exported `WeekMain` (id `hafta-main`). Header "Hafta N · 26 Eki–1 Kas" + phase, plain-text previous/next week links (`?tarih=`). Volume panel: "Haftalık hacim X / Y dk", mobile "X / Y seans · N taşındı", desktop 260px figure column with a Seans/Taşınan `dl` (no Seri). Seven volume bars (done ink, moved hatch+ring, today sand+ochre ring, planned line, rest 3px stub). Mobile list rows and desktop table (Gün, Seans, Sayfa, Süre, Durum) with status words.
- Row action (`rowControl`): today/past-done/past-missed rows get a "Bitir"/"Geri al" button; future/moved/planned/rest rows have none. Posts to `POST /hafta/oturum/{id}/{tamamla|geri-al}` (source web); htmx swaps the whole main content (`#hafta-main`) so the figures and the row update together; a plain form POST (hidden `_csrf`) redirects to `/hafta?tarih=<weekStart>`.
- `present.go` added: `rowName`, `rowMeta`, `pagesPlain`, `sureCell`, `rowActionable`/`rowActionSlug`/`rowActionLabel`, `rowActive`, `weekForm`, `phaseTail`. `web.go` added `screenWeek`, `weekComplete`/`weekUndo`, `buildWeekPage`, `weekDateParam`, `shiftDays`.
- Tests (`week_test.go`): every state (done, moved, missed-without-link, today, planned, rest) + "Cumartesiye taşındı"; no "Seri"; other-week-by-date; undo on a past row via htmx returns the `#hafta-main` fragment and clears completion; future row action → 422. Web test harness refactored to seed through a Monday-clock service on a shared store (past days must be created before today freezes them); also hardened `TestTamperedCookieRejected` (flip the first base64 char, not the last — the last char's low bits are padding).
- Checks green (Go 1.27.3): templ generate, Tailwind, `go build ./...`, `go vet ./...`, `gofmt -l .` empty, `go test ./internal/web/...` (incl. `-count=3`). Verified live at 390 and 1280: on-brand, bars/state vocabulary correct, prev/next links, table statuses, row action; console clean.
