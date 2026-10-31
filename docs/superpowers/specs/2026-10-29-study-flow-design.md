# Workey — çalışmaya başla ve kaldığın yerden dön

Tarih: 2026-10-29
Durum: 2026-10-29 SF-01..SF-05 uygulandı; sonraki görev SF-06.
Görev sonuçları .agent/work/study-flow/tasks/ altında kayıtlıdır.

## Amaç

Workey'nin planı okuma ve tamamlandı işaretleme arasındaki boşluğu doldurmak:
başlamayı kolaylaştırmak, yarım kalan emeği görünür tutmak ve ara sonrası
devam edilecek yeri korumak. Başarı ölçüsü uygulamada geçirilen süre değildir.
Fiziksel kitapla çalışmaya geçiş ve geri dönüş kolaylığıdır.

## Kullanıcının seçtikleri

- “Masaya geç” ve “Kaldığım yer” yaklaşımlarının birleşimi.
- Başlama ritüeli; bitirme veya kaldığı yeri bırakma; sonraki ziyarette devam.
- Kaldığı yeri tanımak için kitap/sayfa ve çalışma aşaması yeterli.
- Rozet, streak, suçluluk yaratan mekanikler ve sosyal çalışma odaları yok.
- Bu aşamada ürün kodu değil, yazılımsal uygulama planı hazırlanacak.

Aşağıdaki ayrıntılar mühendislik önerisidir; kullanıcı tarafından tek tek
onaylanmış kararlar olarak DECISIONS.md'ye taşınmayacak.

## Mevcut gerçeklik ve kaynaklar

- PRODUCT.md: tek kullanıcı, kağıt kitap, çalışma öncesi/sonrası kullanım,
  içerik üretmeyen uygulama, dış asistandan gelen plan.
- internal/web/views/today.templ: ilk seans, seans profili, tamamla/geri al;
  tamamlamada TodayMain ve özetler birlikte yenileniyor.
- internal/plan/read.go: durumlar planned/today/done/missed/moved;
  ilerleme tamamlanan seansların planlı dakikalarından hesaplanıyor.
- internal/plan/complete.go: bugün/geçmiş tamamlanabilir; gelecek yasak;
  gerçek değişiklikler audit kaydı üretir.
- internal/plan/service.go: save_plan açık bugünkü/gelecek seansları değiştirir
  ve eksiklerini siler; geçmiş ve tamamlanmış seanslar korunur.
- internal/store/migrations/0001_init.sql: kısmi çalışma veya açık çalışma yok;
  session_items seans güncellemesinde silinip yeniden oluşturulabilir.
- internal/web/views/usage_content.go: kaynak referansları ve kişisel koşulları
  isteyen mevcut plan istemi var, materyal hazırlama rehberi eksik.

PRODUCT.md/CLAUDE.md içindeki repo-boş ifadeleri güncel kodla çelişiyor;
taslak mevcut uygulamayı esas alır. Kişisel plan/veritabanı okunmadı.

## Yaklaşımlar

1. Masaya geç: somut ilk adım ve başlatma eylemi. Giriş eşiğini düşürür;
   tek başına yarım kalan çalışmayı korumaz.
2. Açık masa: ortam sesi ve çalışma boyunca eşlik. Kullanıcı bunu seçmedi;
   telefonun açık kalmasını teşvik edebilir ve kapsamı büyütür.
3. Kaldığım yer: sayfa/aşama kaydı ve dönüş bağlamı. Emeği görünür tutar;
   başlama desteği tek başına sınırlıdır.

Seçim: 1 + 3. İlk teslim sunucuda saklanan çalışma durumu ve yer imi üzerine
kurulur. Canlı zamanlayıcı, ses, bildirim ve sosyal özellikler kapsam dışıdır.

## Kullanıcı akışı — önerilen ayrıntılar

### Başlama

Bugünün planı ilk bakışta kitap, bölüm, sayfa ve ilk çalışma aşamasını gösterir.
“Çalışmaya geç” seansı başlatır. İkincil “İlk 10 dakikayla başla” aynı işlemi
küçük bir başlangıç hedefiyle yapar; plan süresini değiştirmez. Sürenin dolması
hiçbir işi tamamlamaz. İlk teslimde canlı geri sayım yok; başlangıç saati ve
başlangıç hedefi sunucudan okunur. Kullanıcı istediği an devam edebilir.

Çalışma görünümü Bugün içinde sade bir durumdur; ayrı uygulama veya yeni
navigasyon bölümü değildir. Kaynak/aşama ve kapatma eylemleri öne çıkar.
Telefonun kapanması, sekmenin kapanması ve uygulama yeniden başlaması kaydı
silmez; bunlar çalışmanın sürdüğünün veya sürenin ölçüldüğünün kanıtı sayılmaz.

### Burada bırakma

“Burada bıraktım” formu kaynak, aşama ve sayfa ister. Tek kaynak otomatik
seçilir; çok kaynaklı seanslarda kullanıcı ilgili kaynağı seçer. Sayfa alanı
“Devam edeceğin sayfa” olarak adlandırılır: sayfa bitmişse sonraki sayfa,
yarım kalmışsa aynı sayfa girilir. Uygulama otomatik +1 yapmaz.
Bir aşamanın veya kaynağın son sayfası bittiyse sonraki aşama/kaynak seçilir;
seansın tamamı bittiyse “Bitirdim” kullanılır. Son sayfa yarım kaldıysa aynı
sayfa geçerlidir. Sırf sonraki sayfa diye plan aralığı dışına yer imi yazılmaz.

Aşamalar akşam için konu/soru/yanlış defteri, uzun_blok ve deneme için tek
“Çalışma” aşamasıdır. Yanlış defterinde de sayfa, ilgili kitap sayfasıdır;
defterin fiziksel sayfası değildir. Aşamaya özel sayfa aralığı tanımlıysa o,
yoksa kaynak öğesinin ana aralığı doğrulanır. Aralık dışı/boş giriş kayıt
yapmaz; formdaki değerleri koruyan Türkçe hata döndürür.

“Çalışmayı kapat” sayfa girmeden de mümkündür: başlangıç kaydı korunur,
yer imi uydurulmaz. Görünür metin “Çalışma kapatıldı; devam yeri kaydedilmedi.”
Tamamlandı işareti sayfa veya başlatma zorunluluğu getirmeden kullanılabilir.

### Dönüş ve seçim

Sayfa yenilenince veya başka cihazdan girilince sunucudaki açık çalışma ve
son yer imi okunur. Gece yarısı çalışma kendiliğinden bitmez; günün planı
İstanbul tarihine göre değişir, eski seansın tarihi değiştirilmez.

Öncelik: açık çalışma varsa onun bağlamı; yoksa bugünün planı. Son yarım
çalışma, bugünkü planı gizlemeyen kısa bir devam bağlamı olarak gösterilir.
Eski işe devam etmek açık bir kullanıcı seçimidir; takvim değişmez.
Dinlenme gününde de varsa yarım işe isteğe bağlı devam sunulur; yoksa dinlenme
durumu kalır. Gelecek tarihli seans başlatılamaz; plan/sınav sonrası yeni
başlatma kapalı, mevcut açık çalışma kapatma/tamamlama erişilebilir kalır.

Tek açık çalışma vardır. Başka bir seansa geçmeden mevcut çalışma kapatılır;
bir seansın açılması ötekini sessizce kapatmaz. make_up bağlantısı olan eski
seans otomatik başlatılmaz; telafi seansına yönlendirilir. Kullanıcı eski
seansı tamamlamak isterse mevcut tamamlama davranışı korunur; telafi otomatik
tamamlanmaz veya kaldırılmaz. Tek eylemle ikinci bir tamamlama kaydı üretilmez.
Kullanıcı eski ve telafi seanslarını ayrı ayrı tamamladığını beyan ederse mevcut
seans tabanlı toplam ikisini de sayar; içerik/sayfa bazında tekilleştirme bu
teslimin kapsamı değildir. İlerleme öğrenilen benzersiz içerik ölçüsü değildir.

“Bitirdim” mevcut Complete işlemini çağırır ve açık çalışma durumunu aynı
transaction içinde kapatır. “Geri al” yer imini korur, çalışmayı paused yapar;
yer imi yoksa boş paused kaydı döner. Çalışma kaydı olmayan eski seansların
tamamla/geri al akışı ek form istemez.

## Veri ve mimari — öneri

Mevcut Go + templ + htmx, standalone Tailwind, SQLite ve internal/plan servis
katmanı korunur. Sunucu gerçeğin kaynağıdır; localStorage'a bağlı kayıt yok.

Yeni 0002_study_flow.sql migration'ı session_study tablosu ekler:

| Alan | Anlam |
| --- | --- |
| session_id | PK + sessions FK; silme RESTRICT |
| status | active / paused / closed; closed yalnız tamamlanmış seans |
| first_started_at, last_started_at, updated_at | UTC zaman damgaları |
| entry_mode | normal / ten_minutes; son başlatmanın hedefi |
| item_ordinal, stage, resume_page | birlikte NULL veya birlikte dolu yer imi |
| version | eski sekme yazmalarını önleyen artan sürüm |

status=active için unique partial index bütün uygulamada en fazla bir açık
çalışma sağlar. Tarihler mevcut enjekte edilen clock ile üretilir.
İlk sürüm gerçek çalışma dakikası, deneme puanı veya soru sayısı toplamaz.

Yer imi session_items.id'ye bağlanmaz: kayıtlar plan güncellemesinde yeniden
oluşabilir. Seans başlamışsa kaynakları ve takvim slotu korunduğundan ordinal
kararlıdır. Okuma çıktısında kitap/bölüm/sayfa referansı da çözümlenir.

### Plan güncellemesiyle çakışma

Herhangi bir session_study kaydı olan seans started kabul edilir. save_plan
bu seansın kaynaklarını, tarihini, süresini veya slotunu değiştiremez/silemez.
Aynı içerik unchanged; değişen hedef skipped reason=started; istekten eksik
seans korunur. Başka slota yeni seans yazmak eski çalışmayı taşımış sayılmaz.
Geçmiş/completed mevcut önceliklerini korur. Bütün kontrol ve replace aynı
transaction'dadır. Guide, OpenAPI, MCP şemaları ve istemler bu sözleşmeyi
açıklar; eski istemlerin kısmi başarıyı başarı saymaması korunur.

Bu koruma telafi planı yazmayı engellemez: geçmiş seansı değiştirmek yerine
makeup_of ile yeni seans eklenebilir. Asistan kullanıcıya kapsamı açıklar;
uygulama yer imini telafiye kendiliğinden kopyalamaz.

### Servis ve eşzamanlılık

Önerilen servis işlemleri: StartStudy(sessionID, mode, expectedVersion),
PauseStudy(sessionID, optionalBookmark, expectedVersion), ReadStudyContext.
Başlatma ilk defa version=0 ile istenir; her değişiklik sürümü artırır.
Tekrarlanan istek aynı sürümle gelirse mevcut hedef durum ve veri aynıysa
no-op olarak yanıtlanır; farklıysa conflict ve güncel durumu yenileme istenir.
Başka seans aktifse StartStudy conflict döndürür. Tüm değişiklikler kaynak,
önce/sonra verisiyle audit edilir; no-op ek audit üretmez.

Paused seans yeniden açılırken yer imi korunur; kullanıcı düzeltmedikçe
geriye/ileriye otomatik taşınmaz. Durumu okumak çalışma başlatmaz.
Complete/Undo aynı transaction içinde study version artırır; stale pause
tamamlanmış çalışmayı tekrar açamaz. Yeni web Complete/Undo formları da beklenen
study version gönderir; servis bunu kontrol eden opsiyonel bir seçenek veya
ayrı sürümlü servis girişi sunar. Çalışma kaydı yokken beklenen sürüm 0'dır.
Eski REST/MCP Complete istemcileri sürüm göndermeden mevcut davranışta çalışır;
bu istemciler için eski sekme niyeti tespit edilemez. Bu uyumluluk sınırı belgelenir.

REST/MCP ilk teslimde study verisini mevcut SessionView içinde ek, opsiyonel
study alanıyla okur. Yeni study yazma API/MCP aracı kapsam dışı; kullanıcı
webden başlatır/bırakır. SessionState enum'u korunur. Kısmi çalışma ayrı alandır;
webde geçmiş kısmi kayıt için “Kısmen çalışıldı” görünür, “Yapılamadı” olmaz.
makeup bağlantısı ayrıca görünür. Eksik kayıt “hiç çalışılmadı” kanıtı sayılmaz.

## İlerleme ve asistan girdileri

Kısmi sayfa veya 10 dakika hedefi biten seans/dakika yüzdesini artırmaz.
Hafta ve Bugün, yer imini ayrı metinle gösterir. Mevcut ilerleme göstergelerinin
“tamamlanan seansların planlı süresi” olduğu açıklanır; ölçülmüş süre veya
öğrenme yüzdesi iddiası kurulmaz. Telafi ve geçmiş seans birlikte otomatik
tamamlanmaz; mevcut kullanıcı tamamlama beyanı esas alınır.

Kullanım rehberine materyal hazırlama bölümü eklenir: kitap adı/baskısı,
okunaklı içindekiler, gerekli örnek sayfalar veya PDF, basılı sayfa/PDF sayfası
ayrımı, mevcut seviye ve uygun zaman. PDF/OCR yükleme altyapısı eklenmez;
materyal bağlı asistana verilir. Özel kaynak yüklenmiş gibi davranılmaz.

Asistan istemine ilk uygulama haftasından sonra kullanıcıyla yük kontrolü ve
yer imlerini dikkate alma adımı önerilir. İlk haftayı ayrıntılı, uzak dönemi
kaba planlama tercihi mevcut tüm-dönem istemini sessizce değiştirmez; ayrı
isteğe bağlı başlangıç yolu olarak sunulur. Gerçek tempo bilgisi kullanıcıdan
alınır; açık kalma zamanından hesaplanmaz. Geri dönüş için kaynak referansı,
yer imi, ilgili seanslar ve mevcut kapsamı koruma kuralları içeren kopyalanabilir
yeniden planlama istemi sunulur; kopyalama işleminde plan değişmez.

## Kapsam dışı ve sınırlar

- Sosyal oda, rozet/streak, kutlama, kırmızı borç listesi, baskıcı bildirim.
- Ses, müzik entegrasyonu, canlı sayaç ve arka planda gerçek süre ölçümü.
- İçerik üretme, otomatik takvim kaydırma, önkoşul grafiği tahmin etme.
- PDF depolama/OCR ve çok kullanıcılı ürün dönüşümü.
- Deployment; WKY-12 beklemesi ve WKY-20 zoom kabulü ayrı kalır.

## Kabul koşulları

1. Bugün veya geçmiş açık seans normal/10 dakika hedefiyle açılabilir; gelecek
   reddedilir. Aynı anda iki açık seans oluşmaz.
2. Başlatılmış kayıt sekme/telefon/sunucu kapanışından sonra yeniden okunur.
3. Tek/çok kaynak ve bütün seans türlerinde geçerli yer imi kaydolur; geçersiz
   sayfa/aşama başka kaydı bozmaz. Sayfa bırakma anlamı belirsiz değildir.
4. Kısmi çalışma tamamlandı sayılmaz; geçmişte kısmi emek doğru metinle görünür.
5. Tamamla/geri al study durumu ve özetlerle atomiktir; eski istemci çalışır.
6. save_plan started seansı açık veya eksik payload ile yok edemez; şema ve
   asistan yönergeleri yeni reason değerini taşır.
7. Gece yarısı, dinlenme, sınav sonrası kapanış ve telafi dönüşleri veri kaybetmez.
8. İki sekmenin çakışan yazması sessizce yer imini ezmez; ağ hatasında form
   değerleri korunur ve kayıt doğrulanmadan başarı bildirilmez.
9. 390x844'te ilk adım görünür; 320px/desktop/klavye/azaltılmış hareket ve
   JavaScript kapalı temel form yolları kullanılabilir.

## Uygulayıcıya bırakılanlar

SQL ve Go isimlerinin son hali, dosya bölme ve düz metin mikro düzenlemeleri.
Kullanıcı seçimini etkileyen kapsam değişiklikleri taslağa geri getirilir.
Görev sırası: .agent/work/study-flow/TASKS.md.

## Taslak inceleme sonucu

Bağımsız builder incelemesi yapıldı. Üç mühendislik açığı kapatıldı: yeni web
tamamla/geri al işlemlerinde sürüm kontrolü; eski+telafi ayrı tamamlanırsa mevcut
seans tabanlı toplamın anlamı; son sayfa sonrası aşama/kaynak seçimi.
Bu çözümler mühendislik önerisidir; ek kullanıcı tercihi olarak kaydedilmedi.

## Onay ve araç durumu

İlk taslak uygulama onayı içermiyordu. Kullanıcının sonraki SF-01'e başlama
talimatı bu görevin uygulama yetkisidir; commit/push istenmedi.
brainstorming skill'inin istediği superpowers:writing-plans skill'i bu makinedeki
.agents, .codex/skills ve .codex/plugins taramasında bulunamadı. Görev taslağı
mevcut repo görev biçimiyle elle hazırlandı; o skill çalıştırılmış sayılmaz.
