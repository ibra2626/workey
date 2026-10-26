package store

import (
	"context"
	"path/filepath"
	"testing"
)

// newStore opens a migrated store on a fresh temp-dir database.
func newStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "workey.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if err := s.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return s
}

func ptrI64(v int64) *int64   { return &v }
func ptrInt(v int) *int       { return &v }
func ptrStr(v string) *string { return &v }

func TestMigrateIdempotent(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	// Second run is a no-op.
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	var applied int
	if err := s.DB().QueryRowContext(ctx,
		`SELECT COUNT(1) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if applied != 1 {
		t.Fatalf("schema_migrations rows = %d, want 1", applied)
	}

	// Every v1 table exists.
	for _, tbl := range []string{
		"phases", "topic_groups", "sessions", "session_items",
		"audit_log", "idempotency_keys",
	} {
		var name string
		err := s.DB().QueryRowContext(ctx,
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, tbl).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", tbl, err)
		}
	}
}

func TestForeignKeysOn(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	var fk int
	if err := s.DB().QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatalf("pragma foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys = %d, want 1", fk)
	}

	// An item under a missing session violates the FK.
	err := s.WithTx(ctx, func(tx dbtx) error {
		return s.InsertSessionItem(ctx, tx, &SessionItem{
			SessionID: 999, Ordinal: 1, Book: "M1",
			ChapterNo: 1, ChapterTitle: "x", PageFrom: 1, PageTo: 2,
		})
	})
	if err == nil {
		t.Fatal("expected FK violation inserting orphan item, got nil")
	}
}

func TestSessionsByDateRangeWithItems(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	now := "2026-10-26T09:00:00Z"

	// Three sessions on three dates; the range query should return the two
	// inside [2026-11-02, 2026-11-08] in date/time order with their items.
	mk := func(date, start, kind string) int64 {
		var id int64
		err := s.WithTx(ctx, func(tx dbtx) error {
			sess := &Session{
				Date: date, StartTime: start, DurationMin: 90, Kind: kind,
				CreatedAt: now, UpdatedAt: now,
			}
			if err := s.InsertSession(ctx, tx, sess); err != nil {
				return err
			}
			id = sess.ID
			return s.InsertSessionItem(ctx, tx, &SessionItem{
				SessionID: sess.ID, Ordinal: 1, Book: "M2",
				ChapterNo: 17, ChapterTitle: "Türev", PageFrom: 210, PageTo: 225,
				SoruFrom: ptrInt(1), SoruTo: ptrInt(40),
			})
		})
		if err != nil {
			t.Fatalf("create session %s: %v", date, err)
		}
		return id
	}

	mk("2026-11-01", "20:00", "aksam") // before range
	inA := mk("2026-11-02", "20:00", "aksam")
	inB := mk("2026-11-08", "09:00", "uzun_blok")
	mk("2026-11-09", "20:00", "aksam") // after range

	got, err := s.SessionsByDateRange(ctx, s.DB(), "2026-11-02", "2026-11-08")
	if err != nil {
		t.Fatalf("SessionsByDateRange: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d sessions, want 2", len(got))
	}
	if got[0].ID != inA || got[1].ID != inB {
		t.Fatalf("order wrong: got ids %d,%d want %d,%d", got[0].ID, got[1].ID, inA, inB)
	}
	if len(got[0].Items) != 1 || got[0].Items[0].Book != "M2" || got[0].Items[0].PageTo != 225 {
		t.Fatalf("session items wrong: %+v", got[0].Items)
	}
	if got[0].Items[0].SoruTo == nil || *got[0].Items[0].SoruTo != 40 {
		t.Fatalf("soru_to not round-tripped: %+v", got[0].Items[0])
	}
}

func TestCascadeDeleteItems(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	now := "2026-10-26T09:00:00Z"

	var sessID int64
	if err := s.WithTx(ctx, func(tx dbtx) error {
		sess := &Session{Date: "2026-11-02", StartTime: "20:00", DurationMin: 90, Kind: "aksam", CreatedAt: now, UpdatedAt: now}
		if err := s.InsertSession(ctx, tx, sess); err != nil {
			return err
		}
		sessID = sess.ID
		return s.InsertSessionItem(ctx, tx, &SessionItem{
			SessionID: sess.ID, Ordinal: 1, Book: "M1", ChapterNo: 1, ChapterTitle: "x", PageFrom: 1, PageTo: 2,
		})
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := s.DB().ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, sessID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	var items int
	if err := s.DB().QueryRowContext(ctx, `SELECT COUNT(1) FROM session_items WHERE session_id = ?`, sessID).Scan(&items); err != nil {
		t.Fatalf("count items: %v", err)
	}
	if items != 0 {
		t.Fatalf("items after cascade = %d, want 0", items)
	}
}

func TestCheckConstraints(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	now := "2026-10-26T09:00:00Z"

	// Bad kind.
	if err := s.WithTx(ctx, func(tx dbtx) error {
		return s.InsertSession(ctx, tx, &Session{Date: "2026-11-02", StartTime: "20:00", DurationMin: 90, Kind: "yok", CreatedAt: now, UpdatedAt: now})
	}); err == nil {
		t.Fatal("expected CHECK failure for bad kind")
	}

	// page_from > page_to.
	err := s.WithTx(ctx, func(tx dbtx) error {
		sess := &Session{Date: "2026-11-03", StartTime: "20:00", DurationMin: 90, Kind: "aksam", CreatedAt: now, UpdatedAt: now}
		if err := s.InsertSession(ctx, tx, sess); err != nil {
			return err
		}
		return s.InsertSessionItem(ctx, tx, &SessionItem{SessionID: sess.ID, Ordinal: 1, Book: "M1", ChapterNo: 1, ChapterTitle: "x", PageFrom: 50, PageTo: 10})
	})
	if err == nil {
		t.Fatal("expected CHECK failure for page_from > page_to")
	}
}

func TestUniqueDateStartTime(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	now := "2026-10-26T09:00:00Z"

	ins := func() error {
		return s.WithTx(ctx, func(tx dbtx) error {
			return s.InsertSession(ctx, tx, &Session{Date: "2026-11-02", StartTime: "20:00", DurationMin: 90, Kind: "aksam", CreatedAt: now, UpdatedAt: now})
		})
	}
	if err := ins(); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if err := ins(); err == nil {
		t.Fatal("expected UNIQUE(date, start_time) violation")
	}
}

func TestMakeupSelfReferenceAndSetNull(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	now := "2026-10-26T09:00:00Z"

	var missedID, makeupID int64
	if err := s.WithTx(ctx, func(tx dbtx) error {
		missed := &Session{Date: "2026-11-02", StartTime: "20:00", DurationMin: 90, Kind: "aksam", CreatedAt: now, UpdatedAt: now}
		if err := s.InsertSession(ctx, tx, missed); err != nil {
			return err
		}
		missedID = missed.ID
		makeup := &Session{Date: "2026-11-07", StartTime: "20:00", DurationMin: 90, Kind: "aksam", MakeupOf: ptrI64(missedID), CreatedAt: now, UpdatedAt: now}
		if err := s.InsertSession(ctx, tx, makeup); err != nil {
			return err
		}
		makeupID = makeup.ID
		return nil
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Deleting the missed session nulls the makeup link (ON DELETE SET NULL).
	if _, err := s.DB().ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, missedID); err != nil {
		t.Fatalf("delete missed: %v", err)
	}
	var makeupOf *int64
	if err := s.DB().QueryRowContext(ctx, `SELECT makeup_of FROM sessions WHERE id = ?`, makeupID).Scan(&makeupOf); err != nil {
		t.Fatalf("read makeup_of: %v", err)
	}
	if makeupOf != nil {
		t.Fatalf("makeup_of = %v, want nil after SET NULL", *makeupOf)
	}
}

func TestAuditInsert(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	entry := &AuditEntry{
		AtUTC: "2026-10-26T09:00:00Z", Source: "mcp", Action: "save_plan",
		Entity: "session", EntityID: ptrI64(7),
		BeforeJSON: nil, AfterJSON: ptrStr(`{"id":7}`), RequestHash: ptrStr("abc123"),
	}
	if err := s.WithTx(ctx, func(tx dbtx) error {
		return s.InsertAudit(ctx, tx, entry)
	}); err != nil {
		t.Fatalf("InsertAudit: %v", err)
	}
	if entry.ID == 0 {
		t.Fatal("audit ID not set")
	}

	var source, action string
	var after *string
	if err := s.DB().QueryRowContext(ctx,
		`SELECT source, action, after_json FROM audit_log WHERE id = ?`, entry.ID,
	).Scan(&source, &action, &after); err != nil {
		t.Fatalf("read audit: %v", err)
	}
	if source != "mcp" || action != "save_plan" || after == nil || *after != `{"id":7}` {
		t.Fatalf("audit round-trip wrong: %s %s %v", source, action, after)
	}

	// Bad source is rejected.
	if err := s.WithTx(ctx, func(tx dbtx) error {
		return s.InsertAudit(ctx, tx, &AuditEntry{AtUTC: "2026-10-26T09:00:00Z", Source: "cron", Action: "x", Entity: "session"})
	}); err == nil {
		t.Fatal("expected CHECK failure for bad audit source")
	}
}

func TestIdempotencyRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	// Missing key returns (nil, nil).
	got, err := s.GetIdempotency(ctx, s.DB(), "missing")
	if err != nil {
		t.Fatalf("GetIdempotency(missing): %v", err)
	}
	if got != nil {
		t.Fatalf("missing key returned %+v, want nil", got)
	}

	rec := IdempotencyRecord{
		Key: "req-1", RequestHash: "hash-1",
		ResponseJSON: `{"ok":true}`, CreatedAt: "2026-10-26T09:00:00Z",
	}
	if err := s.WithTx(ctx, func(tx dbtx) error {
		return s.PutIdempotency(ctx, tx, rec)
	}); err != nil {
		t.Fatalf("PutIdempotency: %v", err)
	}

	got, err = s.GetIdempotency(ctx, s.DB(), "req-1")
	if err != nil {
		t.Fatalf("GetIdempotency(req-1): %v", err)
	}
	if got == nil || got.RequestHash != "hash-1" || got.ResponseJSON != `{"ok":true}` {
		t.Fatalf("idempotency round-trip wrong: %+v", got)
	}

	// Same key again is rejected (PK), so callers must read-before-write.
	if err := s.WithTx(ctx, func(tx dbtx) error {
		return s.PutIdempotency(ctx, tx, rec)
	}); err == nil {
		t.Fatal("expected PK violation on duplicate idempotency key")
	}
}
