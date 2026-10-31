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
		"save_plan / PUT /api/v1/plan, from–to dahil aralıktaki bugün ve geleceğe ait başlanmamış açık seansların TAM listesini değiştirir. İstekten çıkarılan başlanmamış açık seanslar silinir; korunacak seansları da gönder. Eşleştirme date + start_time ile yapılır.",
		"Geçmiş tarihler, tamamlanmış seanslar ve study kaydı olan başlanmış seanslar plan yazımında korunur. Başlanmış seansın aynı içeriği unchanged; değişikliği skipped reason=started döner; istekten eksikse silinmez. Geçmiş/completed önceliği korunur. skipped kayıtları başarıyla değişmiş sayma; reason alanını kullanıcıya açıkla.",
		"save_outline / PUT /api/v1/outline faz ve konu gruplarının TAM listesini key üzerinden değiştirir. Korunacak kayıtları da gönder. Seansların kullandığı grup silinemez.",
		"get_plan yanıtındaki topic_group_key yazarken topic_group alanına taşınır. id, state, completed, notation ve study okuma alanlarıdır; yazma isteğine taşıma. items referanslarını koru.",
		"Telafi için makeup_of = {date, start_time} kullan; hedef geçmiş ve tamamlanmamış bir seans olmalı. Kaynak geçmiş seansı değiştirme. Okuma yanıtı makeup_of taşımaz; moved_to bilgisi kaynak seansta bulunur. Telafi seansını yeniden yazarken kaynak bağlantısını get_plan sonuçlarından doğrula; belirsizse kullanıcıdan sor.",
		"Aynı gövdeyi tekrar göndermek kayıt veya audit çoğaltmaz. idempotency_key ilk sonucu tekrar döndürür; aynı anahtar farklı içerikle kullanılamaz. REST'te boş olmayan Idempotency-Key başlığı gövdedeki anahtardan önceliklidir. MCP ve REST arasında aynı anahtarı paylaşma.",
		StudyReadRules(),
		"complete_session ve REST tamamlama/geri alma işlemlerini yalnızca kullanıcı istediğinde yap. Kimliği get_plan sonucundan al. Gelecek seanslar tamamlanamaz veya geri alınamaz. Tamamlama study durumunu atomik closed yapar; geri alma yer imini koruyup paused yapar. Study kaydı olmayan eski seanslar ek veri istemez. Mevcut REST/MCP yazmaları study.version denetlemez; eski sekme niyetini saptayamaz.",
		"validation_failed alan hatalarını düzeltmeden isteği tekrarlama. idempotency_conflict durumunda önceki sonucu oku ve değişmiş istek için yeni anahtar kullan. Yetkilendirme hatasında tokenı veya gizli bilgileri yanıta yazma.",
	}
}

func GuideWorkflow() []string {
	return []string{
		"Önce get_agent_guide / GET /api/v1/agent-guide, get_outline, ilgili aralık için get_plan ve get_progress çağır.",
		"Kullanıcının hedefini ve kitap/bölüm/sayfa referanslarını doğrula; eksikleri sor.",
		"Yer imlerini ve study durumunu incele; başlanmış seansların kaynak, tarih, saat ve süresini koru. İstenen değişikliği hazırla; korunacak açık seansları ve mevcut faz/grupları tam isteğe dahil et. Gereksiz geniş tarih aralığı kullanma.",
		"Gerekliyse önce save_outline, sonra save_plan çağır. Tamamlama/geri almayı sadece açık kullanıcı isteğiyle uygula.",
		"created, updated, deleted, unchanged, skipped sonuçlarını incele; atlanan veya başarısız işlemleri tamamlandı diye bildirme.",
		"İlgili planı ve ilerlemeyi tekrar oku; sonucu, silinen ve atlanan kayıtları Türkçe özetle.",
	}
}

// StudyReadRules is reused in the agent guide, usage pages and checked docs.
func StudyReadRules() string {
	return "SessionView.study opsiyonel, salt okunur çalışma bağlamıdır: status=active/paused/closed, first_started_at/last_started_at/updated_at (UTC), entry_mode=normal/ten_minutes ve version. bookmark varsa item_ordinal (sıfır tabanlı), stage (akşam: konu/soru/yanlis_defteri; diğer: calisma), resume_page, çözülmüş item ve notation içerir. resume_page devam edilecek kitap sayfasıdır; yarım sayfa aynı kalır, otomatik +1 yapılmaz. Yanlış defteri aşamasında da kitap sayfasıdır. Yer imi yoksa uydurma; study yokluğu hiç çalışılmadığını kanıtlamaz. SessionState planned/today/done/missed/moved değişmez; geçmiş kısmi emeği ayrıca study üzerinden oku. GET /api/v1/today active_study ve latest_paused_study ile tarih dışındaki devam bağlamını da sunar; aktif önceliklidir, yarım işe dönüş açık seçimdir, bugünün planı korunur. get_plan yalnız istenen aralığı döndürür; gerekirse geçmiş aralığı da oku. Yeni çalışma yazma REST endpoint'i veya MCP aracı yoktur. Kısmi sayfa ve ten_minutes hedefi ilerlemeyi artırmaz; done_min tamamlanan seansların planlı süresidir, ölçülmüş süre veya öğrenme yüzdesi değildir. Telafi bağlantısı yer imini veya tamamlamayı otomatik taşımaz; eski ve telafi ayrı tamamlanırsa mevcut toplam ikisini de sayar."
}
