# workey-v2-katalog — tasks

Shared rules for every task: `README.md` in this folder (new concept model,
amended invariants, layout). Direction decision: `.agent/state/DECISIONS.md`
`2026-11-02 — v2 kaynak kataloğu`.

Plan content (İbrahim's own DGS catalog, if re-entered later) is not part of this
feature; these tasks deliver the generalized machinery. Work one task per session.

| ID | Task | Status | Depends on |
| --- | --- | --- | --- |
| WKY-21 | [Katalog şeması ve store katmanı (sources/sections/units/unit_progress)](tasks/WKY-21.md) | done | — |
| WKY-22 | [Plan domain modelini kataloğa bağla (Item → kaynak/bölüm/birim, notasyon)](tasks/WKY-22.md) | done | WKY-21 |
| WKY-23 | [Birim ilerlemesi ve D/Y/B servis yolu (tamamlamadan bağımsız)](tasks/WKY-23.md) | done | WKY-22 |
| WKY-24 | [MCP: save_source/get_sources, bookEnum kaldırma, ajan rehberi](tasks/WKY-24.md) | done | WKY-22, WKY-23 |
| WKY-25 | [REST API: katalog uçları ve plan item sözleşmesi](tasks/WKY-25.md) | done | WKY-22, WKY-23 |
| WKY-26 | [Web: Kaynaklar düzeltme sayfası ve D/Y/B raporu](tasks/WKY-26.md) | done | WKY-22, WKY-23 |
| WKY-27 | [Web: Bugün/Hafta/İlerleme birim görünümü ve işaretleme](tasks/WKY-27.md) | done | WKY-23, WKY-26 |
| WKY-28 | [Uçtan uca doğrulama ve v2 kapısı](tasks/WKY-28.md) | done | WKY-24, WKY-25, WKY-27 |

## Suggested session order
WKY-21 → 22 → 23 are the backbone (store → domain → unit progress) and must go in
order. WKY-24 (MCP) and WKY-25 (REST) are independent of each other once 22/23
land. WKY-26 then WKY-27 build the web flows. WKY-28 closes the gate end to end.

## Whole-tree build is red between WKY-22 and WKY-24..27 (user split, 2026-11-02)
WKY-22 rebinds `plan.Item` to the catalog and keeps `internal/plan` + `internal/store`
building and green, but it deliberately does **not** touch the layers that render the
v1 book/chapter/page concepts. Each downstream task now also carries the job of making
its own layer compile again against the catalog `plan.Item`/notation:
- WKY-24 → `internal/mcpserver`, `internal/apidoc` — DONE (both build/vet/test green;
  `internal/api` non-test also builds now as a side effect of the apidoc fix).
- WKY-25 → `internal/api` + `internal/apidoc` — DONE (catalog endpoints added; all
  _test files updated to seed the catalog; build/vet/test green).
- WKY-26 → `internal/web`, `internal/web/views` (present.go, study.go, *.templ;
  the in-app API usage guide text in usage_content.go and the verbatim-synced
  study-contract docs) — DONE. The whole tree now compiles: `go build ./...`,
  `CGO_ENABLED=0 go build ./...` and `go vet ./...` are green again.
Remaining red after WKY-26: none. WKY-27 adds the Bugün/Hafta/İlerleme unit
marking; WKY-28 closes the end-to-end gate.
