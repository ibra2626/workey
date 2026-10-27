# MCP/API kullanım doğrulaması — 27 Ekim 2026

## WKY-19 — birleşik rehber kabulü

27 Ekim 2026'da uygulama içi tarayıcıda ayrı, boş SQLite test veritabanıyla
`/kullanim/mcp` ve `/kullanim/api` yeniden kontrol edildi. Mevcut `.env`, token,
kullanıcı planı veya kalıcı istemci ayarları kullanılmadı. Önizleme ayrı çalışma
dizininde `WORKEY_DEV=1` ile yalnızca loopback üzerinde çalıştı.

| Kontrol | Sonuç / kanıt |
| --- | --- |
| 390×844 ve 1280×800 | İki sayfanın Rehber ve Markdown görünümleri incelendi. MCP'de üç istemci ve tüm kapalı/açık bölümler kontrol edildi. Belge genişliği sırasıyla 375 ve 1265 px; sayfa yatay taşmadı. |
| Uzun içerik | Mobil API'de açılan 11 endpoint ve 12 taşan kod kutusu kendi alanında kaldı; masaüstünde altı kod kutusu kendi içinde kaydırılabilir. MCP'nin uzun Claude komutu masaüstünde de kendi kutusunda kaldı. Uzun plan istemi kendi metin alanında kaydırıldı. |
| Klavye | Görünüm ve istemci bağlantıları Enter ile, açılır bölümler Enter/Boşluk ile çalıştı. Tab ile kod kutusuna ulaşıldı. Metin alanında görünür 2 px odak, kodda tarayıcının 1 px odağı görüldü. Salt okunur metin Ctrl+A ile tamamen seçildi. |
| Başarılı kopyalama | MCP'nin tam Markdown'ı ve altı ayrı istemi, API'nin tam Markdown'ı panodan okunup textarea değerleriyle eşleşti. Windows satır sonları karşılaştırmada LF'ye normalize edildi. Başarı bildirimi beklenerek okundu. |
| Pano hatası | Test vekili `copy.js` öncesinde `writeText` için `NotAllowedError` üretti; ikinci senaryo Clipboard API'yi kaldırdı. Her iki sayfanın tam Markdown'ı ve sınav isteminde alan görünür/odaklı/tam seçili oldu, düğme yeniden etkinleşti; yalnızca elle kopyalama mesajı çıktı. |
| Betiksiz kullanım | Test yanıtlarına `Content-Security-Policy: script-src 'none'` eklenerek tüm sayfa betikleri engellendi. Üç istemci, görünüm bağlantıları, native açılır bölümler ve Markdown/plan istemi için gerçek Ctrl+A/Ctrl+C ile pano eşitliği geçti. Kopyalama düğmeleri gizli kaldı. |
| HTTP ve sözleşme | Tüm istemci/görünüm kombinasyonlarında oturumsuz erişim 303; oturumlu erişim 200/no-store. OpenAPI indirmesi no-store, attachment ve mevcut sözleşmeyle bayt eşitliği kontrollerini geçti. |
| Dinamik içerik | İstanbul gece sınırı, alternatif sınav tarihi, ortak HTML/Markdown kaynakları, API örneklerinin sözleşmeyle eşitliği ve tüm istemlerin kaynak metniyle eşitliği Go testleriyle geçti. |

Çalıştırılan kontroller (Go 1.27.3):

```powershell
go tool templ generate
go test ./internal/web/... ./internal/apidoc/...
go test ./internal/api/...
go vet ./...
go build ./...
$env:CGO_ENABLED = '0'
go build ./...
gofmt -l .
git diff --check
```

templ üretimi değişiklik oluşturmadı. Sınıf/token veya uygulama kaynağı
değişmediğinden CSS yeniden derlenmedi. Build'ler başarılı; Go'nun modül stat
önbelleğine yazarken verdiği erişim uyarısı çıkış kodunu etkilemedi.
Eklenen regresyon testleri: `TestUsagePagesAndDownload` için tüm istemciler ve
indirme başlıkları/sözleşme eşitliği; `TestAllUsagePromptTargets` için üç istemcide
altı ayrı textarea'nın ortak kaynak metinleriyle tam eşitliği.

Sınırlar: pano reddi test betiğiyle üretildi; tarayıcının kalıcı izin ayarı
değiştirilmedi. Betiksiz kabul CSP ile yapıldı, tarayıcı ayarından JavaScript
kapatılmadı. Ayrı ekran okuyucu, Safari/Firefox veya macOS testi yapılmadı.
Görseller araç çıktılarında incelendi; repoya ekran görüntüsü eklenmedi.
Önizleme dosyaları `bin/wky19-preview` altında Git dışında tutuldu.
Bu çalışma yeni bir AI plan yazma/istemci kurulum kabulü değildir; aşağıdaki
gerçek Codex CLI kabulü ve Desktop sınırı ayrı sonuçlardır. Tüm paket testleri
ve güvenlik taraması bu rehber kabulünde yeniden çalıştırılmadı.
WKY-12 deploy beklemesi korunur; commit, push ve deploy yapılmadı.

## Otomatik kontroller

- Rehber sayfaları ve OpenAPI indirmesi web oturumu ister; API rehberi ve sözleşme Bearer doğrulaması ister.
- MCP initialize talimatı önce `get_agent_guide` çağırır. Keşif yedi araç ve okuma/yazma, idempotency, silme etkisi metadata'sı içerir.
- Rehber tarihi İstanbul gece sınırında plan servisinin saatiyle test edilir.
- `TestOpenAPIContract`: temsili outline/plan yazımı, tüm boş/dolu okuma şemaları, tam başarılı yanıt örneklerinin gerçek API yanıtlarıyla eşitliği, tamamlama/geri alma ve idempotency başlığının önceliği.
- Mevcut plan testleri geçmiş/tamamlanmış kayıtları koruma, açık seans silme, telafi, idempotency ve alan hatalarını kapsar.

## Gerçek Codex CLI oturumu

Kurulu Codex ile `exec --ignore-user-config --ephemeral --approve-for-me` kullanıldı.
Yalnızca o çağrı için HTTP MCP adresi ve bir test-token ortam değişkeni tanımlandı;
kullanıcının kalıcı MCP yapılandırması değiştirilmedi. Sunucu mevcut
`TestVisualReview` geçici SQLite veritabanıydı; gerçek veritabanlarına dokunulmadı.

Araç sırası:

```text
get_agent_guide → get_outline / get_plan / get_progress
→ save_plan → get_plan / get_progress
```

Kullanıcı referansları test isteminde açıkça verildi: M1, bölüm 1, Temel kavramlar,
s. 1–10. Ajan 2 Kasım 2026 20:30 için 90 dakikalık aksam seansını yalnızca o günü
kapsayan istekle oluşturdu. Yeni kimlik 6; plan 5 seans/540 dakikadan 6 seans/630
dakikaya çıktı; silinen/atlanan kayıt yok. Tekrar okuma ile doğrulandı. Dosya veya
shell araçları kullanılmadı; test oturumu yalnızca Workey MCP araçlarını çağırdı.

İlk salt okunur CLI oturumunun onay politikası `never` olduğundan `save_plan`
onay denetiminde reddedildi. Onayları otomatik incelemeye ileten oturumla test
tekrarlandı ve başarılı oldu. Güvenlik kontrolleri kapatılmadı.

## Görsel kontrol ve sınırlar

Geçici localhost önizlemesinde 390×844 mobil ve 1280×800 masaüstü görünümleri
incelendi. Menü bağlantıları ve mobil Kullanım aktifliği doğrulandı. Mobil belge
genişliği 375 px, viewport 390 px; uzun komutlar `overflow-x:auto` kod blokları
içinde kaydırılır. Kullanım ana sayfası, MCP/API sayfaları ve örnek yanıt açılımları
kontrol edildi.

Codex Desktop için bağlantı rehberi eklendi; ayrı bir Desktop ajan oturumu
çalıştırılmadı. Bu kabul testi, uygulamanın token ortamıyla yeniden başlatılıp
rehberdeki akışla ayrıca denenmesini gerektirir. ChatGPT bağlantısı kapsam dışıdır.
SDK/HTTP testleri ve gerçek CLI kabulü Desktop kabulü yerine sayılmaz.

## Yeniden çalıştırma

```powershell
go tool templ generate
./bin/tailwindcss.exe -i web/styles/input.css -o web/static/app.css --minify
go test ./...
go vet ./...
go build ./...
$env:CGO_ENABLED = '0'
go build ./...
```

Görsel inceleme için `docs/mcp.md` içindeki geçici önizleme yönergesini kullanın.
Commit, push, yayınlama ve kalıcı istemci kurulumu bu çalışmanın parçası değildir.

## WKY-20 — rehber düzeni ve içerik hiyerarşisi (2026-10-27)

Uygulama tamamlandı; gerçek tarayıcı %200 yakınlaştırma kabulü bekliyor.
Önceki WKY-19 kanıtları bu düzenin yerine kullanılmadı.

- Başlangıç: güncel örnek-verili Bugün, Hafta, İlerleme ve iki rehber uygulama içi
  tarayıcıda karşılaştırıldı. Yeni rehberler aynı Barlow, gri-yeşil zemin, beyaz
  14px paneller, 20/28px iç boşluk, 12/32px dış boşlukları kullanıyor. Kardeş
  paneller ve içeride ayırıcılar var; iç içe beyaz panel yok.
- MCP: istemci seçimi bağlantı bölümünde, alt başlık h3. Kısa doğrulama beklenen
  sonuç olarak anlatılıyor. 3945 karakterlik plan istemi kapalı önizlemeden tam
  kopyalandı; tam-liste uyarısı dışarıda görünür. İstem metni değiştirilmedi.
- API: ilk okuma ve token girişi referanstan önce; okuma/yazma/korunacak kayıtlar/
  tam-liste uyarısı/tekrar okuma gruplandı. Sözleşmeden gelen JSON örnekleri aynı.
  OpenAPI indirmesi referansın başında; endpoint yanıtları kapalı açılır içerik.
- Bölüm bağlantıları benzersiz ve doğru; geniş ekranda sağdaki dizin, dar ekranda
  native details listesi. Görünüm bağlantıları istemciyi, istemci seçimi #connect
  konumunu koruyor. Markdown tüm istemcileri ve kapalı içerikleri kapsıyor.
- 320×844, 390×844, 1280×800: üç MCP istemcisi × iki görünüm ve iki API görünümü
  tarandı. Belge scrollWidth/clientWidth eşit: 305/375/1265px. Uzun satırlar
  yalnız pre içinde kayıyor. 390×844 ve 1280×800 görüntüleri görsel olarak
  incelendi; mobil sayfa sonunda metin alt gezinmenin üzerinde kalıyor.
- Klavye: bölüm listesi, bölüm bağlantısı, istemci seçimi, görünüm değişimi,
  istem önizlemesi ve endpoint disclosures Enter ile çalıştı. Odak çizgisi
  görünür. Görünen rehber a/button/summary hedeflerinde 44px altı yükseklik yok.
  DOM üzerinden ölçülen metin renklerinde en düşük kontrast 4,60:1; h1 28px,
  gövde 15px. Başlıklar h1 → h2 → h3; API örnek başlıkları da h3.
- Başarı: iki tam Markdown, altı ayrı istem ve 22 kod hedefi (toplam 30) gerçek
  pano okumasıyla kaynak metne eşit; CRLF normalleştirildi. Diğer istemcilerin
  kod hedefleri ve altı istemin tüm istemcilerde kaynak eşitliği Go testiyle
  kontrol edildi. Kapalı istem önizlemesinden kopyalama ayrıca doğrulandı.
- Hata: loopback proxy Clipboard writeText reddi ve API yokluğunu zorladı.
  Kod, tam Markdown ve uzun istem için doğru odak/seçim, hata metni, yeniden
  etkin düğme ve başarı bildirimi yokluğu doğrulandı. Kapalı istem açıldı ve
  3945 karakter seçildi; tam Markdown paneli açıldı; kod Range ile seçildi.
- Scriptsiz: CSP script-src 'none' altında sıfır görünür kopyalama düğmesi.
  Mobil bölüm/istemci/görünüm gezinmesi ve API yanıt açılımı çalıştı; ayrı istem
  ve tam Markdown gerçek Ctrl+A/Ctrl+C sonrası pano metniyle eşleşti.

Go 1.27.3: templ ve Tailwind üretimi; go test ./internal/web/... ./internal/apidoc/...;
go build ./...; CGO_ENABLED=0 go build ./...; go vet ./...; gofmt -l . (boş);
git diff --check geçti. Bölüm/kod hedefi/istem sözleşmesi regresyonları eklendi;
oturumsuz erişim, no-store ve OpenAPI eşitliği mevcut test matrisinde geçti.
Impeccable layout detector boş sonuç verdi; görsel kontrolün yerine sayılmadı.

Sınır: IAB Ctrl+plus girişinden sonra innerWidth ve devicePixelRatio değişmedi;
gerçek %200 zoom doğrulanmadı. 1280×800 için eşdeğer 640×400 yeniden akış alanında
iki rehber okunur ve yatay taşmasız (625px belge genişliği) incelendi. Bu eşdeğer
kontrol gerçek zoom kabulünün yerine geçirilmedi; WKY-20 done işaretlenmedi.
Ekran okuyucu, diğer tarayıcı motorları, tam test paketi ve güvenlik taraması bu
görevde çalıştırılmadı. Pano ret/yokluk durumları kontrollü enjeksiyondur; JS
kapalılığı CSP ile sağlandı. Gerçek kullanıcı .env, token ve planına dokunulmadı.
Örnek fixture sunucuları ve proxy'ler kabul bitiminde kapatıldı. Commit/push/deploy
bu görevin kapsamında değil; WKY-12 beklemesi sürüyor.
