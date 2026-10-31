package views

import (
	"github.com/ibra2626/workey/internal/apidoc"
	"github.com/ibra2626/workey/internal/plan"
)

func usageTitle(key string) string {
	switch key {
	case "mcp":
		return "MCP Kullanımı"
	case "api":
		return "API Kullanımı"
	default:
		return "Kullanım"
	}
}

func planningInputSections() []UsageBlock {
	return []UsageBlock{
		namedSection("materials", "Plan için kaynaklarını hazırla",
			paragraph("Materyalleri Workey'e bağlı asistanına ver. Workey dosya yüklemez, PDF saklamaz veya OCR yapmaz; asistanın görmediği özel bir kaynağı okuduğunu varsayma."),
			UsageBlock{Kind: UsageList, Items: []string{
				"Her kitabın adı, kitap kodu ve baskısı; okunaklı içindekiler ve bölüm sırası.",
				"Gerekli örnek sayfalar veya erişebildiğin PDF; basılı sayfa numarası ile PDF görüntüleyicisinin sayfa sırası arasındaki fark.",
				"Doğrulanabilen bölüm ve sayfa aralıkları; varsa konu, soru ve antrenman aralıkları. Eksik referansı asistana sor; tahminle doldurtma.",
				"Mevcut seviye, bitirilen bölümler, hedef, uygun gün ve saatler, dinlenme ve diğer zaman kısıtları.",
			}},
			paragraph("İlk haftadan sonra gerçek temponu ve yükün uygunluğunu asistana anlat. Açık sekme süresi veya başlangıç hedefi çalışma süresi sayılmaz.")),
		namedSection("gradual-plan", "İsteğe bağlı: ilk haftayla başla",
			paragraph("Sınava kadar plan oluşturma istemi kullanılabilir. Alternatif olarak ilk haftayı ayrıntılı, uzak dönemi yalnız kaba kapsam olarak konuşabilirsin. Uzak dönem için doğrulanmamış sayfa veya seans yazılmaz. Kopyalama planı değiştirmez; asistan önce okur, yazdıktan sonra yeniden okuyarak doğrular."),
			UsageBlock{Kind: UsagePrompt, ID: "first-week", Text: firstWeekPrompt}),
		namedSection("replan", "Ara sonrası planı yeniden değerlendir",
			paragraph("Bu istemi kopyalayıp Workey'e bağlı asistanına gönder. Yer imi sunucudan okunur; kimlik veya token ekleme. Kopyalama plan yazmaz ve eski işi otomatik bugüne taşımaz. Asistan okuma → yazma → yeniden okuma sırasını izler."),
			UsageBlock{Kind: UsagePrompt, ID: "replan", Text: replanPrompt}),
	}
}

const firstWeekPrompt = `Workey bağlantısıyla önce get_agent_guide (REST: GET /api/v1/agent-guide), get_outline, ilgili get_plan ve get_progress eşdeğerlerini oku. Güncel today, timezone ve exam_date değerlerini esas al.
Hedefim/seviyem: [doldur]. Uygun gün ve saatlerim: [doldur]. Doğrulanabilir kitap/baskı/içindekiler/sayfa kaynaklarım: [asistana ver].
İlk haftayı ayrıntılı, uzak dönemi kaba kapsam olarak birlikte hazırlayalım. Eksik bölüm veya sayfa referansı uydurma; eksikleri sor ve yanıtımı bekle. Uzak dönem için doğrulanmamış seans yazma. İlk uygulama haftasından sonra gerçek tempomu ve yükü bana sor; açık kalma zamanından süre hesaplama.
Yazılacak aralığı ve korunacak kayıtları benimle netleştir. Yazmadan hemen önce outline, ilgili plan ve study/yer imini yeniden oku. save_outline faz ve grupların, save_plan aralıktaki başlanmamış açık seansların TAM listesini değiştirir; eksik açık seanslar silinir. Korunacak kayıtları tüm kaynaklarıyla dahil et. Geçmiş, tamamlanmış ve başlanmış seansları koru; başlanmış seansın kaynak/tarih/saat/süresini değiştirme. Telafi bağlantılarını koru.
Her yazma sonucunda created/updated/deleted/unchanged/skipped listelerini incele. skipped reason=started dahil atlanan değişikliği başarı sayma; hata veya belirsizlikte dur, körlemesine tekrar yazma. İlgili aralığı, outline ve ilerlemeyi yeniden okuyup doğrula; kalan kapsamı bildir. Planlama sırasında seans tamamlama çağrısı yapma.`

const replanPrompt = `Workey bağlantısıyla ara sonrası planımı değerlendirelim.
Gerçek tempom, uygun zamanım ve değişen koşullarım: [doldur]. Eksik kaynaklarım veya henüz çalışmadığım kapsam: [bildiğimi yazacağım; eksikleri bana sor].
Önce get_agent_guide (REST: GET /api/v1/agent-guide) ile today, timezone ve exam_date değerlerini oku. Sınav geçmişse yazmadan hedefi benimle netleştir. get_outline, ilgili geçmiş/bugün/gelecek get_plan aralıkları ve get_progress eşdeğerlerini oku; büyük aralıkları haftalara böl.
study durumunu, yer iminin kitap/bölüm/aşama/resume_page değerlerini ve makeup_of/moved_to ilişkilerini incele. Yer imi devam edeceğim sayfadır; otomatik +1 yapma veya telafiye kopyalama. Kısmi çalışmayı tamamlanmış ya da ölçülmüş dakika sayma. Eski işi otomatik bugüne yükleme; eksik kapsamı ve yetişmeyecek işi benimle netleştir. Kitap/bölüm/sayfa uydurma; kaynak eksikse sor ve yanıtımı bekle.
Yazmadan hemen önce ilgili aralığı yeniden oku. Her study kaydı begun/başlanmış kabul edilir: kaynak, tarih, start_time, duration_min ve items korunur. save_plan aralıktaki başlanmamış açık seansların TAM listesini değiştirir; listede bulunmayanlar silinir. Korunacak açık seansları tüm items ile dahil et; silme/değiştirme kararını benimle netleştir. Geçmiş, tamamlanmış ve başlanmış seansları koru. Outline değişirse korunacak faz ve grupları da içeren tam listeyi gönder; kullanılan grupları silme.
Telafi gerekiyorsa kaynak seansı değiştirmeden makeup_of ile ayrı seans öner ve bağlantıyı doğrula. Eski seans ve telafisini otomatik tamamlamayacaksın; ikisini ayrı tamamlamak mevcut seans tabanlı toplamda ikisini de sayar, benzersiz içerik ölçmez.
Okuma alanları id/state/completed/notation/study yazma gövdesine girmez; topic_group_key yazmada topic_group olur. Her save_plan sonucunun created/updated/deleted/unchanged/skipped listelerini incele; skipped reason=started dahil atlananları başarı sayma. Hata, belirsiz yazma veya doğrulanamayan sonuçta dur; kalan haftaları yazma. Tekrar denemeden önce oku; aynı gövde için aynı idempotency_key kullan.
Yazdıktan sonra aynı get_plan aralığını, outline ve get_progress eşdeğerlerini yeniden oku. Yer iminin, korunacak kayıtların ve telafi ilişkilerinin korunduğunu doğrula; değişen ve eksik kalan kapsamı Türkçe bildir. Bu planlama sırasında complete_session veya REST tamamlama/geri alma çağrısı yapma.`

var mcpToolDescriptions = []struct{ Name, Description string }{
	{"get_agent_guide", "İlk çağrı: sunucu tarihi, desteklenen değerler, kurallar ve işlem sırası."},
	{"get_outline", "Fazları ve konu gruplarını oku."},
	{"save_outline", "Fazların ve grupların tam listesini değiştir."},
	{"get_plan", "from–to dahil aralıktaki seansları ve kimliklerini oku."},
	{"save_plan", "Aralıktaki başlanmamış açık seansların tam listesini değiştir; eksikleri silinir. Başlanmış seans korunur; değişikliği skipped reason=started döner."},
	{"complete_session", "session_id ile tamamla; done=false ile geri al. Yalnızca kullanıcı isteğiyle."},
	{"get_progress", "Seans, dakika, konu grubu ve 38 haftalık ilerlemeyi oku."},
}

// One setup explanation serves both HTML/Markdown guides. Tokens are supplied
// by the operator; Workey does not issue them through a login or API endpoint.
func apiTokenSetup() UsageBlock {
	return namedSection("api-token", "API tokenını oluşturun veya mevcut tokenı alın",
		paragraph("Workey API tokenını bir siteden almazsınız; ilk kurulumda kendiniz oluşturursunuz. Web giriş parolası, API tokenı ve oturum imzalama anahtarı farklı değerlerdir. Mevcut bir kurulum kullanıyorsanız yeniden token üretmeyin: proje dizinindeki .env dosyasının WORKEY_API_TOKEN değerini alın veya sunucuyu kuran kişiden güvenli yoldan isteyin."),
		paragraph("İlk kurulum için PowerShell 7 terminalinde aşağıdaki komutu çalıştırın. Komut 32 rastgele baytı Base64 metnine çevirir ve ekranda gösterir; dosya veya sunucu ayarını değiştirmez."),
		code(randomSecretCommand, "powershell"),
		paragraph("Çıktıyı .env dosyasındaki WORKEY_API_TOKEN=change-me satırında change-me yerine yapıştırın. Sunucuyu bu ayarla başlatın; zaten çalışıyorsa yeniden başlatın. Ardından istemcinin bağlantı adımlarında aynı değeri kullanın. Sunucu süreç ortamında WORKEY_API_TOKEN tanımlıysa .env yerine o değer kullanılır; iki tarafta aynı olmalıdır. Gerçek tokenı paylaşmayın veya Git'e eklemeyin."),
	)
}

const randomSecretCommand = `$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
try {
    $bytes = [byte[]]::new(32)
    $rng.GetBytes($bytes)
    [Convert]::ToBase64String($bytes)
} finally {
    $rng.Dispose()
}`

func mcpSections(guide plan.AgentGuide) []UsageBlock {
	const tokenInput = `$env:WORKEY_API_TOKEN = Read-Host 'Workey API tokenı' -MaskInput`
	const codexAdd = `codex mcp add workey --url http://127.0.0.1:8080/mcp --bearer-token-env-var WORKEY_API_TOKEN
codex mcp list`
	tokenHelp := paragraph("Bir önceki hazırlık adımında oluşturup .env dosyasının WORKEY_API_TOKEN alanına kaydettiğiniz tokenı kullanın. Mevcut kurulumda aynı alandaki değeri alın; sunucu ortam değişkeniyle başlatılıyorsa o etkin değeri kullanın veya kuran kişiden isteyin. Aşağıdaki Read-Host komutu yeni token üretmez; bu değeri gizli girişle alıp istemcinin ortamına tanımlar. Workey'in .env okuması, ayrı başlatılan istemciye ortam değişkeni aktarmaz. Tokenı URL'ye veya config.toml içine yazmayın.")
	desktop := section("Codex Desktop",
		paragraph("Ayarlar → MCP sunucuları → Sunucu ekle yolunu açın. Workey zaten ekliyse mevcut kaydı düzenleyin."),
		UsageBlock{Kind: UsageList, Items: []string{
			"Ad: Workey",
			"Tür: Akış destekli HTTP (Streamable HTTP)",
			"URL: http://127.0.0.1:8080/mcp",
			"Taşıyıcı token için ortam değişkeni: boş bırakın. Doluysa içeriğini silin.",
			"Başlıklar → Anahtar: Authorization",
			"Başlıklar → Değer: Bearer SENİN_TOKENIN",
		}},
		paragraph("SENİN_TOKENIN yerine Workey projesindeki .env dosyasının WORKEY_API_TOKEN= satırındaki değeri yazın. Bearer kelimesinden sonra bir boşluk bırakın. Sunucu ortam değişkeniyle başlatılıyorsa o etkin tokenı kullanın. Tokenı sohbete göndermeyin."),
		paragraph("Kaydet'e basın. Codex Desktop'ı tamamen kapatıp yeniden açın ve yeni sohbet başlatın. Workey sunucusu çalışır durumda kalmalı. Ardından aşağıdaki bağlantıyı doğrulama istemini gönderin."),
	)
	desktop.Client = "codex-desktop"
	cli := section("Codex CLI", tokenHelp, code(tokenInput+"\n"+codexAdd+"\ncodex", "powershell"),
		paragraph("workey kaydı zaten varsa tekrar eklemeyin. Codex'i tokenı tanımladığınız aynı terminalden başlatın; /mcp ile bağlantıyı kontrol edin."))
	cli.Client = "codex-cli"
	claude := section("Claude Code", tokenHelp,
		paragraph("Proje dizininde aşağıdaki komutları aynı terminalde çalıştırın. Komut tokenı yerel Claude yapılandırmasına yazar; bu dosyayı Git'e eklemeyin. Claude Code'u aynı terminalden başlatın."),
		code(tokenInput+`
claude mcp add --scope local --transport http workey http://127.0.0.1:8080/mcp --header "Authorization: Bearer $env:WORKEY_API_TOKEN"
claude mcp get workey
claude`, "powershell"))
	claude.Client = "claude-code"
	alternative := usageSection("Alternatif: Codex CLI için config.toml", []string{"CLI komutuyla ekleme yerine kullanıcı dizininizdeki .codex/config.toml dosyasına aşağıdaki kaydı ekleyebilirsiniz. workey kaydı varsa ikinci kez eklemeyin. Codex CLI'yi WORKEY_API_TOKEN değişkenini tanımladığınız aynı terminalden başlatın. Desktop için yukarıdaki ayarlar ekranını kullanın."},
		`[mcp_servers.workey]
url = "http://127.0.0.1:8080/mcp"
bearer_token_env_var = "WORKEY_API_TOKEN"`, "toml", "")
	alternative.Kind = UsageDetails
	troubleshooting := usageSection("Bağlantı sorunları ve kapsam", []string{
		"Connection refused: Workey'in çalıştığını ve portu kontrol edin. 401: istemcideki tokenın Workey sunucusundakiyle aynı olduğunu kontrol edin. Desktop'ta Authorization başlığının değeri Bearer ve bir boşluktan sonra token olmalı; taşıyıcı token için ortam değişkeni alanı boş kalmalı. Web oturum çerezi MCP doğrulaması yerine geçmez.",
		"Araçlar görünmüyorsa istemciyi tamamen yeniden başlatıp yeni sohbet açın. Metadata değiştiğinde bağlantıyı yenileyin. İlk çağrı get_agent_guide olmalı.",
		"Bu rehber Codex Desktop, Codex CLI ve Claude Code içindir. Claude Desktop için istemcinin desteklediği HTTP MCP bağlantısını aynı adresle yapılandırmak gerekir; bu rehberde doğrulanmış ayrı kurulum adımı yoktur. ChatGPT bağlantısı sonraki aşamadır; HTTPS veya desteklenen tünel ve uygun kimlik doğrulama ayrıca hazırlanmalıdır.",
	}, "", "", "")
	troubleshooting.Kind = UsageDetails
	return []UsageBlock{
		namedSection("prepare", "1. Workey'i hazırlayın",
			paragraph("Proje dizininde .env.example dosyasından .env hazırlayın ve gerekli değerleri doldurun. Mevcut .env dosyanız varsa üzerine yazmayın. Giriş parolası hash'i ve oturum anahtarı dahil temel kurulum adımları projenin README.md dosyasındadır."),
			apiTokenSetup(),
			paragraph("Varsayılan bağlantı http://127.0.0.1:8080/mcp. WORKEY_ADDR değiştiyse aşağıdaki bağlantı komutlarında ve config.toml örneğinde 8080 portunu kendi portunuzla değiştirin. Yerel HTTP için WORKEY_COOKIE_SECURE=false kullanın. Sunucuyu başlatın ve çalışır durumda bırakın."),
			code("go run ./cmd/workey serve", "powershell")),
		namedSection("connect", "2. İstemcinizi bağlayın", desktop, cli, claude),
		namedSection("verify", "3. Bağlantıyı doğrulayın", paragraph("Yeni sohbette bu istemi gönderin. Beklenen sonuç: asistan get_agent_guide çağrısıyla sunucu tarihini ve kuralları okur. Bu sayfa bağlantıyı sınamaz; sonucu asistanınızın yanıtında kontrol edin."),
			UsageBlock{Kind: UsagePrompt, ID: "connection", Text: "Workey MCP bağlantısını kullanarak ajan rehberini oku ve bağlantıyı doğrula"}),
		namedSection("create-plan", "4. Sınava kadar çalışma planımı oluştur",
			paragraph("Sınav tarihi: "+guide.ExamDate+". İstemi kopyalayın; hedefinizi, uygun saatlerinizi, sürelerinizi, dinlenme günlerinizi ve kitap referanslarınızı doldurup Workey'e bağlı asistanınıza gönderin. Çalışma ritmi düzenlenebilir bir başlangıç önerisidir. Bu düğme yalnızca metni kopyalar; sayfa plan üretmez veya MCP çağrısı yapmaz."),
			paragraph("Dikkat: save_plan, seçilen aralıktaki başlanmamış açık seansları tam listeyle değiştirir. Listede bulunmayan başlanmamış açık seanslar silinir; korunacak açık seanslar da her haftanın isteğine dahil edilmelidir. Geçmiş, tamamlanmış ve başlanmış kayıtlar korunur; başlanmış seansın değişikliği skipped reason=started döner."),
			UsageBlock{Kind: UsagePrompt, ID: "exam-plan", Text: examPlanPrompt},
		),
		alternative, troubleshooting,
	}
}

const examPlanPrompt = `Workey MCP bağlantısını kullanarak bugünden sınava kadar çalışma planımı hazırla ve kaydet.

Dolduracağım bilgiler:
- Hedefim ve mevcut seviyem: [hedef, güçlü/zayıf konular, bitirdiğim bölümler]
- Uygun gün ve saatlerim: [günler ve saat aralıkları; başlangıç önerisi: hafta içi seçtiğim dört akşam 20:30–22:00 ve seçtiğim bir hafta sonu günü üç saat]
- Seans sürelerim: [başlangıç önerisi: dört akşam × 90 dakika, bir hafta sonu × 180 dakika; bana göre düzenle]
- Dinlenme günlerim: [çalışmayacağım günler ve diğer kısıtlar]
- Kitap/bölüm/sayfa referanslarım: [her kitap için kitap kodu, bölüm numarası, bölüm adı, sayfa başlangıcı ve sonu; varsa konu/soru/antrenman aralıkları; doğrulanabilir içindekiler veya kaynak ekle]

Şu sırayla ilerle:
1. Önce get_agent_guide çağır. Güncel today, timezone ve exam_date değerlerini yeniden oku; dönem ve saatleri bunlara göre belirle. Sabit örnek tarihleri kullanma. Sınav geçmişse yazma, kullanıcıyla hedefi netleştir.
2. get_outline → ilgili get_plan → get_progress sırasıyla mevcut fazları, konu gruplarını, bugünden sınava kadar planı ve ilerlemeyi oku. Aralık büyükse get_plan okumalarını haftalara böl. Mevcut bağlantıları anlamak için gerekirse geçmişteki ilgili seansları da oku.
3. Hedefimi, seviyemi, gün/saatleri, süreleri, dinlenme günlerini ve referanslarımı benimle netleştir. Doldurulmamış alanlarla yazmaya başlama. Eksik kitap içeriği, bölüm veya sayfa referansı uydurma; eksikleri sor ve yanıtımı bekle. Çalışma ritmini otomatik kabul etme. Mevcut seanslarla çakışmayı ve yetişmeyecek kapsamı bildir; kararı benimle netleştir.
4. Bugünden sınava kadar faz, konu grubu ve seansları doğrulanmış referanslarla hazırla. Gerekliyse save_outline çağrısına korunacak mevcut faz ve grupları da içeren TAM listeyi gönder; seansların kullandığı grupları silme. Outline sonuçlarındaki created/updated/deleted/unchanged listelerini incele ve get_outline ile doğrula.
5. Planı Pazartesi–Pazar haftalarına göre haftalık, çakışmayan from–to dahil aralıklarla save_plan üzerinden yaz. İlk aralığın başlangıcı today, son aralığın bitişi exam_date olsun; kısmi ilk/son haftayı bu sınırlara kırp. Bugünden önceye veya sınavdan sonraya seans yazma.
6. HER haftayı yazmadan hemen önce o aralığı get_plan ile yeniden oku. study ve yer imini incele. save_plan başlanmamış açık seansların TAM listesini değiştirir: korunacak açık seansları tüm items referanslarıyla isteğe dahil et; listeden çıkarılan başlanmamış açık seanslar silinir. Değiştirme veya silme kararını kullanıcıyla netleştir. Geçmiş, tamamlanmış ve başlanmış kayıtları koru; başlanmış seansın kaynak, tarih, saat ve süresini değiştirme. Başlanmış seans aynı içerikle unchanged, değişiklikte skipped reason=started döner, eksikse korunur. Tamamlanmış seansın date + start_time konumuna yeni seans koyma.
7. Okumadaki topic_group_key değerini yazmada topic_group alanına taşı; id, state, completed, notation ve study okuma alanlarını gönderme. Telafi bağlantılarını koru: gerekirse kaynak get_plan sonucundaki moved_to ile makeup_of değerini doğrula; bağlantı belirsizse sor ve yazmayı durdur.
8. Her save_plan yanıtındaki created/updated/deleted/unchanged/skipped listelerini incele. skipped kayıtlarını başarı diye sunma; reason değerini açıkla. Her yazmadan sonra ilgili get_plan ve get_progress çağrılarıyla tarih, saat, süre, referans, korunacak kayıtlar ve değişiklikleri doğrula. Planlananın uygulanmasını engelleyen skipped sonucunda da dur ve kalan haftalara geçme.
9. Başarısız hafta, outline hatası veya tekrar okuma/doğrulama hatasında DUR. Sonraki haftaları yazma. Tamamlanan ve kalan aralıkları, hatayı ve gerekli kullanıcı bilgisini bildir. Belirsiz yazma sonucunda önce mevcut durumu oku; körlemesine tekrar yazma. Tekrar denemede aynı gövde ve idempotency_key kullan; değişen içerik için yeni anahtar seç.
10. Sonucu Türkçe özetle: güncel dönem ve saat dilimi, hazırlanan seanslar, created/updated/deleted/unchanged/skipped sayıları ve ilgili kayıtlar, silinen/atlanan kayıtların nedenleri, doğrulanan haftalar ve kalan iş. Kısmi çalışmayı bütünü tamamlandı diye bildirme. Bu planlama işlemi sırasında complete_session çağırma; seansları tamamlanmış işaretleme.`

func withSectionID(id string, block UsageBlock) UsageBlock {
	block.ID = id
	return block
}

func apiSections() []UsageBlock {
	examples := apiExamples()
	first := examples[0]
	first.ID, first.Text = "first-request", "İlk API isteğiniz"
	first.Children = append([]UsageBlock{
		paragraph("Önce Workey sunucusunu başlatın. Taban yol varsayılan olarak http://127.0.0.1:8080/api/v1; WORKEY_ADDR değiştiyse adresi uyarlayın. Tüm endpoint’ler Authorization: Bearer token ister; web oturumu token yerine geçmez."),
		apiTokenSetup(),
		paragraph("Workey’in .env dosyasındaki WORKEY_API_TOKEN değerini PowerShell 7 komutunun gizli girişine girin. Tokenı kod örneğine veya URL’ye yazmayın."),
	}, first.Children...)
	write := namedSection("read-write", "Planı okuyun ve yazın",
		paragraph("Önce ajan rehberi → mevcut outline → ilgili from–to aralığının planı → ilerleme sırasıyla okuyun. Korunacak faz, grup ve açık seansları tüm kitap referanslarıyla yazma gövdesine dahil edin."),
		examples[1], examples[2],
		paragraph("Dikkat: outline faz ve grupların, plan ise seçilen aralıktaki başlanmamış açık seansların tam listesini değiştirir. Listede bulunmayan başlanmamış açık seanslar silinir; korunacak kayıtları dahil edin. Geçmiş, tamamlanmış ve başlanmış kayıtlar korunur; başlanmış seansın değişikliği skipped reason=started döner."),
		examples[3],
		paragraph("Yazmadan sonra get_plan eşdeğeri GET /api/v1/plan ile yazdığınız from–to aralığının tamamını ve GET /api/v1/progress ile ilerlemeyi yeniden okuyun. Yukarıdaki komut yalnız bugünü okur; plan.json farklı tarihler içeriyorsa sorguyu o aralığa uyarlayın. created/updated/deleted/unchanged/skipped listelerini inceleyin; skipped sonucunu başarı diye sunmayın."),
	)
	responses := namedSection("examples", "Temsili yanıtlar", examples[4], examples[5])
	responses.Kind = UsageDetails
	complete := withSectionID("completion", examples[6])
	complete.Text = "Tamamlama ve geri alma"
	curl := withSectionID("curl", examples[7])
	curl.Text, curl.Kind = "curl.exe alternatifi", UsageDetails
	errors := withSectionID("errors", examples[8])
	errors.Text = "Hatalar ve tekrar deneme"
	return []UsageBlock{first, write, responses, complete, curl, errors}
}

func apiExamples() []UsageBlock {
	return []UsageBlock{
		usageSection("1. PowerShell ile okuma", []string{"PowerShell 7 terminalinde sunucuyla aynı tokenı girin. Gerçek token örneklerde bulunmaz."},
			`$env:WORKEY_API_TOKEN = Read-Host 'Workey API tokenı' -MaskInput
$base = 'http://127.0.0.1:8080/api/v1'
$headers = @{ Authorization = "Bearer $env:WORKEY_API_TOKEN" }
$guide = Invoke-RestMethod "$base/agent-guide" -Headers $headers
Invoke-RestMethod "$base/outline" -Headers $headers
Invoke-RestMethod "$base/plan?from=$($guide.today)&to=$($guide.today)" -Headers $headers
Invoke-RestMethod "$base/progress" -Headers $headers`, "powershell", ""),
		usageSection("2. Faz isteği — temsili veri", []string{"PUT /api/v1/outline. Aşağıdaki tarihler ve sayfalar yalnızca sözleşme örneğidir. Gerçek planı göndermeden önce mevcut kayıtları okuyun ve korunacak kayıtları dahil edin. Yanıt, phases ve topic_groups için created/updated/deleted/unchanged anahtar listelerini içerir."}, apidoc.ExampleOutline, "json", ""),
		usageSection("3. Plan isteği — temsili veri", []string{
			"PUT /api/v1/plan. from/to dahil aralık; sessions tam listedir. Her seans date, start_time, pozitif duration_min, kind ve en az bir items kaydı taşır. topic_group ve makeup_of isteğe bağlıdır.",
			"Her item: book, pozitif chapter_no, chapter_title, pozitif page_from/page_to. Bitiş başlangıçtan küçük olamaz. İsteğe bağlı konu_from/to, soru_from/to, antrenman_from/to çift halinde gönderilir. Bu isteğe bağlı aralıkların mevcut doğrulaması yalnızca çift bulunmasını ve sıralamayı denetler.",
			"Sabit tarihli örneği bugünün gerçek planı olarak göndermeyin. Geçmiş tarihler skipped döner. Tüm alanların şemaları indirilebilir OpenAPI belgesindedir.",
		}, apidoc.ExamplePlan, "json", ""),
		usageSection("4. Yazma ve tekrar deneme", []string{
			"Önce yukarıdaki gövdeleri gerçek referanslarınız ve tarihlerinizle outline.json ve plan.json dosyalarına hazırlayın. Dosyalar UTF-8 olmalı. Yeni içerik için yeni idempotency anahtarı seçin. Aşağıdaki plan-v1 anahtarı temsili bir değerdir.",
		}, `Invoke-RestMethod "$base/outline" -Method Put -Headers $headers -ContentType 'application/json; charset=utf-8' -Body ([IO.File]::ReadAllBytes('outline.json'))
Invoke-RestMethod "$base/plan" -Method Put -Headers ($headers + @{ 'Idempotency-Key' = 'plan-v1' }) -ContentType 'application/json; charset=utf-8' -Body ([IO.File]::ReadAllBytes('plan.json'))
Invoke-RestMethod "$base/plan?from=$($guide.today)&to=$($guide.today)" -Headers $headers`, "powershell", ""),
		usageSection("5. Plan yazma yanıtı — temsili", []string{"Boş sonuç listeleri mevcut API'de null olabilir. skipped içindeki reason: past_date veya completed. Aynı gövde anahtarsız tekrarlandığında unchanged; aynı anahtarla tekrarlandığında ilk yanıt döner."},
			exampleResponse("PlanResult"), "json", ""),
		usageSection("6. Plan okuma yanıtı — temsili", []string{"GET /api/v1/plan?from=2026-10-26&to=2026-10-26. Bugünün 26 Ekim 2026 olduğu örnek senaryodur. Tamamlama/geri alma yanıtı aynı session nesnesidir; completed ve state güncellenir."},
			exampleResponse("Plan"), "json", ""),
		usageSection("7. Tamamlama ve geri alma", []string{"Kimliği plan yanıtından alın. Örnek 1 yerine seçilen gerçek seans kimliğini kullanın. Yalnızca kullanıcı talebiyle çalıştırın; gelecek seanslarda 422 döner. Gövde gönderilmez."},
			`$sessionId = 1
Invoke-RestMethod "$base/sessions/$sessionId/complete" -Method Post -Headers $headers
Invoke-RestMethod "$base/sessions/$sessionId/complete" -Method Delete -Headers $headers`, "powershell", ""),
		usageSection("8. curl.exe ile aynı işlemler", []string{"PowerShell'de curl alias'ı yerine curl.exe kullanın. Önce tokenı yukarıdaki gibi ortamda tanımlayın. Tarihleri rehberdeki today ve istediğiniz aralığa göre uyarlayın."},
			`curl.exe -H "Authorization: Bearer $env:WORKEY_API_TOKEN" "$base/agent-guide"
curl.exe -H "Authorization: Bearer $env:WORKEY_API_TOKEN" "$base/plan?from=$($guide.today)&to=$($guide.today)"
curl.exe -X PUT -H "Authorization: Bearer $env:WORKEY_API_TOKEN" -H "Content-Type: application/json" -H "Idempotency-Key: plan-v1" --data-binary '@plan.json' "$base/plan"
curl.exe -X POST -H "Authorization: Bearer $env:WORKEY_API_TOKEN" "$base/sessions/$sessionId/complete"
curl.exe -X DELETE -H "Authorization: Bearer $env:WORKEY_API_TOKEN" "$base/sessions/$sessionId/complete"`, "powershell", ""),
		usageSection("Hatalar", []string{"400: bozuk JSON veya gövde sınırı. 401: eksik/geçersiz token. 404: kayıt yok. 409: idempotency_conflict. 422: validation_failed. 500: internal. fields içindeki path/code/message alanlarına göre isteği düzeltin. Sunucu hatasında önce mevcut durumu okuyun; aynı yazmayı tekrarlarken aynı gövde ve anahtarı koruyun."},
			`{"error":{"code":"validation_failed","message":"the request did not pass validation","fields":[{"path":"sessions[0].date","code":"invalid","message":"date must be a date in YYYY-MM-DD form"}]}}`, "json", ""),
	}
}
