package plan

import (
	"context"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

// actionAuditCount counts audit rows with the given action.
func actionAuditCount(t *testing.T, st *store.Store, action string) int {
	t.Helper()
	var n int
	if err := st.DB().QueryRow("SELECT COUNT(1) FROM audit_log WHERE action = ?", action).Scan(&n); err != nil {
		t.Fatalf("count %s audit: %v", action, err)
	}
	return n
}

func TestCompleteFutureRejected(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	sat := sessionAt(t, st, "2026-10-31", "10:00") // future
	_, err := svc.Complete(ctx, SourceWeb, sat.ID)
	ve, ok := AsValidationError(err)
	if !ok {
		t.Fatalf("want ValidationError, got %v", err)
	}
	if ve.Errors[0].Code != CodeFuture {
		t.Errorf("code = %s, want %s", ve.Errors[0].Code, CodeFuture)
	}
}

func TestCompleteMissingSession(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t, istanbul(t, "2026-10-26 09:00"))
	_, err := svc.Complete(ctx, SourceWeb, 999)
	ve, ok := AsValidationError(err)
	if !ok || ve.Errors[0].Code != CodeNotFound {
		t.Fatalf("want not_found ValidationError, got %v", err)
	}
}

func TestCompleteIdempotentOneAudit(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	mon := sessionAt(t, st, "2026-10-26", "20:30")

	for i := 0; i < 2; i++ {
		view, err := svc.Complete(ctx, SourceWeb, mon.ID)
		if err != nil {
			t.Fatalf("Complete #%d: %v", i, err)
		}
		if view.State != StateDone || !view.Completed {
			t.Errorf("Complete #%d state = %q", i, view.State)
		}
	}
	if n := actionAuditCount(t, st, actionComplete); n != 1 {
		t.Errorf("complete audit rows = %d, want 1 (second call is a no-op)", n)
	}
	again := sessionAt(t, st, "2026-10-26", "20:30")
	if again.CompletedAt == nil || again.CompletedSource == nil || *again.CompletedSource != "web" {
		t.Errorf("session not completed by web: %+v", again)
	}
}

func TestUndoRestores(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	mon := sessionAt(t, st, "2026-10-26", "20:30")

	if _, err := svc.Complete(ctx, SourceWeb, mon.ID); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	view, err := svc.Undo(ctx, SourceWeb, mon.ID)
	if err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if view.Completed || view.State == StateDone {
		t.Errorf("undo left session done: %+v", view)
	}
	again := sessionAt(t, st, "2026-10-26", "20:30")
	if again.CompletedAt != nil {
		t.Errorf("completed_at not cleared: %v", again.CompletedAt)
	}
	// A second undo is a no-op (still one undo audit row).
	if _, err := svc.Undo(ctx, SourceWeb, mon.ID); err != nil {
		t.Fatalf("Undo repeat: %v", err)
	}
	if n := actionAuditCount(t, st, actionUndo); n != 1 {
		t.Errorf("undo audit rows = %d, want 1", n)
	}
}

func TestCompleteMidnightBoundary(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	wed := sessionAt(t, st, "2026-10-28", "20:30")

	// 23:59 on the 27th (Istanbul): the 28th is still the future → rejected.
	late := New(st, clock.Fake{Current: istanbul(t, "2026-10-27 23:59")})
	_, err := late.Complete(ctx, SourceWeb, wed.ID)
	ve, ok := AsValidationError(err)
	if !ok || ve.Errors[0].Code != CodeFuture {
		t.Fatalf("at 23:59 want future rejection, got %v", err)
	}

	// 00:00 on the 28th: now today → allowed.
	midnight := New(st, clock.Fake{Current: istanbul(t, "2026-10-28 00:00")})
	view, err := midnight.Complete(ctx, SourceWeb, wed.ID)
	if err != nil {
		t.Fatalf("at 00:00 Complete: %v", err)
	}
	if view.State != StateDone {
		t.Errorf("state = %q, want done", view.State)
	}
}
