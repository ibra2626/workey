# Çalışmaya başla / kaldığın yer — görev taslağı

Durum: SF-01..SF-05 tamamlandı (2026-10-29); sonraki görev SF-06.
Tasarım: [2026-10-29-study-flow-design.md](../../../docs/superpowers/specs/2026-10-29-study-flow-design.md)

## Bu geliştirmeye özel branch kuralı

Kullanıcı kararı — 2026-10-29:

- SF-01..SF-06 kapsamındaki tüm geliştirmeler ve deney düzeltmeleri
  `codex/study-flow-experiment` branch'inde yapılacak.
- Mevcut proje klasöründe çalışılacak; Git worktree, `.worktree` klasörü veya
  ayrı checkout oluşturulmayacak.
- Bu kural yalnız bu çalışma akışına aittir; projenin diğer işleri veya global
  Codex/agent ayarları için genel bir tercih oluşturmaz.
- Deneme sonucu değerlendirilmeden main'e birleştirme yapılmayacak.
  Commit, push, merge ve deployment ayrıca kullanıcı isteği gerektirir.

Branch oluşturuldu; plan ve görev dosyaları aynı klasörde korunuyor.
Bu kayıt ürün kodu uygulamasına başlandığı anlamına gelmez.

| ID | Görev | Durum | Bağımlılık |
| --- | --- | --- | --- |
| SF-01 | [Veri modeli ve başlanmış seansı koruma](tasks/SF-01.md) | tamamlandı | kullanıcı başlama talimatı |
| SF-02 | [Başlat/bırak/devam servisleri ve atomik tamamlama](tasks/SF-02.md) | tamamlandı | SF-01 |
| SF-03 | [Okuma modelleri, REST/MCP sözleşmeleri](tasks/SF-03.md) | tamamlandı | SF-02 |
| SF-04 | [Bugün çalışma akışı ve yer imi formu](tasks/SF-04.md) | tamamlandı | SF-03 |
| SF-05 | [Hafta, ilerleme ve asistan kaynak rehberi](tasks/SF-05.md) | tamamlandı | SF-03, SF-04 |
| SF-06 | [Entegrasyon, görsel kabul ve doküman eşitleme](tasks/SF-06.md) | taslak | SF-01..05 |

Çalışma kuralları: Go 1.27.3, mevcut Go/templ/htmx/Tailwind mimarisi.
Gerçek plan, kişisel veritabanı veya .env test verisi değildir. Örnek kaynaklar
açıkça temsili olarak etiketlenir. Eski migration değiştirilmez.
Kod aşamasında her görevin odaklı kontrolleri çalıştırılır; SF-06 bütünleştirir.
CURRENT.md SF-05 sonucunu ve SF-06 sonraki adımını gösterir; WKY-20 zoom kabulü ayrı bekler.
Commit, push ve deployment bu planın parçası değildir.
