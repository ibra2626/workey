package views

import "github.com/ibra2626/workey/internal/apidoc"

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

func mcpSections() []UsageBlock {
	const tokenInput = `$env:WORKEY_API_TOKEN = Read-Host 'Workey API tokenı' -MaskInput`
	const codexAdd = `codex mcp add workey --url http://127.0.0.1:8080/mcp --bearer-token-env-var WORKEY_API_TOKEN
codex mcp list`
	tokenHelp := paragraph("Workey'in .env dosyasındaki WORKEY_API_TOKEN değerini alın; sunucu bir başkası tarafından kurulduysa bu değeri kuran kişiden güvenli yoldan isteyin. Aşağıdaki PowerShell 7 komutuna aynı tokenı girin. Workey'in .env okuması, ayrı başlatılan istemciye ortam değişkeni aktarmaz. Tokenı URL'ye veya config.toml içine yazmayın.")
	desktop := section("Codex Desktop",
		paragraph("Codex CLI kurulu olmalı. Codex Desktop'ı tamamen kapatın. PowerShell 7 terminalinde tokenı tanımlayıp bağlantıyı ekleyin; workey kaydı zaten varsa tekrar eklemeyin."),
		tokenHelp, code(tokenInput+"\n"+codexAdd, "powershell"),
		paragraph("Codex Desktop aynı kullanıcı yapılandırmasındaki MCP kaydını kullanır. Uygulamayı tokenı tanımladığınız terminalden kendi kurulumunuzun başlatıcısıyla açın; uygulama süreci WORKEY_API_TOKEN değişkenini miras almalı. Açık uygulamaya sonradan tanımlanan değişken aktarılmaz. Başlatıcıya göre ortam aktarımı farklılaşırsa uygulama sürecinin ortamında bu değişkeni tanımlayın ve tamamen yeniden başlatın."),
		paragraph("MCP ayarlarında workey sunucusunu kontrol edin ve yeni sohbet açın."),
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
	alternative := usageSection("Alternatif: config.toml", []string{"CLI ile ekleme yerine kullanıcı dizininizdeki .codex/config.toml dosyasına aşağıdaki kaydı ekleyebilirsiniz. workey kaydı varsa ikinci kez eklemeyin. Token ortam aktarımı ve Desktop yeniden başlatma adımları yine gereklidir."},
		`[mcp_servers.workey]
url = "http://127.0.0.1:8080/mcp"
bearer_token_env_var = "WORKEY_API_TOKEN"`, "toml", "")
	alternative.Kind = UsageDetails
	troubleshooting := usageSection("Bağlantı sorunları ve kapsam", []string{
		"Connection refused: Workey'in çalıştığını ve portu kontrol edin. 401: tokenın hem Workey hem istemci sürecinde aynı olduğunu kontrol edin. Web oturum çerezi MCP doğrulaması yerine geçmez.",
		"Araçlar görünmüyorsa istemciyi tamamen yeniden başlatıp yeni sohbet açın. Metadata değiştiğinde bağlantıyı yenileyin. İlk çağrı get_agent_guide olmalı.",
		"Bu rehber Codex Desktop, Codex CLI ve Claude Code içindir. Claude Desktop için istemcinin desteklediği HTTP MCP bağlantısını aynı adresle yapılandırmak gerekir; bu rehberde doğrulanmış ayrı kurulum adımı yoktur. ChatGPT bağlantısı sonraki aşamadır; HTTPS veya desteklenen tünel ve uygun kimlik doğrulama ayrıca hazırlanmalıdır.",
	}, "", "", "")
	troubleshooting.Kind = UsageDetails
	return []UsageBlock{
		usageSection("1. Workey'i hazırlayın", []string{
			"Proje dizininde .env.example dosyasından .env hazırlayın ve gerekli değerleri doldurun. Mevcut .env dosyanız varsa üzerine yazmayın. Gerçek tokenları paylaşmayın veya Git'e eklemeyin.",
			"Varsayılan bağlantı http://127.0.0.1:8080/mcp. WORKEY_ADDR değiştiyse aşağıdaki bağlantı komutlarında ve config.toml örneğinde 8080 portunu kendi portunuzla değiştirin. Yerel HTTP için WORKEY_COOKIE_SECURE=false kullanın. Sunucuyu başlatın ve çalışır durumda bırakın.",
		}, "go run ./cmd/workey serve", "powershell", ""),
		section("2. İstemcinizi bağlayın", desktop, cli, claude),
		section("3. Bağlantıyı deneyin", paragraph("Yeni sohbette bu istemi gönderin. Asistanın get_agent_guide çağrısıyla sunucu tarihini ve kuralları okuyabildiğini doğrulayın."),
			UsageBlock{Kind: UsagePrompt, ID: "connection", Text: "Workey MCP bağlantısını kullanarak ajan rehberini oku ve bağlantıyı doğrula"}),
		section("4. Planınızı oluşturun", paragraph("Bağlantı hazır. Sınava kadar çalışma planı oluşturma istemi bir sonraki adımda eklenecek.")),
		alternative, troubleshooting,
	}
}

func apiSections() []UsageBlock {
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
