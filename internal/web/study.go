package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/web/views"
)

func (a *Auth) startStudy(w http.ResponseWriter, r *http.Request) { a.writeStudy(w, r, true) }
func (a *Auth) pauseStudy(w http.ResponseWriter, r *http.Request) { a.writeStudy(w, r, false) }

func (a *Auth) writeStudy(w http.ResponseWriter, r *http.Request, start bool) {
	id, parseErr := strconv.ParseInt(r.PathValue("id"), 10, 64)
	version, err := completionStudyVersion(r)
	// Unlike legacy completion forms, study forms must carry an explicit version.
	if parseErr != nil || id <= 0 || err != nil || len(r.PostForm["expected_study_version"]) != 1 {
		http.Error(w, "Geçersiz seans veya çalışma sürümü. Sayfayı yenileyin.", http.StatusBadRequest)
		return
	}
	input := &views.StudyFormInput{SessionID: id, Version: r.PostFormValue("expected_study_version"),
		Ordinal: r.PostFormValue("item_ordinal"), Stage: r.PostFormValue("stage"), Page: r.PostFormValue("resume_page")}
	message := "Çalışma odası açıldı. Kitabını aç; ekran açık kalabilir ama gerekmiyor."
	parseMessage := ""
	if start {
		_, err = a.svc.StartStudy(r.Context(), plan.SourceWeb, id, r.PostFormValue("entry_mode"), version, a.examDate)
	} else {
		var bookmark *plan.Bookmark
		switch r.PostFormValue("pause_mode") {
		case "bookmark":
			ordinal, e1 := strconv.Atoi(input.Ordinal)
			page, e2 := strconv.Atoi(input.Page)
			if e1 != nil || e2 != nil {
				parseMessage = "Kaynak ve devam edeceğin sayfa alanlarına geçerli sayılar girin."
				err = errors.New(parseMessage)
			} else {
				bookmark = &plan.Bookmark{ItemOrdinal: ordinal, Stage: input.Stage, ResumePage: page}
			}
			message = "Kaldığın yer kaydedildi. İstediğinde buradan devam edebilirsin."
		case "close":
			message = "Çalışma kapatıldı; devam yeri kaydedilmedi. Önceki yer imi varsa korundu."
		default:
			parseMessage = "Geçerli bir kapatma yolu seçin."
			err = errors.New(parseMessage)
		}
		if err == nil {
			_, err = a.svc.PauseStudy(r.Context(), plan.SourceWeb, id, bookmark, version)
		}
	}
	status := http.StatusOK
	if err != nil {
		status = http.StatusUnprocessableEntity
		message = "İşlem kaydedilemedi. Girilen değerleri kontrol edin."
		if parseMessage != "" {
			message = parseMessage
		} else if ve, ok := plan.AsValidationError(err); ok {
			var messages []string
			for _, field := range ve.Errors {
				messages = append(messages, field.Message)
			}
			message = strings.Join(messages, " ")
		} else if errors.Is(err, plan.ErrStudyConflict) {
			status = http.StatusConflict
			message = "Çalışma başka bir sekmede değişmiş veya başka bir çalışma açık. Girilen değerler korunuyor; güncel durumu görmek için sayfayı yenileyin."
		} else {
			var moved *plan.StudyMovedError
			if errors.As(err, &moved) {
				message = "Bu seans için telafi planlandı: " + moved.Date + " " + moved.StartTime + ". Hafta ekranından telafi seansını açın."
			} else {
				status = http.StatusInternalServerError
				message = "İşlem doğrulanamadı. Yeniden denemeden önce güncel durumu okuyun."
			}
		}
	}
	if err == nil && r.Header.Get("HX-Request") == "" && (start || r.PostFormValue("pause_mode") != "close") {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	p, readErr := a.buildTodayPage(r.Context(), csrfTokenFromRequest(r))
	if readErr != nil {
		http.Error(w, "Sunucu hatası; kaydı yeniden okuyun.", http.StatusInternalServerError)
		return
	}
	p.StudyFeedback, p.StudyInvalid = message, err != nil
	if err != nil && !start && r.PostFormValue("pause_mode") == "bookmark" {
		p.StudyInput = input
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status < 500 {
		w.Header().Set("X-Workey-Study-Form", "true")
	}
	w.WriteHeader(status)
	if r.Header.Get("HX-Request") != "" {
		_ = views.TodayMain(p).Render(r.Context(), w)
	} else {
		_ = views.Today(p).Render(r.Context(), w)
	}
}
