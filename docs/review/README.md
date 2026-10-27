# v1 ekran incelemesi — 27 Ekim 2026

Örnek veriyle, uygulama içi tarayıcıda 390×844 ve 1280×800 ölçülerinde incelendi.
Kaynak: DESIGN.md ve mevcut v1 kararları. Gerçek plan/veritabanı değiştirilmedi.
Görsel senaryonun tarihi 28 Ekim; dinlenme 30, uzun blok 31 Ekim. İncelemede
27 ekran görüntüsü ve erişilebilirlik ölçüm çıktısı kullanıldı. Geçici çıktılar
kullanıcının isteğiyle commit öncesinde kaldırıldı; bulgular bu raporda tutuldu.
Senaryoları yeniden açma yönergesi docs/mcp.md içindedir.

## Bulgular ve düzeltmeler

- Giriş: tek alan, Parola etiketi, görünür 2px odak ve tek düğme doğrulandı.
  Paroladan Tab ile düğmeye geçiş denendi. Gerçek parola kullanılmadı.
- Bugün: htmx tamamlama/geri alma sonrası eski kalan özet rakamları TodayMain
  değiştirilerek düzeltildi. Tamamlama sonrası 90/540 dakika, 1/5 seans; geri
  alındığında sıfır gösterildi. Uzun bloklarda tüm kitap referansları listelendi.
- Hafta: yedi gün, dinlenme, yapılamayan ve taşınan seanslar kontrol edildi.
  Mobil başlık h1 yapıldı; tarama dokusu durum metniyle destekleniyor.
- İlerleme: bir tamamlanan seans sonrası 1,5/9 saat, 1/5 seans ve ilgili
  konuda %50 doğrulandı. Mobil başlık h1 yapıldı; skor/deneme sonuç paneli yok.
- Mobil kenar boşluğu 12px'e uyarlandı; yatay taşma bulunmadı. Boş durum metni
  sadeleştirildi. DESIGN.md mevcut v1 kararlarıyla uyumlu hale getirildi.

## Erişilebilirlik ve sınırlar

Görünür mobil metinlerde hesaplanan en düşük düz renk kontrastı
Bugün/Hafta/İlerleme için 4,60:1, Giriş için 5,23:1; 12px altında görünür metin
yoktu. Giriş ve gezinmede klavye odağı doğrulandı. Tamamlandı bildirimi
role=status taşıyor. Türkçe karakterler, lang=tr, ondalık virgül, font yüklemesi
ve konsol kontrol edildi. Bu inceleme tam WCAG sertifikasyonu değildir.
Reduced-motion koruması kaynakta incelendi; emülasyon ve canlı ekran okuyucu
testi yapılmadı.

## Teknik doğrulama

Go 1.27.3: templ/Tailwind üretimi, tüm testler, vet, normal ve CGO kapalı derleme,
gofmt ve diff boşluk kontrolü geçti. Tamamlama/geri alma sayaçları ve çok kaynaklı
blok için regresyon testleri eklendi. govulncheck v1.9.0: No vulnerabilities found.

WKY-12 kullanıcı kararıyla beklemede; Linux dağıtımı yapılmadı. İlk MCP doğrulaması
Go SDK ile yapıldı; ayrı Claude Code model oturumu çalıştırılmadı. Sonraki gerçek
Codex CLI kabulü ve Desktop sınırı docs/usage-verification.md içinde kayıtlıdır.
