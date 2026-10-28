# Workey: Codex instructions

## Proje bağlamı
- Proje kuralları ve odaklı doğrulama komutları için bu klasördeki `CLAUDE.md` dosyasını oku. Oradaki global Claude/claude-kit atıfları Codex kurulumu sayılmaz.
- İşe göre `PRODUCT.md`, `DESIGN.md` ve `.agent/state/CURRENT.md` kaynaklarını kullan; mevcut kullanıcı değişikliklerini koru.

<!-- beyin-codex-bridge:start -->
## Beyin köprüsü: otomatik receipt
- Bu proje `D:/Github/Atlas` kasasına bağlıdır; proje etiketi `workey`, receipt kaynağı `notes/projeler/workey.md`.
- Kullanıcı hafıza kullanmamanı istemediyse, anlamlı iş tamamlandığında final yanıtından önce kullanıcıdan ayrıca istemesini beklemeden receipt gönder. Uygulama, düzeltme, doğrulama sonucu, engel veya devir bu kapsamdadır; salt okuma soruları ve selamlaşma için receipt yazma. Aynı sonucu tekrar kaydetme.
- Komut: `& 'C:/Users/SUBU/AppData/Local/Programs/Python/Python313/python.exe' 'D:/Github/Atlas/beyin.py' receipt --harness codex --session SESSION --event-id EVENT_ID --summary 'Sonuç özeti' --ref notes/projeler/workey.md --json`.
- SESSION için açılıştaki `Optional Beyin bridge` bağlamında gelen `Receipt session=...` değerini aynen kullan; oturum kimliği uydurma. Değer yoksa aşağıdaki yöntem yalnız gerçek `CODEX_THREAD_ID` mevcutsa kullanılabilir: köprü oturumu, `json.dumps([str(Path('D:/Github/workey').resolve()), os.environ['CODEX_THREAD_ID']])` UTF-8 metninin SHA-256 özetinin ilk 24 karakteridir. Kimlik bulunamazsa bağlam eksikliğini bildir; sahte oturumla doğrulandı deme.
- EVENT_ID sonuca özgü olsun. Özet 1–3 cümlede değişikliği, geçen/çalıştırılmayan kontrolleri ve açık adımı anlatsın. `Öğrenilen: yok (rutin proje işi)` ya da kaynaklarıyla damıtılmış gerçek öğrenimi ayrıca belirt. Sır, .env değeri, ham log, kod veya sohbet dökümü aktarma.
- Karmaşık/çok satırlı özet için UTF-8 geçici JSON dosyası ve `receipt --file DOSYA --harness codex` kullan; JSON alanları `event_id`, `session`, `summary`, `refs` olmalı. `--file` ile alan bayraklarını karıştırma; geçici dosyayı iş bitince kaldır.
- Kaynak araması gerektiğinde aynı Python ve CLI ile `context 'konu' --project workey --json` çalıştır. Kaynağını okumadan tarihsel iddiayı doğrulanmış kabul etme.
- Receipt komutu runtime/kasa erişim hatası verirse kaydedildi deme; ortamın izin mekanizmasıyla yetkili çalıştırmayı dene. Kullanıcıya sonucu tek satırla bildir.
<!-- beyin-codex-bridge:end -->
