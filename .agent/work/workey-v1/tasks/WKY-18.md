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
- [ ] Şablon tüm kullanıcı alanlarını, güncel sınav tarihini ve değiştirilebilir çalışma ritmini içeriyor.
- [ ] Okuma → eksikleri sorma → outline → haftalık plan yazımı → tekrar okuma sırası açık.
- [ ] Tam liste semantiği, kayıtları koruma, referans uydurmama ve hata halinde durma talimatları mevcut.
- [ ] Kopyalanan prompt ile gösterilen prompt aynı; tam rehber Markdown'ında da bulunuyor.
- [ ] Seans tamamlama veya sayfa üzerinden otomatik plan yazma yok.

## Checks
- Sahte saat/sınav tarihiyle prompt içeriği ve kritik talimatlar için odaklı testler ekle.
- `go tool templ generate`; gereken Tailwind üretimi.
- `go test ./internal/web/...`
- Mevcut veriyi okuma, eksik referans, kısmi son hafta, skipped ve yarıda kalan yazma senaryolarını prompt metni üzerinden gözden geçir. Gerçek plana yazma yapma.

## Progress
Todo. Sonraki görev: WKY-19.
