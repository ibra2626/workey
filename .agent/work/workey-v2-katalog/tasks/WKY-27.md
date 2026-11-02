# WKY-27 — Web: Bugün/Hafta/İlerleme birim görünümü ve işaretleme

## Goal
Surface units in the study screens: the session detail shows which units
(Antrenman, Test, Ders…) to work, each can be marked "bitti" with optional D/Y/B,
and İlerleme reflects unit-level progress — all without disturbing the one-tap
session "tamamlandı" moment.

## Scope
- **Bugün / Hafta (`internal/web/views/today.templ`, `week.templ`):** render each
  session's items as catalog refs (`M1·4, Antrenman 5–7, Test 1–2`) via the
  formatter. Under the session, list its units with a per-unit "bitti" control and
  optional D/Y/B inputs, posting to `set_unit_result` (source `web`, htmx-first,
  CSRF, no-JS fallback). The session's own completion control stays separate and
  primary (independent per WKY-23).
- **İlerleme (`progress.templ`):** add unit-level done and D/Y/B aggregates
  alongside the existing minutes/weeks figures, honest denominators from saved data
  only (no hard-coded totals). Keep the single-column layout and DESIGN.md calm
  tone; no gamification.
- **`present.go`:** helpers for unit rows, done state, D/Y/B display; tokens only.
- Reuse the mark-done control shared with WKY-26 if one was introduced there.
- Non-goals: catalog editing (WKY-26).

## Acceptance criteria
- [ ] `go test ./internal/web/...`: Bugün/Hafta show unit refs and controls;
  marking a unit done does NOT complete the session and vice versa; İlerleme shows
  unit aggregates from seeded data; no-JS path + CSRF work.
- [ ] Visual check (built-in browser) 390×844 and 1280×800: no overflow, Turkish
  glyphs, tokens only, ≥44px targets, contrast; completion animation/feedback
  unchanged from v1.
- [ ] templ + Tailwind regenerated; builds, vet, gofmt clean.

## Checks
- `go tool templ generate` then Tailwind rebuild
- `go test ./internal/web/...`
- `go vet ./...`, `gofmt -l .`, builds (normal + `CGO_ENABLED=0`)

## Notes
- Depends on WKY-23 (unit result path) and WKY-26 (shared control + Kaynaklar).

## Progress
(planned)
