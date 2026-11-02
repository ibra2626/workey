package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

// seedSource inserts a source with one section holding two units and returns
// the ids (source, section, unit1, unit2).
func seedSource(t *testing.T, s *Store, key string) (src *Source, secID, u1, u2 int64) {
	t.Helper()
	ctx := context.Background()
	src = &Source{Key: key, Name: key + " adı", Kind: "book", LocationUnit: "page",
		Ordinal: 1, CreatedAt: "t0", UpdatedAt: "t0"}
	if err := s.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.InsertSource(ctx, tx, src); err != nil {
			return err
		}
		sec := &Section{SourceID: src.ID, Ordinal: 1, Name: "Bölüm 1", LocFrom: ptrInt(1), LocTo: ptrInt(40)}
		if err := s.InsertSection(ctx, tx, sec); err != nil {
			return err
		}
		secID = sec.ID
		a := &Unit{SectionID: sec.ID, Ordinal: 1, Label: "Antrenman", Number: ptrInt(1)}
		b := &Unit{SectionID: sec.ID, Ordinal: 2, Label: "Test", Number: ptrInt(1), LocFrom: ptrInt(10), LocTo: ptrInt(15)}
		if err := s.InsertUnit(ctx, tx, a); err != nil {
			return err
		}
		if err := s.InsertUnit(ctx, tx, b); err != nil {
			return err
		}
		u1, u2 = a.ID, b.ID
		return nil
	}); err != nil {
		t.Fatalf("seedSource %s: %v", key, err)
	}
	return src, secID, u1, u2
}

func TestSourceCRUDAndList(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	a, _, _, _ := seedSource(t, s, "M1")
	b, _, _, _ := seedSource(t, s, "G1")
	// Insert order (M1 ordinal 1, G1 ordinal 1) but list orders by ordinal; give
	// G1 a later ordinal to pin the order.
	b.Ordinal = 2
	b.UpdatedAt = "t1"
	if err := s.UpdateSourceByKey(ctx, s.DB(), *b); err != nil {
		t.Fatalf("update G1: %v", err)
	}

	got, err := s.ListSources(ctx, s.DB())
	if err != nil {
		t.Fatalf("ListSources: %v", err)
	}
	if len(got) != 2 || got[0].Key != "M1" || got[1].Key != "G1" {
		t.Fatalf("list wrong: %+v", got)
	}
	if got[1].Ordinal != 2 || got[1].UpdatedAt != "t1" {
		t.Fatalf("update not applied: %+v", got[1])
	}

	one, err := s.SourceByKey(ctx, s.DB(), "M1")
	if err != nil || one == nil || one.ID != a.ID || one.LocationUnit != "page" {
		t.Fatalf("SourceByKey(M1): %+v, %v", one, err)
	}
	missing, err := s.SourceByKey(ctx, s.DB(), "nope")
	if err != nil || missing != nil {
		t.Fatalf("SourceByKey(nope): %+v, %v", missing, err)
	}

	if err := s.DeleteSourceByKey(ctx, s.DB(), "G1"); err != nil {
		t.Fatalf("delete G1: %v", err)
	}
	got, _ = s.ListSources(ctx, s.DB())
	if len(got) != 1 || got[0].Key != "M1" {
		t.Fatalf("after delete: %+v", got)
	}
}

func TestSourceCascadeDelete(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	src, _, u1, _ := seedSource(t, s, "M1")

	// Mark a unit done, then delete the source: sections, units and progress go.
	if err := s.UpsertUnitProgress(ctx, s.DB(), UnitProgress{
		UnitID: u1, DoneAt: ptrStr("t2"), Correct: ptrInt(30), Wrong: ptrInt(8), Blank: ptrInt(2), CompletedSource: "web",
	}); err != nil {
		t.Fatalf("upsert progress: %v", err)
	}

	if err := s.DeleteSourceByKey(ctx, s.DB(), "M1"); err != nil {
		t.Fatalf("delete source: %v", err)
	}
	for _, tc := range []struct {
		name, query string
		arg         any
	}{
		{"sections", `SELECT COUNT(1) FROM sections WHERE source_id = ?`, src.ID},
		{"units", `SELECT COUNT(1) FROM units WHERE id = ?`, u1},
		{"progress", `SELECT COUNT(1) FROM unit_progress WHERE unit_id = ?`, u1},
	} {
		var n int
		if err := s.DB().QueryRowContext(ctx, tc.query, tc.arg).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", tc.name, err)
		}
		if n != 0 {
			t.Fatalf("%s after cascade = %d, want 0", tc.name, n)
		}
	}
}

func TestUnitProgressSurvivesContentRewrite(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	_, secID, u1, _ := seedSource(t, s, "M1")

	if err := s.UpsertUnitProgress(ctx, s.DB(), UnitProgress{
		UnitID: u1, DoneAt: ptrStr("t2"), Correct: ptrInt(30), Wrong: ptrInt(8), Blank: ptrInt(2), CompletedSource: "mcp",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// A content re-write that keeps the unit id (UpdateUnit, not delete+insert)
	// must preserve the unit's progress — progress is keyed by unit id.
	if err := s.UpdateUnit(ctx, s.DB(), Unit{
		ID: u1, SectionID: secID, Ordinal: 1, Label: "Alıştırma", Number: ptrInt(2), LocFrom: ptrInt(3), LocTo: ptrInt(9),
	}); err != nil {
		t.Fatalf("update unit: %v", err)
	}

	got, err := s.UnitProgressByUnit(ctx, s.DB(), u1)
	if err != nil || got == nil {
		t.Fatalf("progress gone after rewrite: %+v, %v", got, err)
	}
	if got.Correct == nil || *got.Correct != 30 || got.CompletedSource != "mcp" {
		t.Fatalf("progress changed: %+v", got)
	}

	// Re-read the unit to confirm its content did change.
	units, err := s.UnitsBySection(ctx, s.DB(), secID)
	if err != nil || len(units) != 2 || units[0].Label != "Alıştırma" {
		t.Fatalf("unit content not rewritten: %+v, %v", units, err)
	}

	// By contrast, ReplaceUnits drops progress (new ids): document that too.
	if err := s.WithTx(ctx, func(tx *sql.Tx) error {
		return s.ReplaceUnits(ctx, tx, secID, []Unit{{Ordinal: 1, Label: "Yeni", Number: ptrInt(1)}})
	}); err != nil {
		t.Fatalf("replace units: %v", err)
	}
	if got, err := s.UnitProgressByUnit(ctx, s.DB(), u1); err != nil || got != nil {
		t.Fatalf("progress should be gone after ReplaceUnits: %+v, %v", got, err)
	}
}

func TestUnitProgressUpsertAndReadBySource(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	src, _, u1, u2 := seedSource(t, s, "M1")

	// Insert, then overwrite (upsert) the same unit's result.
	if err := s.UpsertUnitProgress(ctx, s.DB(), UnitProgress{UnitID: u1, DoneAt: ptrStr("t1"), Correct: ptrInt(10), CompletedSource: "web"}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := s.UpsertUnitProgress(ctx, s.DB(), UnitProgress{UnitID: u1, DoneAt: ptrStr("t2"), Correct: ptrInt(20), Wrong: ptrInt(5), CompletedSource: "api"}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	// A video-style unit: done with no D/Y/B.
	if err := s.UpsertUnitProgress(ctx, s.DB(), UnitProgress{UnitID: u2, DoneAt: ptrStr("t3"), CompletedSource: "mcp"}); err != nil {
		t.Fatalf("u2 upsert: %v", err)
	}

	one, err := s.UnitProgressByUnit(ctx, s.DB(), u1)
	if err != nil || one == nil || one.Correct == nil || *one.Correct != 20 || one.CompletedSource != "api" {
		t.Fatalf("u1 progress after overwrite: %+v, %v", one, err)
	}

	bySource, err := s.UnitProgressBySource(ctx, s.DB(), src.ID)
	if err != nil {
		t.Fatalf("UnitProgressBySource: %v", err)
	}
	if len(bySource) != 2 {
		t.Fatalf("progress rows = %d, want 2", len(bySource))
	}
	if p := bySource[u2]; p.Correct != nil || p.CompletedSource != "mcp" {
		t.Fatalf("u2 D/Y/B should be empty: %+v", p)
	}

	// Removing progress marks the unit not done again.
	if err := s.DeleteUnitProgress(ctx, s.DB(), u1); err != nil {
		t.Fatalf("delete progress: %v", err)
	}
	if got, err := s.UnitProgressByUnit(ctx, s.DB(), u1); err != nil || got != nil {
		t.Fatalf("u1 still done: %+v, %v", got, err)
	}
}

func TestCatalogConstraints(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	src, secID, u1, _ := seedSource(t, s, "M1")

	cases := []struct {
		name string
		fn   func(tx *sql.Tx) error
	}{
		{"duplicate source key", func(tx *sql.Tx) error {
			return s.InsertSource(ctx, tx, &Source{Key: "M1", Name: "x", Kind: "book", LocationUnit: "page", Ordinal: 9, CreatedAt: "t", UpdatedAt: "t"})
		}},
		{"bad source kind", func(tx *sql.Tx) error {
			return s.InsertSource(ctx, tx, &Source{Key: "X1", Name: "x", Kind: "audio", LocationUnit: "page", Ordinal: 9, CreatedAt: "t", UpdatedAt: "t"})
		}},
		{"bad location_unit", func(tx *sql.Tx) error {
			return s.InsertSource(ctx, tx, &Source{Key: "X2", Name: "x", Kind: "book", LocationUnit: "chapter", Ordinal: 9, CreatedAt: "t", UpdatedAt: "t"})
		}},
		{"duplicate section ordinal", func(tx *sql.Tx) error {
			return s.InsertSection(ctx, tx, &Section{SourceID: src.ID, Ordinal: 1, Name: "çakışma"})
		}},
		{"section range order", func(tx *sql.Tx) error {
			return s.InsertSection(ctx, tx, &Section{SourceID: src.ID, Ordinal: 5, Name: "ters", LocFrom: ptrInt(9), LocTo: ptrInt(2)})
		}},
		{"duplicate unit ordinal", func(tx *sql.Tx) error {
			return s.InsertUnit(ctx, tx, &Unit{SectionID: secID, Ordinal: 1, Label: "çakışma"})
		}},
		{"unit range order", func(tx *sql.Tx) error {
			return s.InsertUnit(ctx, tx, &Unit{SectionID: secID, Ordinal: 9, Label: "ters", LocFrom: ptrInt(9), LocTo: ptrInt(2)})
		}},
		{"orphan section", func(tx *sql.Tx) error {
			return s.InsertSection(ctx, tx, &Section{SourceID: 999, Ordinal: 1, Name: "yetim"})
		}},
		{"orphan unit", func(tx *sql.Tx) error {
			return s.InsertUnit(ctx, tx, &Unit{SectionID: 999, Ordinal: 1, Label: "yetim"})
		}},
		{"bad progress source", func(tx *sql.Tx) error {
			return s.UpsertUnitProgress(ctx, tx, UnitProgress{UnitID: u1, DoneAt: ptrStr("t"), CompletedSource: "cron"})
		}},
		{"negative correct", func(tx *sql.Tx) error {
			return s.UpsertUnitProgress(ctx, tx, UnitProgress{UnitID: u1, DoneAt: ptrStr("t"), Correct: ptrInt(-1), CompletedSource: "web"})
		}},
		{"orphan progress", func(tx *sql.Tx) error {
			return s.UpsertUnitProgress(ctx, tx, UnitProgress{UnitID: 999, DoneAt: ptrStr("t"), CompletedSource: "web"})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := s.WithTx(ctx, tc.fn); err == nil {
				t.Fatalf("%s: expected constraint error", tc.name)
			}
		})
	}
}

func TestUpdateMissingUnitAndSection(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	if err := s.UpdateUnit(ctx, s.DB(), Unit{ID: 999, SectionID: 1, Ordinal: 1, Label: "x"}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("UpdateUnit(missing): %v, want ErrNoRows", err)
	}
	if err := s.UpdateSection(ctx, s.DB(), Section{ID: 999, Ordinal: 1, Name: "x"}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("UpdateSection(missing): %v, want ErrNoRows", err)
	}
}
