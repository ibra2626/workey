package plan

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

const (
	StudyActive        = "active"
	StudyPaused        = "paused"
	StudyClosed        = "closed"
	EntryNormal        = "normal"
	EntryTenMinutes    = "ten_minutes"
	StageKonu          = "konu"
	StageSoru          = "soru"
	StageYanlisDefteri = "yanlis_defteri"
	StageCalisma       = "calisma"
)

var ErrStudyConflict = errors.New("çalışma değişti; sayfayı yenileyip tekrar deneyin")

// StudyMovedError directs the caller to the existing makeup session without
// starting either session. Completion of the original remains available.
type StudyMovedError struct {
	SessionID       int64
	Date, StartTime string
}

func (e *StudyMovedError) Error() string {
	return "bu seans için telafi planlandı; telafi seansını açın"
}

// Bookmark always names the next page to work on, including a partly read page.
// Ordinals are zero based; pages are never automatically incremented.
type Bookmark struct {
	ItemOrdinal int
	Stage       string
	ResumePage  int
}

// StudyContextEntry resolves a persisted study and bookmark against its frozen
// session sources. It is internal service data, not the REST/MCP read contract.
type StudyContextEntry struct {
	Session      store.Session
	Study        store.Study
	BookmarkItem *store.SessionItem
	MovedTo      *SessionRef
}
type StudyContext struct{ Active, LatestPaused *StudyContextEntry }

// ReadStudyContext reads a single snapshot, without starting or moving work.
// Callers give Active priority; LatestPaused must remain an explicit choice.
func (s *Service) ReadStudyContext(ctx context.Context) (*StudyContext, error) {
	var out *StudyContext
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		var err error
		out, err = s.readStudyContextTx(ctx, tx)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// At most two session lookups: one active and one latest paused, independent
// of the number of study rows. Day uses the same transaction as its sessions.
func (s *Service) readStudyContextTx(ctx context.Context, tx *sql.Tx) (*StudyContext, error) {
	out := &StudyContext{}
	studies, err := s.st.OpenStudies(ctx, tx)
	if err != nil {
		return nil, err
	}
	moved, err := s.movedByMissedTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	for _, v := range studies {
		if v.Status == StudyPaused && out.LatestPaused != nil {
			continue
		}
		ss, err := s.st.SessionByID(ctx, tx, v.SessionID)
		if err != nil {
			return nil, err
		}
		if ss == nil {
			return nil, fmt.Errorf("plan: study session %d missing", v.SessionID)
		}
		entry := &StudyContextEntry{Session: *ss, Study: v}
		if v.ItemOrdinal != nil {
			for i := range ss.Items {
				if ss.Items[i].Ordinal == *v.ItemOrdinal {
					it := ss.Items[i]
					entry.BookmarkItem = &it
					break
				}
			}
		}
		if link, ok := moved[ss.ID]; ok {
			entry.MovedTo = &SessionRef{ID: link.SessionID, Date: link.Date, StartTime: link.StartTime}
		}
		if v.Status == StudyActive {
			out.Active = entry
		} else {
			out.LatestPaused = entry
		}
	}
	return out, nil
}

func studyVersion(v *store.Study) int64 {
	if v == nil {
		return 0
	}
	return v.Version
}

// StartStudy opens today/past work. examDate comes from server configuration,
// as with Progress; its zero value means no configured exam. A configured plan
// ends at its last phase. Neither the slot nor planned duration is changed.
func (s *Service) StartStudy(ctx context.Context, source Source, sessionID int64, mode string, expectedVersion int64, examDate clock.Date) (*store.Study, error) {
	if !source.valid() {
		return nil, fmt.Errorf("plan: invalid source %q", source)
	}
	if mode != EntryNormal && mode != EntryTenMinutes {
		return nil, fieldError("entry_mode", CodeInvalid, "geçerli bir başlangıç hedefi seçin")
	}
	if expectedVersion < 0 {
		return nil, fieldError("expected_study_version", CodeInvalid, "geçersiz çalışma sürümü")
	}
	var result *store.Study
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		ss, v, err := s.studyTarget(ctx, tx, sessionID)
		if err != nil {
			return err
		}
		if ss.CompletedAt != nil {
			return ErrStudyConflict
		}
		same := v != nil && v.Status == StudyActive && v.EntryMode == mode
		if !studyVersionMatches(v, expectedVersion, same) {
			return ErrStudyConflict
		}
		if same {
			result = v
			return nil
		}
		today := clock.Today(s.clk).String()
		if ss.Date > today {
			return fieldError("session_id", CodeFuture, "gelecek seans başlatılamaz")
		}
		if err := s.studyPeriodError(ctx, tx, examDate); err != nil {
			return err
		}
		moved, err := s.movedByMissedTx(ctx, tx)
		if err != nil {
			return err
		}
		if link, ok := moved[ss.ID]; ok {
			return &StudyMovedError{SessionID: link.SessionID, Date: link.Date, StartTime: link.StartTime}
		}
		open, err := s.st.OpenStudies(ctx, tx)
		if err != nil {
			return err
		}
		for _, other := range open {
			if other.Status == StudyActive && other.SessionID != ss.ID {
				return ErrStudyConflict
			}
		}
		before := v
		now := s.nowUTC()
		if v == nil {
			v = &store.Study{SessionID: sessionID, FirstStartedAt: now}
		} else {
			copy := *v
			v = &copy
		}
		v.Status, v.EntryMode, v.LastStartedAt, v.UpdatedAt = StudyActive, mode, now, now
		v.Version++
		if before == nil {
			err = s.st.InsertStudy(ctx, tx, v)
		} else {
			err = s.st.UpdateStudy(ctx, tx, v)
		}
		if err != nil {
			return err
		}
		if err := s.auditStudy(ctx, tx, source, "start_study", before, v); err != nil {
			return err
		}
		result = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// StudyPeriodOpen lets the web hide new starts using the same bounds as writes.
// It says nothing about a particular session or whether another study is active.
func (s *Service) StudyPeriodOpen(ctx context.Context, examDate clock.Date) (bool, error) {
	open := false
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		err := s.studyPeriodError(ctx, tx, examDate)
		if _, ok := AsValidationError(err); ok {
			return nil
		}
		open = err == nil
		return err
	})
	return open, err
}

func (s *Service) studyPeriodError(ctx context.Context, tx *sql.Tx, examDate clock.Date) error {
	today := clock.Today(s.clk).String()
	if examDate != (clock.Date{}) && today > examDate.String() {
		return fieldError("session_id", CodeOutOfRange, "sınav sonrası çalışma başlatılamaz")
	}
	phases, err := s.st.ListPhases(ctx, tx)
	if err != nil {
		return err
	}
	end := ""
	for _, p := range phases {
		if p.EndDate > end {
			end = p.EndDate
		}
	}
	if end != "" && today > end {
		return fieldError("session_id", CodeOutOfRange, "plan dönemi sonrası çalışma başlatılamaz")
	}
	return nil
}

// PauseStudy closes active work without marking it completed. A nil bookmark
// preserves the existing bookmark; explicit input can also correct paused work.
// Pausing is allowed after midnight, on rest days and after the plan/exam.
func (s *Service) PauseStudy(ctx context.Context, source Source, sessionID int64, bookmark *Bookmark, expectedVersion int64) (*store.Study, error) {
	if !source.valid() {
		return nil, fmt.Errorf("plan: invalid source %q", source)
	}
	if expectedVersion < 0 {
		return nil, fieldError("expected_study_version", CodeInvalid, "geçersiz çalışma sürümü")
	}
	var result *store.Study
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		ss, v, err := s.studyTarget(ctx, tx, sessionID)
		if err != nil {
			return err
		}
		if v == nil || ss.CompletedAt != nil || v.Status == StudyClosed {
			return ErrStudyConflict
		}
		if bookmark != nil {
			if err := validateBookmark(*ss, *bookmark); err != nil {
				return err
			}
		}
		same := v.Status == StudyPaused && bookmarkMatches(v, bookmark)
		if !studyVersionMatches(v, expectedVersion, same) {
			return ErrStudyConflict
		}
		if same {
			result = v
			return nil
		}
		before := *v
		v.Status, v.UpdatedAt = StudyPaused, s.nowUTC()
		v.Version++
		if bookmark != nil {
			ordinal, stage, page := bookmark.ItemOrdinal, bookmark.Stage, bookmark.ResumePage
			v.ItemOrdinal, v.Stage, v.ResumePage = &ordinal, &stage, &page
		}
		if err := s.st.UpdateStudy(ctx, tx, v); err != nil {
			return err
		}
		if err := s.auditStudy(ctx, tx, source, "pause_study", &before, v); err != nil {
			return err
		}
		result = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) studyTarget(ctx context.Context, tx *sql.Tx, id int64) (*store.Session, *store.Study, error) {
	ss, err := s.st.SessionByID(ctx, tx, id)
	if err != nil {
		return nil, nil, err
	}
	if ss == nil {
		return nil, nil, fieldError("session_id", CodeNotFound, "seans bulunamadı")
	}
	v, err := s.st.StudyBySession(ctx, tx, id)
	return ss, v, err
}

// Only an immediately repeated request may acknowledge an older version. This
// prevents an old tab from succeeding after an active/paused/active ABA cycle.
func studyVersionMatches(v *store.Study, expected int64, same bool) bool {
	current := studyVersion(v)
	return expected == current || (same && expected == current-1)
}
func bookmarkMatches(v *store.Study, b *Bookmark) bool {
	return b == nil || (v.ItemOrdinal != nil && v.Stage != nil && v.ResumePage != nil && *v.ItemOrdinal == b.ItemOrdinal && *v.Stage == b.Stage && *v.ResumePage == b.ResumePage)
}

func validateBookmark(ss store.Session, b Bookmark) error {
	var item *store.SessionItem
	for i := range ss.Items {
		if ss.Items[i].Ordinal == b.ItemOrdinal {
			item = &ss.Items[i]
			break
		}
	}
	if item == nil {
		return fieldError("bookmark.item_ordinal", CodeNotFound, "geçerli bir kaynak seçin")
	}
	from, to := item.PageFrom, item.PageTo
	if ss.Kind == KindAksam {
		switch b.Stage {
		case StageKonu:
			if item.KonuFrom != nil && item.KonuTo != nil {
				from, to = *item.KonuFrom, *item.KonuTo
			}
		case StageSoru:
			if item.SoruFrom != nil && item.SoruTo != nil {
				from, to = *item.SoruFrom, *item.SoruTo
			}
		case StageYanlisDefteri:
		default:
			return fieldError("bookmark.stage", CodeInvalid, "konu, soru veya yanlış defteri aşamasını seçin")
		}
	} else if b.Stage != StageCalisma {
		return fieldError("bookmark.stage", CodeInvalid, "çalışma aşamasını seçin")
	}
	if b.ResumePage < from || b.ResumePage > to {
		return fieldError("bookmark.resume_page", CodeOutOfRange, "devam edeceğin sayfa seçilen kaynak ve aşama aralığında olmalı")
	}
	return nil
}

func (s *Service) auditStudy(ctx context.Context, tx *sql.Tx, source Source, action string, before, after *store.Study) error {
	var beforeJSON *string
	if before != nil {
		b, err := json.Marshal(before)
		if err != nil {
			return err
		}
		value := string(b)
		beforeJSON = &value
	}
	b, err := json.Marshal(after)
	if err != nil {
		return err
	}
	afterJSON := string(b)
	return s.st.InsertAudit(ctx, tx, &store.AuditEntry{AtUTC: after.UpdatedAt, Source: string(source), Action: action, Entity: "session_study", EntityID: &after.SessionID, BeforeJSON: beforeJSON, AfterJSON: &afterJSON})
}
