package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStudyUpgradeAndPersistence(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "upgrade.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	// Build a real pre-upgrade database, with the original migration recorded.
	body, err := migrationsFS.ReadFile("migrations/0001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().ExecContext(ctx, string(body)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().ExecContext(ctx, `CREATE TABLE schema_migrations (version TEXT PRIMARY KEY);
		INSERT INTO schema_migrations VALUES ('0001_init.sql')`); err != nil {
		t.Fatal(err)
	}
	sess := &Session{Date: "2026-10-29", StartTime: "20:30", DurationMin: 90, Kind: "aksam", CreatedAt: "before", UpdatedAt: "before"}
	if err := s.InsertSession(ctx, s.DB(), sess); err != nil {
		t.Fatal(err)
	}
	it := &SessionItem{SessionID: sess.ID, Ordinal: 0, Book: "M1", ChapterNo: 1, ChapterTitle: "Temsili kaynak", PageFrom: 1, PageTo: 10}
	if err := s.InsertSessionItem(ctx, s.DB(), it); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := s.Migrate(ctx); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.SessionByID(ctx, s.DB(), sess.ID)
	if err != nil || got == nil || len(got.Items) != 1 || got.Items[0].ID != it.ID || got.UpdatedAt != "before" {
		t.Fatalf("legacy read: %+v, %v", got, err)
	}
	if v, err := s.StudyBySession(ctx, s.DB(), sess.ID); err != nil || v != nil {
		t.Fatalf("legacy study: %+v, %v", v, err)
	}
	v := &Study{SessionID: sess.ID, Status: "active", FirstStartedAt: "2026-10-29T17:30:00Z", LastStartedAt: "2026-10-29T17:30:00Z", UpdatedAt: "2026-10-29T17:30:00Z", EntryMode: "normal", Version: 1}
	if err := s.WithTx(ctx, func(tx *sql.Tx) error { return s.InsertStudy(ctx, tx, v) }); err != nil {
		t.Fatal(err)
	}
	v.Status, v.EntryMode, v.Version = "paused", "ten_minutes", 2
	v.ItemOrdinal, v.Stage, v.ResumePage = ptrInt(0), ptrStr("konu"), ptrInt(5)
	if err := s.WithTx(ctx, func(tx *sql.Tx) error { return s.UpdateStudy(ctx, tx, v) }); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	read, err := s.StudyBySession(ctx, s.DB(), sess.ID)
	if err != nil || !reflect.DeepEqual(read, v) {
		t.Fatalf("persisted study: %+v, %v", read, err)
	}
	missing := *v
	missing.SessionID = 999
	if err := s.UpdateStudy(ctx, s.DB(), &missing); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing update: %v", err)
	}
}

func TestStudyConstraints(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	for _, slot := range []string{"20:30", "22:00"} {
		if err := s.InsertSession(ctx, s.DB(), &Session{Date: "2026-10-29", StartTime: slot, DurationMin: 90, Kind: "aksam", CreatedAt: "now", UpdatedAt: "now"}); err != nil {
			t.Fatal(err)
		}
	}
	base := Study{SessionID: 1, Status: "active", EntryMode: "normal", FirstStartedAt: "now", LastStartedAt: "now", UpdatedAt: "now", Version: 1}
	if err := s.InsertStudy(ctx, s.DB(), &base); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*Study)
	}{
		{"second active", func(v *Study) { v.Status = "active" }},
		{"orphan", func(v *Study) { v.SessionID = 999; v.Status = "paused" }},
		{"status", func(v *Study) { v.Status = "unknown" }},
		{"mode", func(v *Study) { v.EntryMode = "unknown" }},
		{"stage", func(v *Study) { v.ItemOrdinal = ptrInt(0); v.Stage = ptrStr("unknown"); v.ResumePage = ptrInt(1) }},
		{"version", func(v *Study) { v.Version = 0 }},
		{"ordinal", func(v *Study) { v.ItemOrdinal = ptrInt(-1); v.Stage = ptrStr("konu"); v.ResumePage = ptrInt(1) }},
		{"page", func(v *Study) { v.ItemOrdinal = ptrInt(0); v.Stage = ptrStr("konu"); v.ResumePage = ptrInt(0) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := base
			v.SessionID = 2
			v.Status = "paused"
			tc.change(&v)
			if err := s.InsertStudy(ctx, s.DB(), &v); err == nil {
				t.Fatal("expected constraint error")
			}
		})
	}
	// All six partially filled bookmark combinations must fail, even paused.
	for mask := 1; mask < 7; mask++ {
		v := base
		v.SessionID = 2
		v.Status = "paused"
		if mask&1 != 0 {
			v.ItemOrdinal = ptrInt(0)
		}
		if mask&2 != 0 {
			v.Stage = ptrStr("konu")
		}
		if mask&4 != 0 {
			v.ResumePage = ptrInt(1)
		}
		if err := s.InsertStudy(ctx, s.DB(), &v); err == nil {
			t.Fatalf("partial bookmark %d accepted", mask)
		}
	}
	if err := s.DeleteSession(ctx, s.DB(), 1); err == nil {
		t.Fatal("started session deleted")
	}
	// Paused and closed records are not restricted by the active index.
	base.Status = "paused"
	if err := s.UpdateStudy(ctx, s.DB(), &base); err != nil {
		t.Fatal(err)
	}
	base.SessionID = 2
	base.Status = "active"
	if err := s.InsertStudy(ctx, s.DB(), &base); err != nil {
		t.Fatal(err)
	}
	base.Status = "closed"
	if err := s.UpdateStudy(ctx, s.DB(), &base); err != nil {
		t.Fatal(err)
	}
}
