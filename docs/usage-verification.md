# MCP/API kullanım doğrulaması — 27 Ekim 2026

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
