package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Study is the persisted lifecycle of a started session. Bookmark fields are
// either all nil or all set. The plan layer owns lifecycle and source rules.
type Study struct {
	SessionID      int64
	Status         string
	FirstStartedAt string
	LastStartedAt  string
	UpdatedAt      string
	EntryMode      string
	ItemOrdinal    *int
	Stage          *string
	ResumePage     *int
	Version        int64
}

const studyColumns = `session_id, status, first_started_at, last_started_at,
	updated_at, entry_mode, item_ordinal, stage, resume_page, version`

// StudiesByDateRange batches all study states for the caller's read snapshot.
// It avoids a per-session query on plan/day/week reads.
func (s *Store) StudiesByDateRange(ctx context.Context, q dbtx, from, to string) (map[int64]*Study, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+studyColumns+` FROM session_study
		WHERE session_id IN (SELECT id FROM sessions WHERE date BETWEEN ? AND ?)`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]*Study{}
	for rows.Next() {
		var v Study
		if err := rows.Scan(&v.SessionID, &v.Status, &v.FirstStartedAt, &v.LastStartedAt,
			&v.UpdatedAt, &v.EntryMode, &v.ItemOrdinal, &v.Stage, &v.ResumePage, &v.Version); err != nil {
			return nil, err
		}
		out[v.SessionID] = &v
	}
	return out, rows.Err()
}

// OpenStudies returns active first, then paused studies, newest update first.
// Audit order breaks equal-second timestamps, then session id handles old rows
// without study audit history deterministically.
func (s *Store) OpenStudies(ctx context.Context, q dbtx) ([]Study, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+studyColumns+` FROM session_study
		WHERE status IN ('active', 'paused')
		ORDER BY CASE status WHEN 'active' THEN 0 ELSE 1 END, updated_at DESC,
		(SELECT MAX(id) FROM audit_log WHERE entity = 'session_study' AND entity_id = session_id) DESC,
		session_id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Study
	for rows.Next() {
		var v Study
		if err := rows.Scan(&v.SessionID, &v.Status, &v.FirstStartedAt, &v.LastStartedAt,
			&v.UpdatedAt, &v.EntryMode, &v.ItemOrdinal, &v.Stage, &v.ResumePage, &v.Version); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// StudyBySession returns nil when the session has never been started.
func (s *Store) StudyBySession(ctx context.Context, q dbtx, sessionID int64) (*Study, error) {
	var v Study
	err := q.QueryRowContext(ctx, `SELECT `+studyColumns+` FROM session_study WHERE session_id = ?`, sessionID).
		Scan(&v.SessionID, &v.Status, &v.FirstStartedAt, &v.LastStartedAt,
			&v.UpdatedAt, &v.EntryMode, &v.ItemOrdinal, &v.Stage, &v.ResumePage, &v.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: read study %d: %w", sessionID, err)
	}
	return &v, nil
}

// StartedSessionIDs returns all sessions with a study record in the date range,
// including paused/closed records. It uses the caller's transaction snapshot.
func (s *Store) StartedSessionIDs(ctx context.Context, q dbtx, from, to string) (map[int64]bool, error) {
	rows, err := q.QueryContext(ctx, `SELECT ss.session_id FROM session_study ss
		JOIN sessions s ON s.id = ss.session_id WHERE s.date BETWEEN ? AND ?`, from, to)
	if err != nil {
		return nil, fmt.Errorf("store: read started sessions: %w", err)
	}
	defer rows.Close()
	ids := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("store: scan started session: %w", err)
		}
		ids[id] = true
	}
	return ids, rows.Err()
}

func (s *Store) InsertStudy(ctx context.Context, q dbtx, v *Study) error {
	_, err := q.ExecContext(ctx, `INSERT INTO session_study (`+studyColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.SessionID, v.Status, v.FirstStartedAt, v.LastStartedAt, v.UpdatedAt,
		v.EntryMode, v.ItemOrdinal, v.Stage, v.ResumePage, v.Version)
	if err != nil {
		return fmt.Errorf("store: insert study %d: %w", v.SessionID, err)
	}
	return nil
}

// UpdateStudy writes an existing study. Callers own version and state checks.
func (s *Store) UpdateStudy(ctx context.Context, q dbtx, v *Study) error {
	res, err := q.ExecContext(ctx, `UPDATE session_study SET status = ?, first_started_at = ?,
		last_started_at = ?, updated_at = ?, entry_mode = ?, item_ordinal = ?, stage = ?,
		resume_page = ?, version = ? WHERE session_id = ?`,
		v.Status, v.FirstStartedAt, v.LastStartedAt, v.UpdatedAt, v.EntryMode,
		v.ItemOrdinal, v.Stage, v.ResumePage, v.Version, v.SessionID)
	if err != nil {
		return fmt.Errorf("store: update study %d: %w", v.SessionID, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: update study rows: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("store: update study %d: %w", v.SessionID, sql.ErrNoRows)
	}
	return nil
}
