# WKY-11 — İlerleme screen (mobile + desktop)

## Goal
Honest progress against the exam date, matching the İlerleme artboards minus the deneme panels.

## Scope
- `internal/web/views/progress.templ` + handler on `Progress()`; route `/ilerleme`.
- Mobile: "N gün · 18 Tem", Plan X / 38 hafta, X / Y sa, X / Y seans, 38-week strip (current week done fraction in ink, rest of current phase Phase Mist, later phases Hairline) with phase/date labels, Konu grupları list with progress tracks, condensed %, "8–10 soru" (never "s.").
- Desktop: header figures, strip with legend (Yapılan, Faz N planı, Sonraki fazlar) and date axis, 380px side column for topic groups.
- Removed per DECISIONS: "İlk ölçüm · seviye testi" panel and "Denemeler · net" chart; the layout must not leave a hole (rebalance columns, note how).
- Weeks beyond the saved plan count as unplanned, not zero-done; denominators come from the saved plan, never hard-coded 342/190.

## Acceptance criteria
- [ ] Render tests: percentages with decimal comma, zero values in Lichen Ink, no deneme/seviye strings.
- [ ] Partial plan (only week 1 saved) renders sensibly.
- [ ] Screenshots 390 and 1280 compared with the mockup; fixes or reasons in Progress.

## Checks
- `go tool templ generate && go build ./...`
- `go test ./internal/web/...`
- Preview screenshots at 390 and 1280.

## Progress
Not started.
