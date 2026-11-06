package plan

import (
	"context"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

// sectionUnitIDs returns the ids of a source section's units, ordered by
// ordinal, failing if the section is absent.
func sectionUnitIDs(t *testing.T, st *store.Store, sourceKey string, sectionOrdinal int) []int64 {
	t.Helper()
	ctx := context.Background()
	db := st.DB()
	src, err := st.SourceByKey(ctx, db, sourceKey)
	if err != nil || src == nil {
		t.Fatalf("SourceByKey %s: %v", sourceKey, err)
	}
	secs, err := st.SectionsBySource(ctx, db, src.ID)
	if err != nil {
		t.Fatalf("SectionsBySource: %v", err)
	}
	for _, sec := range secs {
		if sec.Ordinal != sectionOrdinal {
			continue
		}
		units, err := st.UnitsBySection(ctx, db, sec.ID)
		if err != nil {
			t.Fatalf("UnitsBySection: %v", err)
		}
		ids := make([]int64, len(units))
		for i, u := range units {
			ids[i] = u.ID
		}
		return ids
	}
	t.Fatalf("source %s has no section with ordinal %d", sourceKey, sectionOrdinal)
	return nil
}

// progressRow reads a unit's stored progress directly, or nil when not done.
func progressRow(t *testing.T, st *store.Store, unitID int64) *store.UnitProgress {
	t.Helper()
	p, err := st.UnitProgressByUnit(context.Background(), st.DB(), unitID)
	if err != nil {
		t.Fatalf("UnitProgressByUnit %d: %v", unitID, err)
	}
	return p
}

func TestSetUnitResultDoneNoDYB(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)
	unit := sectionUnitIDs(t, st, "M1", 3)[0]

	view, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: unit, Done: true})
	if err != nil {
		t.Fatalf("SetUnitResult: %v", err)
	}
	if !view.Done || view.DoneAt == nil {
		t.Fatalf("view not done: %+v", view)
	}
	if view.Correct != nil || view.Wrong != nil || view.Blank != nil {
		t.Errorf("D/Y/B should be nil: %+v", view)
	}
	if view.Notation != "M1·3, Antrenman 1" {
		t.Errorf("notation = %q, want %q", view.Notation, "M1·3, Antrenman 1")
	}
	row := progressRow(t, st, unit)
	if row == nil || row.DoneAt == nil || row.CompletedSource != "web" {
		t.Errorf("stored row wrong: %+v", row)
	}
}

func TestSetUnitResultDoneWithDYB(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)
	unit := sectionUnitIDs(t, st, "M1", 3)[1]

	view, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{
		UnitID: unit, Done: true, Correct: ptr(12), Wrong: ptr(3), Blank: ptr(1),
	})
	if err != nil {
		t.Fatalf("SetUnitResult: %v", err)
	}
	if view.Correct == nil || *view.Correct != 12 || view.Wrong == nil || *view.Wrong != 3 || view.Blank == nil || *view.Blank != 1 {
		t.Errorf("D/Y/B not recorded: %+v", view)
	}
	row := progressRow(t, st, unit)
	if row == nil || row.Correct == nil || *row.Correct != 12 {
		t.Errorf("stored D/Y/B wrong: %+v", row)
	}
}

func TestSetUnitResultIdempotentStableDoneAt(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)
	unit := sectionUnitIDs(t, st, "M1", 3)[0]

	first, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: unit, Done: true, Correct: ptr(5)})
	if err != nil {
		t.Fatalf("SetUnitResult #1: %v", err)
	}
	// A later clock, same payload: no new state, no second audit row, done_at stable.
	later := New(st, clock.Fake{Current: istanbul(t, "2026-10-27 09:00")})
	second, err := later.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: unit, Done: true, Correct: ptr(5)})
	if err != nil {
		t.Fatalf("SetUnitResult #2: %v", err)
	}
	if first.DoneAt == nil || second.DoneAt == nil || *first.DoneAt != *second.DoneAt {
		t.Errorf("done_at changed on repeat: %v -> %v", first.DoneAt, second.DoneAt)
	}
	if n := unitAuditCount(t, st); n != 1 {
		t.Errorf("unit audit rows = %d, want 1 (repeat is a no-op)", n)
	}
}

func TestSetUnitResultUpdateKeepsDoneAtAndAudits(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)
	unit := sectionUnitIDs(t, st, "M1", 3)[0]

	first, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: unit, Done: true, Correct: ptr(5)})
	if err != nil {
		t.Fatalf("SetUnitResult #1: %v", err)
	}
	later := New(st, clock.Fake{Current: istanbul(t, "2026-10-27 09:00")})
	second, err := later.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: unit, Done: true, Correct: ptr(9)})
	if err != nil {
		t.Fatalf("SetUnitResult #2: %v", err)
	}
	if *second.Correct != 9 {
		t.Errorf("correct not updated: %+v", second)
	}
	if *first.DoneAt != *second.DoneAt {
		t.Errorf("done_at must stay the first completion: %v -> %v", first.DoneAt, second.DoneAt)
	}
	if n := unitAuditCount(t, st); n != 2 {
		t.Errorf("unit audit rows = %d, want 2 (a real change)", n)
	}
}

func TestSetUnitResultUndoClearsRow(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)
	unit := sectionUnitIDs(t, st, "M1", 3)[0]

	if _, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: unit, Done: true}); err != nil {
		t.Fatalf("set done: %v", err)
	}
	view, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: unit, Done: false})
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	if view.Done || view.DoneAt != nil {
		t.Errorf("view still done after undo: %+v", view)
	}
	if progressRow(t, st, unit) != nil {
		t.Errorf("progress row not cleared")
	}
	// A second undo is a no-op: still one undo audit row.
	if _, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: unit, Done: false}); err != nil {
		t.Fatalf("undo repeat: %v", err)
	}
	if n := actionAuditCount(t, st, actionUndo); n != 1 {
		t.Errorf("undo audit rows = %d, want 1", n)
	}
}

func TestSetUnitResultNegativeRejected(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)
	unit := sectionUnitIDs(t, st, "M1", 3)[0]

	_, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: unit, Done: true, Wrong: ptr(-1)})
	ve, ok := AsValidationError(err)
	if !ok || ve.Errors[0].Code != CodeInvalid || ve.Errors[0].Path != "wrong" {
		t.Fatalf("want invalid wrong, got %v", err)
	}
	if progressRow(t, st, unit) != nil {
		t.Errorf("row written despite validation error")
	}
}

func TestSetUnitResultMissingUnit(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)
	_, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: 9999, Done: true})
	ve, ok := AsValidationError(err)
	if !ok || ve.Errors[0].Code != CodeNotFound {
		t.Fatalf("want not_found, got %v", err)
	}
}

func TestSetUnitResultVideoRejectsDYB(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)
	videoUnit := seedVideoUnit(t, st)

	// D/Y/B on a video unit is rejected.
	_, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: videoUnit, Done: true, Correct: ptr(1)})
	if ve, ok := AsValidationError(err); !ok || ve.Errors[0].Code != CodeInvalid {
		t.Fatalf("want invalid for video D/Y/B, got %v", err)
	}
	// But a video unit can still be marked done (D/Y/B left null).
	view, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: videoUnit, Done: true})
	if err != nil {
		t.Fatalf("video done: %v", err)
	}
	if !view.Done || view.Correct != nil {
		t.Errorf("video unit view wrong: %+v", view)
	}
}

// TestUnitAndSessionIndependence verifies the two directions of independence:
// marking every unit of a session done does not complete the session, and
// completing a session does not touch its units.
func TestUnitAndSessionIndependence(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)

	// A session on today referencing M1·3 Antrenman 1–3.
	req := SavePlanRequest{From: "2026-10-26", To: "2026-10-26", Sessions: []SessionInput{{
		Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: KindUzunBlok,
		Items: []Item{{SourceKey: "M1", SectionOrdinal: ptr(3), UnitFrom: ptr(1), UnitTo: ptr(3)}},
	}}}
	if _, err := svc.SavePlan(ctx, SourceAPI, req); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	units := sectionUnitIDs(t, st, "M1", 3)

	// Mark every unit done.
	for _, id := range units {
		if _, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: id, Done: true}); err != nil {
			t.Fatalf("SetUnitResult %d: %v", id, err)
		}
	}
	// The session is still open.
	if ss := sessionAt(t, st, "2026-10-26", "20:30"); ss.CompletedAt != nil {
		t.Errorf("session completed by unit marks: %v", ss.CompletedAt)
	}

	// Completing the session must not touch the units.
	ss := sessionAt(t, st, "2026-10-26", "20:30")
	before := progressRow(t, st, units[0])
	if _, err := svc.Complete(ctx, SourceWeb, ss.ID); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	after := progressRow(t, st, units[0])
	if before == nil || after == nil || *before.DoneAt != *after.DoneAt {
		t.Errorf("unit progress changed by session completion: %+v -> %+v", before, after)
	}
}

func TestUnitProgressAggregates(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	seedCatalog(t, st)
	units := sectionUnitIDs(t, st, "M1", 3) // three Antrenman units in section 3

	// Two done with D/Y/B, one done without.
	if _, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: units[0], Done: true, Correct: ptr(10), Wrong: ptr(2), Blank: ptr(1)}); err != nil {
		t.Fatalf("set 0: %v", err)
	}
	if _, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: units[1], Done: true, Correct: ptr(8), Wrong: ptr(4)}); err != nil {
		t.Fatalf("set 1: %v", err)
	}
	if _, err := svc.SetUnitResult(ctx, SourceWeb, UnitResultRequest{UnitID: units[2], Done: true}); err != nil {
		t.Fatalf("set 2: %v", err)
	}

	bySource, err := svc.SourceProgress(ctx)
	if err != nil {
		t.Fatalf("SourceProgress: %v", err)
	}
	srcTotals := bySource[sourceID(t, st, "M1")]
	if srcTotals.DoneCount != 3 || srcTotals.Correct != 18 || srcTotals.Wrong != 6 || srcTotals.Blank != 1 {
		t.Errorf("source totals = %+v, want {3 18 6 1}", srcTotals)
	}

	bySection, err := svc.SectionProgress(ctx)
	if err != nil {
		t.Fatalf("SectionProgress: %v", err)
	}
	secID := sectionID(t, st, "M1", 3)
	secTotals := bySection[secID]
	if secTotals.DoneCount != 3 || secTotals.Correct != 18 || secTotals.Wrong != 6 || secTotals.Blank != 1 {
		t.Errorf("section totals = %+v, want {3 18 6 1}", secTotals)
	}

	// Read-by-units returns exactly the done units.
	results, err := svc.UnitResults(ctx, units)
	if err != nil {
		t.Fatalf("UnitResults: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("UnitResults len = %d, want 3", len(results))
	}
}

// --- test helpers specific to unit progress ---

func unitAuditCount(t *testing.T, st *store.Store) int {
	t.Helper()
	var n int
	if err := st.DB().QueryRow("SELECT COUNT(1) FROM audit_log WHERE entity = 'unit'").Scan(&n); err != nil {
		t.Fatalf("count unit audit: %v", err)
	}
	return n
}

// sectionID returns the stored id of a source's section with the given ordinal.
func sectionID(t *testing.T, st *store.Store, sourceKey string, ordinal int) int64 {
	t.Helper()
	ctx := context.Background()
	db := st.DB()
	src, err := st.SourceByKey(ctx, db, sourceKey)
	if err != nil || src == nil {
		t.Fatalf("SourceByKey %s: %v", sourceKey, err)
	}
	secs, err := st.SectionsBySource(ctx, db, src.ID)
	if err != nil {
		t.Fatalf("SectionsBySource: %v", err)
	}
	for _, sec := range secs {
		if sec.Ordinal == ordinal {
			return sec.ID
		}
	}
	t.Fatalf("source %s has no section %d", sourceKey, ordinal)
	return 0
}

// seedVideoUnit inserts a video source with one section and one "Ders" unit,
// returning the unit id, for the video D/Y/B rule.
func seedVideoUnit(t *testing.T, st *store.Store) int64 {
	t.Helper()
	ctx := context.Background()
	db := st.DB()
	src := &store.Source{Key: "V1", Name: "YouTube serisi", Kind: "video",
		LocationUnit: "minute", Ordinal: 9, CreatedAt: "2026-10-01T00:00:00Z", UpdatedAt: "2026-10-01T00:00:00Z"}
	if err := st.InsertSource(ctx, db, src); err != nil {
		t.Fatalf("InsertSource V1: %v", err)
	}
	sec := &store.Section{SourceID: src.ID, Ordinal: 1, Name: "Bölüm 1"}
	if err := st.InsertSection(ctx, db, sec); err != nil {
		t.Fatalf("InsertSection V1: %v", err)
	}
	u := &store.Unit{SectionID: sec.ID, Ordinal: 1, Label: "Ders", Number: ptr(1)}
	if err := st.InsertUnit(ctx, db, u); err != nil {
		t.Fatalf("InsertUnit V1: %v", err)
	}
	return u.ID
}
