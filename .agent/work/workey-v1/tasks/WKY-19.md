# WKY-19 — Kullanım rehberleri kabul kontrolleri ve dokümantasyon

## Goal
Markdown kontrolleri, sade MCP rehberi ve sınava kadar plan promptunun birlikte çalıştığını doğrula; kullanım belgelerini güncelle.

## Scope
- Önkoşul: WKY-18; WKY-15..18 tamamlanmış olmalı. Ortak kurallar: `../README.md`.
- `docs/mcp.md` ve `docs/api.md` içinde yeni görünüm/kopyalama akışını anlat. MCP belgesine istemci seçimi ve plan şablonunun nasıl kullanılacağını ekle; tarih ve kuralları ikinci bağımsız uygulama kaynağına dönüştürme.
- İki sayfayı 390×844 ve 1280×800 boyutlarında incele: normal/Markdown görünümü, tüm MCP istemcileri, açılır bölümler, uzun kodlar ve prompt metni.
- Klavyeyle görünüm/istemci seçimi, açılır bölüm kullanımı, görünür odak, salt okunur metni seçme ve kopyalama bildirimlerini doğrula. Sayfa seviyesinde yatay taşma olmasın; uzun metin/kod kendi alanında kaydırılabilsin.
- Panoya kopyalanan tam Markdown ve ayrı istem metinlerini beklenen kaynaklarla karşılaştır. Clipboard API erişimi reddedildiğinde/desteklenmediğinde elle kopyalama yolunu ve yanlış başarı bildirimi olmadığını kontrol et.
- JavaScript kapalıyken görünüm ve istemci bağlantıları ile elle kopyalama çalışsın. Oturumsuz erişim, no-store ve mevcut OpenAPI indirmesi tekrar doğrulansın.
- Bulunan sorunları düzelt; ilgili kontrolleri yeniden çalıştır. Kanıt ve kalan sınırlamaları bu dosyanın Progress bölümünde özetle; çalıştırılmayan kontrolü başarılı sayma.
- TASKS.md durumlarını ve `.agent/state/CURRENT.md` devir bilgisini gerçekleşen sonuçlara göre güncelle. WKY-12 deploy beklemesini koru; deploy/commit/push bu işin kapsamında değil.

## Acceptance criteria
- [ ] HTTP/içerik testleri tüm görünüm, auth, dinamik tarih, API örneği ve prompt senaryolarını kapsıyor ve geçiyor.
- [ ] Mobil/masaüstü, klavye, kopyalama ve pano hatası kontrollerinin sonuçları kaydedildi.
- [ ] Kullanım belgeleri nihai davranışla uyumlu; gerçek token veya kullanıcı planı test verisi olarak kullanılmadı.
- [ ] Generated templ/CSS güncel; build ve vet başarılı; sınırlamalar açıkça raporlandı.

## Checks
- `go tool templ generate`
- Sınıf/token değiştiyse `./bin/tailwindcss.exe -i web/styles/input.css -o web/static/app.css --minify`
- `go test ./internal/web/... ./internal/apidoc/...`
- `go build ./...` ve `CGO_ENABLED=0` ortamıyla `go build ./...`
- `go vet ./...`
- `gofmt -l .` ve `git diff --check`
- Yukarıdaki tarayıcı kabul senaryoları; mevcut kişisel veriden ayrı geçici test ortamı kullan.

## Progress
Todo. Test sonuçları, tarayıcı kanıtı ve varsa doğrulanamayan davranışlar tamamlandığında buraya yazılacak.
