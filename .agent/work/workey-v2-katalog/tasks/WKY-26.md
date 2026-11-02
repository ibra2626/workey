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
(planned)
