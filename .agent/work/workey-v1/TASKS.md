# workey-v1 — tasks

Shared rules for every task: `README.md` in this folder (invariants, layout, conventions).

| ID | Task | Status | Depends on |
| --- | --- | --- | --- |
| WKY-01 | [Go skeleton, config, HTTP server](tasks/WKY-01.md) | done | — |
| WKY-02 | [Frontend toolchain: templ, Tailwind, fonts, htmx, base layout](tasks/WKY-02.md) | done | WKY-01 |
| WKY-03 | [SQLite store and schema](tasks/WKY-03.md) | done | WKY-01 |
| WKY-04 | [Plan service: write path (save plan, outline, audit)](tasks/WKY-04.md) | done | WKY-03 |
| WKY-05 | [Plan service: read models and completion](tasks/WKY-05.md) | done | WKY-04 |
| WKY-06 | [REST API /api/v1 with Bearer auth](tasks/WKY-06.md) | done | WKY-05 |
| WKY-07 | [MCP server (Streamable HTTP)](tasks/WKY-07.md) | todo | WKY-06 |
| WKY-08 | [Web auth: Giriş screen, session cookie, CSRF](tasks/WKY-08.md) | todo | WKY-02, WKY-03 |
| WKY-09 | [Bugün screen (mobile + desktop)](tasks/WKY-09.md) | todo | WKY-05, WKY-06, WKY-08 |
| WKY-10 | [Hafta screen (mobile + desktop)](tasks/WKY-10.md) | todo | WKY-09 |
| WKY-11 | [İlerleme screen (mobile + desktop)](tasks/WKY-11.md) | todo | WKY-09 |
| WKY-12 | [Deploy: Linux build, systemd unit, Nginx example](tasks/WKY-12.md) | blocked | WKY-07, WKY-08 |
| WKY-13 | [End-to-end: Claude Code MCP writes a week](tasks/WKY-13.md) | todo | WKY-07, WKY-10, WKY-11 |
| WKY-14 | [Visual review of all screens and v1 gate](tasks/WKY-14.md) | todo | WKY-13 |
