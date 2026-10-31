package plan

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

// Completion audit actions.
const (
	actionComplete = "complete"
	actionUndo     = "undo"
)

// Complete marks the session done. It is allowed for today and past sessions
// and rejected for future ones (a typed validation error). The call is
// idempotent: completing an already-completed session changes nothing and
// writes no audit row. Every real change is audited with its source and
// before/after JSON.
// An optional expected study version protects web writes. Omitting it keeps
// legacy REST/MCP behavior: those clients cannot detect stale study intent.
func (s *Service) Complete(ctx context.Context, source Source, sessionID int64, expectedStudyVersion ...int64) (*SessionView, error) {
	return s.setCompletion(ctx, source, sessionID, true, expectedStudyVersion)
}

// Undo clears a session's completion, under the same today-or-past rule and the
// same idempotency and audit guarantees as Complete. Undoing a session that is
// not completed changes nothing and writes no audit row.
func (s *Service) Undo(ctx context.Context, source Source, sessionID int64, expectedStudyVersion ...int64) (*SessionView, error) {
	return s.setCompletion(ctx, source, sessionID, false, expectedStudyVersion)
}

// setCompletion is the shared body of Complete (done=true) and Undo
// (done=false).
func (s *Service) setCompletion(ctx context.Context, source Source, sessionID int64, done bool, expected []int64) (*SessionView, error) {
	if !source.valid() {
		return nil, fmt.Errorf("plan: invalid source %q", source)
	}
	if len(expected) > 1 || (len(expected) == 1 && expected[0] < 0) {
		return nil, fieldError("expected_study_version", CodeInvalid, "geçersiz çalışma sürümü")
	}
	today := clock.Today(s.clk).String()
	now := s.nowUTC()

	var view *SessionView
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		ss, err := s.st.SessionByID(ctx, tx, sessionID)
		if err != nil {
			return err
		}
		if ss == nil {
			return fieldError("session_id", CodeNotFound, "session does not exist")
		}
		if ss.Date > today {
			return fieldError("session_id", CodeFuture, "a future session cannot be marked")
		}
		study, err := s.st.StudyBySession(ctx, tx, sessionID)
		if err != nil {
			return err
		}
		alreadyDone := ss.CompletedAt != nil
		same := done == alreadyDone
		if study != nil {
			wanted := StudyPaused
			if done {
				wanted = StudyClosed
			}
			same = same && study.Status == wanted
		}
		if len(expected) == 1 && !studyVersionMatches(study, expected[0], same) {
			return ErrStudyConflict
		}

		tgName, tgKey, err := s.topicGroupNamesTx(ctx, tx)
		if err != nil {
			return err
		}
		movedBy, err := s.movedByMissedTx(ctx, tx)
		if err != nil {
			return err
		}

		if done == alreadyDone {
			// No-op: already in the requested state. No write, no audit row.
			sv := buildSessionView(*ss, today, movedBy, tgName, tgKey)
			attachStudy(&sv, study, ss.Items)
			view = &sv
			return nil
		}

		before, err := snapshotJSON(*ss)
		if err != nil {
			return err
		}

		action := actionUndo
		if done {
			action = actionComplete
			src := string(source)
			ss.CompletedAt = &now
			ss.CompletedSource = &src
		} else {
			ss.CompletedAt = nil
			ss.CompletedSource = nil
		}
		ss.UpdatedAt = now
		if err := s.st.SetSessionCompletion(ctx, tx, ss.ID, ss.CompletedAt, ss.CompletedSource, now); err != nil {
			return err
		}
		if study != nil {
			beforeStudy := *study
			study.Status = StudyPaused
			if done {
				study.Status = StudyClosed
			}
			study.UpdatedAt = now
			study.Version++
			if err := s.st.UpdateStudy(ctx, tx, study); err != nil {
				return err
			}
			if err := s.auditStudy(ctx, tx, source, action, &beforeStudy, study); err != nil {
				return err
			}
		}

		after, err := snapshotJSON(*ss)
		if err != nil {
			return err
		}
		if err := s.auditCompletion(ctx, tx, source, action, ss.ID, before, after, now); err != nil {
			return err
		}

		sv := buildSessionView(*ss, today, movedBy, tgName, tgKey)
		attachStudy(&sv, study, ss.Items)
		view = &sv
		return nil
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

// auditCompletion appends one audit row for a completion change. Completions
// carry no request hash (unlike plan writes), so request_hash stays NULL.
func (s *Service) auditCompletion(ctx context.Context, tx *sql.Tx, source Source, action string, id int64, before, after, now string) error {
	src := string(source)
	entry := store.AuditEntry{
		AtUTC:      now,
		Source:     src,
		Action:     action,
		Entity:     "session",
		EntityID:   &id,
		BeforeJSON: &before,
		AfterJSON:  &after,
	}
	return s.st.InsertAudit(ctx, tx, &entry)
}

// topicGroupNamesTx is topicGroupNames bound to a transaction.
func (s *Service) topicGroupNamesTx(ctx context.Context, tx *sql.Tx) (name, key map[int64]string, err error) {
	groups, err := s.st.ListTopicGroups(ctx, tx)
	if err != nil {
		return nil, nil, err
	}
	name = map[int64]string{}
	key = map[int64]string{}
	for _, g := range groups {
		name[g.ID] = g.Name
		key[g.ID] = g.Key
	}
	return name, key, nil
}

// movedByMissedTx is movedByMissed bound to a transaction.
func (s *Service) movedByMissedTx(ctx context.Context, tx *sql.Tx) (map[int64]store.MakeupLink, error) {
	links, err := s.st.ListMakeupLinks(ctx, tx)
	if err != nil {
		return nil, err
	}
	byMissed := make(map[int64]store.MakeupLink, len(links))
	for _, l := range links {
		byMissed[l.MakeupOf] = l
	}
	return byMissed, nil
}
