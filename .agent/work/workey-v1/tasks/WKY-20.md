# WKY-20 — API ve MCP kullanım sayfalarının tasarım ve içerik hiyerarşisi

## Status
In progress — 2026-10-27. Uygulama ve odaklı kontroller tamamlandı; gerçek %200 tarayıcı yakınlaştırma kabulü bekliyor.
Depends on: WKY-19. Ortak kurallar: `../README.md`.

## Goal
API Kullanımı ve MCP Kullanımı sayfalarını Bugün, Hafta ve İlerleme ile aynı görsel dilde sunmak; kullanıcının ilk bağlantı/istek adımını kolayca bulmasını ve gerektiğinde teknik referansa ulaşmasını sağlamak. Yüzey modu Read; marka yönü mevcut sakin, net Workey tasarımıdır.

## Evidence and limits
- Ortak şablon uygulama kabuğunu kullanıyor, fakat içerik beyaz panel yapısı yerine doğrudan zemin üzerinde akıyor.
- MCP başlığında rehber seçimi, görünüm seçimi ve istemci seçimi üst üste geliyor. Uzun sınav planı istemi ana akışta büyük bir metin alanı oluşturuyor.
- API referansı ilk çalıştırılabilir okuma örneğinden önce geliyor. Uzun belgede bölüm bağlantıları ve kod bloğu bazında kopyalama yok.
- Ön inceleme kaynak koduna dayanır. `impeccable detect --json internal/web/views/usage.templ` boş sonuç verdi; görsel kalite veya erişilebilirlik onayı değildir.
- Hazır görsel fixture port erişim hatası verdi; alternatif yerel önizlemeye uygulama içi tarayıcı erişemedi ve Windows URL doğrulaması bilgisayar kullanımını durdurdu. İki bağımsız değerlendirme kullanım limitinden tamamlanamadı. Önceki WKY-19 kanıtları bu yeni tasarımın doğrulaması sayılmaz.

## Scope and design decisions

### 1. Ortak rehber çerçevesi — P1
- Mevcut menü ve aktif sayfa davranışını koru. Başlık, açıklama ve içerik başlangıcını diğer sayfaların hizasına getir.
- MCP/API geçişini tek rehber gezinme satırında sun. Seçili rehberi yalnızca alt çizgiye dayanmayan belirgin metin/zemin ve `aria-current` ile göster.
- Rehber/Markdown görünümünü ve tam Markdown kopyalama eylemini ikincil araç alanında topla; görünüm değiştirme ile kopyalamanın farklı işler olduğu anlaşılsın.
- Ana bölümleri kardeş beyaz panellerde sun. Alt içerik için iç içe kartlar yerine başlıklar ve ayırıcılar kullan.
- Masaüstünde 28px başlık, mevcut Barlow ailesi ve 15px gövde; uzun açıklamalarda yaklaşık 65–75 karakterlik satır ölçüsü. Kod eş aralıklı fontta kalır.
- DESIGN.md tokenları: gri-yeşil zemin, beyaz panel, koyu yeşil metin, 14px panel köşesi; mobil 20px, masaüstü 24–32px panel iç boşluğu. Mevcut mobil 12px ve masaüstü 32px dış boşluklara uy.
- Seanslara ayrılan renkleri teknik etiketlerde kullanma; gölge, gradyan, emoji, yeni font veya süsleyici sayaç ekleme.

### 2. MCP içerik akışı — P1
1. Başlık ve tek cümlelik amaç.
2. Workey'i hazırlayın: sunucuyu başlatma, adres ve gerekli önkoşullar.
3. İstemcinizi bağlayın: Codex Desktop / Codex CLI / Claude Code seçimi bu bölümde; yalnız seçili istemcinin kurulumu görünür.
4. Bağlantıyı doğrulayın: kısa istem ve beklenen sonuç. Beklenen sonucu canlı bağlantı durumu gibi sunma; sayfa MCP çağrısı yapmaz.
5. Planınızı oluşturun: doldurulacak alanları anlatan kısa açıklama, görünür tam-liste uyarısı ve doğrudan erişilebilir İstemi kopyala düğmesi. Uzun istemin tamamını açılabilir önizlemede göster; kopyalama kapalı önizlemede de çalışsın.
6. Araçlar, ek istemler, alternatif yapılandırma, bağlantı sorunları ve ortak kurallar ikincil referans bölümleri olarak açılsın.
- Mevcut istem metninin koruma/doğrulama talimatlarını ve AgentGuide tarihlerini koru. İstemi kısaltmak için kritik kuralları silme.

### 3. API içerik akışı — P1
1. Başlık ve tek cümlelik amaç.
2. İlk API isteğiniz: taban yol, Bearer doğrulaması, PowerShell 7 önkoşulu ve mevcut çalıştırılabilir okuma örneği. Tokenın nasıl girileceği görünür olsun.
3. Planı okuyun ve yazın: önce okuma, korunacak kayıtları dahil etme, outline/plan gövdeleri, yazma ve tekrar okuma sırası. Tam-liste değiştirme uyarısı yazma örneğinin yanında bulunsun.
4. Endpoint referansı: yöntem, yol ve kısa amaçla taranabilir liste; ayrıntılı yanıtlar açılabilir. OpenAPI indirmesi bu bölümün başında erişilebilir olsun.
5. Temsili istek/yanıtlar, tamamlama/geri alma, curl alternatifi, hatalar ve tekrar deneme kuralları ilgili başlıklarda gruplansın.
- Temsili tarih, kimlik ve referans etiketlerini koru; sabit örnekleri güncel plan gibi sunma. API sözleşmesi veya örneklerin teknik anlamı değişmesin.

### 4. Bölüm gezinmesi ve kopyalama — P2
- Yeterli genişlikte dar bir Bu sayfada dizini; mobilde ve dar masaüstünde açılabilir bölüm listesi. Uzun etiketler taşmadan sarsın; kodun okunabilir alanını daraltma.
- Bölüm kimlikleri benzersiz ve kararlı olsun; iç içe başlıklarda doğru h2/h3 hiyerarşisini üret. Gerçek bağlantılarla JS kapalıyken de çalışsın.
- Kod örneklerine Kodu kopyala ekle; hedef metni değiştirmeden, dil etiketi ve kopyalama denetimiyle sun. Uzun satırlar yalnız kod alanında yatay kaydırılsın.
- Promptu kopyala / İstemi kopyala etiketlerini İstemi kopyala olarak birleştir. Tam rehber eylemi Markdown olarak kopyala olarak kalsın.
- Mevcut tam Markdown tüm istemcileri içerir; kapsamı arayüzde açıkla. Seçili istemci yalnız Rehber görünümünü filtrelemeye devam etsin.
- Kopyalama başarısı/hatası erişilebilir canlı durum metniyle bildirilsin. Pano kullanılamazsa metin seçimi ve elle kopyalama yolu bulunsun; JS olmadan çalışmayan düğmeler görünmesin.

## Implementation plan
- [x] Önce güncel iki rehberi ve Bugün/Hafta/İlerleme referanslarını ayrı test verisiyle aç; 390×844 ve 1280×800 başlangıç görüntülerini karşılaştır. Önceki tarayıcı kısıtını mevcut izinli yöntemlerle çöz; çözülemezse görsel kontrolü tamamlanmış sayma.
- [x] UI düzenlemesinden hemen önce Impeccable craft-floor referansını oku; ortak panel, başlık ve araç alanı iskeletini kur.
- [x] Ortak içerik modeline gerekli bölüm kimliği, başlık düzeyi ve sunum bilgisini ekle. HTML/Markdown için ikinci içerik kaynağı oluşturma; sunum metadatasını Markdown içeriğine sızdırma.
- [x] MCP ve API bloklarını yukarıdaki sıraya taşı; istemci/görünüm URL'leri ve seçimin korunmasını doğrula.
- [x] Bölüm dizini, kod kopyalama ve uzun istem önizlemesini ekle. Kopyalama hedeflerinin kapalı details içinde de erişilebilir olduğunu kontrol et.
- [x] templ/CSS üret; etkilenen davranışlar için odaklı regresyonları güncelle. Salt görsel sınıfları birebir sınayan testler ekleme.
- [ ] Görsel ve etkileşim kabulünü birlikte çalıştır; çıkan sorunları düzelt, etkilenen kontrolleri tekrarla.
- [x] Nihai davranışa göre kullanım belgelerini, görev durumunu ve CURRENT devir bilgisini güncelle; çalıştırılmayan kontrolleri ve sınırlamaları açıkça kaydet.

## Likely files
- `internal/web/views/usage.templ`, üretilen `usage_templ.go`
- `internal/web/views/usage_document.go`, `usage_content.go`
- `web/styles/input.css`, üretilen `web/static/app.css`, gerekirse `web/static/copy.js`
- `internal/web/views/usage_document_test.go`, `internal/web/usage_test.go`
- `docs/api.md`, `docs/mcp.md`, `docs/usage-verification.md`
- Genel kabuğa ancak rehberler için zorunluysa dokun; diğer sayfalara görsel yan etki oluşturma.

## Acceptance criteria
- [x] Her iki rehber diğer uygulama sayfalarıyla aynı renk, tipografi, panel ve boşluk dilini kullanıyor; iç içe panel yok.
- [x] İlk ekranda amaç ve başlangıç bölümü anlaşılıyor; MCP istemci seçimi bağlantı adımında, API ilk okuma örneği endpoint referansından önce.
- [x] Bölüm dizini doğru başlıklara gidiyor; görünüm değişimi ve istemci seçimi normal bağlantılarla, JS kapalıyken de çalışıyor.
- [x] Uzun istemin tamamı açılabiliyor ve kapalı önizlemeden eksiksiz kopyalanabiliyor; kritik kayıt koruma uyarıları kapalı içerikte kaybolmuyor.
- [x] HTML/Markdown ortak kaynaktan geliyor; tam Markdown bütün istemcileri içeriyor ve arayüz bunu açıkça bildiriyor. Üç istemci × iki görünümde mevcut içerik ve dinamik tarihler korunuyor.
- [x] Kod, ayrı istem ve tam Markdown kopyalama hedefleri doğru; başarı, ret/destek yokluğu ve JS kapalı elle kopyalama senaryoları yanlış başarı bildirmiyor.
- [ ] 390×844 ve 1280×800 görsel kontrolü ile 320px dar ekran ve %200 yakınlaştırmada okunabilirlik kontrolü kaydedildi. Sayfa seviyesinde yatay taşma ve sabit alt menü altında kaybolan içerik yok.
- [x] Klavye ile tüm bağlantılar, details ve düğmeler erişilebilir; görünür odak, anlamlı etiketler, en az 44px dokunma alanı ve doğru başlık hiyerarşisi var. Normal metin kontrastı en az 4,5:1.
- [x] Oturumsuz erişim, no-store ve OpenAPI indirmesi korunuyor; plan yazma/bağlantı denemesi gibi yeni sunucu eylemi eklenmiyor.
- [x] İlgili testler, derleme ve üretim kontrolleri geçiyor; görsel doğrulama engellenmişse görev done olarak işaretlenmiyor.

## Checks
- `go tool templ generate`
- `./bin/tailwindcss.exe -i web/styles/input.css -o web/static/app.css --minify`
- `go test ./internal/web/... ./internal/apidoc/...`
- `go build ./...`; ayrı ortamda `CGO_ENABLED=0` ile `go build ./...`
- `go vet ./...`
- `gofmt -l .` ve `git diff --check`
- Yukarıdaki görünüm/istemci/boyut/klavye/pano/JS matrisi; gerçek kullanıcı verisinden ayrı geçici DB ve loopback önizleme. Başlatılan süreçleri sonlandır; geçici dosyaları temizle.

## Out of scope
Backend/API/MCP sözleşmesi, yeni istemci desteği, gerçek plan yazımı, otomatik plan üretimi, auth değişikliği, yeni tasarım dünyası, deploy, commit ve push. WKY-12 beklemesi korunur. DESIGN.md yan dosyasının genel yenilenmesi bu göreve eklenmez.

## Progress
2026-10-27: Görev ve kabul planı oluşturuldu. Uygulama kaynakları değiştirilmedi; test veya görsel kabul henüz çalıştırılmadı.

2026-10-27: Rehber çerçevesi, kardeş paneller, doğru başlık düzeyleri, bölüm dizini, istemci seçimi, API akışı, kod kopyalama ve kapalı istem önizlemesi uygulandı. Teknik testler ve 320/390/1280px tarayıcı matrisi geçti. Gerçek pano ile 30 hedef eşleşti; ret/yokluk/scriptsiz senaryoları doğrulandı. Kanıt ve sınırlar: docs/usage-verification.md WKY-20 bölümü. IAB gerçek zoom komutunu uygulamadığından %200 kabulü açık; 640×400 eşdeğer yeniden akış kontrolü geçti. Bu nedenle done değil. Commit/push/deploy yapılmadı.
