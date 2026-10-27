package plan

import (
	"github.com/ibra2626/workey/internal/clock"
	"testing"
	"time"
)

func TestGuideUsesIstanbulDate(t *testing.T) {
	s := New(nil, clock.Fake{Current: time.Date(2026, 10, 26, 22, 0, 0, 0, time.UTC)})
	g := s.AgentGuide(clock.Date{Year: 2027, Month: 7, Day: 18})
	if g.Today != "2026-10-27" || g.ExamDate != "2027-07-18" {
		t.Fatalf("%+v", g)
	}
}
