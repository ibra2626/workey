# WKY-18 — Sınava kadar çalışma planı oluşturma promptu

## Goal
Kullanıcı, doldurup bağlı asistanına gönderebileceği tek bir promptla bugünden sınava kadar çalışma planını MCP üzerinden hazırlatabilsin.

## Scope
- Önkoşul: WKY-17. Ortak kurallar: `../README.md`.
- MCP rehberindeki son adıma Sınava kadar çalışma planımı oluştur başlığı ve ayrı Promptu kopyala düğmesi ekle. Prompt ortak içerik modelinde yer alsın; tam Markdown'a dahil edilsin.
- Açıklamada kullanıcıya alanları doldurup Workey'e bağlı asistanına göndermesini söyle. Düğme yalnızca metin kopyalasın; web sayfası plan üretmesin veya MCP çağrısı yapmasın.
- Doldurulabilir alanlar: mevcut seviye, uygun gün/saatler, seans süreleri, dinlenme günleri, kitap/bölüm/sayfa referansları. Dört akşam × 90 dakika ve bir hafta sonu × 180 dakika, düzenlenebilir başlangıç önerisi olarak sunulsun.
- Sayfadaki sınav tarihi mevcut AgentGuide'dan gelsin. Prompt çalıştırıldığında güncel today, timezone ve exam_date değerlerinin get_agent_guide ile yeniden okunmasını istesin; sabit örnek tarihleri kullanmasın.
- Prompt sırası: get_agent_guide → get_outline → ilgili get_plan → get_progress; hedef ve eksik alan/referansları kullanıcıyla netleştir; bugünden sınava kadar faz, konu grubu ve seansları hazırla. Eksik kitap içeriği, bölüm veya sayfa referansı uydurulmasın.
- Gerekliyse korunacak mevcut faz/grupları da içeren tam listeyle save_outline çağrılsın. Ardından plan haftalık, çakışmayan aralıklarla yazılsın; ilk/son aralık bugüne ve sınav tarihine göre sınırlansın.
- Her hafta yazmadan önce mevcut planı oku ve korunacak açık seansları tam listeye dahil et. Promptun yanında açık seansların tam listeyle değiştirildiğini görünür metinle belirt. Geçmiş ve tamamlanmış kayıtlar korunsun.
- Her yazmada created/updated/deleted/unchanged/skipped sonuçları incelensin; plan ve ilerleme tekrar okunarak doğrulansın. Başarısız hafta veya doğrulama hatasında durulsun; tamamlanan ve kalan aralıklar bildirilsin. Skipped başarı diye sunulmasın.
- Sonuç Türkçe özetlensin: dönem, seanslar, değişen/silinen/atlanan kayıtlar ve kalan iş. complete_session çağrısı istenmesin; doldurulmamış alanlarla yazmaya başlanmasın.

## Acceptance criteria
- [x] Şablon tüm kullanıcı alanlarını, güncel sınav tarihini ve değiştirilebilir çalışma ritmini içeriyor.
- [x] Okuma → eksikleri sorma → outline → haftalık plan yazımı → tekrar okuma sırası açık.
- [x] Tam liste semantiği, kayıtları koruma, referans uydurmama ve hata halinde durma talimatları mevcut.
- [x] Kopyalanan prompt ile gösterilen prompt aynı; tam rehber Markdown'ında da bulunuyor.
- [x] Seans tamamlama veya sayfa üzerinden otomatik plan yazma yok.

## Checks
- Sahte saat/sınav tarihiyle prompt içeriği ve kritik talimatlar için odaklı testler ekle.
- `go tool templ generate`; gereken Tailwind üretimi.
- `go test ./internal/web/...`
- Mevcut veriyi okuma, eksik referans, kısmi son hafta, skipped ve yarıda kalan yazma senaryolarını prompt metni üzerinden gözden geçir. Gerçek plana yazma yapma.

## Progress
Completed 2026-10-27. Step four now contains the exam-plan prompt, editable
placeholders and a separate Promptu kopyala control. The shared UsagePrompt
block supplies the displayed text, copy target and full Markdown. ExamDate is
read from AgentGuide for visible guidance; the prompt reads current dates via
get_agent_guide at execution time. Existing shorter prompts retain three rows;
the planning prompt uses eighteen rows and the existing copy.js behavior.

Passed on Go 1.27.3: go tool templ generate; Tailwind rebuild (no CSS diff);
go test ./internal/web/...; go vet ./internal/web/...;
go build -o bin/wky18-preview/workey.exe ./cmd/workey; gofmt; git diff --check.
Build emitted a module stat-cache access warning but exited successfully.
Focused tests use a fake Istanbul midnight and alternate exam date, cover
critical instructions and confirm prompt/Markdown/copy-target parity for all clients.
Headless Edge at 390x844 and 1280x844 passed all clients, keyboard/touch,
clipboard readback (Windows CRLF normalized), Markdown parity, no overflow,
closed details and JS-disabled manual selection. Screenshots visually inspected.
The first sandboxed browser launch failed; browser and disposable test server
then ran together outside the sandbox. Both test servers stopped. Temporary
binaries, DBs, checks and screenshots are ignored under bin/wky18-preview.

Prompt scenarios reviewed: existing open sessions are reread before each week
and preserved in the full list; missing references stop before writes; partial
last week ends at exam_date; skipped entries are reported and blocking skips
stop subsequent writes; interrupted writes report verified and remaining ranges.
No real plan writes, secrets or real DB access. Full suite, CGO-disabled build
and vulnerability scan not run for this guide-only change. User requested a local
commit after verification; no push requested.
Next task: WKY-19.
