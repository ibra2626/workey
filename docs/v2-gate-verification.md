# v2 kaynak kataloğu — uçtan uca kapı doğrulaması (WKY-28)

4 Kasım 2026, Go 1.27.3. Bu not, workey'in sabit DGS/kitap takipçisinden genel
kaynak kataloğuna genelleştirilmesini uçtan uca kanıtlar ve v2 kapısını kapatır:
bir kullanıcı, ajan üzerinden kendi kaynaklarını tanımlayıp planlayabilir —
kodda hiçbir değişiklik ve hiçbir DGS/kitap varsayımı olmadan.

## 1. Uçtan uca koşu (canlı /mcp + web, tek store)

Kalıcı ve yeniden koşulabilir kayıt: `internal/mcpserver/webworkflow_test.go`
içindeki `TestCatalogEndToEnd`. Gerçek localhost HTTP üzerinden `/mcp` uç noktası
MCP istemcisiyle sürülür; web ekranları aynı store'dan gerçek HTTP GET/POST ile
okunur. Hiçbir gerçek veritabanı, `.env`, token veya kullanıcı planı kullanılmaz;
geçici SQLite ve sabit saat (2026-10-26 09:00 İstanbul) ile çalışır.

Araç sırası ve kanıtlanan sözleşme:

1. `save_source` — **kitap** türü kaynak (`kind=book`, `location_unit=page`):
   "Tenses" bölümü, iki `Antrenman` ve bir `Test` birimi, sayfa aralıklarıyla.
   `created=true`.
2. `save_source` — **video** türü kaynak (`kind=video`, `location_unit=lesson`):
   "Limit" bölümü, üç `Ders` birimi, ders numarası aralıklarıyla.
3. `get_sources` — iki kaynak; `book/page` ve `video/lesson` türleri; birim
   label'ları kaynağın kendi sözcüğü (`Antrenman`, `Test`, `Ders`) olarak geri
   okundu; birim id'leri `set_unit_result` için alındı.
4. `save_plan` — her iki kaynağa referans veren hafta: bugünün akşam bloğu kitabı
   (GR·1, Antrenman/Test birim aralığı), yarınki blok videoyu (VD·1, Ders aralığı)
   çalışır. İki seans oluşturuldu; `get_plan` notasyonu `GR·1 … Antrenman` ve
   `VD·1 … Ders` referanslarını taşıdı.
5. `set_unit_result` — kitabın ilk Antrenman'ı D/Y/B ile (doğru 8, yanlış 2, boş
   0); videonun ilk Ders'i D/Y/B **olmadan**. Dönen görünüm: kitap biriminde
   doğru=8 ve notasyon `GR·1`; video biriminde D/Y/B alanları null.
6. Ekranlar ve bağımsızlık:
   - Bugün (`/`): birim listesi başlığı, `GR·1, Antrenman 1` ve `GR·1, Test 1`
     tam birim notasyonu, birim "Bitti" kontrolü, kitap kaynağının D/Y/B girişleri
     (`name="correct"`) göründü.
   - Hafta (`/hafta`): videonun `VD·1, Ders 1` birim referansı göründü.
   - İki birim bitti işaretliyken `get_progress` **SessionsDone=0** — birim
     işaretleme seansı tamamlamadı.
   - Bugünün seansı gerçek web formuyla (CSRF ile `/oturum/{id}/tamamla`)
     tamamlandı → `get_progress` **SessionsDone=1**, ancak `unit_progress` satır
     sayısı 2'de kaldı: seans tamamlama birim ilerlemesine dokunmadı.
   - İlerleme (`/ilerleme`): "Birim ilerlemesi", dürüst payda "6 birim" (katalogdaki
     toplam birim = 3 kitap + 3 video) ve "Çözülen" özeti göründü.

Bu test WKY-28 ile varsayılan koşuya alındı. WKY-26/27 web view'lerini yeniden
kurana kadar `-tags webworkflow` ardında bekliyordu; build tag kaldırıldı,
`newMCP`/`newWorkflow` yorumları güncellendi. `TestVisualReview` hâlâ
`WORKEY_VISUAL_REVIEW=1` ile opt-in'dir.

## 2. DGS kalıntıları için grep kapısı

Kodda (`*.go`) `bookEnum`, `dgs_question_*` ve `dgsQuestion*` tanımlayıcısı yok;
eşleşmeler yalnızca `.agent/` görev/karar dosyalarında (tarihçe). `M1/M2/M3/G1/G2`
eşleşmeleri yalnızca kullanıcı-tanımlı kaynak `key` örnekleri (test fixture'ları,
yorumlar, notasyon örnekleri) ve generic `kind="book"` etiketi — sabit beş-kitap
enum'u değil.

Şema (`internal/store/migrations/0001_init.sql`) temiz v2:
- `sources.key` serbest `TEXT UNIQUE` (sabit liste yok).
- `sources.kind` CHECK yalnızca generic `book|video|course|other`.
- `sources.location_unit` CHECK yalnızca `page|minute|lesson|other`.
- `units.label` serbest `TEXT` (kod label sözcüklerini saymaz).
- `topic_groups.dgs_question_*` → generic `question_*`.
- `sessions.kind` CHECK düşürüldü (açık `TEXT` etiket).

`get_agent_guide` çıktısı (`internal/plan/guide.go`) tümüyle katalog tabanlı:
`source_kinds`, `location_units`, `session_kinds` örnek olarak listelenir; rehber
kuralları/iş akışı save_source → get_sources → save_plan → set_unit_result
akışını anlatır ve sabit kitap/DGS sözcüğü içermez (bir test bunu zorlar:
`TestAgentGuide`). `docs/mcp.md` ve `docs/api.md` da katalog akışını anlatır.

Bu görevde düzeltilen tek kalıntı: `get_agent_guide` **araç açıklaması** hâlâ
"supported books/session kinds" diyordu → "supported source kinds, location units
and session kinds" olarak güncellendi (hiçbir test bu dizgeyi zorlamıyordu).

Sabit sınav tarihi ve 38 hafta çerçevesi **bilerek** korundu — bunlar şema kısıtı
değil, kaydedilmiş plan içeriği (faz tarihleri) ve kurulum başına yapılandırmadır:
`WORKEY_EXAM_DATE` (varsayılan `2027-07-18`, `internal/config/config.go`) ve
`phases` tablosundaki tarih aralıkları. 38 hafta `get_progress` / İlerleme'de bu
plan içeriğinden türetilir; şema çalışması gerektirmeden genelleşir.

## 3. v1 kapısının yeniden koşusu

| Kontrol | Sonuç |
| --- | --- |
| `go tool templ generate` | 0 güncelleme (üretilen `*_templ.go` güncel) |
| `./bin/tailwindcss.exe … --minify` | `web/static/app.css` yeniden derlendi |
| `go build ./...` | başarılı |
| `CGO_ENABLED=0 go build ./...` | başarılı (modernc.org/sqlite saf Go) |
| `go vet ./...` | temiz |
| `gofmt -l .` | LF üzerinde temiz; çalışma ağacındaki CRLF listesi autocrlf artefaktı (her dosya CR soyulduktan sonra gofmt-temiz doğrulandı) |
| odaklı test seti | config/clock/store/plan/web/web·views/api/apidoc/mcpserver/cmd → tümü ok |
| `govulncheck@v1.9.0 ./...` | **No vulnerabilities found** |

Odaklı test komutu (mcpserver artık varsayılan olarak cross-transport web akışını
da içerir):

```bash
go test ./internal/config/... ./internal/clock/... ./internal/store/... \
  ./internal/plan/... ./internal/web/... ./internal/api/... \
  ./internal/apidoc/... ./internal/mcpserver/... ./cmd/...
```

## Sınırlar

Uçtan uca koşu, gerçek bir dış AI istemcisi (Codex/Claude Desktop) değil, resmi
Go MCP SDK istemcisiyle canlı `/mcp` HTTP uç noktasını sürer; WKY-19'daki gerçek
CLI kabulü ayrı bir sonuçtur ve burada tekrarlanmadı. Görsel inceleme bu görevde
yeniden yapılmadı (WKY-27'de 390×844 ve 1280×800 geçti). Commit, push ve PR
yapılmadı; WKY-12 deploy beklemesi sürüyor.
