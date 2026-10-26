# Architecture map

Bootstrapped on 2026-10-26. A navigation map, not a full audit. No application code exists yet; everything below except the design artifact is intended, not verified.

- Entry points: — (none yet; intended: a Go HTTP server rendering templ components, htmx for the one-tap "Tamamlandı" action)
- Modules/ownership: — (intended screens: Bugün, Hafta, İlerleme, per `design/directions/04-antrenman-gunlugu.html`)
- Data and persistence: — (undecided; plan items are book·chapter + page range per day, written by an AI assistant via MCP/API)
- Security boundaries: — (single user; the MCP/API write interface is the main boundary; auth undecided)
