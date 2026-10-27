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
- [x] HTTP/içerik testleri tüm görünüm, auth, dinamik tarih, API örneği ve prompt senaryolarını kapsıyor ve geçiyor.
- [x] Mobil/masaüstü, klavye, kopyalama ve pano hatası kontrollerinin sonuçları kaydedildi.
- [x] Kullanım belgeleri nihai davranışla uyumlu; gerçek token veya kullanıcı planı test verisi olarak kullanılmadı.
- [x] Generated templ/CSS güncel; build ve vet başarılı; sınırlamalar açıkça raporlandı.

## Checks
- `go tool templ generate`
- Sınıf/token değiştiyse `./bin/tailwindcss.exe -i web/styles/input.css -o web/static/app.css --minify`
- `go test ./internal/web/... ./internal/apidoc/...`
- `go build ./...` ve `CGO_ENABLED=0` ortamıyla `go build ./...`
- `go vet ./...`
- `gofmt -l .` ve `git diff --check`
- Yukarıdaki tarayıcı kabul senaryoları; mevcut kişisel veriden ayrı geçici test ortamı kullan.

## Progress
Completed 2026-10-27. docs/mcp.md ve docs/api.md görünüm, tam/ayrı istem
kopyalama, istemci seçimi, betiksiz kullanım ve plan şablonu akışıyla güncellendi.
Tarih/kuralların asıl kaynağı AgentGuide ve ortak içerik modeli olarak korundu.

TestUsagePagesAndDownload üç istemci × iki görünüm için auth/no-store
kontrolünü genişletir; OpenAPI attachment/no-store ve sözleşme bayt eşitliğini
doğrular. TestAllUsagePromptTargets üç istemcide altı ayrı istemin textarea ve
ortak kaynak eşitliğini doğrular. Mevcut testler dinamik İstanbul/sınav tarihini,
tam Markdown'ı, API örneklerini ve kritik sınav istemi talimatlarını kapsar.

Passed (Go 1.27.3): templ generate (no diff), go test ./internal/web/...
./internal/apidoc/..., go test ./internal/api/..., go vet ./..., go build ./...,
CGO_ENABLED=0 go build ./..., gofmt -l . (empty), git diff --check.
Build modül stat-cache erişim uyarısı verdi; çıkış kodları 0.
Sınıf/token değişmedi; CSS rebuild gerekmedi, generated dosyalar değişmedi.

Uygulama içi tarayıcıda 390×844 ve 1280×800: iki sayfa, iki görünüm, tüm MCP
istemcileri ve kapalı/açık bölümler geçti. Sayfa genişliği 375/1265 px, yatay
taşma yok; uzun kod kendi alanında kayar. Enter/Boşluk/Tab, görünür odak,
readonly seçim ve canlı durum metinleri doğrulandı. MCP tam Markdown + altı
istem ve API tam Markdown pano geri okumasıyla eşleşti (CRLF → LF).
Pano reddi ve destek yokluğu test vekiliyle üretildi: doğru elle kopyalama
bildirimi, tam seçim/odak, tekrar etkin düğme; yanlış başarı bildirimi yok.
Tüm betikler CSP ile engellendiğinde istemci/görünüm bağlantıları, açılır
bölümler ve Ctrl+A/Ctrl+C ile manuel pano eşitliği geçti.

Ayrıntılı kanıt/matris: docs/usage-verification.md, WKY-19 bölümü.
Gerçek .env/token/plan kullanılmadı; ayrı boş DB ve loopback sunucuları,
Git dışındaki bin/wky19-preview altında. İlk sandbox sunucusuna tarayıcı
erişemedi; aynı ayrı sunucu sandbox dışında çalıştırılarak testler tamamlandı.

Sınırlar: tarayıcının kalıcı pano/JavaScript ayarı değiştirilmedi; reddetme
test betiğiyle, betiksizlik CSP ile üretildi. Ayrı ekran okuyucu, diğer tarayıcı
motorları/macOS, yeni AI plan yazımı/kurulumu, tüm paket testleri ve güvenlik
taraması çalıştırılmadı. Ekran görüntüleri araç çıktılarında incelendi.
Uygulama hatası bulunmadı; değişiklikler belgeler ve regresyon testleriyle sınırlı.
WKY-12 beklemesi korunur; deploy/commit/push yapılmadı.
