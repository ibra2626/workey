package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Session is a planned study session with its ordered items. Dates, times and
// timestamps are stored as strings; the plan layer owns their parsing and
// meaning.
type Session struct {
	ID              int64
	Date            string // 'YYYY-MM-DD'
	StartTime       string // 'HH:MM'
	DurationMin     int
	Kind            string // 'aksam' | 'uzun_blok' | 'deneme'
	TopicGroupID    *int64
	MakeupOf        *int64
	CompletedAt     *string // ISO-8601 UTC, or nil
	CompletedSource *string // 'mcp' | 'api' | 'web', or nil
	CreatedAt       string
	UpdatedAt       string
	Items           []SessionItem
}

// SessionItem is one book reference inside a session.
type SessionItem struct {
	ID            int64
	SessionID     int64
	Ordinal       int
	Book          string // 'M1' | 'M2' | 'M3' | 'G1' | 'G2'
	ChapterNo     int
	ChapterTitle  string
	PageFrom      int
	PageTo        int
	KonuFrom      *int
	KonuTo        *int
	SoruFrom      *int
	SoruTo        *int
	AntrenmanFrom *int
	AntrenmanTo   *int
}

// AuditEntry is one row of the change log.
type AuditEntry struct {
	ID          int64
	AtUTC       string // ISO-8601 UTC
	Source      string // 'mcp' | 'api' | 'web'
	Action      string
	Entity      string
	EntityID    *int64
	BeforeJSON  *string
	AfterJSON   *string
	RequestHash *string
}

// IdempotencyRecord stores a request's hash and its stored response so a repeat
// returns the same result without re-running the write.
type IdempotencyRecord struct {
	Key          string
	RequestHash  string
	ResponseJSON string
	CreatedAt    string // ISO-8601 UTC
}

// InsertSession inserts s and sets s.ID to the new row id. It does not insert
// items; call InsertSessionItem for each.
func (s *Store) InsertSession(ctx context.Context, q dbtx, sess *Session) error {
	res, err := q.ExecContext(ctx, `
		INSERT INTO sessions
			(date, start_time, duration_min, kind, topic_group_id, makeup_of,
			 completed_at, completed_source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.Date, sess.StartTime, sess.DurationMin, sess.Kind,
		sess.TopicGroupID, sess.MakeupOf, sess.CompletedAt, sess.CompletedSource,
		sess.CreatedAt, sess.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store: insert session: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store: insert session id: %w", err)
	}
	sess.ID = id
	return nil
}

// UpdateSession updates the mutable fields of the session with sess.ID. It does
// not touch items; callers replace items with DeleteSessionItems +
// InsertSessionItem.
func (s *Store) UpdateSession(ctx context.Context, q dbtx, sess *Session) error {
	if _, err := q.ExecContext(ctx, `
		UPDATE sessions SET
			duration_min = ?, kind = ?, topic_group_id = ?, makeup_of = ?, updated_at = ?
		WHERE id = ?`,
		sess.DurationMin, sess.Kind, sess.TopicGroupID, sess.MakeupOf,
		sess.UpdatedAt, sess.ID,
	); err != nil {
		return fmt.Errorf("store: update session %d: %w", sess.ID, err)
	}
	return nil
}

// DeleteSession deletes the session with the given id. Its items are removed by
// the ON DELETE CASCADE foreign key.
func (s *Store) DeleteSession(ctx context.Context, q dbtx, id int64) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id); err != nil {
		return fmt.Errorf("store: delete session %d: %w", id, err)
	}
	return nil
}

// DeleteSessionItems removes every item of a session, so the caller can
// re-insert the new set in one transaction.
func (s *Store) DeleteSessionItems(ctx context.Context, q dbtx, sessionID int64) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM session_items WHERE session_id = ?`, sessionID); err != nil {
		return fmt.Errorf("store: delete session items of %d: %w", sessionID, err)
	}
	return nil
}

// SessionByDateAndTime returns the single session at (date, start_time) with its
// items, or (nil, nil) if there is none. UNIQUE(date, start_time) guarantees at
// most one.
func (s *Store) SessionByDateAndTime(ctx context.Context, q dbtx, date, startTime string) (*Session, error) {
	var ss Session
	err := q.QueryRowContext(ctx, `
		SELECT id, date, start_time, duration_min, kind, topic_group_id, makeup_of,
		       completed_at, completed_source, created_at, updated_at
		FROM sessions WHERE date = ? AND start_time = ?`, date, startTime,
	).Scan(
		&ss.ID, &ss.Date, &ss.StartTime, &ss.DurationMin, &ss.Kind,
		&ss.TopicGroupID, &ss.MakeupOf, &ss.CompletedAt, &ss.CompletedSource,
		&ss.CreatedAt, &ss.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get session %s %s: %w", date, startTime, err)
	}

	itemRows, err := q.QueryContext(ctx, `
		SELECT id, session_id, ordinal, book, chapter_no, chapter_title,
		       page_from, page_to, konu_from, konu_to, soru_from, soru_to,
		       antrenman_from, antrenman_to
		FROM session_items WHERE session_id = ? ORDER BY ordinal`, ss.ID)
	if err != nil {
		return nil, fmt.Errorf("store: query items of %d: %w", ss.ID, err)
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var it SessionItem
		if err := itemRows.Scan(
			&it.ID, &it.SessionID, &it.Ordinal, &it.Book, &it.ChapterNo, &it.ChapterTitle,
			&it.PageFrom, &it.PageTo, &it.KonuFrom, &it.KonuTo, &it.SoruFrom, &it.SoruTo,
			&it.AntrenmanFrom, &it.AntrenmanTo,
		); err != nil {
			return nil, fmt.Errorf("store: scan item of %d: %w", ss.ID, err)
		}
		ss.Items = append(ss.Items, it)
	}
	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate items of %d: %w", ss.ID, err)
	}
	return &ss, nil
}

// InsertSessionItem inserts it under its SessionID and sets it.ID.
func (s *Store) InsertSessionItem(ctx context.Context, q dbtx, it *SessionItem) error {
	res, err := q.ExecContext(ctx, `
		INSERT INTO session_items
			(session_id, ordinal, book, chapter_no, chapter_title,
			 page_from, page_to, konu_from, konu_to, soru_from, soru_to,
			 antrenman_from, antrenman_to)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		it.SessionID, it.Ordinal, it.Book, it.ChapterNo, it.ChapterTitle,
		it.PageFrom, it.PageTo, it.KonuFrom, it.KonuTo, it.SoruFrom, it.SoruTo,
		it.AntrenmanFrom, it.AntrenmanTo,
	)
	if err != nil {
		return fmt.Errorf("store: insert session item: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store: insert session item id: %w", err)
	}
	it.ID = id
	return nil
}

// SessionsByDateRange returns sessions whose date is within [from, to]
// inclusive, ordered by date then start time, each with its items ordered by
// ordinal. q may be the store's DB (via DB()) or a transaction.
func (s *Store) SessionsByDateRange(ctx context.Context, q dbtx, from, to string) ([]Session, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, date, start_time, duration_min, kind, topic_group_id, makeup_of,
		       completed_at, completed_source, created_at, updated_at
		FROM sessions
		WHERE date >= ? AND date <= ?
		ORDER BY date, start_time`, from, to)
	if err != nil {
		return nil, fmt.Errorf("store: query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	index := map[int64]int{}
	for rows.Next() {
		var ss Session
		if err := rows.Scan(
			&ss.ID, &ss.Date, &ss.StartTime, &ss.DurationMin, &ss.Kind,
			&ss.TopicGroupID, &ss.MakeupOf, &ss.CompletedAt, &ss.CompletedSource,
			&ss.CreatedAt, &ss.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan session: %w", err)
		}
		index[ss.ID] = len(sessions)
		sessions = append(sessions, ss)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate sessions: %w", err)
	}
	if len(sessions) == 0 {
		return nil, nil
	}

	itemRows, err := q.QueryContext(ctx, `
		SELECT i.id, i.session_id, i.ordinal, i.book, i.chapter_no, i.chapter_title,
		       i.page_from, i.page_to, i.konu_from, i.konu_to, i.soru_from, i.soru_to,
		       i.antrenman_from, i.antrenman_to
		FROM session_items i
		JOIN sessions s ON s.id = i.session_id
		WHERE s.date >= ? AND s.date <= ?
		ORDER BY i.session_id, i.ordinal`, from, to)
	if err != nil {
		return nil, fmt.Errorf("store: query session items: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var it SessionItem
		if err := itemRows.Scan(
			&it.ID, &it.SessionID, &it.Ordinal, &it.Book, &it.ChapterNo, &it.ChapterTitle,
			&it.PageFrom, &it.PageTo, &it.KonuFrom, &it.KonuTo, &it.SoruFrom, &it.SoruTo,
			&it.AntrenmanFrom, &it.AntrenmanTo,
		); err != nil {
			return nil, fmt.Errorf("store: scan session item: %w", err)
		}
		if i, ok := index[it.SessionID]; ok {
			sessions[i].Items = append(sessions[i].Items, it)
		}
	}
	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate session items: %w", err)
	}
	return sessions, nil
}

// SessionByID returns the session with the given id and its ordered items, or
// (nil, nil) if there is none.
func (s *Store) SessionByID(ctx context.Context, q dbtx, id int64) (*Session, error) {
	var ss Session
	err := q.QueryRowContext(ctx, `
		SELECT id, date, start_time, duration_min, kind, topic_group_id, makeup_of,
		       completed_at, completed_source, created_at, updated_at
		FROM sessions WHERE id = ?`, id,
	).Scan(
		&ss.ID, &ss.Date, &ss.StartTime, &ss.DurationMin, &ss.Kind,
		&ss.TopicGroupID, &ss.MakeupOf, &ss.CompletedAt, &ss.CompletedSource,
		&ss.CreatedAt, &ss.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get session %d: %w", id, err)
	}
	items, err := s.sessionItems(ctx, q, ss.ID)
	if err != nil {
		return nil, err
	}
	ss.Items = items
	return &ss, nil
}

// sessionItems returns the items of one session ordered by ordinal.
func (s *Store) sessionItems(ctx context.Context, q dbtx, sessionID int64) ([]SessionItem, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, session_id, ordinal, book, chapter_no, chapter_title,
		       page_from, page_to, konu_from, konu_to, soru_from, soru_to,
		       antrenman_from, antrenman_to
		FROM session_items WHERE session_id = ? ORDER BY ordinal`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("store: query items of %d: %w", sessionID, err)
	}
	defer rows.Close()
	var items []SessionItem
	for rows.Next() {
		var it SessionItem
		if err := rows.Scan(
			&it.ID, &it.SessionID, &it.Ordinal, &it.Book, &it.ChapterNo, &it.ChapterTitle,
			&it.PageFrom, &it.PageTo, &it.KonuFrom, &it.KonuTo, &it.SoruFrom, &it.SoruTo,
			&it.AntrenmanFrom, &it.AntrenmanTo,
		); err != nil {
			return nil, fmt.Errorf("store: scan item of %d: %w", sessionID, err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate items of %d: %w", sessionID, err)
	}
	return items, nil
}

// AllSessions returns every session ordered by date then start time, each with
// its items ordered by ordinal. It powers the plan-wide read models (progress,
// the 38-week strip) for this single-user app.
func (s *Store) AllSessions(ctx context.Context, q dbtx) ([]Session, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, date, start_time, duration_min, kind, topic_group_id, makeup_of,
		       completed_at, completed_source, created_at, updated_at
		FROM sessions
		ORDER BY date, start_time`)
	if err != nil {
		return nil, fmt.Errorf("store: query all sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	index := map[int64]int{}
	for rows.Next() {
		var ss Session
		if err := rows.Scan(
			&ss.ID, &ss.Date, &ss.StartTime, &ss.DurationMin, &ss.Kind,
			&ss.TopicGroupID, &ss.MakeupOf, &ss.CompletedAt, &ss.CompletedSource,
			&ss.CreatedAt, &ss.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan session: %w", err)
		}
		index[ss.ID] = len(sessions)
		sessions = append(sessions, ss)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate sessions: %w", err)
	}
	if len(sessions) == 0 {
		return nil, nil
	}

	itemRows, err := q.QueryContext(ctx, `
		SELECT id, session_id, ordinal, book, chapter_no, chapter_title,
		       page_from, page_to, konu_from, konu_to, soru_from, soru_to,
		       antrenman_from, antrenman_to
		FROM session_items
		ORDER BY session_id, ordinal`)
	if err != nil {
		return nil, fmt.Errorf("store: query all session items: %w", err)
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var it SessionItem
		if err := itemRows.Scan(
			&it.ID, &it.SessionID, &it.Ordinal, &it.Book, &it.ChapterNo, &it.ChapterTitle,
			&it.PageFrom, &it.PageTo, &it.KonuFrom, &it.KonuTo, &it.SoruFrom, &it.SoruTo,
			&it.AntrenmanFrom, &it.AntrenmanTo,
		); err != nil {
			return nil, fmt.Errorf("store: scan session item: %w", err)
		}
		if i, ok := index[it.SessionID]; ok {
			sessions[i].Items = append(sessions[i].Items, it)
		}
	}
	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate session items: %w", err)
	}
	return sessions, nil
}

// SetSessionCompletion updates a session's completion fields and updated_at.
// Pass non-nil completedAt and completedSource to mark it done, or nil/nil to
// clear completion (undo).
func (s *Store) SetSessionCompletion(ctx context.Context, q dbtx, id int64, completedAt, completedSource *string, updatedAt string) error {
	if _, err := q.ExecContext(ctx, `
		UPDATE sessions SET completed_at = ?, completed_source = ?, updated_at = ?
		WHERE id = ?`,
		completedAt, completedSource, updatedAt, id,
	); err != nil {
		return fmt.Errorf("store: set completion %d: %w", id, err)
	}
	return nil
}

// MakeupLink records that a session (SessionID at Date/StartTime) makes up the
// missed session MakeupOf.
type MakeupLink struct {
	SessionID int64
	Date      string
	StartTime string
	MakeupOf  int64
}

// ListMakeupLinks returns every makeup link (sessions whose makeup_of is set),
// ordered by date then start time. The plan layer uses these to mark a missed
// session as "moved" and to name the day it was moved to.
func (s *Store) ListMakeupLinks(ctx context.Context, q dbtx) ([]MakeupLink, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, date, start_time, makeup_of
		FROM sessions WHERE makeup_of IS NOT NULL
		ORDER BY date, start_time`)
	if err != nil {
		return nil, fmt.Errorf("store: query makeup links: %w", err)
	}
	defer rows.Close()
	var links []MakeupLink
	for rows.Next() {
		var l MakeupLink
		if err := rows.Scan(&l.SessionID, &l.Date, &l.StartTime, &l.MakeupOf); err != nil {
			return nil, fmt.Errorf("store: scan makeup link: %w", err)
		}
		links = append(links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate makeup links: %w", err)
	}
	return links, nil
}

// InsertAudit appends an audit row and sets a.ID.
func (s *Store) InsertAudit(ctx context.Context, q dbtx, a *AuditEntry) error {
	res, err := q.ExecContext(ctx, `
		INSERT INTO audit_log
			(at_utc, source, action, entity, entity_id, before_json, after_json, request_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.AtUTC, a.Source, a.Action, a.Entity, a.EntityID,
		a.BeforeJSON, a.AfterJSON, a.RequestHash,
	)
	if err != nil {
		return fmt.Errorf("store: insert audit: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store: insert audit id: %w", err)
	}
	a.ID = id
	return nil
}

// GetIdempotency returns the record for key, or (nil, nil) if none exists.
func (s *Store) GetIdempotency(ctx context.Context, q dbtx, key string) (*IdempotencyRecord, error) {
	var r IdempotencyRecord
	err := q.QueryRowContext(ctx, `
		SELECT key, request_hash, response_json, created_at
		FROM idempotency_keys WHERE key = ?`, key,
	).Scan(&r.Key, &r.RequestHash, &r.ResponseJSON, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get idempotency key: %w", err)
	}
	return &r, nil
}

// PutIdempotency stores r. The key is the primary key, so a repeated key is an
// error (INSERT, not upsert) — callers check with GetIdempotency first.
func (s *Store) PutIdempotency(ctx context.Context, q dbtx, r IdempotencyRecord) error {
	if _, err := q.ExecContext(ctx, `
		INSERT INTO idempotency_keys (key, request_hash, response_json, created_at)
		VALUES (?, ?, ?, ?)`,
		r.Key, r.RequestHash, r.ResponseJSON, r.CreatedAt,
	); err != nil {
		return fmt.Errorf("store: put idempotency key: %w", err)
	}
	return nil
}
