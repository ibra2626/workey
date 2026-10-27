package plan

import "github.com/ibra2626/workey/internal/clock"

// AgentGuide is shared by MCP, REST and the human-facing usage pages.
type AgentGuide struct {
	Today        string   `json:"today"`
	Timezone     string   `json:"timezone"`
	ExamDate     string   `json:"exam_date"`
	Books        []string `json:"books"`
	SessionKinds []string `json:"session_kinds"`
	Rules        []string `json:"rules"`
	Workflow     []string `json:"workflow"`
}

// AgentGuide uses the same injectable clock as plan writes, including at midnight.
func (s *Service) AgentGuide(examDate clock.Date) AgentGuide {
	return AgentGuide{
		Today: clock.Today(s.clk).String(), Timezone: "Europe/Istanbul", ExamDate: examDate.String(),
		Books:        []string{"M1", "M2", "M3", "G1", "G2"},
		SessionKinds: []string{KindAksam, KindUzunBlok, KindDeneme},
		Rules:        GuideRules(), Workflow: GuideWorkflow(),
	}
}

func GuideRules() []string {
	return []string{
		"Workey çalışma içeriği üretmez; yalnızca kullanıcının sağladığı kitap, bölüm ve sayfa referanslarını kaydeder. Eksik referansları kullanıcıdan iste; uydurma.",
		"Tarihler YYYY-MM-DD, saatler HH:MM biçiminde ve Europe/Istanbul saat dilimindedir. Bugünü rehberin today alanından al; örneklerdeki sabit tarihleri gerçek plan olarak kullanma.",
		"save_plan / PUT /api/v1/plan, from–to dahil aralıktaki bugün ve geleceğe ait açık seansların TAM listesini değiştirir. İstekten çıkarılan açık seanslar silinir; korunacak seansları da gönder. Eşleştirme date + start_time ile yapılır.",
		"Geçmiş tarihler ve tamamlanmış seanslar plan yazımında korunur. skipped kayıtları başarıyla değişmiş sayma; reason alanını kullanıcıya açıkla.",
		"save_outline / PUT /api/v1/outline faz ve konu gruplarının TAM listesini key üzerinden değiştirir. Korunacak kayıtları da gönder. Seansların kullandığı grup silinemez.",
		"get_plan yanıtındaki topic_group_key yazarken topic_group alanına taşınır. id, state, completed ve notation okuma alanlarıdır. items referanslarını koru.",
		"Telafi için makeup_of = {date, start_time} kullan; hedef geçmiş ve tamamlanmamış bir seans olmalı. Kaynak geçmiş seansı değiştirme. Okuma yanıtı makeup_of taşımaz; moved_to bilgisi kaynak seansta bulunur. Telafi seansını yeniden yazarken kaynak bağlantısını get_plan sonuçlarından doğrula; belirsizse kullanıcıdan sor.",
		"Aynı gövdeyi tekrar göndermek kayıt veya audit çoğaltmaz. idempotency_key ilk sonucu tekrar döndürür; aynı anahtar farklı içerikle kullanılamaz. REST'te boş olmayan Idempotency-Key başlığı gövdedeki anahtardan önceliklidir. MCP ve REST arasında aynı anahtarı paylaşma.",
		"complete_session ve REST tamamlama/geri alma işlemlerini yalnızca kullanıcı istediğinde yap. Kimliği get_plan sonucundan al. Gelecek seanslar tamamlanamaz veya geri alınamaz.",
		"validation_failed alan hatalarını düzeltmeden isteği tekrarlama. idempotency_conflict durumunda önceki sonucu oku ve değişmiş istek için yeni anahtar kullan. Yetkilendirme hatasında tokenı veya gizli bilgileri yanıta yazma.",
	}
}

func GuideWorkflow() []string {
	return []string{
		"Önce get_agent_guide / GET /api/v1/agent-guide, get_outline, ilgili aralık için get_plan ve get_progress çağır.",
		"Kullanıcının hedefini ve kitap/bölüm/sayfa referanslarını doğrula; eksikleri sor.",
		"İstenen değişikliği hazırla; korunacak açık seansları ve mevcut faz/grupları tam isteğe dahil et. Gereksiz geniş tarih aralığı kullanma.",
		"Gerekliyse önce save_outline, sonra save_plan çağır. Tamamlama/geri almayı sadece açık kullanıcı isteğiyle uygula.",
		"created, updated, deleted, unchanged, skipped sonuçlarını incele; atlanan veya başarısız işlemleri tamamlandı diye bildirme.",
		"İlgili planı ve ilerlemeyi tekrar oku; sonucu, silinen ve atlanan kayıtları Türkçe özetle.",
	}
}
