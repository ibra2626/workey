package views

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
)

const progressMeasure = "Tamamlanan seansların planlı süresi gösterilir; ölçülmüş çalışma süresi değildir. Kısmi çalışma ve 10 dakika başlangıç hedefi toplamı artırmaz."

const progressCounting = "Eski seans ve telafisi ayrı ayrı tamamlanırsa ikisi de toplama katılır. Yüzdeler benzersiz öğrenilen sayfa veya içerik oranı değildir."

func partialStudy(sv plan.SessionView) bool {
	return sv.State != plan.StateDone && sv.Study != nil
}

func weekStudyContext(sv plan.SessionView) string {
	if sv.Study == nil {
		return ""
	}
	context := "Çalışma başlangıcı kayıtlı; devam yeri kaydedilmedi."
	if b := sv.Study.Bookmark; b != nil {
		context = "Kaldığın yer: " + b.Notation + " · " + stageName(b.Stage) + " · Devam edeceğin sayfa: s. " + strconv.Itoa(b.ResumePage)
	}
	if sv.State == plan.StateDone && sv.MovedTo != nil {
		context += " · Telafi bağlantısı: " + sv.MovedTo.Date
	}
	return context
}

// Raw submitted strings intentionally survive validation, including the version.
type StudyFormInput struct {
	SessionID                     int64
	Version, Ordinal, Stage, Page string
}

func studyInput(sv plan.SessionView, submitted *StudyFormInput) StudyFormInput {
	if submitted != nil && submitted.SessionID == sv.ID {
		return *submitted
	}
	v := StudyFormInput{SessionID: sv.ID, Version: strconv.FormatInt(sv.StudyVersion, 10), Stage: plan.StageCalisma}
	if len(sv.Items) == 1 {
		v.Ordinal = "0"
	}
	if sv.Kind == plan.KindAksam {
		v.Stage = plan.StageKonu
	}
	if sv.Study != nil && sv.Study.Bookmark != nil {
		b := sv.Study.Bookmark
		v.Ordinal, v.Stage, v.Page = strconv.Itoa(b.ItemOrdinal), b.Stage, strconv.Itoa(b.ResumePage)
	}
	return v
}

func stageName(stage string) string {
	switch stage {
	case plan.StageKonu:
		return "Konu"
	case plan.StageSoru:
		return "Soru"
	case plan.StageYanlisDefteri:
		return "Yanlış defteri"
	default:
		return "Çalışma"
	}
}
func studyStartLabel(sv plan.SessionView) string {
	if sv.Study != nil {
		return "Kaldığım yerden devam et"
	}
	return "Çalışmaya geç"
}
func sourceRanges(item plan.Item, kind string) string {
	base := plan.FormatItem(item)
	if kind != plan.KindAksam {
		return base + " · Çalışma: s. " + numSpan(item.PageFrom, item.PageTo)
	}
	kf, kt, sf, st := item.PageFrom, item.PageTo, item.PageFrom, item.PageTo
	if item.KonuFrom != nil && item.KonuTo != nil {
		kf, kt = *item.KonuFrom, *item.KonuTo
	}
	if item.SoruFrom != nil && item.SoruTo != nil {
		sf, st = *item.SoruFrom, *item.SoruTo
	}
	return base + " · Konu: s. " + numSpan(kf, kt) + "; Soru: s. " + numSpan(sf, st) + "; Yanlış defteri: s. " + numSpan(item.PageFrom, item.PageTo)
}
func studyFirstStep(sv plan.SessionView) string {
	if sv.Study != nil && sv.Study.Bookmark != nil {
		b := sv.Study.Bookmark
		return b.Notation + " · " + stageName(b.Stage) + ": s. " + strconv.Itoa(b.ResumePage) + " sayfasından devam et."
	}
	if len(sv.Items) == 0 {
		return "Seans kaynaklarını aç."
	}
	it := sv.Items[0]
	page, stage := it.PageFrom, "Çalışma"
	if sv.Kind == plan.KindAksam {
		stage = "Konu"
		if it.KonuFrom != nil {
			page = *it.KonuFrom
		}
	}
	return plan.FormatItem(it) + " · " + stage + ": s. " + strconv.Itoa(page) + " sayfasını aç."
}
func studyStartTime(value string) string {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return t.In(clock.Location).Format("02.01.2006 15:04")
}
func studyCanStart(p TodayPage, sv plan.SessionView) bool {
	return p.StudyPeriodOpen && sv.Date <= p.Day.Date && sv.State != plan.StateDone && sv.MovedTo == nil && p.Day.ActiveStudy == nil
}

func studyFeedbackRole(invalid bool) string {
	if invalid {
		return "alert"
	}
	return "status"
}
func validStudyOrdinal(sv plan.SessionView, value string) bool {
	i, err := strconv.Atoi(value)
	return err == nil && i >= 0 && i < len(sv.Items) && strconv.Itoa(i) == value
}
func validStudyStage(sv plan.SessionView, value string) bool {
	if sv.Kind != plan.KindAksam {
		return value == plan.StageCalisma
	}
	return value == plan.StageKonu || value == plan.StageSoru || value == plan.StageYanlisDefteri
}

// studyTopic names what the session is about beyond the chapter title: the
// topic group from the outline, then the book and chapter, e.g.
// "Konu: Temel kavramlar · Antrenmanlarla Matematik 1, 3. bölüm".
func studyTopic(sv plan.SessionView) string {
	var parts []string
	if sv.TopicGroupName != "" {
		parts = append(parts, "Konu: "+sv.TopicGroupName)
	}
	var item *plan.Item
	if sv.Study != nil && sv.Study.Bookmark != nil {
		item = &sv.Study.Bookmark.Item
	} else if len(sv.Items) > 0 {
		item = &sv.Items[0]
	}
	if item != nil {
		book := bookName(item.Book)
		if item.ChapterNo > 0 {
			book += ", " + strconv.Itoa(item.ChapterNo) + ". bölüm"
		}
		parts = append(parts, book)
	}
	return strings.Join(parts, " "+middleDot+" ")
}

func studyTitle(sv plan.SessionView) string {
	if sv.Study != nil && sv.Study.Bookmark != nil {
		return sv.Study.Bookmark.Item.ChapterTitle
	}
	return chapterTitle(sv)
}

// roomSegment is one block of the study room timeline. Minutes come from the
// same split as the session profile; the room script only reads label, stage
// and minutes, and nothing it computes is stored.
type roomSegment struct {
	Label string `json:"label"`
	Stage string `json:"stage"`
	Min   int    `json:"min"`
	Fill  string `json:"-"`
	Pages string `json:"-"`
}

func roomSegments(sv plan.SessionView) []roomSegment {
	stages := []string{plan.StageKonu, plan.StageSoru, plan.StageYanlisDefteri}
	var out []roomSegment
	for i, b := range sessionProfile(sv) {
		seg := roomSegment{Label: b.Label, Stage: plan.StageCalisma, Min: b.Flex, Fill: b.Fill, Pages: b.Pages}
		if b.Single {
			seg.Min = sv.DurationMin
		} else if i < len(stages) {
			seg.Stage = stages[i]
		}
		out = append(out, seg)
	}
	return out
}

func roomSegmentsJSON(sv plan.SessionView) string {
	b, err := json.Marshal(roomSegments(sv))
	if err != nil {
		return "[]"
	}
	return string(b)
}

// roomOffset starts the timeline at the bookmarked stage, so resuming in soru
// does not show konu again.
func roomOffset(sv plan.SessionView) int {
	if sv.Study == nil || sv.Study.Bookmark == nil {
		return 0
	}
	offset := 0
	for _, seg := range roomSegments(sv) {
		if seg.Stage == sv.Study.Bookmark.Stage {
			return offset
		}
		offset += seg.Min
	}
	return 0
}

func roomPlanMin(sv plan.SessionView) int {
	total := 0
	for _, seg := range roomSegments(sv) {
		total += seg.Min
	}
	return total
}

func roomAutoStage(sv plan.SessionView, submitted *StudyFormInput) bool {
	return (submitted == nil || submitted.SessionID != sv.ID) && (sv.Study == nil || sv.Study.Bookmark == nil)
}

func singleStudySource(sv plan.SessionView, input StudyFormInput) bool {
	return len(sv.Items) == 1 && input.Ordinal == "0"
}

func studyStages(sv plan.SessionView) []string {
	if sv.Kind == plan.KindAksam {
		return []string{plan.StageKonu, plan.StageSoru, plan.StageYanlisDefteri}
	}
	return []string{plan.StageCalisma}
}

// roomRow is one line of the study room's session schedule: a stage with the
// clock times it covers, counted from when this study actually started (moved
// back to the start of the bookmarked stage on resume).
type roomRow struct {
	roomSegment
	From, To string
}

func roomRows(sv plan.SessionView) []roomRow {
	segs := roomSegments(sv)
	rows := make([]roomRow, 0, len(segs))
	var t time.Time
	known := false
	if sv.Study != nil {
		if started, err := time.Parse(time.RFC3339, sv.Study.LastStartedAt); err == nil {
			t, known = started.In(clock.Location).Add(-time.Duration(roomOffset(sv))*time.Minute), true
		}
	}
	for _, seg := range segs {
		if seg.Pages == "" {
			seg.Pages = "Bugünün yanlışları"
			if seg.Stage == plan.StageCalisma {
				seg.Pages = headPages(sv)
			}
		}
		row := roomRow{roomSegment: seg}
		if known {
			row.From = t.Format("15:04")
			t = t.Add(time.Duration(seg.Min) * time.Minute)
			row.To = t.Format("15:04")
		}
		rows = append(rows, row)
	}
	return rows
}
