# Workey REST API — `/api/v1`

## Uygulama içi rehber ve makine sözleşmesi

Menüden **API Kullanımı** (`/kullanim/api`) sayfasına ulaşılır. Sayfa oturum
doğrulaması ister; endpoint listesi, tam temsili yanıtlar, PowerShell/curl.exe
komutları ve ortak ajan kurallarını gösterir. Mobilde **Kullanım** sekmesini açın.

- `GET /api/v1/agent-guide`: Bearer korumalı, salt okunur ajan başlangıç rehberi.
  `today`, `timezone`, `exam_date`, `books`, `session_kinds`, `rules`, `workflow`
  döner. Tarih, plan yazımıyla aynı enjekte edilebilir İstanbul saatinden gelir.
- `GET /api/v1/openapi.json`: Bearer korumalı OpenAPI 3.1 sözleşmesi. Tüm REST
  işlemlerini, JSON alanlarını, hata zarflarını ve örnekleri kapsar.
- `GET /kullanim/api/openapi.json`: aynı belgenin web oturumu korumalı indirmesi.

Önce ajan rehberini, fazları, ilgili plan aralığını ve ilerlemeyi okuyun.
Korunacak seansları tam yazma isteğine dahil edin; yazdıktan sonra tekrar okuyun.
Boş olmayan `Idempotency-Key` başlığı gövdedeki `idempotency_key` alanını geçersiz
kılar. Kaynak, istek hash'ine dahil olduğundan MCP ve REST arasında anahtar paylaşmayın.

Şemalar gerçek Go taşıma tiplerinden üretilir. Boş listelerin mevcut `null`
gösterimi korunur. REST'in bilinmeyen alanları kabul eden ve eksik koleksiyonları
boş sayan mevcut davranışı belgelenir; bu nedenle seans/faz/grup listelerini her
zaman açıkça gönderin. İsteğe bağlı konu/soru/antrenman aralıklarında backend
yalnızca çift bulunmasını ve sıralamayı denetler; yeni pozitiflik kuralı eklenmez.

`TestOpenAPIContract`, örnek outline/plan isteklerini geçici veritabanında çalıştırır;
boş/dolu yanıtları JSON Schema ile doğrular ve tüm temsili başarılı yanıtları
gerçek API sonuçlarıyla karşılaştırır. Örnekler sabit 26 Ekim 2026 senaryosudur.

The REST API exposes the same plan operations as the MCP server, thin over
`internal/plan` (source `api`). It is server-rendered JSON; there is no web
cookie auth here (that is the Giriş screen, WKY-08).

## Conventions

- **Base path:** `/api/v1`.
- **Auth:** every route requires `Authorization: Bearer <WORKEY_API_TOKEN>`. The
  token is compared in constant time on SHA-256 digests; a missing, malformed or
  wrong token returns `401` with `WWW-Authenticate: Bearer`. If no token is
  configured the API is locked (every request `401`).
- **Dates & times:** civil dates `YYYY-MM-DD` and times `HH:MM`, all interpreted
  in Europe/Istanbul. Notation is produced server-side (`M2·17, s. 210–225`).
- **Request bodies** are JSON, capped at 1 MiB. Malformed JSON → `400`.
- **Responses** are `application/json; charset=utf-8` with `Cache-Control:
  no-store`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`.
- **Errors** use one envelope:

  ```json
  { "error": { "code": "validation_failed", "message": "…", "fields": [
    { "path": "sessions[0].date", "code": "invalid", "message": "…" }
  ] } }
  ```

  Status mapping: validation → `422` (or `404` when every field is `not_found`);
  reused idempotency key with a different body → `409`
  (`code: "idempotency_conflict"`); unexpected failure → `500` (`code:
  "internal"`, no internals leaked).

## Session state

Read responses derive each session's `state`: `planned` (future), `today`,
`done` (completed), `missed` (past, not done) or `moved` (missed but referenced
by a makeup session; the row carries `moved_to: {date, weekday}`).

## Routes

### `GET /api/v1/plan?from=YYYY-MM-DD&to=YYYY-MM-DD`
Sessions in `[from, to]` with items, derived `state` and `notation`. Both
params are required (`422` otherwise).

```
GET /api/v1/plan?from=2026-10-26&to=2026-11-01
200 → { "sessions": [
  { "id": 1, "date": "2026-10-26", "start_time": "20:30", "duration_min": 90,
    "kind": "aksam", "completed": false, "state": "today",
    "topic_group_key": "temel-kavramlar", "topic_group_name": "Temel kavramlar",
    "items": [ { "book": "M1", "chapter_no": 1, "chapter_title": "Temel kavramlar",
                 "page_from": 1, "page_to": 22 } ],
    "notation": "M1·1, s. 1–22" }
] }
```

### `PUT /api/v1/plan`
Declarative range replace for `[from, to]` (today-or-later open sessions become
exactly the given set; past and completed sessions are frozen and returned in
`skipped`). Optional `Idempotency-Key` header. Body is a `SavePlanRequest`.

```
PUT /api/v1/plan
Idempotency-Key: week-1            (optional)
{ "from": "2026-10-26", "to": "2026-11-01", "sessions": [ … ] }
200 → { "created": [ {"id":1,"date":"2026-10-26","start_time":"20:30"}, … ],
        "updated": [], "deleted": [], "unchanged": [], "skipped": [] }
```

Re-sending the same body reports every session under `unchanged` and writes no
audit rows. Re-using the same `Idempotency-Key` with a different body → `409`.

### `GET /api/v1/outline`
The saved phases and topic groups, ordered by ordinal.

```
200 → { "phases": [ { "key": "faz-1", "ordinal": 1, "name": "Faz 1 · İşlem temeli",
                      "start_date": "2026-10-26", "end_date": "2026-12-13" } ],
        "topic_groups": [ { "key": "temel-kavramlar", "name": "Temel kavramlar",
                            "ordinal": 1, "dgs_question_min": 8, "dgs_question_max": 10 }, … ] }
```

### `PUT /api/v1/outline`
Replace phases and topic groups by key (create/update/unchanged/delete),
idempotent and audited. Deleting a topic group that sessions still reference →
`422` with `code: "referenced"`. Optional `Idempotency-Key`. Body is a
`SaveOutlineRequest`.

```
200 → { "phases": { "created": ["faz-1"], "updated": [], "deleted": [], "unchanged": [] },
        "topic_groups": { "created": ["temel-kavramlar", …], … } }
```

### `GET /api/v1/today`
The Bugün model for the server's current date.

```
200 → { "date": "2026-10-26", "weekday": "Pazartesi", "heading": "Pazartesi, 26 Ekim",
        "is_rest": false, "sessions": [ … ], "tomorrow": { … }, "notes": [] }
```

### `GET /api/v1/week?date=YYYY-MM-DD`
The Hafta model for the Monday–Sunday week containing `date` (defaults to
today). Seven `days` (rest days included), weekly totals, plan week number,
phase and generated `notes`.

```
200 → { "start": "2026-10-26", "end": "2026-11-01", "range_label": "26 Eki–1 Kas",
        "week_number": 1, "phase_name": "Faz 1 · İşlem temeli", "days": [ … ],
        "planned_min": 540, "done_min": 90, "sessions_planned": 5, "sessions_done": 1,
        "moved_count": 1, "notes": ["Salı yapılamadı; M1·2 cumartesi bloğuna eklendi."] }
```

### `GET /api/v1/progress`
The İlerleme model: plan position, totals, hours (decimal comma), per-group
completion, exam countdown (from `WORKEY_EXAM_DATE`) and the 38-week strip.

```
200 → { "has_plan": true, "week_number": 1, "weeks_total": 38,
        "sessions_done": 1, "sessions_planned": 5, "done_min": 90, "planned_min": 540,
        "done_hours": "1,5", "planned_hours": "9", "current_phase": "Faz 1 · İşlem temeli",
        "days_to_exam": 265, "exam_date": "18 Tem",
        "groups": [ { "key": "temel-kavramlar", "name": "Temel kavramlar",
                      "dgs_min": 8, "dgs_max": 10, "planned_min": 180, "done_min": 90,
                      "done_fraction": 0.5, "percent": 50 }, … ],
        "strip": [ { "week_number": 1, "phase_key": "faz-1",
                     "phase_name": "Faz 1 · İşlem temeli", "planned_min": 540,
                     "done_min": 90, "done_fraction": 0.1667 }, … ] }
```

### `POST /api/v1/sessions/{id}/complete`
Mark a session done. Allowed for today and past; a future session → `422` with
`code: "future"`; an unknown id → `404`. Idempotent (completing twice writes one
audit row). Returns the updated session view.

```
200 → { "id": 1, "date": "2026-10-26", "state": "done", "completed": true, … }
```

### `DELETE /api/v1/sessions/{id}/complete`
Undo completion, under the same rules. Returns the updated session view
(`completed: false`).
