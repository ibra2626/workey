# Workey MCP bağlantısı

Menüde **MCP Kullanımı** (`/kullanim/mcp`), mobilde **Kullanım** sekmesi altında
Claude Code, Codex CLI ve Codex Desktop bağlantı adımları bulunur. Sayfalar web
oturumu ister; ortak kurallar doğrudan ajan rehberiyle aynı kaynaktan gelir.

## Rehber görünümü, istemci seçimi ve kopyalama

Sayfa Codex Desktop seçili açılır. **Codex CLI** veya **Claude Code** bağlantısını
seçtiğinizde yalnızca o istemcinin kurulum adımları gösterilir. Seçim URL'deki
`client` parametresinde tutulur; görünüm değiştirirken korunur. Teknik ayrıntılar
ve ek örnekler başlangıçta kapalıdır; başlıklarına tıklayarak ya da klavyeyle
odaklayıp Enter/Boşluk ile açabilirsiniz.

**Rehber** biçimlendirilmiş adımları, **Markdown** salt okunur tam metni gösterir.
**Markdown olarak kopyala**, seçili istemciden bağımsız olarak tüm istemcileri,
kapalı bölümleri ve istemleri içeren rehberi kopyalar. **İstemi kopyala** ve
**Promptu kopyala** düğmeleri yalnızca yanlarındaki istemi kopyalar. İşlem
tamamlandığında ilgili düğmenin yanında **Kopyalandı** bildirimi görünür.

Pano erişimi yoksa veya reddedilirse metin görünür hale gelir, odaklanır ve
seçilir; elle Ctrl+C (macOS: ⌘C) kullanın. Başarı bildirimi verilmez.
JavaScript kapalıyken kopyalama düğmeleri gizlidir; görünüm/istemci bağlantıları
ve açılır bölümler çalışır. Markdown görünümündeki ya da ayrı istem alanındaki
metni odaklayıp Ctrl+A ve Ctrl+C ile elle kopyalayabilirsiniz. Uzun kodlar kendi
kutusunda yatay, uzun istemler metin alanında dikey kaydırılır.

## Sınava kadar plan şablonu

Dördüncü adımda **Sınava kadar çalışma planımı oluştur** alanındaki
**Promptu kopyala** düğmesini kullanın. Kopyayı asistanınıza göndermeden önce
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

## Codex CLI ve Desktop

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

Workey'in `.env` okuması Codex'e ortam değişkeni aktarmaz. Codex Desktop'ı tamamen
kapatıp token ortamını miras alacağı süreçten yeniden başlatın; açık uygulamaya
sonradan terminalde tanımlanan değişken aktarılmaz. MCP ayarlarını kontrol edin
ve yeni sohbette `get_agent_guide` çağrılmasını isteyin. Kalıcı kullanıcı
yapılandırması bu geliştirme sırasında değiştirilmedi.

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
| `save_plan` | Aralıktaki açık seansların tam listesini yaz. İstekten çıkarılan açık gelecek seanslar silinir. |
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
