# Tamamlama ve hareket geri bildirimi — 27 Ekim 2026

Impeccable animate/polish ve mevcut DESIGN.md kullanıldı. Gerçek plan yerine
test fixture'larıyla ayrı, geçici SQLite üzerinde incelendi.

- Bugün: eylem “Seansı tamamla”, sonuç “Tamamlandı” olarak ayrıldı.
  54/56px hedef, pointer, hover, basılma ve görünür klavye odağı korundu.
- Bugün ve Hafta: tamamlama/geri alma, 0 → 90 → 0 dakika ve 0 → 1 → 0 seans
  değişimleri doğrulandı. İçerik güncellemesi sonrası klavye odağı aynı
  seansın yeni düğmesine taşındı.
- İki saniye geciktirilen örnek istekte spinner, Kaydediliyor metni, üst yükleme
  çizgisi, aria-busy ve devre dışı düğmeler doğrulandı. Yapay 503 sonrasında
  hata mesajı gösterildi; düğmeler ve odak geri geldi, tekrar deneme başarılıydı.
- 390×844, 768×1024 ve 1280×800 Bugün görünümleri incelendi. Mobilde ilk panel
  kaydırmadan erişilebilir; ara genişlikteki taşma tek sütunla düzeltildi.
  Hafta mobil görünümünde 44px işlem hedefleri ve taşma kontrol edildi.
- Sayfa geçişleri, giriş, rehber açılması ve kopyalama sonucu ortak kısa hareket
  kurallarını kullanır. Azaltılmış hareket CSS'i kaynakta incelendi; işletim
  sistemi tercihi değiştirilerek canlı emülasyon ve ekran okuyucu testi yapılmadı.
- Statik sürüm özeti tüm gömülü dosyalardan hesaplanarak betik değişikliklerinin
  eski önbellekte kalması önlendi. Yeni animasyon kütüphanesi eklenmedi.

Templ/Tailwind çıktıları yeniden üretildi. Geçici önizleme sunucusu ve test
dosyası kaldırıldı; ürün verisi değiştirilmedi.

## Grafik yükselmesi

Kullanıcı isteğiyle seans profili, haftalık çubuklar, faz blokları, 38 haftalık
şerit ve konu dolumlarına tabandan yukarı 600ms clip-path açılışı eklendi.
Dinlenme çizgileri sabit. Sayfa açılışı ve htmx tamamlama sonrası animasyon
sınıfı/süresi tarayıcıda doğrulandı; mobil Bugün/Hafta/İlerleme görünümleri
ve masaüstü Bugün incelendi. Mobilde yatay taşma yok. Azaltılmış hareket
tercihinde bu animasyon tanımlanmaz; grafik doğrudan görünür.

## Yükleme ve akıcılık düzeltmesi

Fontlar CSS sonrasında keşfediliyordu. İlk ekranda kullanılan Barlow ağırlıkları
ve Türkçe alt kümeleri HTML'den preload edilir. `font-display: optional`, yavaş
ilk yüklemede yazının sonradan değişip yerleşimi oynatmasını önler. Rehber ve
giriş sayfaları daha küçük preload kümesini kullanır.

Sayfa görüntüsünü donduran View Transition ve tüm içerik fade'i kaldırıldı.
Grafik dolguları clip-path yerine 650ms transform ile hareket eder; seans
etiketleri ayrı, sabit katmanda kalır. Dikey çubuklar alttan, yatay konu
çizgileri soldan açılır. Kalıcı will-change/GPU katmanı kullanılmaz.

Aynı masaüstü tarayıcıda, farklı localhost portlarıyla önbelleksiz iki örnek:

| Ölçüm | Önce | Sonra |
| --- | ---: | ---: |
| İlk font isteğinin başlangıcı | 184 ms | 23 ms |
| Fontların hazır olması | 311 ms | 277 ms |
| İlk içerik çizimi | 268 ms | 280 ms |
| CLS | 0,00133 | 0 |
| En uzun kare aralığı (90 örnek) | 33,5 ms | 33,5 ms |
| 25 ms üstündeki kare sayısı | 2 | 2 |

Tek yerel örnek genel FPS veya ağ hızını temsil etmez: toplam ilk çizim/FPS
iyileşmesi iddia edilmiyor. Doğrulanan değişim, font isteğinin erken başlaması,
font kaynaklı kaymanın bu örnekte kalkması ve çakışan geçişlerin kaldırılmasıdır.
