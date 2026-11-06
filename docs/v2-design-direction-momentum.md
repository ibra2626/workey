# Workey · Tasarım Yönü — "İvme / Momentum"

> Durum: **yön planı (taslak)**, 2026-11-05. Henüz hiçbir şey uygulanmadı.
> Onay verilmeden `DESIGN.md`, token katmanı veya `.templ` dosyaları değişmez.
> Kullanıcı bu oturumda PRODUCT.md'deki bağlayıcı "asla" listesini (oyunlaştırma,
> kutlama, seri, canlı gradyan, mor/menekşe) bu yön için açıkça serbest bıraktı.
> Değişmeyen ürün gerçeği: Türkçe karakter + `M2·17, s. 210–225` notasyonu,
> "bu akşam ne çalışacağım / tek dokunuşla tamamla / 38 haftalık ilerleme",
> erişilebilirlik ve reduced-motion, templ+Tailwind+htmx (JS framework yok).

---

## 1. Neden yeni bir dünya

Bugünkü "Training Log" dünyası doğru bir fikir (maratona benzer periyotlama:
faz, haftalık hacim, seanslar) ama ifadesi kasıtlı olarak **sessiz**: greige
zemin, düz beyaz paneller, üç soluk bölge rengi, tek clay aksan, neredeyse hiç
hareket. Sakin — ama kullanıcının kendi deyişiyle "kurumsal, renksiz" tarafa
kayıyor ve 38 haftalık yolculuğun en çok ihtiyaç duyduğu şeyi vermiyor:
**görülebilir momentum.** Akşam 20:30'da telefonun başında "bu uzun yolda
ilerliyorum" hissini yaşatan bir şey yok; emek harcandığında sistem görünür
biçimde yanıt vermiyor.

**Kuzey Yıldızı: "İvme" — yaşayan bir çalışma sezonu.**
Workey artık emekle *canlanan* bir sistem. Her tamamlanan seans sistemde
görünür bir enerji bırakır: bölge renkleri ışıldar, ilerleme yaylanarak büyür,
yol üzerindeki konumun ileri gider, seri yüklenir. Sakinliğini kaybetmeden
*canlı* olur. Atletik DNA kalır ama sessiz günlükten **parlak bir sezon
panosuna** döner. Kaçırılan gün hâlâ suçlamaz; başarı ise artık görünür biçimde
kutlanır.

---

## 2. İmza fikir: Gün + Gece teması

Ürünün birincil anı **akşam 20:30, telefonda.** Bu tek gerçek, renk ve
animasyon için en güçlü gerekçe: Workey'in **imza teması karanlık "Gece
çalışması" temasıdır.**

- **Gündüz teması** (gündüz/masaüstü planlama anı): sıcak kâğıt zemin, aydınlık,
  canlı ama yumuşak. Hafif radyal sıcaklık (artık gradyan serbest) zemine
  derinlik katar.
- **Gece teması** (varsayılan akşam, `prefers-color-scheme` + saat ipucuyla):
  derin orman-gece zemini, ışıldayan bölge renkleri, lamba ışığı sıcaklığı.
  Çalışma odası bu temada bir "stüdyo" gibi aydınlanır. Parıltı (glow) burada
  dekorasyon değil; dikkatin nereye gideceğini söyleyen tek araç.

Tema, kullanıcının tercihini hatırlar; ikisi de birinci sınıf, "gece" sonradan
eklenmiş bir ışık düğmesi değil.

---

## 3. Renk sistemi

Renk, zaten **anlam taşıyan** üç seans bölgesinin etrafında örgütlenir; bu
yüzden "rastgele renklendirme" değil, var olan anlamı canlandırmaktır. Üstüne
tek bir "momentum" aksanı biner ve seriyle birlikte yoğunlaşır.

### Üç bölge (canlandırılmış, anlam sabit)
| İş | Eski (soluk) | Yeni Gündüz | Yeni Gece (ışıldar) |
|---|---|---|---|
| **Konu** (20 dk) | `#66a69d` | Teal `#12B8A6` | `#2DD4BF` |
| **Soru** (50 dk, ana set) | `#d0963a` | Marigold `#F59E0B` | `#FBBF24` |
| **Yanlış defteri** (20 dk) | `#8aa7c4` | İndigo `#6366F1` | `#818CF8` |

Bölge renkleri hâlâ **yalnızca** kendi seans parçasını boyar (kural korunur);
tek fark artık canlı ve ışıldayabilir olmaları.

### Momentum aksanı (seriyle yoğunlaşır)
Sıcak mercandan sıcak magentaya bir aralık; gradyan artık serbest. Momentum
ölçeri, seri alevi ve tamamlama patlaması bu aileden beslenir.
- `momentum-1` Mercan `#FF6B4A`
- `momentum-2` Gün batımı `#F43F8E`
- Gradyan `linear-gradient(135deg, #FF6B4A, #F43F8E)` — yalnızca momentum/kutlama
  yüzeylerinde; asla gövde metninde veya nötr panelde.

### Nötr + zemin
- **Gündüz:** zemin `#F5F1E8` (sıcak kâğıt, hafif radyal sıcaklık), yüzey
  `#FFFFFF`, yükseltilmiş `#FFFDF9`, ink `#15302A`.
- **Gece:** zemin `#0D1A16` (hafif aurora gradyanı), yüzey `#15241E`,
  yükseltilmiş `#1C2E27`, metin `#E8F0EC`, ikincil `#9DB3AC`.

### Başarı / kutlama
Zümrüt-lime `#22C55E` yalnızca seri, kilometre taşı ve "tamamlandı" doğrulama
parıltısında. Hata/eksik durumu **asla kırmızı değil** — kaçırılan gün nötr
taranmış kalır (suçlamama ilkesi korunur).

### Erişilebilirlik notu
Canlı renkler metin kontrastını zorlar. Kural: bölge dolgularının üzerindeki
etiketler ink (gündüz) / koyu-üstü-açık (gece) kalır; beyaz metin yalnızca
kontrastı geçen momentum/başarı dolgularında kullanılır. Her renk çifti ≥ 4.5:1
doğrulanır (uygulama aşamasında contrast-check ile).

---

## 4. Hareket sistemi

Amaçlı, fizik tabanlı, **yorulmayan.** Her hareketin bir işi var: durumu
doğrulamak, dikkati yönlendirmek, emeği ödüllendirmek. Varsayılan eğri yumuşak
bir yay (overshoot'lu): `cubic-bezier(.34, 1.56, .64, 1)` kısa UI geçişleri,
`cubic-bezier(.22, .61, .36, 1)` grafik büyümeleri için.

**İmza anlar:**
1. **"Seansı tamamla"** → buton dolgusu bir süpürme ile momentum gradyanına
   döner, hafif bir parçacık patlaması (ölçülü konfeti, serbest), günün çubuğu
   yaylanarak ink'e kilitlenir, seri sayacı +1 tıklar ve momentum ölçeri dolar.
   Tek, tatmin edici, ~900 ms'lik bir an — sonra sakin "Tamamlandı" durumuna oturur.
2. **İlerleme çubukları** yaydan hafif overshoot ile büyür.
3. **38 hafta şeridi** → aşağıdaki "Zirveye giden yol": mevcut hafta nazikçe
   nabız atar; hafta tamamlanınca sıvı bir süpürmeyle dolar.
4. **Çalışma odası** → canlı geçen-süre halkası yavaşça "nefes alır"; aşama
   geçişleri renk kaymasıyla çapraz geçiş yapar.
5. **Sekme göstergesi** aktif sekmeler arasında kayar (shared indicator).
6. **Sayfa geçişi** → View Transitions API destekleyen tarayıcıda hero sayısı
   (zirveye gün, haftalık hacim) paylaşılan öğe olarak taşınır; desteklenmeyende
   düz render (mevcut davranış).

**Reduced-motion:** tüm uzamsal hareket, döngü, parçacık ve geçiş kapanır;
yalnızca opaklık ve son durum kalır. Performans: yalnızca transform/opacity,
layout thrash yok; parçacıklar sayıca sınırlı ve GPU-composited; kalıcı
`will-change` yok, animasyon katmanları bitişte serbest bırakılır.

---

## 5. Oyunlaştırma (dürüst ve zevkli)

Serbest bırakıldı — ama ucuz değil. İlke: **sahte puan yok; gösterilen her sayı
gerçek.** Hacim gerçek planlı dakikadır, seri gerçek çalışılan günlerdir.

- **Seri ("İvme"):** ardışık çalışma günleri, yüklenen bir momentum ölçeri /
  alev olarak. Kırılınca **suçlama yok** — "ara verdin; bugün yeniden başlat"
  çerçevesi, nazik bir "kurtarma" anı. Seri bir ödül değil, bir ivme göstergesi.
- **Kilometre taşları:** bir Faz'ın bitişi, her 5 haftalık blok, zirveye varış.
  Ölçülü kutlama anı + 38 haftalık yolda kalıcı bir "mihenk taşı" işareti.
- **Zirveye giden yol:** 38 haftalık şerit, sınav zirvesine (18 Tem 2027) doğru
  bir tırmanış olarak yeniden tasarlanır. Arkanda kalan yol momentum rengiyle
  dolar, konumun bir işaretle gösterilir, zirvede sınav tarihi parlar.
- **Konu ustalığı:** İlerleme'deki konu grupları, canlı dolum animasyonuyla
  "ustalık" izleri olarak; tamamlanan konular ışıldar.

Kutlama **yalnızca gerçek kazanımda** tetiklenir, asla boş durumda; ve her
zaman reduced-motion ile sönümlenebilir.

---

## 6. Ekran ekran plan

### Shell + navigasyon
- Sol ray / alt sekme: aktif durum momentum parıltısıyla; sekmeler arası kayan
  gösterge. Ray ayağındaki "zirveye __ gün" canlı, hero bir sayaç olur.
- Gündüz/gece teması shell düzeyinde; geçiş yumuşak çapraz solma.

### Bugün (amiral ekran)
- İlk panel kuralı korunur: kaynak · bölüm · sayfa · seans profili · birincil
  aksiyon, 390×844'te kaydırmasız. Ama artık **canlı**: seans profili ışıldayan
  bölge çubuklarıyla, akşam gece temasında stüdyo gibi.
- Üstte ince bir momentum/seri şeridi (bugünün durumu, bu akşamın ivmesi).
- Birincil aksiyon = imza tamamlama anı (§4.1).
- Boş durum ("Plan henüz yazılmadı") bile sıcak ve davetkâr: asistanın planı
  yazmasını bekleyen, nazikçe nabız atan bir yer tutucu.

### Çalışma odası (imza yüzey)
- Gece temasında aydınlanan bir stüdyo. Canlı nefes alan geçen-süre halkası,
  aşama ilerlemesi renkle, masadaki eşlikçi figürü lamba ışığında.
- Aşama geçişi: renk kayması + yumuşak çan (mevcut) + canlı bölgeye yumuşak vurgu.
- Tamamlama: imza kutlama, sonra sakin kapanış.
- Mevcut "Burada bıraktım / Çalışmayı kapat" mantığı ve sürümlemesi aynen korunur.

### Hafta
- Hacim çubukları canlı bölge renkleriyle, yüklenişte yaylanarak büyür.
- Bugün nabız atan bir vurguyla; taşınan gün nötr taranmış (kırmızı yok).

### İlerleme
- "Zirveye giden yol" ana görsel (§5). Kilometre taşları, konu ustalığı izleri.
- Gerçek sayılar: hacim, tamamlanan seans, zirveye hafta — hepsi dürüst.

### Kaynaklar
- Verimli açılır/özetli yapı korunur (yüzlerce birim). Üstüne: kaynak başına
  canlı dolan ilerleme halkası, renkli tür rozetleri (video = momentum rozeti),
  açık bölümde yumuşak vurgu. İç içe kart yok; ayraçlar kalır.

### Kullanım / API
- Araçsal ve sakin kalır; yalnızca yeni aksan ve tipografiyle tutarlı. Burada
  kutlama/animasyon aranmaz (Operate kipinde kısıtlılık = kalite).

---

## 7. Teknik yaklaşım

- **Token katmanı:** DESIGN.md yeniden yazılır; yeni palet + gece teması
  tokenları + hareket tokenları. Tailwind standalone CLI ile derlenir; değişken
  renk üzerinde `/35` opaklık yok (kural korunur), yumuşak tonlar açık token.
- **Tema:** CSS değişkenleri + `data-theme` / `prefers-color-scheme`; FOUC'suz.
- **Hareket:** çoğunlukla saf CSS; yaylar ve parçacıklar için ufak vanilla JS
  (framework yok); View Transitions progressive enhancement.
- **Erişilebilirlik:** her etkileşimli öğede görünür focus; ≥4.5:1; tam
  reduced-motion yolu; Türkçe casing CSS'e bırakılmaz.
- **Performans:** transform/opacity, composited katmanlar, sınırlı parçacık,
  kritik font önyükleme korunur.

---

## 8. Riskler ve açık kararlar

- **Gece temasının varsayılan tetiği:** yalnızca saat mi, yalnızca
  `prefers-color-scheme` mi, yoksa ikisi + manuel geçiş mi? (Öneri: ikisi +
  manuel.)
- **Konfeti/parçacık dozu:** "zevkli" ile "çocuksu" ince çizgi; ilk yapımda
  düşük dozla başlayıp canlı bakışla ayarlamak.
- **Kontrast:** indigo/marigold üzerinde etiketler; uygulamada doğrulanacak.
- **Kapsam/sıra:** aşağıdaki fazlama önerisi.

## 9. Önerilen uygulama sırası (onay sonrası)

1. Token katmanı + gece teması + hareket tokenları (DESIGN.md yeniden yazımı).
2. Shell + navigasyon hareketi + tema geçişi.
3. Bugün hero + imza tamamlama anı.
4. Çalışma odası (imza yüzey).
5. Hafta + İlerleme grafikleri + "zirveye giden yol" + oyunlaştırma.
6. Kaynaklar + Kullanım cilası.
7. Erişilebilirlik + reduced-motion + performans geçişi; 1280 ve 390'da canlı
   doğrulama; dedektör temiz.
