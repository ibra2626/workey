package views

import (
	"encoding/json"
	"strings"

	"github.com/ibra2626/workey/internal/apidoc"
	"github.com/ibra2626/workey/internal/plan"
)

func studyContractSection() UsageBlock {
	paused, _ := json.MarshalIndent(apidoc.StudyResponseExample(), "", "  ")
	skipped, _ := json.MarshalIndent(apidoc.StartedResultExample(), "", "  ")
	return namedSection("study-contract", "Çalışma durumu ve yer imini okuma",
		paragraph(plan.StudyReadRules()),
		paragraph("26 Ekim 2026 tarihli temsili okuma → yazma → tekrar okuma örneği; gerçek kullanıcı planı değildir. Önce ajan rehberini, outline, ilgili get_plan / GET /api/v1/plan aralığını ve ilerlemeyi okuyun. Aşağıda yarım kalmış örnek seansın okuma yanıtı var."),
		code(string(paused), "json"),
		paragraph("Plan yazarken study ve diğer okuma alanlarını göndermeyin. Başlanmış seansın items, date, start_time ve duration_min değerlerini koruyun: aynı içerik unchanged, değiştirme denemesi skipped reason=started, istekten eksikse korunur. past_date ve completed nedenleri önceliklidir. Örnekte duration_min=100 yapmayı deneyen eski bir istemci aşağıdaki sonucu alır; seans 90 dakika ve yer imi aynı kalır."),
		code(string(skipped), "json"),
		paragraph("Bu sonucu başarıyla uygulanmış değişiklik saymayın. Yazdıktan sonra aynı aralığı ve ilerlemeyi tekrar okuyun; seansın süre, kaynak ve yer iminin korunduğunu doğrulayıp atlanan değişikliği kullanıcıya açıklayın. Telafi gerekiyorsa kaynak geçmiş seansı değiştirmeden makeup_of ile ayrı seans hazırlayın; yer imi otomatik taşınmaz."),
		paragraph("Mevcut complete_session ve REST tamamlama/geri alma çağrıları ek alan istemez: tamamla mevcut study durumunu atomik closed yapar; geri al yer imini koruyup paused yapar. Eski seans için çalışma kaydı oluşturmaz. Bu REST/MCP yazmaları study.version denetlemez; eski sekme niyeti tespit edilemez."),
	)
}

// StudyContractMarkdown is checked verbatim in docs/api.md and docs/mcp.md.
func StudyContractMarkdown() string {
	var out strings.Builder
	writeMarkdown(&out, []UsageBlock{studyContractSection()}, 2)
	return out.String()
}
