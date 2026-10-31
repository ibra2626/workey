# Workey REST API — `/api/v1`

## Uygulama içi rehber ve makine sözleşmesi

İlk kurulum, API tokenı üretimi, web parolası ve oturum anahtarı için
[README'nin yerel kurulum adımlarını](../README.md#yerel-kurulum) izleyin.
Tokenı kendiniz oluşturup `.env` dosyasının `WORKEY_API_TOKEN` alanına kaydedersiniz;
Workey token dağıtan bir endpoint sunmaz. Mevcut kurulumda etkin sunucu tokenını
kullanın: süreç ortamı `.env` ayarını geçersiz kılabilir. `Read-Host` yalnızca
bu mevcut değeri girmenizi sağlar; yeni token oluşturmaz.

Menüden **API Kullanımı** (`/kullanim/api`) sayfasına ulaşılır. Sayfa oturum
doğrulaması ister; endpoint listesi, tam temsili yanıtlar, PowerShell/curl.exe
komutları ve ortak ajan kurallarını gösterir. Mobilde **Kullanım** sekmesini açın.

**Rehber** görünümünde endpoint başlıkları kapalı açılır bölümlerdir; tıklama,
Enter veya Boşluk ile tam yanıt örneklerini açın. Uzun komutlar ve JSON örnekleri
kendi kod kutusunda yatay kaydırılır; kutular klavyeyle odaklanabilir.
**Markdown** bağlantısı salt okunur tam rehberi açar.
**Markdown olarak kopyala**, kapalı endpoint örnekleri dahil tüm Markdown'ı
kopyalar; başarıdan sonra **Kopyalandı** bildirimi görünür. Bu işlem API isteği
göndermez. OpenAPI indirme bağlantısı ayrı makine sözleşmesini indirir.

Pano desteklenmiyorsa veya erişim reddedilirse Markdown alanı açılıp odaklanır
ve metin seçilir; hata bildiriminin ardından Ctrl+C (macOS: ⌘C) ile elle
kopyalayın. JavaScript kapalıyken düğme gizlenir; **Markdown** bağlantısı ve
native açılır bölümler kullanılabilir. Salt okunur alanda Ctrl+A ve Ctrl+C
ile tam metni seçip kopyalayabilirsiniz. Her iki görünüm ve OpenAPI indirmesi
web oturumu ister ve `Cache-Control: no-store` döner.

**İlk API isteğiniz** bölümünde sunucu adresi, PowerShell 7 ve gizli token girişi
ile ilk okuma örneği bulunur. **Planı okuyun ve yazın** bölümünde mevcut kayıtları
okuma, tam-liste gövdeleri, silme uyarısı, yazma ve ilgili aralığın tamamını tekrar
okuma sırası gösterilir. Ardından **Endpoint referansı** ve OpenAPI indirmesi gelir.
Temsili yanıtlar, tamamlama/geri alma, curl alternatifi ve hatalar ayrı gruplardır.

**Bu sayfada** dizini geniş ekranda sağda, dar ekranda açılır listededir; gerçek
bölüm bağlantıları JS olmadan çalışır. Her kod örneğinin **Kodu kopyala** düğmesi
vardır. Pano reddinde/yokluğunda kod odaklanıp seçilir ve elle kopyalama bildirimi
gösterilir. JS kapalıyken bu düğmeler gizlidir; kod metni seçilebilir kalır.

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
exactly the given unstarted set; past and completed sessions are frozen and returned in
`skipped`). A session with any study record is also protected: identical input is
`unchanged`, changed input is `skipped` with `reason=started`, and omission retains
it. Past/completed reasons take priority. Optional `Idempotency-Key` header.
Body is a `SavePlanRequest`.

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

<!-- study-contract:start -->
## Çalışma durumu ve yer imini okuma

SessionView\.study opsiyonel, salt okunur çalışma bağlamıdır: status=active/paused/closed, first\_started\_at/last\_started\_at/updated\_at (UTC), entry\_mode=normal/ten\_minutes ve version\. bookmark varsa item\_ordinal (sıfır tabanlı), stage (akşam: konu/soru/yanlis\_defteri; diğer: calisma), resume\_page, çözülmüş item ve notation içerir\. resume\_page devam edilecek kitap sayfasıdır; yarım sayfa aynı kalır, otomatik \+1 yapılmaz\. Yanlış defteri aşamasında da kitap sayfasıdır\. Yer imi yoksa uydurma; study yokluğu hiç çalışılmadığını kanıtlamaz\. SessionState planned/today/done/missed/moved değişmez; geçmiş kısmi emeği ayrıca study üzerinden oku\. GET /api/v1/today active\_study ve latest\_paused\_study ile tarih dışındaki devam bağlamını da sunar; aktif önceliklidir, yarım işe dönüş açık seçimdir, bugünün planı korunur\. get\_plan yalnız istenen aralığı döndürür; gerekirse geçmiş aralığı da oku\. Yeni çalışma yazma REST endpoint'i veya MCP aracı yoktur\. Kısmi sayfa ve ten\_minutes hedefi ilerlemeyi artırmaz; done\_min tamamlanan seansların planlı süresidir, ölçülmüş süre veya öğrenme yüzdesi değildir\. Telafi bağlantısı yer imini veya tamamlamayı otomatik taşımaz; eski ve telafi ayrı tamamlanırsa mevcut toplam ikisini de sayar\.

26 Ekim 2026 tarihli temsili okuma → yazma → tekrar okuma örneği; gerçek kullanıcı planı değildir\. Önce ajan rehberini, outline, ilgili get\_plan / GET /api/v1/plan aralığını ve ilerlemeyi okuyun\. Aşağıda yarım kalmış örnek seansın okuma yanıtı var\.

```json
{
  "study": {
    "status": "paused",
    "first_started_at": "2026-10-26T17:30:00Z",
    "last_started_at": "2026-10-26T17:30:00Z",
    "updated_at": "2026-10-26T17:40:00Z",
    "entry_mode": "ten_minutes",
    "version": 2,
    "bookmark": {
      "item_ordinal": 0,
      "stage": "soru",
      "resume_page": 5,
      "item": {
        "book": "M1",
        "chapter_no": 1,
        "chapter_title": "Temel kavramlar",
        "page_from": 1,
        "page_to": 10
      },
      "notation": "M1·1, s. 1–10"
    }
  },
  "id": 1,
  "date": "2026-10-26",
  "start_time": "20:30",
  "duration_min": 90,
  "kind": "aksam",
  "completed": false,
  "state": "today",
  "topic_group_key": "temel",
  "topic_group_name": "Temel kavramlar",
  "items": [
    {
      "book": "M1",
      "chapter_no": 1,
      "chapter_title": "Temel kavramlar",
      "page_from": 1,
      "page_to": 10
    }
  ],
  "notation": "M1·1, s. 1–10"
}
```

Plan yazarken study ve diğer okuma alanlarını göndermeyin\. Başlanmış seansın items, date, start\_time ve duration\_min değerlerini koruyun: aynı içerik unchanged, değiştirme denemesi skipped reason=started, istekten eksikse korunur\. past\_date ve completed nedenleri önceliklidir\. Örnekte duration\_min=100 yapmayı deneyen eski bir istemci aşağıdaki sonucu alır; seans 90 dakika ve yer imi aynı kalır\.

```json
{
  "created": null,
  "updated": null,
  "deleted": null,
  "unchanged": null,
  "skipped": [
    {
      "date": "2026-10-26",
      "start_time": "20:30",
      "reason": "started"
    }
  ]
}
```

Bu sonucu başarıyla uygulanmış değişiklik saymayın\. Yazdıktan sonra aynı aralığı ve ilerlemeyi tekrar okuyun; seansın süre, kaynak ve yer iminin korunduğunu doğrulayıp atlanan değişikliği kullanıcıya açıklayın\. Telafi gerekiyorsa kaynak geçmiş seansı değiştirmeden makeup\_of ile ayrı seans hazırlayın; yer imi otomatik taşınmaz\.

Mevcut complete\_session ve REST tamamlama/geri alma çağrıları ek alan istemez: tamamla mevcut study durumunu atomik closed yapar; geri al yer imini koruyup paused yapar\. Eski seans için çalışma kaydı oluşturmaz\. Bu REST/MCP yazmaları study\.version denetlemez; eski sekme niyeti tespit edilemez\.

<!-- study-contract:end -->

<!-- planning-inputs:start -->
## Plan için kaynaklarını hazırla

Materyalleri Workey'e bağlı asistanına ver\. Workey dosya yüklemez, PDF saklamaz veya OCR yapmaz; asistanın görmediği özel bir kaynağı okuduğunu varsayma\.

- Her kitabın adı, kitap kodu ve baskısı; okunaklı içindekiler ve bölüm sırası\.
- Gerekli örnek sayfalar veya erişebildiğin PDF; basılı sayfa numarası ile PDF görüntüleyicisinin sayfa sırası arasındaki fark\.
- Doğrulanabilen bölüm ve sayfa aralıkları; varsa konu, soru ve antrenman aralıkları\. Eksik referansı asistana sor; tahminle doldurtma\.
- Mevcut seviye, bitirilen bölümler, hedef, uygun gün ve saatler, dinlenme ve diğer zaman kısıtları\.

İlk haftadan sonra gerçek temponu ve yükün uygunluğunu asistana anlat\. Açık sekme süresi veya başlangıç hedefi çalışma süresi sayılmaz\.

## İsteğe bağlı: ilk haftayla başla

Sınava kadar plan oluşturma istemi kullanılabilir\. Alternatif olarak ilk haftayı ayrıntılı, uzak dönemi yalnız kaba kapsam olarak konuşabilirsin\. Uzak dönem için doğrulanmamış sayfa veya seans yazılmaz\. Kopyalama planı değiştirmez; asistan önce okur, yazdıktan sonra yeniden okuyarak doğrular\.

```
Workey bağlantısıyla önce get_agent_guide (REST: GET /api/v1/agent-guide), get_outline, ilgili get_plan ve get_progress eşdeğerlerini oku. Güncel today, timezone ve exam_date değerlerini esas al.
Hedefim/seviyem: [doldur]. Uygun gün ve saatlerim: [doldur]. Doğrulanabilir kitap/baskı/içindekiler/sayfa kaynaklarım: [asistana ver].
İlk haftayı ayrıntılı, uzak dönemi kaba kapsam olarak birlikte hazırlayalım. Eksik bölüm veya sayfa referansı uydurma; eksikleri sor ve yanıtımı bekle. Uzak dönem için doğrulanmamış seans yazma. İlk uygulama haftasından sonra gerçek tempomu ve yükü bana sor; açık kalma zamanından süre hesaplama.
Yazılacak aralığı ve korunacak kayıtları benimle netleştir. Yazmadan hemen önce outline, ilgili plan ve study/yer imini yeniden oku. save_outline faz ve grupların, save_plan aralıktaki başlanmamış açık seansların TAM listesini değiştirir; eksik açık seanslar silinir. Korunacak kayıtları tüm kaynaklarıyla dahil et. Geçmiş, tamamlanmış ve başlanmış seansları koru; başlanmış seansın kaynak/tarih/saat/süresini değiştirme. Telafi bağlantılarını koru.
Her yazma sonucunda created/updated/deleted/unchanged/skipped listelerini incele. skipped reason=started dahil atlanan değişikliği başarı sayma; hata veya belirsizlikte dur, körlemesine tekrar yazma. İlgili aralığı, outline ve ilerlemeyi yeniden okuyup doğrula; kalan kapsamı bildir. Planlama sırasında seans tamamlama çağrısı yapma.
```

## Ara sonrası planı yeniden değerlendir

Bu istemi kopyalayıp Workey'e bağlı asistanına gönder\. Yer imi sunucudan okunur; kimlik veya token ekleme\. Kopyalama plan yazmaz ve eski işi otomatik bugüne taşımaz\. Asistan okuma → yazma → yeniden okuma sırasını izler\.

```
Workey bağlantısıyla ara sonrası planımı değerlendirelim.
Gerçek tempom, uygun zamanım ve değişen koşullarım: [doldur]. Eksik kaynaklarım veya henüz çalışmadığım kapsam: [bildiğimi yazacağım; eksikleri bana sor].
Önce get_agent_guide (REST: GET /api/v1/agent-guide) ile today, timezone ve exam_date değerlerini oku. Sınav geçmişse yazmadan hedefi benimle netleştir. get_outline, ilgili geçmiş/bugün/gelecek get_plan aralıkları ve get_progress eşdeğerlerini oku; büyük aralıkları haftalara böl.
study durumunu, yer iminin kitap/bölüm/aşama/resume_page değerlerini ve makeup_of/moved_to ilişkilerini incele. Yer imi devam edeceğim sayfadır; otomatik +1 yapma veya telafiye kopyalama. Kısmi çalışmayı tamamlanmış ya da ölçülmüş dakika sayma. Eski işi otomatik bugüne yükleme; eksik kapsamı ve yetişmeyecek işi benimle netleştir. Kitap/bölüm/sayfa uydurma; kaynak eksikse sor ve yanıtımı bekle.
Yazmadan hemen önce ilgili aralığı yeniden oku. Her study kaydı begun/başlanmış kabul edilir: kaynak, tarih, start_time, duration_min ve items korunur. save_plan aralıktaki başlanmamış açık seansların TAM listesini değiştirir; listede bulunmayanlar silinir. Korunacak açık seansları tüm items ile dahil et; silme/değiştirme kararını benimle netleştir. Geçmiş, tamamlanmış ve başlanmış seansları koru. Outline değişirse korunacak faz ve grupları da içeren tam listeyi gönder; kullanılan grupları silme.
Telafi gerekiyorsa kaynak seansı değiştirmeden makeup_of ile ayrı seans öner ve bağlantıyı doğrula. Eski seans ve telafisini otomatik tamamlamayacaksın; ikisini ayrı tamamlamak mevcut seans tabanlı toplamda ikisini de sayar, benzersiz içerik ölçmez.
Okuma alanları id/state/completed/notation/study yazma gövdesine girmez; topic_group_key yazmada topic_group olur. Her save_plan sonucunun created/updated/deleted/unchanged/skipped listelerini incele; skipped reason=started dahil atlananları başarı sayma. Hata, belirsiz yazma veya doğrulanamayan sonuçta dur; kalan haftaları yazma. Tekrar denemeden önce oku; aynı gövde için aynı idempotency_key kullan.
Yazdıktan sonra aynı get_plan aralığını, outline ve get_progress eşdeğerlerini yeniden oku. Yer iminin, korunacak kayıtların ve telafi ilişkilerinin korunduğunu doğrula; değişen ve eksik kalan kapsamı Türkçe bildir. Bu planlama sırasında complete_session veya REST tamamlama/geri alma çağrısı yapma.
```

<!-- planning-inputs:end -->
