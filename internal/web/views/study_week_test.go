package views

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/plan"
)

func TestWeekPartialContextAndCompletion(t *testing.T) {
	sv := plan.SessionView{ID: 42, Kind: plan.KindAksam, State: plan.StateMissed,
		StudyVersion: 2, Items: []plan.Item{{Book: "M1", ChapterNo: 1, ChapterTitle: "Temsili bölüm", PageFrom: 10, PageTo: 20}},
		Study: &plan.StudyView{Status: plan.StudyPaused, Bookmark: &plan.BookmarkView{Notation: "G1·2, s. 100–120", Stage: plan.StageSoru, ResumePage: 105}}}
	for _, moved := range []bool{false, true} {
		if moved {
			sv.State = plan.StateMoved
			sv.MovedTo = &plan.MovedInfo{Date: "2026-10-31", Weekday: "Cumartesi"}
		}
		for _, done := range []bool{false, true} {
			v := sv
			if done {
				v.State, v.StudyVersion = plan.StateDone, 3
			}
			var b bytes.Buffer
			if err := WeekMain(WeekPage{Week: &plan.WeekView{Days: []plan.DayRow{{DayNum: 26, Sessions: []plan.SessionView{v}}}}}).Render(context.Background(), &b); err != nil {
				t.Fatal(err)
			}
			body := b.String()
			for _, want := range []string{"G1·2, s. 100–120", "Soru", "Devam edeceğin sayfa: s. 105", progressMeasure, progressCounting} {
				if !strings.Contains(body, want) {
					t.Errorf("missing %q", want)
				}
			}
			if !done && (strings.Contains(body, "Yapılamadı") || strings.Count(body, "Kısmen çalışıldı") != 2) {
				t.Fatal("partial state lost on mobile/desktop")
			}
			if !done && moved && !strings.Contains(body, "cumartesiye taşındı") {
				t.Fatal("makeup relation lost")
			}
			if !strings.Contains(body, "/hafta/oturum/42/"+rowActionSlug(v)) || !strings.Contains(body, `name="expected_study_version"`) {
				t.Fatal("completion/undo control missing")
			}
		}
	}
	sv.Study.Bookmark = nil
	if !strings.Contains(weekStudyContext(sv), "devam yeri kaydedilmedi") {
		t.Fatal("invented bookmark")
	}
	sv.Study = nil
	sv.State = plan.StateMissed
	if statusWord(sv) != "Yapılamadı" {
		t.Fatal("legacy row changed")
	}
}
