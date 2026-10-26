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
Done (2026-10-26).

- `internal/web/views/progress.templ`: `Progress(ProgressPage)`. Header "İlerleme" + "N gün · 18 Tem" (mobile) / "X / Y seans · X / Y saat" (desktop). Plan panel: "N / 38 hafta", mobile totals (sa, seans), desktop legend (Yapılan / "Faz N planı" / Sonraki fazlar). 38-week strip: each column is its phase tint (current phase Phase Mist, later Hairline) with a Forest-Ink overlay for the done fraction; labels "Faz 1 · H1–7" and "H38 · 18 Tem". Topic-group rows: name, Hairline track with the ink done share, DGS weight "8–10 soru" (never "s."), condensed % (Lichen Ink at zero). Empty state.
- Removed per DECISIONS: the "İlk ölçüm · seviye testi" panel and the "Denemeler · net" chart. **Rebalance:** İlerleme is a single column — the strip panel and the topic-group panel each span the full width (no empty 380px side column, no hole). Noted in DECISIONS.
- Denominators come from the saved plan (`PlannedHours`, `SessionsPlanned`, `WeeksTotal`), never the mockup's 342/190. Weeks beyond the saved plan are unplanned (Hairline, zero done).
- `present.go` added `dgsLabel`, `currentPhaseRange`, `phaseRangeLabel`. `web.go` added `screenProgress`; the old `page`/`Placeholder` scaffold was removed (all three screens are real now).
- Tests (`progress_test.go`): "/ 38 hafta", "1,5" (decimal-comma hours), " soru"; zero percentage in Lichen Ink (`text-ink-3">%0`); no "deneme"/"seviye"; partial plan (week 1 only) renders and shows no 342/190; empty state.
- Checks green (Go 1.27.3): templ generate, Tailwind, build, vet, gofmt, `go test ./internal/web/...`. Verified live at 390 and 1280: on-brand, strip + legend + tracks correct, decimal comma, single-column rebalance, console clean.
