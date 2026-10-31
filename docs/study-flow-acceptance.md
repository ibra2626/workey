# SF-06 — çalışma akışı birleşik kabul raporu

Tarih: 2026-10-29 (Europe/Istanbul). Durum: **devam ediyor**; teknik kontroller
geçti, tarayıcı matrisi tamamlanmadı. SF-06 ve ayrı WKY-20 zoom kabulü kapanmadı.

## Ortam ve veri sınırı

Go 1.27.3 windows/amd64; aynı checkout, `codex/study-flow-experiment`.
SF-01..05 değişiklikleri korunmuştur. Kişisel .env, veritabanı veya gerçek plan
test girdisi olarak kullanılmadı. SQLite dosyaları disposable; kaynaklar
`testdata/sample-{outline,week}.json` ve açıkça **ÖRNEK VERİ** etiketli ek
G1·1, s. 100–110 kaydıdır. Bu ek referans tamamen temsili, doğrulanmış kitap
içeriği değildir. Commit/push/merge/deployment yapılmadı.

## Otomatik kabul

Yeni `internal/web/study_restart_test.go` her HTTP isteğini ayrı OS sürecinde
aynı geçici SQLite dosyasına karşı çalıştırır. Gerçek web handler, parola girişi,
bağımsız imzalı cookie'ler, CSRF ve sürüm alanları kullanılır. Döngüsel soket
kısıtına bağımlı olmaması için HTTP taşıması `httptest.NewRecorder` kullanır;
bu test gerçek ağ üzerinden sunucu kill/relaunch testi diye sunulmaz.

Yeni testte geçen sıra:

1. İki bağımsız oturum oluştur; 23:59 İstanbul'da 10 dakika hedefiyle başlat.
2. Yeni süreç/diğer oturum aktif çalışmayı ve hedefi yeniden okusun.
3. UTC hâlâ 28 Ekim iken 21:01Z = İstanbul 29 Ekim 00:01'e geç.
   Başlık Perşembe 29 Ekim, açık seans tarihi 28 Ekim, bugünün M1·4 planı görünür.
4. Soru aşaması s. 58 yer imini bırak; yeni süreç/diğer oturumdan yeniden oku.
   Bugünün planı ve eski işe devam seçimi birlikte görünür.
5. Eski işi açık seçimle sürdür; eski sürümlü sekmenin s. 57 yazmasını 409 reddet.
6. Bitir → geri al; son SQLite kaydı paused/version=5/s. 58 olarak kalsın.

Birleşik takımda yeniden geçen mevcut regresyonlar:

| Koşul | Kalıcı kanıt |
| --- | --- |
| Tek aktif seans, paralel başlatma, iki sekme | plan: `TestStudySingleActiveAndConcurrentStart`, `TestStudyTwoTabsCannotOverwriteBookmark` |
| Yer imi aralıkları, aşama ve kaynak sınırları | plan: `TestStudyBookmarkValidation`; web: `TestStudyBookmarkKindsAndInvalidSelections` |
| save_plan değiştirme/silme/unchanged ve skipped önceliği | plan: `TestSavePlanStartedProtection`, `TestStartedSkipPrecedence` |
| Gece yarısı, gelecek, dinlenme/sınav dönemi, telafi | plan: `TestStudyTemporalRulesAndMakeup`, `TestStudyReadSnapshotsAndProgress`; web: `TestStudyActiveSurvivesRestAndExamAndCannotSwitch` |
| Complete/Undo atomikliği ve rollback | plan: `TestStudyRollbackOnAuditFailure`, `TestStudyCompletionRollsBackIfSessionAuditFails` |
| Yer imsiz kapatma/bitir/geri al | plan: `TestStudyNoBookmarkCompletionUndo`; web lifecycle |
| HTTP formu, bağımsız auth, stale sürüm | web: `TestStudyFormsLifecycleAndIndependentSessions`, `TestWebCompletionStudyVersion` |
| Hafta yer imi, ayrı eski+telafi toplamı | web: `TestWeekPartialBookmarkSurvivesCompletionAndUndo`; plan: `TestPartialOriginalAndMakeupTotals` |
| Eski REST/MCP sürümsüz Complete ve study okuma sözleşmesi | API/MCP: `TestStudyRESTContractAndLegacyWrites`, `TestStudyMCPOutputContractAndLegacyWrites` |
| API/MCP rehber, şema ve kopya metni eşliği | views: `TestStudyContractDocsStayInSync`, `TestPlanningInputsSharedAndCopyable` |
| Tekrarlanan yazma, no-op audit, hata sonrası yeniden deneme | plan: `TestStudyLifecycleVersionsAndBookmarks` |

## Bu oturumdaki canlı tarayıcı kabulü

Geçici Go fixture sunucusu 127.0.0.1:8097 üzerinde, enjekte edilen dosya saati
ve kendi örnek SQLite dosyasıyla açıldı; config.Load/.env kullanılmadı.
8098 ayrı uç yalnız bu fixture için `script-src 'none'` başlığı taşıyordu,
ancak bu uçta canlı kabul gerçekleştirilemedi. Sunucu durduruldu ve geçici
`cmd/sf06preview/main.go` kaldırıldı; `bin/sf06` gitignored örnek veridir.

| Kontrol | Sonuç |
| --- | --- |
| IAB 390×844, başlangıç | Geçti: somut ilk sayfa, Çalışmaya geç, 10 dakika ve bitir ilk ekranda; Türkçe uzun çok kaynak metni sarılıyor. innerWidth=390, document scrollWidth=375 |
| IAB 10 dakika başlat | Geçti: açık çalışma/başlangıç hedefi ve 0/540 dk; telefonun açık kalması gerekmiyor metni |
| Çok kaynak formu | Geçti: kaynak seçimi, Soru aşaması, iki kaynağın aralıkları ve sonraki aşama/kaynak açıklaması |
| Geçersiz s. 111 | Geçti: 422 geri bildirim, kaynak=1/aşama=soru/sayfa=111 korunuyor; document.activeElement.id=study-feedback |
| IAB 320×844 hata ve açık form | Geçti: yatay sayfa taşması yok; innerWidth=320, scrollWidth=305; hata ve alanlar erişilebilir |
| Düzeltip s. 105 kaydet | Geçti: başarı mesajı, G1·1/Soru/s. 105 devam bağlamı, kısmi çalışma; 0/540 dk |
| Bağımsız Chrome oturumu | Geçti: ayrı parola girişi, aynı s. 105 yer imi ve paused bağlamı okundu |
| 1280×800 Bugün/Hafta/İlerleme ve uzun_blok | Bu oturumda tamamlanmadı; SF-04/05 kanıtları aşağıda ayrı tutulur |
| Tam klavye/focus sırası | Htmx hata odağı geçti; tam Tab/Shift+Tab matrisi tamamlanmadı |
| Reduced-motion gerçek ayarı | Çalıştırılmadı; CSS varlığı canlı ayar kabulü sayılmaz |
| JS kapalı tam lifecycle | Bu oturumda çalıştırılmadı; önceki SF-04 kanıtı ayrı |
| Ağ hatası/tekrar deneme | Servis retry regresyonları geçti; canlı bağlantı kesme bu oturumda yapılmadı |
| Gerçek %200 zoom | Tamamlanmadı. Chrome Control++ denemesinde innerWidth=2560, innerHeight=1271, DPR=1 değişmedi; zoom kanıtı değildir |

Chrome penceresi küçültülmüştü. Yerel pencere kontrolü denenince Computer Use,
Windows'ta etkin URL'yi politika kontrolü için yeterli güvenle belirleyemediğini
bildirip bu tur için durduruldu. Bu noktadan sonra tarayıcı işlemi yapılmadı.
Bu nedenle yukarıdaki açık satırlar tamamlandı sayılmadı; WKY-20 aynen açık.

Ekran kanıtı repo dışında, sohbetin visualization klasöründe:
`sf06-320-error.jpg` (320px, tam sayfa; geçersiz örnek yer imi).

Önceki kanıtlar yeni çalıştırma değildir: `.agent/work/study-flow/tasks/SF-04.md`
390/320/1280 yerleşim, klavye, stale sekme, düşen POST ve CSP no-JS lifecycle
sonuçlarını; SF-05.md Hafta/İlerleme ve rehber regresyonlarını kaydeder.
`docs/usage-verification.md` ayrı WKY-20 rehber kabulünün kaynağıdır.

## Teknik kapılar

- `go version`: go1.27.3 windows/amd64.
- Sıralı `go tool templ generate` → Tailwind standalone v4.3.3 üretimi geçti.
- SF-06'da listelenen tüm config/clock/store/plan/web/views/api/apidoc/mcpserver/cmd
  test paketleri geçti (son birleşik çalıştırma yetkili ortamda).
- `go build ./...` ve ayrı `CGO_ENABLED=0 go build ./...` exit 0.
  Go VCS modül stat önbelleği yazımına sandbox erişim uyarısı verdi; build başarısız olmadı.
- `go vet ./...` geçti.
- `gofmt -w .` ile CRLF→LF normalleştirildi; Git'te ilgisiz içerik farkı
  oluşmadı. Son `gofmt -l .` boş; `git diff --check` geçti.
- `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...`:
  **No vulnerabilities found.** Ağ/modül erişimli yetkili ortamda geçti.
- İlk sandbox çalıştırmasında MCP loopback testleri ve güvenlik taraması
  engellendi. İlk escalation otomatik incelemenin kullanım sınırı nedeniyle
  çalışmadı; kullanıcının “devam” talimatından sonraki yetkili tekrarlar geçti.
- Race, ekran okuyucu ve diğer tarayıcı motorları çalıştırılmadı.

## Açık kabul işi

Tarayıcı kontrolünün kullanılabildiği bir turda 320/390/1280 matrisini Bugün,
Hafta, İlerleme, akşam/uzun blok, klavye ve hata yollarıyla tamamla; gerçek
reduced-motion ayarı, JS kapalı lifecycle, bağlantı hatası sonrası retry ve
gerçek %200 zoom kanıtı al. Güncel kaynaklardan PRODUCT/DESIGN eşitlemesi yapıldı;
yeni bir kullanıcı kararı uydurulmadı. SF-06 bu satırlar doğrulanınca kapanabilir.
