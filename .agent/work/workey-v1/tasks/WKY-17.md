# WKY-17 — MCP rehberini istemci seçimine göre sadeleştir

## Goal
Kullanıcı yalnızca kendi istemcisinin gerekli kurulum adımlarını görerek Workey MCP bağlantısını kurabilsin.

## Scope
- Önkoşul: WKY-16. Ortak kurallar: `../README.md`.
- Açılış açıklaması: “Asistanınız Workey’e bağlanır, çalışma planınızı kaydeder ve ilerlemenizi okur.”
- Ana akış: Workey'i hazırlayın → İstemcinizi bağlayın → Bağlantıyı deneyin → Planınızı oluşturun. Son adım WKY-18'de doldurulacak.
- Codex Desktop, Codex CLI ve Claude Code seçimi sun; varsayılan Codex Desktop. Seçimi `client=codex-desktop|codex-cli|claude-code` sorgu parametresiyle sunucuda işle; bilinmeyen değer varsayılana dönsün. Görünüm geçişlerinde seçimi koru, kalıcı kaydetme.
- Ortak sunucu hazırlığını bir kez göster; rehber görünümünde yalnızca seçilen istemcinin kurulumunu göster. Tam Markdown tüm istemcileri ayrı başlıklarla içersin.
- Tokenın sunucudaki WORKEY_API_TOKEN ile aynı olması gerektiğini, nereden alınacağını ve ilgili istemci sürecine nasıl aktarılacağını gerekli adımın yanında açıkla. Gerçek tokenı okuma veya rehbere yerleştirme.
- Mevcut yerel bağlantı adresi, özel port uyarlaması, token ortam aktarımı ve Desktop yeniden başlatma bilgilerini kısa, eyleme dönük adımlara taşı. Bağlantı komutlarını değiştirmen gerekirse resmi istemci belgeleriyle doğrula.
- Kopyalanabilir bağlantı deneme istemi: “Workey MCP bağlantısını kullanarak ajan rehberini oku ve bağlantıyı doğrula”. WKY-16'nın kopyalama mekanizmasını kullan.
- Araç listesi, alternatif config.toml, ayrıntılı işlem sırası/kuralları ve sorun giderme kapalı açılır bölümlerde yer alsın. Mevcut destek kapsamını koru; ChatGPT veya Claude Desktop'ı doğrulanmış yeni istemci seçeneği olarak ekleme.
- API rehberinin içeriğini değiştirme.

## Acceptance criteria
- [x] İlk açılışta ortak hazırlık ve yalnızca Codex Desktop adımları görünür.
- [x] Üç istemci seçimi hem klavye hem mobil kullanımla çalışıyor; görünüm geçişinde seçim kaybolmuyor.
- [x] Markdown çıktısı aktif istemciden bağımsız tüm kurulum seçeneklerini ve teknik ayrıntıları içeriyor.
- [x] Bağlantı deneme istemi ayrı kopyalanabiliyor; teknik ayrıntılar ana akışı uzatmıyor.
- [x] Token bilgileri somut ve tutarlı; gerçek gizli bilgiler veya doğrulanmamış destek iddiaları yok.

## Checks
- Seçim varsayılanı, geçersiz seçim, istemciye özgü görünüm ve tam Markdown içeriği için odaklı HTTP/içerik testleri.
- `go tool templ generate`; gereken Tailwind üretimi.
- `go test ./internal/web/...`
- Masaüstü ve mobilde dört adımlı akışı incele; son kabul WKY-19'da.

## Progress
Tamamlandı (2026-10-27). Sonraki görev: WKY-18.

- Dört adımlı MCP rehberi; varsayılan Codex Desktop, sunucuda normalize edilen client sorgusu, görünüm bağlantılarında korunan seçim. Tercih kaydı yok.
- Rehber yalnızca seçili istemciyi gösterir; tam Markdown seçime bağlı değildir. Tokenın kaynağı, eşitliği, süreç ortamına aktarımı, özel port ve Desktop yeniden başlatma bilgileri ilgili adımdadır. Gerçek .env/token/plan verisi okunmadı.
- Teknik ayrıntılar kapalı details bölümlerinde. Bağlantı deneme istemi WKY-16 copy.js mekanizmasıyla ayrı kopyalanır; JS yokken readonly metin elle seçilebilir.
- API rehberi içeriği değiştirilmedi. Dördüncü adımın plan oluşturma istemi WKY-18'e bırakıldı.
- Kontroller geçti: go tool templ generate; Tailwind üretimi; go test ./internal/web/...; go vet ./internal/web/...; geçici cmd/workey derlemesi; gofmt; git diff --check.
- Geçici dev veritabanında headless Edge: 390×844 ve 1280×844; üç istemci, dokunma/klavye, görünüm geçişi, tam Markdown eşitliği, ayrı istem kopyalama ve pano geri okuma, kapalı ayrıntılar, yatay taşma ve JS kapalı kullanım doğrulandı. İki genişlikte ekran görüntüleri incelendi. Test sunucuları durduruldu.
- Mevcut bağlantı komutları korundu; Desktop adımına mevcut Codex kayıt komutu da kondu. Ortak yapılandırma ve bearer_token_env_var resmî kaynakla kontrol edildi: https://learn.chatgpt.com/docs/extend/mcp?surface=cli (2026-10-27).
- Tam test paketi, CGO kapalı derleme ve güvenlik taraması bu odaklı rehber değişikliğinde yeniden çalıştırılmadı. Birleşik son kabul WKY-19'da. Commit/push yapılmadı.
