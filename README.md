# Workey

DGS sayısal hazırlığı için kişisel çalışma planı takip uygulaması. Bugün çalışılacak
kitap, bölüm ve sayfaları gösterir; seansları tamamlamayı ve ilerlemeyi izlemeyi sağlar.
Planı dışarıdaki yapay zekâ asistanınız MCP veya REST API üzerinden yazar. Workey
çalışma içeriği üretmez. Arayüz Türkçedir; takvim `Europe/Istanbul` kullanır.

Go + templ + htmx, Tailwind CSS ve SQLite ile geliştirilmiştir. Statik dosyalar
uygulamaya gömülür. SQLite için ayrı veritabanı sunucusu veya C derleyicisi gerekmez.
Mevcut kullanım hedefi Windows üzerinde yerel çalıştırmadır.

## Gereksinimler

- Git ve `go.mod` içinde hedeflenen Go **1.27.3** araç zinciri.
- Aşağıdaki komutlar için **PowerShell 7** (`pwsh`).
- MCP kullanacaksanız seçtiğiniz istemci: Codex CLI / Codex Desktop / Claude Code.

Normal kurulum ve derleme için Node.js, templ CLI veya Tailwind CLI kurmanız gerekmez;
üretilmiş Go şablonları ve CSS depoda bulunur. Arayüz geliştirme araçları aşağıdadır.

## Yerel kurulum

### 1. Projeyi alın

```powershell
git clone https://github.com/ibra2626/workey.git
Set-Location workey
go mod download
if (-not (Test-Path .env)) {
    Copy-Item .env.example .env
}
```

Sonraki adımları proje kökünde çalıştırın. Mevcut `.env` dosyanız varsa ayarlarını
koruyun; yeni kurulum komutlarıyla üzerine yazmayın.

### 2. API tokenını ve oturum anahtarını oluşturun

**Tokenı bir siteden veya giriş ekranından almazsınız; kendiniz oluşturursunuz.**
Workey, `.env` dosyasındaki `WORKEY_API_TOKEN` değerini MCP ve REST isteklerinin
Bearer tokenı olarak kabul eder. Bir token üretme/dağıtma endpoint'i yoktur.

İlk kurulumda aşağıdaki komutu **iki kez** çalıştırın. Her çalıştırma kriptografik
rastgelelikten 32 bayt üretip Base64 metni olarak ekranda gösterir:

```powershell
$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
try {
    $bytes = [byte[]]::new(32)
    $rng.GetBytes($bytes)
    [Convert]::ToBase64String($bytes)
} finally {
    $rng.Dispose()
}
```

`.env` dosyasını metin düzenleyicinizde açın. İlk çıktıyı `WORKEY_API_TOKEN`, ikinci
çıktıyı `WORKEY_SESSION_KEY` satırında `change-me` yerine yapıştırın. Bu komut
`.env` dosyasına kendiliğinden yazmaz. İki alan için ayrı değer kullanın.

Mevcut bir kuruluma bağlanıyorsanız yeni token üretmeyin: sunucunun `.env`
dosyasındaki `WORKEY_API_TOKEN` değerini alın veya kuran kişiden güvenli yoldan
isteyin. Sunucu süreç ortamında bu değişken tanımlıysa etkin değer `.env` yerine
ortamdan gelir. İstemci ve sunucu aynı tokenı kullanmalıdır.

### 3. Web giriş parolasını ayarlayın

```powershell
go run ./cmd/workey hash-password
```

Girişte kullanacağınız parolayı sorulduğunda yazın. Çıktıdaki PBKDF2 hash'ini
`.env` dosyasının `WORKEY_PASSWORD_HASH` alanına kaydedin. Bu alana düz parolayı
veya API tokenını yazmayın. Web'e giriş yaparken ise seçtiğiniz parolayı kullanın.

Üç farklı değer vardır:

| Ayar | İşlev |
| --- | --- |
| `WORKEY_API_TOKEN` | MCP ve REST istemcisinin kullandığı ortak gizli token. |
| `WORKEY_PASSWORD_HASH` | Web giriş parolasının hash'i. |
| `WORKEY_SESSION_KEY` | Web oturum çerezlerini imzalayan ayrı gizli anahtar. |

`.env` içindeki `change-me` yer tutucularını değiştirin. `.env`, gerçek tokenlar ve
SQLite verisi Git'e eklenmez; dosyalar `.gitignore` kapsamındadır.

### 4. Sunucuyu başlatın

```powershell
go run ./cmd/workey serve
```

[Workey'i açın](http://127.0.0.1:8080) ve seçtiğiniz web parolasıyla giriş yapın.
İlk açılışta SQLite dosyası ve tablolar otomatik oluşturulur. Başlangıç planı
boştur; örnek veya gerçek plan kendiliğinden yüklenmez. Durdurmak için Ctrl+C.

Ayrı bir PowerShell terminalinde sağlık kontrolü:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/healthz
```

Beklenen yanıt `ok`. Bu kontrol kimlik doğrulaması istemez ve planı değiştirmez.

## Temel kullanım

- **Bugün:** günün kitap/bölüm/sayfa referansları ve seans tamamlama.
- **Hafta:** haftalık program ve tamamlandı/geri al işlemleri.
- **İlerleme:** kayıtlı seans ve dakika üzerinden ilerleme.
- **Kullanım:** mobilde rehberler; masaüstünde MCP Kullanımı ve API Kullanımı.

MCP için `/kullanim/mcp` sayfasında istemcinizi seçin. Codex Desktop'ta ayarlar
ekranından HTTP sunucusu ekleyip `Authorization` başlığına `Bearer TOKEN`
yazın; [Desktop bağlantı adımları](docs/mcp.md#codex-desktop) alanları tek tek açıklar.
Codex CLI veya Claude Code için istemciyi başlatacağınız PowerShell terminalinde
aşağıdaki komut mevcut tokenı gizli girişle alır:

```powershell
$env:WORKEY_API_TOKEN = Read-Host 'Workey API tokenı' -MaskInput
```

Bu komut **token üretmez**; ikinci adımda oluşturup sunucuya kaydettiğiniz değeri
girin. Workey'in `.env` okuması ayrı başlatılan istemciye ortam değişkeni aktarmaz.
Sonraki bağlantı adımları [MCP rehberinde](docs/mcp.md), ilk API isteği ve sözleşme
[API rehberinde](docs/api.md) bulunur. Web oturumu MCP/API tokenı yerine geçmez.

Plan yazarken önce mevcut kayıtları okuyun. `save_plan` / `PUT /api/v1/plan`,
seçilen aralıktaki açık seansların **tam listesini** değiştirir; gönderilmeyen açık
seanslar silinir. Geçmiş ve tamamlanmış kayıtlar korunur. Gerçek kitap referanslarını
sağlayın; yazdıktan sonra aralığın tamamını ve ilerlemeyi tekrar okuyun.
`testdata/sample-*.json` yalnızca test/örnek veridir.

## Yapılandırma ve veri

Workey, `.env` dosyasını **çalıştırıldığı dizinden** okur. Süreç ortamındaki değerler
aynı adlı `.env` ayarından önce gelir. Ayar değişikliğinden sonra sunucuyu yeniden
başlatın; token değiştiyse istemcinin değerini de güncelleyin.

| Değişken | Varsayılan / kullanım |
| --- | --- |
| `WORKEY_ADDR` | `127.0.0.1:8080`; HTTP dinleme adresi. |
| `WORKEY_DB_PATH` | `workey.db`; göreli yol çalışma dizinine göredir. |
| `WORKEY_API_TOKEN` | Normal çalıştırmada gerekli; yukarıda oluşturulur. |
| `WORKEY_PASSWORD_HASH` | Normal çalıştırmada gerekli; `hash-password` çıktısı. |
| `WORKEY_SESSION_KEY` | Normal çalıştırmada gerekli; ayrı rastgele anahtar. |
| `WORKEY_EXAM_DATE` | `2027-07-18`; `YYYY-MM-DD`. |
| `WORKEY_COOKIE_SECURE` | Kod varsayılanı `true`; örnek `.env` yerel HTTP için `false` kullanır. HTTPS'de `true`. |
| `WORKEY_DEV` | İsteğe bağlı geliştirme modu. `1` ve boş parola hash'i birlikte web girişini atlar. |

Varsayılan loopback adresine yalnız aynı bilgisayardan erişilir; telefondan erişim
ve Linux/HTTPS dağıtımı ayrıca yapılandırılmalıdır. Mevcut Linux dağıtım işi beklemededir.
Veriyi yedeklemek için sunucuyu Ctrl+C ile kapatıp kullanılan SQLite dosyasını
kopyalayın; çalışan WAL veritabanının yalnız `.db` dosyasını kopyalamayın.

## Derleme ve geliştirme

Yerel Windows binary'si üretmek için:

```powershell
New-Item -ItemType Directory -Force bin | Out-Null
go build -o ./bin/workey.exe ./cmd/workey
./bin/workey.exe serve
```

Binary'yi de `.env` bulunan proje kökünden çalıştırın. Çalışma anında Go veya
ayrı statik dosya sunucusu gerekmez; `.env` ve SQLite verisi binary'ye gömülmez.

`.templ` değiştiğinde Go modülündeki araçla yeniden üretin:

```powershell
go tool templ generate
```

CSS sınıfları veya tokenları değiştiğinde Tailwind **4.3.3** standalone CLI'yi
[resmî sürümden](https://github.com/tailwindlabs/tailwindcss/releases/tag/v4.3.3)
platformunuza uygun indirin; Windows dosyasını `bin/tailwindcss.exe` olarak yerleştirin.
Bu araç depoya dahil değildir. Ardından:

```powershell
./bin/tailwindcss.exe -i web/styles/input.css -o web/static/app.css --minify
go build ./...
```

Üretilen `*_templ.go` dosyaları ve `web/static/app.css` depoda tutulur. Bash
kullanıyorsanız `scripts/build.sh` aynı üretim sırasını çalıştırır.

Temel kontroller:

```powershell
go test ./...
go vet ./...
```

Ürün kapsamı [PRODUCT.md](PRODUCT.md), görsel kurallar [DESIGN.md](DESIGN.md),
kabul kanıtları ve açık sınırlamalar [doğrulama raporunda](docs/usage-verification.md).
