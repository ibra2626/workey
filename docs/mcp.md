# Workey MCP bağlantısı

Menüde **MCP Kullanımı** (`/kullanim/mcp`), mobilde **Kullanım** sekmesi altında
Claude Code, Codex CLI ve Codex Desktop bağlantı adımları bulunur. Sayfalar web
oturumu ister; ortak kurallar doğrudan ajan rehberiyle aynı kaynaktan gelir.

## Token nereden alınır?

Workey tokenı bir siteden alınmaz; ilk kurulumda kendiniz rastgele bir değer
oluşturup sunucunun `.env` dosyasındaki `WORKEY_API_TOKEN` alanına kaydedersiniz.
Komutlar ve parola/oturum anahtarı dahil temel kurulum için
[README'nin yerel kurulum adımlarını](../README.md#yerel-kurulum) izleyin.
Uygulama içi rehberin **Workey'i hazırlayın** bölümü de token üretme komutunu gösterir.

Mevcut bir kurulumda yeniden token üretmeyin; `.env` dosyasındaki değeri alın veya
sunucuyu kuran kişiden isteyin. Sunucunun süreç ortamında `WORKEY_API_TOKEN`
tanımlıysa `.env` yerine o değer geçerlidir. Token değiştiğinde sunucuyu yeniden
başlatıp istemcide de aynı değeri kullanın. Aşağıdaki `Read-Host` token üretmez;
var olan tokenı gizli girişle alıp istemcinin ortamına aktarır. Web parolası ve
oturum anahtarı farklı değerlerdir.

## Rehber görünümü, istemci seçimi ve kopyalama

Sayfa Codex Desktop seçili açılır. **İstemcinizi bağlayın** bölümünde **Codex CLI** veya **Claude Code** bağlantısını
seçtiğinizde yalnızca o istemcinin kurulum adımları gösterilir. Seçim URL'deki
`client` parametresinde tutulur; görünüm değiştirirken korunur. Teknik ayrıntılar
ve ek örnekler başlangıçta kapalıdır; başlıklarına tıklayarak ya da klavyeyle
odaklayıp Enter/Boşluk ile açabilirsiniz.

**Rehber** biçimlendirilmiş adımları, **Markdown** salt okunur tam metni gösterir.
**Markdown olarak kopyala**, seçili istemciden bağımsız olarak tüm istemcileri,
kapalı bölümleri ve istemleri içeren rehberi kopyalar. **İstemi kopyala**
düğmeleri yalnızca yanlarındaki istemi, **Kodu kopyala** yalnızca ilgili kodu kopyalar. İşlem
tamamlandığında ilgili düğmenin yanında **Kopyalandı** bildirimi görünür.

Pano erişimi yoksa veya reddedilirse metin görünür hale gelir, odaklanır ve
seçilir; elle Ctrl+C (macOS: ⌘C) kullanın. Başarı bildirimi verilmez.
JavaScript kapalıyken kopyalama düğmeleri gizlidir; görünüm/istemci bağlantıları
ve açılır bölümler çalışır. Markdown görünümündeki ya da ayrı istem alanındaki
metni odaklayıp Ctrl+A ve Ctrl+C ile elle kopyalayabilirsiniz. Uzun kodlar kendi
kutusunda yatay, uzun istemler metin alanında dikey kaydırılır.

Ana bölümler uygulamanın beyaz panellerinde gösterilir. **Bu sayfada** dizini
geniş ekranda sağdadır; dar ekranda açılır listedir. Bölüm bağlantıları JS
olmadan da çalışır. İstemci seçimi sizi bağlantı bölümünde tutar.

## Sınava kadar plan şablonu

Dördüncü adımda **Sınava kadar çalışma planımı oluştur** alanındaki
**İstemi kopyala** düğmesini kullanın. Tam metin **İstemin tamamını göster**
önizlemesindedir; düğme kapalı önizlemeden de eksiksiz kopyalar. Pano hatasında
önizleme açılır ve metin elle kopyalamak üzere seçilir. Tam-liste değiştirme
uyarısı önizlemenin dışında görünür kalır. Kopyayı asistanınıza göndermeden önce
köşeli parantezli hedef/seviye, uygun gün/saat, süre, dinlenme ve doğrulanabilir
kitap/bölüm/sayfa alanlarını doldurun. Önerilen çalışma ritmini kendinize göre
düzenleyin. Sayfa yalnızca metin kopyalar; plan yazmak için Workey'e bağlı
asistanınızla devam edin.

Güncel tarihler ve işlem kuralları için asistan önce `get_agent_guide` çağırır;
sayfada gösterilen sınav tarihi de aynı kaynaktan gelir. Şablonun asıl metni
uygulamadaki ortak içerik modelidir (`internal/web/views/usage_content.go`);
bu belgeye ayrı bir tarih veya şablon kopyası eklenmez. Eksik referanslar
netleşmeden yazma başlamaz. Mevcut plan okunur, korunacak açık kayıtlar her
haftanın tam listesine dahil edilir ve her yazma tekrar okumayla doğrulanır.
Kısmi sonuçlar ve hatalar asistan tarafından açıklanmalıdır.

## Codex Desktop

1. **Ayarlar → MCP sunucuları → Sunucu ekle** yolunu açın. Workey zaten
   ekliyse mevcut kaydı düzenleyin.
2. **Ad:** `Workey`
3. **Tür:** `Akış destekli HTTP` (Streamable HTTP)
4. **URL:** `http://127.0.0.1:8080/mcp` (portu değiştirdiyseniz kendi portunuzu yazın).
5. **Taşıyıcı token için ortam değişkeni:** boş bırakın; doluysa içeriğini silin.
6. **Başlıklar** bölümüne şunları girin:
   - **Anahtar:** `Authorization`
   - **Değer:** `Bearer SENİN_TOKENIN`

`SENİN_TOKENIN` yerine Workey projesindeki `.env` dosyasının `WORKEY_API_TOKEN=`
satırındaki değeri yazın. **`Bearer` kelimesinden sonra bir boşluk bırakın.**
Sunucu süreç ortamında token tanımlıysa o etkin değeri kullanın. Tokenı sohbete
göndermeyin.

**Kaydet** düğmesine basın, Codex Desktop'ı tamamen kapatıp yeniden açın ve yeni
sohbet başlatın. Bu sırada **Workey sunucusu çalışır durumda kalmalı.** Yeni
sohbette “Workey MCP bağlantısını kullanarak ajan rehberini oku ve bağlantıyı
doğrula” istemini gönderin.

## Codex CLI

İstemciyi başlatacağınız PowerShell 7 terminalinde sunucuyla aynı tokenı girin:

```powershell
$env:WORKEY_API_TOKEN = Read-Host 'Workey API tokenı' -MaskInput
codex mcp add workey --url http://127.0.0.1:8080/mcp --bearer-token-env-var WORKEY_API_TOKEN
codex mcp list
codex
```

Alternatif kullanıcı `.codex/config.toml` kaydı:

```toml
[mcp_servers.workey]
url = "http://127.0.0.1:8080/mcp"
bearer_token_env_var = "WORKEY_API_TOKEN"
```

Workey'in `.env` okuması Codex CLI'ye ortam değişkeni aktarmaz. Codex CLI'yi tokenı
tanımladığınız aynı terminalden başlatın. `workey` kaydı zaten varsa tekrar
eklemeyin. `/mcp` ile bağlantıyı kontrol edin ve `get_agent_guide` çağrılmasını
isteyin. Desktop için yukarıdaki ayarlar ekranını kullanın.

[OpenAI resmî MCP dokümantasyonu](https://learn.chatgpt.com/docs/extend/mcp).
ChatGPT bağlantısı sonraki aşamadır; bu sürümde çalışır bağlantı olarak sunulmaz.
HTTPS/tünel ve uygun kimlik doğrulama ayrıca hazırlanmalıdır.

Workey, `/mcp` adresinde Bearer doğrulamalı Streamable HTTP sunar. REST ve web ile aynı SQLite verisini ve plan servisini kullanır. Uygulama çalışma içeriği üretmez; asistan yalnızca kitap, bölüm ve sayfa referanslarını kaydeder.

## Yerel bağlantı

Sunucu çalışma dizinindeki `.env` dosyasını otomatik okur; terminalde önceden tanımlı ortam değişkenleri (boş değerler dahil) önceliklidir. `.env.example` dosyasını `.env` olarak kopyalayıp yer tutucuları doldurun, ardından `go run ./cmd/workey serve` çalıştırın. Yerel HTTP için `WORKEY_COOKIE_SECURE=false`; `.env` Git tarafından yok sayılır. Gerçek tokenları paylaşılan dosyalara veya ekran görüntülerine koymayın. `.env` düz `KEY=value`, tek/çift tırnak ve yorumları destekler; değerlerde `$` değişken genişletmesi veya komut çalıştırma yapılmaz.

Claude Code için proje dizininde, token ortamda tanımlıyken PowerShell:

```powershell
claude mcp add --scope local --transport http workey http://127.0.0.1:8080/mcp --header "Authorization: Bearer $env:WORKEY_API_TOKEN"
claude mcp get workey
```

Bu komut tokenı yerel Claude yapılandırmasına yazar; yapılandırmayı Git'e eklemeyin. Sunucu portunu `WORKEY_ADDR` ile değiştirdiyseniz adresi de uyarlayın. Web oturum çerezi MCP yetkilendirmesi yerine geçmez.

## Araçlar

| Araç | İşlev |
| --- | --- |
| `get_agent_guide` | İlk çağrı: İstanbul bugünü, sınav tarihi, kitap/tür listeleri, işlem kuralları ve önerilen akış. |
| `get_outline` | Fazları ve konu gruplarının anahtarlarını oku. |
| `save_outline` | Faz ve grupların tam listesini anahtara göre değiştir. Kullanılan grup silinemez. |
| `get_plan` | `from`–`to` dahil aralıktaki seansları, kimliklerini ve durumlarını oku. |
| `save_plan` | Aralıktaki başlanmamış açık seansların tam listesini yaz. Eksikleri silinir. Study kaydı olan seans aynıysa unchanged, değişiklikte skipped reason=started; eksikse korunur. |
| `complete_session` | `session_id` için tamamla; `done=false` ile geri al. Gelecek seans tamamlanamaz. |
| `get_progress` | Seans, dakika, saat, konu grubu ve 38 haftalık ilerlemeyi oku. |

Tarihler `YYYY-MM-DD`, saatler `HH:MM`, takvim Europe/Istanbul. Kitaplar `M1`, `M2`, `M3`, `G1`, `G2`; türler `aksam`, `uzun_blok`, `deneme`. `tools/list` açıklamaları ve şemaları alan sınırlarını bildirir.

Geçmiş tarihler ve tamamlanan seanslar plan yazımında korunur; sonuçta `skipped` olarak bildirilir. Aynı istek tekrarlandığında kayıt/denetim satırı çoğalmaz. `idempotency_key` verilirse ilk sonuç tekrar döner; aynı anahtar farklı içerikle kullanılamaz. Telafi için `makeup_of` geçmiş, tamamlanmamış seansın tarih ve saatini taşır.

## Örnek istemler

- “Önce araç açıklamalarını ve mevcut faz/planı oku. Verdiğim kitap ve sayfa referanslarıyla bugünden başlayan haftayı yaz. Mevcut açık seansları değiştirmeden önce tam hafta listesini koru; geçmiş veya tamamlanmış seansları değiştirme.”
- “Bu haftanın planını oku. Yapılamayan salı seansını cumartesi için telafi seansına bağla; mevcut cumartesi içeriğini koru. Ardından planı ve ilerlemeyi tekrar okuyup sonucu bildir.”
- “İlerlemeyi oku; tamamlanan dakika ve seansları özetle. Çalışma içeriği veya soru üretme.”

`testdata/sample-*.json` yalnızca örnek veridir. Sabit 26 Ekim 2026 haftasını bugünün gerçek planı gibi göndermeyin: geçmiş kayıtlar korunacağı için `skipped` dönebilir. Gerçek kitap referanslarını kullanıcı sağlar.

## Tekrarlanabilir doğrulama (2026-10-27)

```powershell
go test ./internal/mcpserver/... ./internal/plan/...
go test ./internal/mcpserver -run '^TestMCPWebWorkflow$' -v
```

`TestMCPWebWorkflow`, resmî Go MCP istemcisiyle gerçek Streamable HTTP üzerinden `save_outline → save_plan → get_progress → web POST tamamla → get_progress → save_plan tekrar → web POST geri-al → get_progress` zincirini çalıştırır. Sahte saat 2026-10-26 09:00 İstanbul; geçici veritabanı kullanılır. Beş seans/540 dakika kaydolur; web tamamlama sonrası bir seans/90 dakika/1,5 saat görünür; geri al sonrası sıfıra döner. Tekrarda beş kayıt korunur, tamamlanan seans atlanır, dört seans değişmeden kalır ve yeni audit satırı yazılmaz.

Denetim sorgusu:

```sql
SELECT at_utc, source, action, COUNT(*)
FROM audit_log
GROUP BY at_utc, source, action ORDER BY source, action;
```

Tamamlama sonrası çıktı (sabit test saati):

```text
2026-10-26T06:00:00Z | mcp | create   | 11
2026-10-26T06:00:00Z | web | complete | 1
```

Görsel incelemede MCP ile yazılan örnek plan Bugün/Hafta'da, web tamamlama sonucu İlerleme'de görüldü. Geçici ekran görüntüleri ve ham ölçüm çıktıları commit öncesi temizlendi; bulgular [ekran inceleme raporunda](review/README.md) saklanır. Claude Code içinde ayrı bir model oturumu çalıştırılmadı; istemci uyumu Go SDK ile doğrulandı. Daha sonra gerçek Codex CLI modeli yeni ajan rehberi ve araç açıklamalarıyla plan yazıp tekrar okuyarak doğruladı. Ayrı Desktop kabulü henüz yapılmadı. Ayrıntılar [kullanım doğrulama raporunda](usage-verification.md).

Geçici görsel senaryoları açmak için:

```powershell
$env:WORKEY_VISUAL_REVIEW='1'
go test ./internal/mcpserver -run '^TestVisualReview$' -v -timeout 35m
```

18114: çarşamba, 18115: dinlenme, 18116: boş plan, 18117: giriş, 18118: uzun blok. Yalnızca loopback; örnek veriler test sonrasında silinir. Ctrl-C ile durdurun. Bu sunucular gerçek çalışma planı için kullanılmaz.

<!-- study-contract:start -->
## Çalışma durumu ve yer imini okuma

SessionView\.study opsiyonel, salt okunur çalışma bağlamıdır: status=active/paused/closed, first\_started\_at/last\_started\_at/updated\_at (UTC), entry\_mode=normal/ten\_minutes ve version\. bookmark varsa item\_ordinal (sıfır tabanlı), stage (akşam: konu/soru/yanlis\_defteri; diğer: calisma), resume\_page, çözülmüş item ve notation içerir\. resume\_page devam edilecek kitap sayfasıdır; yarım sayfa aynı kalır, otomatik \+1 yapılmaz\. Yanlış defteri aşamasında da kitap sayfasıdır\. Yer imi yoksa uydurma; study yokluğu hiç çalışılmadığını kanıtlamaz\. SessionState planned/today/done/missed/moved değişmez; geçmiş kısmi emeği ayrıca study üzerinden oku\. GET /api/v1/today active\_study ve latest\_paused\_study ile tarih dışındaki devam bağlamını da sunar; aktif önceliklidir, yarım işe dönüş açık seçimdir, bugünün planı korunur\. get\_plan yalnız istenen aralığı döndürür; gerekirse geçmiş aralığı da oku\. Yeni çalışma yazma REST endpoint'i veya MCP aracı yoktur\. Kısmi sayfa ve ten\_minutes hedefi ilerlemeyi artırmaz; done\_min tamamlanan seansların planlı süresidir, ölçülmüş süre veya öğrenme yüzdesi değildir\. Telafi bağlantısı yer imini veya tamamlamayı otomatik taşımaz; eski ve telafi ayrı tamamlanırsa mevcut toplam ikisini de sayar\.

26 Ekim 2026 tarihli temsili okuma → yazma → tekrar okuma örneği; gerçek kullanıcı planı değildir\. Önce ajan rehberini, outline, ilgili get\_plan / GET /api/v1/plan aralığını ve ilerlemeyi okuyun\. Aşağıda yarım kalmış örnek seansın okuma yanıtı var\.

```json
{
  "study": {
    "status": "paused",
    "first_started_at": "2026-10-26T17:30:00Z",
    "last_started_at": "2026-10-26T17:30:00Z",
    "updated_at": "2026-10-26T17:40:00Z",
    "entry_mode": "ten_minutes",
    "version": 2,
    "bookmark": {
      "item_ordinal": 0,
      "stage": "soru",
      "resume_page": 5,
      "item": {
        "book": "M1",
        "chapter_no": 1,
        "chapter_title": "Temel kavramlar",
        "page_from": 1,
        "page_to": 10
      },
      "notation": "M1·1, s. 1–10"
    }
  },
  "id": 1,
  "date": "2026-10-26",
  "start_time": "20:30",
  "duration_min": 90,
  "kind": "aksam",
  "completed": false,
  "state": "today",
  "topic_group_key": "temel",
  "topic_group_name": "Temel kavramlar",
  "items": [
    {
      "book": "M1",
      "chapter_no": 1,
      "chapter_title": "Temel kavramlar",
      "page_from": 1,
      "page_to": 10
    }
  ],
  "notation": "M1·1, s. 1–10"
}
```

Plan yazarken study ve diğer okuma alanlarını göndermeyin\. Başlanmış seansın items, date, start\_time ve duration\_min değerlerini koruyun: aynı içerik unchanged, değiştirme denemesi skipped reason=started, istekten eksikse korunur\. past\_date ve completed nedenleri önceliklidir\. Örnekte duration\_min=100 yapmayı deneyen eski bir istemci aşağıdaki sonucu alır; seans 90 dakika ve yer imi aynı kalır\.

```json
{
  "created": null,
  "updated": null,
  "deleted": null,
  "unchanged": null,
  "skipped": [
    {
      "date": "2026-10-26",
      "start_time": "20:30",
      "reason": "started"
    }
  ]
}
```

Bu sonucu başarıyla uygulanmış değişiklik saymayın\. Yazdıktan sonra aynı aralığı ve ilerlemeyi tekrar okuyun; seansın süre, kaynak ve yer iminin korunduğunu doğrulayıp atlanan değişikliği kullanıcıya açıklayın\. Telafi gerekiyorsa kaynak geçmiş seansı değiştirmeden makeup\_of ile ayrı seans hazırlayın; yer imi otomatik taşınmaz\.

Mevcut complete\_session ve REST tamamlama/geri alma çağrıları ek alan istemez: tamamla mevcut study durumunu atomik closed yapar; geri al yer imini koruyup paused yapar\. Eski seans için çalışma kaydı oluşturmaz\. Bu REST/MCP yazmaları study\.version denetlemez; eski sekme niyeti tespit edilemez\.

<!-- study-contract:end -->

<!-- planning-inputs:start -->
## Plan için kaynaklarını hazırla

Materyalleri Workey'e bağlı asistanına ver\. Workey dosya yüklemez, PDF saklamaz veya OCR yapmaz; asistanın görmediği özel bir kaynağı okuduğunu varsayma\.

- Her kitabın adı, kitap kodu ve baskısı; okunaklı içindekiler ve bölüm sırası\.
- Gerekli örnek sayfalar veya erişebildiğin PDF; basılı sayfa numarası ile PDF görüntüleyicisinin sayfa sırası arasındaki fark\.
- Doğrulanabilen bölüm ve sayfa aralıkları; varsa konu, soru ve antrenman aralıkları\. Eksik referansı asistana sor; tahminle doldurtma\.
- Mevcut seviye, bitirilen bölümler, hedef, uygun gün ve saatler, dinlenme ve diğer zaman kısıtları\.

İlk haftadan sonra gerçek temponu ve yükün uygunluğunu asistana anlat\. Açık sekme süresi veya başlangıç hedefi çalışma süresi sayılmaz\.

## İsteğe bağlı: ilk haftayla başla

Sınava kadar plan oluşturma istemi kullanılabilir\. Alternatif olarak ilk haftayı ayrıntılı, uzak dönemi yalnız kaba kapsam olarak konuşabilirsin\. Uzak dönem için doğrulanmamış sayfa veya seans yazılmaz\. Kopyalama planı değiştirmez; asistan önce okur, yazdıktan sonra yeniden okuyarak doğrular\.

```
Workey bağlantısıyla önce get_agent_guide (REST: GET /api/v1/agent-guide), get_outline, ilgili get_plan ve get_progress eşdeğerlerini oku. Güncel today, timezone ve exam_date değerlerini esas al.
Hedefim/seviyem: [doldur]. Uygun gün ve saatlerim: [doldur]. Doğrulanabilir kitap/baskı/içindekiler/sayfa kaynaklarım: [asistana ver].
İlk haftayı ayrıntılı, uzak dönemi kaba kapsam olarak birlikte hazırlayalım. Eksik bölüm veya sayfa referansı uydurma; eksikleri sor ve yanıtımı bekle. Uzak dönem için doğrulanmamış seans yazma. İlk uygulama haftasından sonra gerçek tempomu ve yükü bana sor; açık kalma zamanından süre hesaplama.
Yazılacak aralığı ve korunacak kayıtları benimle netleştir. Yazmadan hemen önce outline, ilgili plan ve study/yer imini yeniden oku. save_outline faz ve grupların, save_plan aralıktaki başlanmamış açık seansların TAM listesini değiştirir; eksik açık seanslar silinir. Korunacak kayıtları tüm kaynaklarıyla dahil et. Geçmiş, tamamlanmış ve başlanmış seansları koru; başlanmış seansın kaynak/tarih/saat/süresini değiştirme. Telafi bağlantılarını koru.
Her yazma sonucunda created/updated/deleted/unchanged/skipped listelerini incele. skipped reason=started dahil atlanan değişikliği başarı sayma; hata veya belirsizlikte dur, körlemesine tekrar yazma. İlgili aralığı, outline ve ilerlemeyi yeniden okuyup doğrula; kalan kapsamı bildir. Planlama sırasında seans tamamlama çağrısı yapma.
```

## Ara sonrası planı yeniden değerlendir

Bu istemi kopyalayıp Workey'e bağlı asistanına gönder\. Yer imi sunucudan okunur; kimlik veya token ekleme\. Kopyalama plan yazmaz ve eski işi otomatik bugüne taşımaz\. Asistan okuma → yazma → yeniden okuma sırasını izler\.

```
Workey bağlantısıyla ara sonrası planımı değerlendirelim.
Gerçek tempom, uygun zamanım ve değişen koşullarım: [doldur]. Eksik kaynaklarım veya henüz çalışmadığım kapsam: [bildiğimi yazacağım; eksikleri bana sor].
Önce get_agent_guide (REST: GET /api/v1/agent-guide) ile today, timezone ve exam_date değerlerini oku. Sınav geçmişse yazmadan hedefi benimle netleştir. get_outline, ilgili geçmiş/bugün/gelecek get_plan aralıkları ve get_progress eşdeğerlerini oku; büyük aralıkları haftalara böl.
study durumunu, yer iminin kitap/bölüm/aşama/resume_page değerlerini ve makeup_of/moved_to ilişkilerini incele. Yer imi devam edeceğim sayfadır; otomatik +1 yapma veya telafiye kopyalama. Kısmi çalışmayı tamamlanmış ya da ölçülmüş dakika sayma. Eski işi otomatik bugüne yükleme; eksik kapsamı ve yetişmeyecek işi benimle netleştir. Kitap/bölüm/sayfa uydurma; kaynak eksikse sor ve yanıtımı bekle.
Yazmadan hemen önce ilgili aralığı yeniden oku. Her study kaydı begun/başlanmış kabul edilir: kaynak, tarih, start_time, duration_min ve items korunur. save_plan aralıktaki başlanmamış açık seansların TAM listesini değiştirir; listede bulunmayanlar silinir. Korunacak açık seansları tüm items ile dahil et; silme/değiştirme kararını benimle netleştir. Geçmiş, tamamlanmış ve başlanmış seansları koru. Outline değişirse korunacak faz ve grupları da içeren tam listeyi gönder; kullanılan grupları silme.
Telafi gerekiyorsa kaynak seansı değiştirmeden makeup_of ile ayrı seans öner ve bağlantıyı doğrula. Eski seans ve telafisini otomatik tamamlamayacaksın; ikisini ayrı tamamlamak mevcut seans tabanlı toplamda ikisini de sayar, benzersiz içerik ölçmez.
Okuma alanları id/state/completed/notation/study yazma gövdesine girmez; topic_group_key yazmada topic_group olur. Her save_plan sonucunun created/updated/deleted/unchanged/skipped listelerini incele; skipped reason=started dahil atlananları başarı sayma. Hata, belirsiz yazma veya doğrulanamayan sonuçta dur; kalan haftaları yazma. Tekrar denemeden önce oku; aynı gövde için aynı idempotency_key kullan.
Yazdıktan sonra aynı get_plan aralığını, outline ve get_progress eşdeğerlerini yeniden oku. Yer iminin, korunacak kayıtların ve telafi ilişkilerinin korunduğunu doğrula; değişen ve eksik kalan kapsamı Türkçe bildir. Bu planlama sırasında complete_session veya REST tamamlama/geri alma çağrısı yapma.
```

<!-- planning-inputs:end -->
