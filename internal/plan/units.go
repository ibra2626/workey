package plan

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ibra2626/workey/internal/store"
)

// kindVideo is the source kind whose units carry no D/Y/B: a video has nothing
// to score right, wrong or blank, so marking it done leaves those fields null.
const kindVideo = "video"

// UnitResultRequest marks one unit done (or not), optionally recording its D/Y/B
// (doğru/yanlış/boş). D/Y/B are non-negative counts; the service does not check
// that they sum to anything — the user reports whatever they did. A video unit
// must leave them null. This path is entirely independent of session completion.
type UnitResultRequest struct {
	UnitID  int64 `json:"unit_id"`
	Done    bool  `json:"done"`
	Correct *int  `json:"correct,omitempty"`
	Wrong   *int  `json:"wrong,omitempty"`
	Blank   *int  `json:"blank,omitempty"`
}

// UnitResultView is a unit's progress after a SetUnitResult call: its catalog
// identity (for display) plus the current stored result. Done is false and the
// result fields are nil when the unit is not done.
type UnitResultView struct {
	UnitID         int64   `json:"unit_id"`
	SourceKey      string  `json:"source_key"`
	SectionOrdinal int     `json:"section_ordinal"`
	Label          string  `json:"label"`
	Number         *int    `json:"number,omitempty"`
	Notation       string  `json:"notation"`
	Done           bool    `json:"done"`
	DoneAt         *string `json:"done_at,omitempty"`
	Correct        *int    `json:"correct,omitempty"`
	Wrong          *int    `json:"wrong,omitempty"`
	Blank          *int    `json:"blank,omitempty"`
}

// UnitResult is a unit's stored progress for a read model, keyed elsewhere by
// unit id. The presence of the entry means the unit is done.
type UnitResult struct {
	DoneAt  *string `json:"done_at,omitempty"`
	Correct *int    `json:"correct,omitempty"`
	Wrong   *int    `json:"wrong,omitempty"`
	Blank   *int    `json:"blank,omitempty"`
}

// ProgressTotals aggregates the done units and their D/Y/B over a source or a
// section, for the İlerleme/Kaynaklar pages. A done unit with null D/Y/B adds to
// DoneCount but contributes zero to the sums.
type ProgressTotals struct {
	DoneCount int `json:"done_count"`
	Correct   int `json:"correct"`
	Wrong     int `json:"wrong"`
	Blank     int `json:"blank"`
}

// SetUnitResult marks a unit done or not done, independently of any session.
//
//   - done=true upserts the unit's progress row, recording optional D/Y/B and the
//     writing source. The first completion stamps done_at; a later change keeps
//     that original done_at (the unit stays "done since" its first mark).
//   - done=false clears the row (undo).
//
// The call is idempotent: the same payload twice leaves the same state and
// writes no second audit row. Every real change is audited (entity "unit",
// before/after JSON, source). D/Y/B must be non-negative when present and must
// be null for a video unit.
func (s *Service) SetUnitResult(ctx context.Context, source Source, req UnitResultRequest) (*UnitResultView, error) {
	if !source.valid() {
		return nil, fmt.Errorf("plan: invalid source %q", source)
	}
	var verr errBuilder
	checkNonNeg(&verr, "correct", req.Correct)
	checkNonNeg(&verr, "wrong", req.Wrong)
	checkNonNeg(&verr, "blank", req.Blank)
	if ve := verr.result(); ve != nil {
		return nil, ve
	}

	now := s.nowUTC()
	var view *UnitResultView
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		u, err := s.st.UnitByID(ctx, tx, req.UnitID)
		if err != nil {
			return err
		}
		if u == nil {
			return fieldError("unit_id", CodeNotFound, "unit does not exist")
		}
		if u.SourceKind == kindVideo && (req.Correct != nil || req.Wrong != nil || req.Blank != nil) {
			return fieldError("correct", CodeInvalid, "video biriminde doğru/yanlış/boş tutulmaz")
		}

		existing, err := s.st.UnitProgressByUnit(ctx, tx, req.UnitID)
		if err != nil {
			return err
		}

		if !req.Done {
			if existing == nil {
				// Already not done: no write, no audit row.
				view = buildUnitResultView(u, nil)
				return nil
			}
			before, err := unitSnapshotJSON(existing)
			if err != nil {
				return err
			}
			if err := s.st.DeleteUnitProgress(ctx, tx, req.UnitID); err != nil {
				return err
			}
			if err := s.auditUnit(ctx, tx, source, actionUndo, req.UnitID, &before, nil, now); err != nil {
				return err
			}
			view = buildUnitResultView(u, nil)
			return nil
		}

		desired := store.UnitProgress{
			UnitID:          req.UnitID,
			Correct:         req.Correct,
			Wrong:           req.Wrong,
			Blank:           req.Blank,
			CompletedSource: string(source),
		}
		if existing != nil && existing.DoneAt != nil {
			desired.DoneAt = existing.DoneAt // keep the first completion instant
		} else {
			n := now
			desired.DoneAt = &n
		}

		if existing != nil && unitProgressEqual(*existing, desired) {
			// Same state already stored: no write, no audit row.
			view = buildUnitResultView(u, existing)
			return nil
		}

		var before *string
		if existing != nil {
			b, err := unitSnapshotJSON(existing)
			if err != nil {
				return err
			}
			before = &b
		}
		if err := s.st.UpsertUnitProgress(ctx, tx, desired); err != nil {
			return err
		}
		after, err := unitSnapshotJSON(&desired)
		if err != nil {
			return err
		}
		if err := s.auditUnit(ctx, tx, source, actionComplete, req.UnitID, before, &after, now); err != nil {
			return err
		}
		view = buildUnitResultView(u, &desired)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

// UnitResults returns the stored progress for the given unit ids, keyed by unit
// id. Units that are not done are simply absent from the map.
func (s *Service) UnitResults(ctx context.Context, unitIDs []int64) (map[int64]UnitResult, error) {
	rows, err := s.st.UnitProgressByUnits(ctx, s.st.DB(), unitIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]UnitResult, len(rows))
	for id, p := range rows {
		out[id] = UnitResult{DoneAt: p.DoneAt, Correct: p.Correct, Wrong: p.Wrong, Blank: p.Blank}
	}
	return out, nil
}

// SourceProgress returns the done-count and D/Y/B sums per source id, over every
// source that has at least one done unit.
func (s *Service) SourceProgress(ctx context.Context) (map[int64]ProgressTotals, error) {
	totals, err := s.st.UnitProgressTotalsBySource(ctx, s.st.DB())
	if err != nil {
		return nil, err
	}
	return toProgressTotals(totals), nil
}

// SectionProgress returns the done-count and D/Y/B sums per section id, over
// every section that has at least one done unit.
func (s *Service) SectionProgress(ctx context.Context) (map[int64]ProgressTotals, error) {
	totals, err := s.st.UnitProgressTotalsBySection(ctx, s.st.DB())
	if err != nil {
		return nil, err
	}
	return toProgressTotals(totals), nil
}

// toProgressTotals maps store aggregates to the plan-level type.
func toProgressTotals(in map[int64]store.UnitProgressTotals) map[int64]ProgressTotals {
	out := make(map[int64]ProgressTotals, len(in))
	for id, t := range in {
		out[id] = ProgressTotals{DoneCount: t.DoneCount, Correct: t.Correct, Wrong: t.Wrong, Blank: t.Blank}
	}
	return out
}

// checkNonNeg records a field error when a present count is negative.
func checkNonNeg(b *errBuilder, field string, v *int) {
	if v != nil && *v < 0 {
		b.add(field, CodeInvalid, field+" negatif olamaz")
	}
}

// unitProgressEqual reports whether two progress rows hold the same state, so a
// repeat leaves no write and no audit row.
func unitProgressEqual(a, b store.UnitProgress) bool {
	return strPtrEqual(a.DoneAt, b.DoneAt) &&
		intPtrEqual(a.Correct, b.Correct) &&
		intPtrEqual(a.Wrong, b.Wrong) &&
		intPtrEqual(a.Blank, b.Blank) &&
		a.CompletedSource == b.CompletedSource
}

func strPtrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// buildUnitResultView assembles the view from the unit's catalog identity and
// its progress (nil when not done).
func buildUnitResultView(u *store.UnitRef, p *store.UnitProgress) *UnitResultView {
	v := &UnitResultView{
		UnitID:         u.ID,
		SourceKey:      u.SourceKey,
		SectionOrdinal: u.SectionOrdinal,
		Label:          u.Label,
		Number:         u.Number,
	}
	num := u.Ordinal
	if u.Number != nil {
		num = *u.Number
	}
	ord := u.SectionOrdinal
	v.Notation = FormatRef(ItemRef{
		SourceKey:      u.SourceKey,
		SectionOrdinal: &ord,
		Units:          []RefUnit{{Label: u.Label, Number: num}},
	})
	if p != nil {
		v.Done = true
		v.DoneAt = p.DoneAt
		v.Correct = p.Correct
		v.Wrong = p.Wrong
		v.Blank = p.Blank
	}
	return v
}

// auditUnit appends one audit row for a unit-progress change. Like completions,
// unit results carry no request hash, so request_hash stays NULL.
func (s *Service) auditUnit(ctx context.Context, tx *sql.Tx, source Source, action string, unitID int64, before, after *string, now string) error {
	src := string(source)
	entry := store.AuditEntry{
		AtUTC:      now,
		Source:     src,
		Action:     action,
		Entity:     "unit",
		EntityID:   &unitID,
		BeforeJSON: before,
		AfterJSON:  after,
	}
	return s.st.InsertAudit(ctx, tx, &entry)
}

// unitSnapshotJSON renders a unit's progress as the JSON kept in audit
// before/after.
func unitSnapshotJSON(p *store.UnitProgress) (string, error) {
	type snap struct {
		UnitID          int64   `json:"unit_id"`
		DoneAt          *string `json:"done_at,omitempty"`
		Correct         *int    `json:"correct,omitempty"`
		Wrong           *int    `json:"wrong,omitempty"`
		Blank           *int    `json:"blank,omitempty"`
		CompletedSource string  `json:"completed_source"`
	}
	buf, err := json.Marshal(snap{
		UnitID:          p.UnitID,
		DoneAt:          p.DoneAt,
		Correct:         p.Correct,
		Wrong:           p.Wrong,
		Blank:           p.Blank,
		CompletedSource: p.CompletedSource,
	})
	if err != nil {
		return "", fmt.Errorf("plan: encode unit snapshot: %w", err)
	}
	return string(buf), nil
}
