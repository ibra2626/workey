package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Phase is one of the plan's periods (e.g. "Faz 1 · İşlem temeli"). Dates are
// stored as 'YYYY-MM-DD' strings; the plan layer owns their meaning.
type Phase struct {
	ID        int64
	Key       string
	Ordinal   int
	Name      string
	StartDate string // 'YYYY-MM-DD'
	EndDate   string // 'YYYY-MM-DD'
}

// TopicGroup is a named group of topics with an optional DGS question-weight
// range. The plan references groups by id from sessions.
type TopicGroup struct {
	ID             int64
	Key            string
	Name           string
	Ordinal        int
	DGSQuestionMin *int
	DGSQuestionMax *int
}

// ListPhases returns all phases ordered by ordinal.
func (s *Store) ListPhases(ctx context.Context, q dbtx) ([]Phase, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, key, ordinal, name, start_date, end_date
		FROM phases ORDER BY ordinal`)
	if err != nil {
		return nil, fmt.Errorf("store: query phases: %w", err)
	}
	defer rows.Close()

	var phases []Phase
	for rows.Next() {
		var p Phase
		if err := rows.Scan(&p.ID, &p.Key, &p.Ordinal, &p.Name, &p.StartDate, &p.EndDate); err != nil {
			return nil, fmt.Errorf("store: scan phase: %w", err)
		}
		phases = append(phases, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate phases: %w", err)
	}
	return phases, nil
}

// InsertPhase inserts p and sets p.ID.
func (s *Store) InsertPhase(ctx context.Context, q dbtx, p *Phase) error {
	res, err := q.ExecContext(ctx, `
		INSERT INTO phases (key, ordinal, name, start_date, end_date)
		VALUES (?, ?, ?, ?, ?)`,
		p.Key, p.Ordinal, p.Name, p.StartDate, p.EndDate,
	)
	if err != nil {
		return fmt.Errorf("store: insert phase: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store: insert phase id: %w", err)
	}
	p.ID = id
	return nil
}

// UpdatePhaseByKey updates the mutable fields of the phase identified by p.Key.
func (s *Store) UpdatePhaseByKey(ctx context.Context, q dbtx, p Phase) error {
	if _, err := q.ExecContext(ctx, `
		UPDATE phases SET ordinal = ?, name = ?, start_date = ?, end_date = ?
		WHERE key = ?`,
		p.Ordinal, p.Name, p.StartDate, p.EndDate, p.Key,
	); err != nil {
		return fmt.Errorf("store: update phase %s: %w", p.Key, err)
	}
	return nil
}

// DeletePhaseByKey deletes the phase with the given key.
func (s *Store) DeletePhaseByKey(ctx context.Context, q dbtx, key string) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM phases WHERE key = ?`, key); err != nil {
		return fmt.Errorf("store: delete phase %s: %w", key, err)
	}
	return nil
}

// ListTopicGroups returns all topic groups ordered by ordinal.
func (s *Store) ListTopicGroups(ctx context.Context, q dbtx) ([]TopicGroup, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, key, name, ordinal, dgs_question_min, dgs_question_max
		FROM topic_groups ORDER BY ordinal`)
	if err != nil {
		return nil, fmt.Errorf("store: query topic groups: %w", err)
	}
	defer rows.Close()

	var groups []TopicGroup
	for rows.Next() {
		var g TopicGroup
		if err := rows.Scan(&g.ID, &g.Key, &g.Name, &g.Ordinal, &g.DGSQuestionMin, &g.DGSQuestionMax); err != nil {
			return nil, fmt.Errorf("store: scan topic group: %w", err)
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate topic groups: %w", err)
	}
	return groups, nil
}

// TopicGroupByKey returns the group with the given key, or (nil, nil) if none.
func (s *Store) TopicGroupByKey(ctx context.Context, q dbtx, key string) (*TopicGroup, error) {
	var g TopicGroup
	err := q.QueryRowContext(ctx, `
		SELECT id, key, name, ordinal, dgs_question_min, dgs_question_max
		FROM topic_groups WHERE key = ?`, key,
	).Scan(&g.ID, &g.Key, &g.Name, &g.Ordinal, &g.DGSQuestionMin, &g.DGSQuestionMax)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get topic group %s: %w", key, err)
	}
	return &g, nil
}

// InsertTopicGroup inserts g and sets g.ID.
func (s *Store) InsertTopicGroup(ctx context.Context, q dbtx, g *TopicGroup) error {
	res, err := q.ExecContext(ctx, `
		INSERT INTO topic_groups (key, name, ordinal, dgs_question_min, dgs_question_max)
		VALUES (?, ?, ?, ?, ?)`,
		g.Key, g.Name, g.Ordinal, g.DGSQuestionMin, g.DGSQuestionMax,
	)
	if err != nil {
		return fmt.Errorf("store: insert topic group: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store: insert topic group id: %w", err)
	}
	g.ID = id
	return nil
}

// UpdateTopicGroupByKey updates the mutable fields of the group identified by
// g.Key.
func (s *Store) UpdateTopicGroupByKey(ctx context.Context, q dbtx, g TopicGroup) error {
	if _, err := q.ExecContext(ctx, `
		UPDATE topic_groups SET name = ?, ordinal = ?, dgs_question_min = ?, dgs_question_max = ?
		WHERE key = ?`,
		g.Name, g.Ordinal, g.DGSQuestionMin, g.DGSQuestionMax, g.Key,
	); err != nil {
		return fmt.Errorf("store: update topic group %s: %w", g.Key, err)
	}
	return nil
}

// DeleteTopicGroupByKey deletes the group with the given key.
func (s *Store) DeleteTopicGroupByKey(ctx context.Context, q dbtx, key string) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM topic_groups WHERE key = ?`, key); err != nil {
		return fmt.Errorf("store: delete topic group %s: %w", key, err)
	}
	return nil
}

// CountSessionsByTopicGroup returns how many sessions reference the group id.
func (s *Store) CountSessionsByTopicGroup(ctx context.Context, q dbtx, id int64) (int, error) {
	var n int
	if err := q.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM sessions WHERE topic_group_id = ?`, id,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("store: count sessions by topic group: %w", err)
	}
	return n, nil
}
