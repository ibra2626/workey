package plan

import (
	"context"
	"database/sql"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

// LeftoverUnit is a unit a closed session covered that is still not done and
// that no open session covers. Workey only reports it; placing it back into the
// plan stays with the assistant (save_plan).
type LeftoverUnit struct {
	UnitID      int64  `json:"unit_id"`
	Notation    string `json:"notation"`
	SessionDate string `json:"session_date"` // the closed session that left it
}

// unitSlot is one session's place on the timeline for a unit.
type unitSlot struct {
	date, start string
}

func (a unitSlot) before(b unitSlot) bool {
	if a.date != b.date {
		return a.date < b.date
	}
	return a.start < b.start
}

// unitTimeline records, per unit, the closed sessions that covered it and the
// open sessions that still do, in date order.
type unitTimeline struct {
	closed map[int64][]unitSlot
	open   map[int64][]unitSlot
}

// closedSession reports whether a session is over for leftover purposes: it was
// completed, or it was started on an earlier day. A past session nobody started
// is a missed session, which the makeup flow handles instead.
func closedSession(ss store.Session, started bool, today string) bool {
	return ss.CompletedAt != nil || (started && ss.Date < today)
}

// forEachTimelineSession walks every session in date order with the units it
// covers, classified as closed or open. A past session neither completed nor
// started belongs to neither group.
func (s *Service) forEachTimelineSession(ctx context.Context, tx *sql.Tx, cat *catalog, fn func(ss store.Session, closed bool, units []store.Unit, item ItemView)) error {
	all, err := s.st.AllSessions(ctx, tx)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return nil
	}
	studies, err := s.st.StudiesByDateRange(ctx, tx, all[0].Date, all[len(all)-1].Date)
	if err != nil {
		return err
	}
	today := clock.Today(s.clk).String()
	for _, ss := range all {
		closed := closedSession(ss, studies[ss.ID] != nil, today)
		if !closed && (ss.CompletedAt != nil || ss.Date < today) {
			continue
		}
		for _, it := range cat.itemViews(ss.Items) {
			fn(ss, closed, cat.coveredUnits(it), it)
		}
	}
	return nil
}

// loadUnitTimeline builds the per-unit timeline within tx.
func (s *Service) loadUnitTimeline(ctx context.Context, tx *sql.Tx, cat *catalog) (*unitTimeline, error) {
	t := &unitTimeline{closed: map[int64][]unitSlot{}, open: map[int64][]unitSlot{}}
	err := s.forEachTimelineSession(ctx, tx, cat, func(ss store.Session, closed bool, units []store.Unit, _ ItemView) {
		slot := unitSlot{date: ss.Date, start: ss.StartTime}
		for _, u := range units {
			if closed {
				t.closed[u.ID] = append(t.closed[u.ID], slot)
			} else {
				t.open[u.ID] = append(t.open[u.ID], slot)
			}
		}
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

// annotate fills an unfinished unit row's carried-over flag and the next open
// session after sv that covers it.
func (t *unitTimeline) annotate(row *UnitRow, sv SessionView) {
	here := unitSlot{date: sv.Date, start: sv.StartTime}
	for _, slot := range t.closed[row.UnitID] {
		if slot.before(here) {
			row.CarriedOver = true
			break
		}
	}
	for _, slot := range t.open[row.UnitID] {
		if here.before(slot) {
			row.NextDate = slot.date
			break
		}
	}
}

// LeftoverUnits returns the unfinished units of closed sessions that no open
// session covers, in the order the sessions occurred.
func (s *Service) LeftoverUnits(ctx context.Context) ([]LeftoverUnit, error) {
	var out []LeftoverUnit
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		cat, err := s.loadCatalog(ctx, tx)
		if err != nil {
			return err
		}
		timeline, err := s.loadUnitTimeline(ctx, tx, cat)
		if err != nil {
			return err
		}
		var candidates []LeftoverUnit
		var ids []int64
		seen := map[int64]bool{}
		err = s.forEachTimelineSession(ctx, tx, cat, func(ss store.Session, closed bool, units []store.Unit, it ItemView) {
			if !closed {
				return
			}
			for _, u := range units {
				if seen[u.ID] || len(timeline.open[u.ID]) > 0 {
					continue
				}
				seen[u.ID] = true
				ids = append(ids, u.ID)
				candidates = append(candidates, LeftoverUnit{
					UnitID: u.ID, Notation: unitRow(cat, it, u).Notation, SessionDate: ss.Date,
				})
			}
		})
		if err != nil {
			return err
		}
		done, err := s.st.UnitProgressByUnits(ctx, tx, ids)
		if err != nil {
			return err
		}
		for _, c := range candidates {
			if _, ok := done[c.UnitID]; !ok {
				out = append(out, c)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
