package views

import "github.com/ibra2626/workey/internal/apidoc"

type usageSectionData struct {
	Title      string
	Paragraphs []string
	Code       string
}

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

var mcpToolDescriptions = []struct{ Name, Description string }{
	{"get_agent_guide", "İlk çağrı: sunucu tarihi, desteklenen değerler, kurallar ve işlem sırası."},
	{"get_outline", "Fazları ve konu gruplarını oku."},
	{"save_outline", "Fazların ve grupların tam listesini değiştir."},
	{"get_plan", "from–to dahil aralıktaki seansları ve kimliklerini oku."},
	{"save_plan", "Aralıktaki açık seansların tam listesini değiştir; eksik seanslar silinir."},
	{"complete_session", "session_id ile tamamla; done=false ile geri al. Yalnızca kullanıcı isteğiyle."},
	{"get_progress", "Seans, dakika, konu grubu ve 38 haftalık ilerlemeyi oku."},
}

func mcpSections() []usageSectionData {
	return []usageSectionData{
		{"1. Workey sunucusunu başlatın", []string{
			"Proje dizininde .env.example dosyasından .env hazırlayın ve gerekli değerleri doldurun. Mevcut .env dosyanız varsa üzerine yazmayın. Gerçek tokenları paylaşmayın veya Git'e eklemeyin.",
			"Varsayılan bağlantı http://127.0.0.1:8080/mcp; WORKEY_ADDR değiştiyse örnek adresleri uyarlayın. Yerel HTTP için WORKEY_COOKIE_SECURE=false kullanın. Sunucu çalışır durumda kalmalı.",
		}, "go run ./cmd/workey serve"},
		{"2. İstemci ortamını hazırlayın", []string{
			"Workey'in .env dosyasını okuması, ayrı başlatılan Claude veya Codex sürecine ortam değişkeni aktarmaz. Aynı WORKEY_API_TOKEN değerini istemciyi başlatacağınız PowerShell 7 terminalinde de tanımlayın.",
		}, `$env:WORKEY_API_TOKEN = Read-Host 'Workey API tokenı' -MaskInput`},
		{"3. Claude Code", []string{"Proje dizininde çalıştırın. Bu komut tokenı yerel Claude yapılandırmasına yazar; dosyayı Git'e eklemeyin. Claude Desktop için istemcinin desteklediği HTTP MCP bağlantısını aynı adresle yapılandırın; aşağıdaki komut Claude Code içindir."},
			`claude mcp add --scope local --transport http workey http://127.0.0.1:8080/mcp --header "Authorization: Bearer $env:WORKEY_API_TOKEN"
claude mcp get workey`},
		{"4. Codex CLI", []string{"Tokenı ortamdan alan HTTP bağlantısını ekleyin. Codex'i aynı terminalden başlatın; /mcp ile bağlantıyı kontrol edin."},
			`codex mcp add workey --url http://127.0.0.1:8080/mcp --bearer-token-env-var WORKEY_API_TOKEN
codex mcp list
codex`},
		{"5. Codex Desktop", []string{
			"Codex Desktop aynı kullanıcı yapılandırmasındaki MCP kaydını kullanır. CLI ile ekledikten sonra uygulamayı tamamen kapatıp token ortam değişkenini miras alacağı terminalden yeniden başlatın. Açık bir uygulamaya terminalde sonradan tanımlanan değişken aktarılmaz.",
			"MCP ayarlarından workey sunucusunu kontrol edin ve yeni sohbette önce get_agent_guide çağrılmasını isteyin. Başlatıcı veya uygulama sürümüne göre ortam aktarımı farklılaşırsa o süreç için WORKEY_API_TOKEN tanımlayın; tokenı URL'ye veya config.toml içine yazmayın.",
		}, ""},
		{"Alternatif: config.toml", []string{"CLI komutu yerine kullanıcı dizininizdeki .codex/config.toml dosyasına aşağıdaki kaydı ekleyebilirsiniz. workey kaydı varsa ikinci kez eklemeyin. CLI ve Desktop aynı endpoint ve araçları kullanır."},
			`[mcp_servers.workey]
url = "http://127.0.0.1:8080/mcp"
bearer_token_env_var = "WORKEY_API_TOKEN"`},
		{"Bağlantı sorunları ve kapsam", []string{
			"Connection refused: Workey'in çalıştığını ve portu kontrol edin. 401: tokenın hem Workey hem istemci sürecinde aynı olduğunu kontrol edin. Web oturum çerezi MCP doğrulaması yerine geçmez.",
			"Araçlar görünmüyorsa istemciyi yeniden başlatıp yeni sohbet açın. Metadata değiştiğinde bağlantıyı yenileyin. İlk çağrı get_agent_guide olmalı.",
			"Bu sürüm Claude ve Codex Desktop/CLI içindir. ChatGPT bağlantısı sonraki aşamadır; HTTPS veya desteklenen tünel ve uygun kimlik doğrulama ayrıca hazırlanmalıdır.",
		}, ""},
	}
}

func apiSections() []usageSectionData {
	return []usageSectionData{
		{"1. PowerShell ile okuma", []string{"PowerShell 7 terminalinde sunucuyla aynı tokenı girin. Gerçek token örneklerde bulunmaz."},
			`$env:WORKEY_API_TOKEN = Read-Host 'Workey API tokenı' -MaskInput
$base = 'http://127.0.0.1:8080/api/v1'
$headers = @{ Authorization = "Bearer $env:WORKEY_API_TOKEN" }
$guide = Invoke-RestMethod "$base/agent-guide" -Headers $headers
Invoke-RestMethod "$base/outline" -Headers $headers
Invoke-RestMethod "$base/plan?from=$($guide.today)&to=$($guide.today)" -Headers $headers
Invoke-RestMethod "$base/progress" -Headers $headers`},
		{"2. Faz isteği — temsili veri", []string{"PUT /api/v1/outline. Aşağıdaki tarihler ve sayfalar yalnızca sözleşme örneğidir. Gerçek planı göndermeden önce mevcut kayıtları okuyun ve korunacak kayıtları dahil edin. Yanıt, phases ve topic_groups için created/updated/deleted/unchanged anahtar listelerini içerir."}, apidoc.ExampleOutline},
		{"3. Plan isteği — temsili veri", []string{
			"PUT /api/v1/plan. from/to dahil aralık; sessions tam listedir. Her seans date, start_time, pozitif duration_min, kind ve en az bir items kaydı taşır. topic_group ve makeup_of isteğe bağlıdır.",
			"Her item: book, pozitif chapter_no, chapter_title, pozitif page_from/page_to. Bitiş başlangıçtan küçük olamaz. İsteğe bağlı konu_from/to, soru_from/to, antrenman_from/to çift halinde gönderilir. Bu isteğe bağlı aralıkların mevcut doğrulaması yalnızca çift bulunmasını ve sıralamayı denetler.",
			"Sabit tarihli örneği bugünün gerçek planı olarak göndermeyin. Geçmiş tarihler skipped döner. Tüm alanların şemaları indirilebilir OpenAPI belgesindedir.",
		}, apidoc.ExamplePlan},
		{"4. Yazma ve tekrar deneme", []string{
			"Önce yukarıdaki gövdeleri gerçek referanslarınız ve tarihlerinizle outline.json ve plan.json dosyalarına hazırlayın. Dosyalar UTF-8 olmalı. Yeni içerik için yeni idempotency anahtarı seçin. Aşağıdaki plan-v1 anahtarı temsili bir değerdir.",
		}, `Invoke-RestMethod "$base/outline" -Method Put -Headers $headers -ContentType 'application/json; charset=utf-8' -Body ([IO.File]::ReadAllBytes('outline.json'))
Invoke-RestMethod "$base/plan" -Method Put -Headers ($headers + @{ 'Idempotency-Key' = 'plan-v1' }) -ContentType 'application/json; charset=utf-8' -Body ([IO.File]::ReadAllBytes('plan.json'))
Invoke-RestMethod "$base/plan?from=$($guide.today)&to=$($guide.today)" -Headers $headers`},
		{"5. Plan yazma yanıtı — temsili", []string{"Boş sonuç listeleri mevcut API'de null olabilir. skipped içindeki reason: past_date veya completed. Aynı gövde anahtarsız tekrarlandığında unchanged; aynı anahtarla tekrarlandığında ilk yanıt döner."},
			`{"created":[{"id":1,"date":"2026-10-26","start_time":"20:30"}],"updated":null,"deleted":null,"unchanged":null,"skipped":null}`},
		{"6. Plan okuma yanıtı — temsili", []string{"GET /api/v1/plan?from=2026-10-26&to=2026-10-26. Bugünün 26 Ekim 2026 olduğu örnek senaryodur. Tamamlama/geri alma yanıtı aynı session nesnesidir; completed ve state güncellenir."},
			`{"sessions":[{"id":1,"date":"2026-10-26","start_time":"20:30","duration_min":90,"kind":"aksam","completed":false,"state":"today","topic_group_key":"temel","topic_group_name":"Temel kavramlar","items":[{"book":"M1","chapter_no":1,"chapter_title":"Temel kavramlar","page_from":1,"page_to":10}],"notation":"M1·1, s. 1–10"}]}`},
		{"7. Tamamlama ve geri alma", []string{"Kimliği plan yanıtından alın. Örnek 1 yerine seçilen gerçek seans kimliğini kullanın. Yalnızca kullanıcı talebiyle çalıştırın; gelecek seanslarda 422 döner. Gövde gönderilmez."},
			`$sessionId = 1
Invoke-RestMethod "$base/sessions/$sessionId/complete" -Method Post -Headers $headers
Invoke-RestMethod "$base/sessions/$sessionId/complete" -Method Delete -Headers $headers`},
		{"8. curl.exe ile aynı işlemler", []string{"PowerShell'de curl alias'ı yerine curl.exe kullanın. Önce tokenı yukarıdaki gibi ortamda tanımlayın. Tarihleri rehberdeki today ve istediğiniz aralığa göre uyarlayın."},
			`curl.exe -H "Authorization: Bearer $env:WORKEY_API_TOKEN" "$base/agent-guide"
curl.exe -H "Authorization: Bearer $env:WORKEY_API_TOKEN" "$base/plan?from=$($guide.today)&to=$($guide.today)"
curl.exe -X PUT -H "Authorization: Bearer $env:WORKEY_API_TOKEN" -H "Content-Type: application/json" -H "Idempotency-Key: plan-v1" --data-binary '@plan.json' "$base/plan"
curl.exe -X POST -H "Authorization: Bearer $env:WORKEY_API_TOKEN" "$base/sessions/$sessionId/complete"
curl.exe -X DELETE -H "Authorization: Bearer $env:WORKEY_API_TOKEN" "$base/sessions/$sessionId/complete"`},
		{"Hatalar", []string{"400: bozuk JSON veya gövde sınırı. 401: eksik/geçersiz token. 404: kayıt yok. 409: idempotency_conflict. 422: validation_failed. 500: internal. fields içindeki path/code/message alanlarına göre isteği düzeltin. Sunucu hatasında önce mevcut durumu okuyun; aynı yazmayı tekrarlarken aynı gövde ve anahtarı koruyun."},
			`{"error":{"code":"validation_failed","message":"the request did not pass validation","fields":[{"path":"sessions[0].date","code":"invalid","message":"date must be a date in YYYY-MM-DD form"}]}}`},
	}
}
