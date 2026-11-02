package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Source is one book / video / course the user studies. Key is a short
// user-given code (no fixed list); LocationUnit names how ranges are counted
// ('page' | 'minute' | 'lesson' | 'other'). The plan layer owns the meaning.
type Source struct {
	ID           int64
	Key          string
	Name         string
	Kind         string // 'book' | 'video' | 'course' | 'other'
	LocationUnit string // 'page' | 'minute' | 'lesson' | 'other'
	Ordinal      int
	CreatedAt    string // ISO-8601 UTC
	UpdatedAt    string // ISO-8601 UTC
}

// Section is an ordered part of a source with an optional location range in the
// source's LocationUnit.
type Section struct {
	ID       int64
	SourceID int64
	Ordinal  int
	Name     string
	LocFrom  *int
	LocTo    *int
}

// Unit is an ordered study piece inside a section. Label is the source's own
// word ("Antrenman" / "Test" / "Föy" / "Ders"); the code never enumerates them.
type Unit struct {
	ID        int64
	SectionID int64
	Ordinal   int
	Label     string
	Number    *int
	LocFrom   *int
	LocTo     *int
}

// UnitProgress is a unit's study result, stored apart from the catalog so a
// source re-write never wipes it. The presence of a row means the unit is done;
// Correct/Wrong/Blank (D/Y/B) are optional.
type UnitProgress struct {
	UnitID          int64
	DoneAt          *string // ISO-8601 UTC, or nil
	Correct         *int
	Wrong           *int
	Blank           *int
	CompletedSource string // 'mcp' | 'api' | 'web'
}

// ListSources returns all sources ordered by ordinal.
func (s *Store) ListSources(ctx context.Context, q dbtx) ([]Source, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, key, name, kind, location_unit, ordinal, created_at, updated_at
		FROM sources ORDER BY ordinal`)
	if err != nil {
		return nil, fmt.Errorf("store: query sources: %w", err)
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var src Source
		if err := rows.Scan(&src.ID, &src.Key, &src.Name, &src.Kind,
			&src.LocationUnit, &src.Ordinal, &src.CreatedAt, &src.UpdatedAt); err != nil {
			return nil, fmt.Errorf("store: scan source: %w", err)
		}
		sources = append(sources, src)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate sources: %w", err)
	}
	return sources, nil
}

// SourceByKey returns the source with the given key, or (nil, nil) if none.
func (s *Store) SourceByKey(ctx context.Context, q dbtx, key string) (*Source, error) {
	var src Source
	err := q.QueryRowContext(ctx, `
		SELECT id, key, name, kind, location_unit, ordinal, created_at, updated_at
		FROM sources WHERE key = ?`, key,
	).Scan(&src.ID, &src.Key, &src.Name, &src.Kind,
		&src.LocationUnit, &src.Ordinal, &src.CreatedAt, &src.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get source %s: %w", key, err)
	}
	return &src, nil
}

// InsertSource inserts src and sets src.ID.
func (s *Store) InsertSource(ctx context.Context, q dbtx, src *Source) error {
	res, err := q.ExecContext(ctx, `
		INSERT INTO sources (key, name, kind, location_unit, ordinal, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		src.Key, src.Name, src.Kind, src.LocationUnit, src.Ordinal, src.CreatedAt, src.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("store: insert source: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store: insert source id: %w", err)
	}
	src.ID = id
	return nil
}

// UpdateSourceByKey updates the mutable fields of the source identified by
// src.Key. The key itself is not changed.
func (s *Store) UpdateSourceByKey(ctx context.Context, q dbtx, src Source) error {
	if _, err := q.ExecContext(ctx, `
		UPDATE sources SET name = ?, kind = ?, location_unit = ?, ordinal = ?, updated_at = ?
		WHERE key = ?`,
		src.Name, src.Kind, src.LocationUnit, src.Ordinal, src.UpdatedAt, src.Key,
	); err != nil {
		return fmt.Errorf("store: update source %s: %w", src.Key, err)
	}
	return nil
}

// DeleteSourceByKey deletes the source with the given key. Its sections, units
// and their unit_progress are removed by ON DELETE CASCADE.
func (s *Store) DeleteSourceByKey(ctx context.Context, q dbtx, key string) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM sources WHERE key = ?`, key); err != nil {
		return fmt.Errorf("store: delete source %s: %w", key, err)
	}
	return nil
}

// SectionsBySource returns a source's sections ordered by ordinal.
func (s *Store) SectionsBySource(ctx context.Context, q dbtx, sourceID int64) ([]Section, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, source_id, ordinal, name, loc_from, loc_to
		FROM sections WHERE source_id = ? ORDER BY ordinal`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("store: query sections of %d: %w", sourceID, err)
	}
	defer rows.Close()
	var sections []Section
	for rows.Next() {
		var sec Section
		if err := rows.Scan(&sec.ID, &sec.SourceID, &sec.Ordinal, &sec.Name, &sec.LocFrom, &sec.LocTo); err != nil {
			return nil, fmt.Errorf("store: scan section of %d: %w", sourceID, err)
		}
		sections = append(sections, sec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate sections of %d: %w", sourceID, err)
	}
	return sections, nil
}

// InsertSection inserts sec and sets sec.ID.
func (s *Store) InsertSection(ctx context.Context, q dbtx, sec *Section) error {
	res, err := q.ExecContext(ctx, `
		INSERT INTO sections (source_id, ordinal, name, loc_from, loc_to)
		VALUES (?, ?, ?, ?, ?)`,
		sec.SourceID, sec.Ordinal, sec.Name, sec.LocFrom, sec.LocTo,
	)
	if err != nil {
		return fmt.Errorf("store: insert section: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store: insert section id: %w", err)
	}
	sec.ID = id
	return nil
}

// UpdateSection updates the mutable fields of the section with sec.ID, keeping
// its id (and so any units and their progress) intact.
func (s *Store) UpdateSection(ctx context.Context, q dbtx, sec Section) error {
	res, err := q.ExecContext(ctx, `
		UPDATE sections SET ordinal = ?, name = ?, loc_from = ?, loc_to = ?
		WHERE id = ?`,
		sec.Ordinal, sec.Name, sec.LocFrom, sec.LocTo, sec.ID,
	)
	if err != nil {
		return fmt.Errorf("store: update section %d: %w", sec.ID, err)
	}
	if err := requireOne(res, "update section", sec.ID); err != nil {
		return err
	}
	return nil
}

// DeleteSection deletes the section with the given id (cascading to its units
// and their progress).
func (s *Store) DeleteSection(ctx context.Context, q dbtx, id int64) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM sections WHERE id = ?`, id); err != nil {
		return fmt.Errorf("store: delete section %d: %w", id, err)
	}
	return nil
}

// UnitsBySection returns a section's units ordered by ordinal.
func (s *Store) UnitsBySection(ctx context.Context, q dbtx, sectionID int64) ([]Unit, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, section_id, ordinal, label, number, loc_from, loc_to
		FROM units WHERE section_id = ? ORDER BY ordinal`, sectionID)
	if err != nil {
		return nil, fmt.Errorf("store: query units of %d: %w", sectionID, err)
	}
	defer rows.Close()
	return scanUnits(rows, sectionID)
}

// UnitsBySource returns every unit of a source (across its sections), ordered by
// section ordinal then unit ordinal. It joins through sections so a reader does
// not need a per-section query.
func (s *Store) UnitsBySource(ctx context.Context, q dbtx, sourceID int64) ([]Unit, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT u.id, u.section_id, u.ordinal, u.label, u.number, u.loc_from, u.loc_to
		FROM units u JOIN sections sec ON sec.id = u.section_id
		WHERE sec.source_id = ?
		ORDER BY sec.ordinal, u.ordinal`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("store: query units of source %d: %w", sourceID, err)
	}
	defer rows.Close()
	return scanUnits(rows, sourceID)
}

func scanUnits(rows *sql.Rows, id int64) ([]Unit, error) {
	var units []Unit
	for rows.Next() {
		var u Unit
		if err := rows.Scan(&u.ID, &u.SectionID, &u.Ordinal, &u.Label, &u.Number, &u.LocFrom, &u.LocTo); err != nil {
			return nil, fmt.Errorf("store: scan unit of %d: %w", id, err)
		}
		units = append(units, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate units of %d: %w", id, err)
	}
	return units, nil
}

// InsertUnit inserts u and sets u.ID.
func (s *Store) InsertUnit(ctx context.Context, q dbtx, u *Unit) error {
	res, err := q.ExecContext(ctx, `
		INSERT INTO units (section_id, ordinal, label, number, loc_from, loc_to)
		VALUES (?, ?, ?, ?, ?, ?)`,
		u.SectionID, u.Ordinal, u.Label, u.Number, u.LocFrom, u.LocTo,
	)
	if err != nil {
		return fmt.Errorf("store: insert unit: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store: insert unit id: %w", err)
	}
	u.ID = id
	return nil
}

// UpdateUnit updates the mutable fields of the unit with u.ID. Keeping the id
// stable is how a source content re-write preserves the unit's progress.
func (s *Store) UpdateUnit(ctx context.Context, q dbtx, u Unit) error {
	res, err := q.ExecContext(ctx, `
		UPDATE units SET section_id = ?, ordinal = ?, label = ?, number = ?, loc_from = ?, loc_to = ?
		WHERE id = ?`,
		u.SectionID, u.Ordinal, u.Label, u.Number, u.LocFrom, u.LocTo, u.ID,
	)
	if err != nil {
		return fmt.Errorf("store: update unit %d: %w", u.ID, err)
	}
	if err := requireOne(res, "update unit", u.ID); err != nil {
		return err
	}
	return nil
}

// DeleteUnit deletes the unit with the given id (cascading to its progress).
func (s *Store) DeleteUnit(ctx context.Context, q dbtx, id int64) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM units WHERE id = ?`, id); err != nil {
		return fmt.Errorf("store: delete unit %d: %w", id, err)
	}
	return nil
}

// ReplaceSections deletes all of a source's sections (cascading to their units
// and progress) and inserts the given set, setting each Section.ID and
// SourceID. This loses unit progress by design; the plan layer uses the
// per-section/per-unit update helpers when it must keep unit ids and their
// progress. The caller owns the transaction.
func (s *Store) ReplaceSections(ctx context.Context, q dbtx, sourceID int64, sections []Section) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM sections WHERE source_id = ?`, sourceID); err != nil {
		return fmt.Errorf("store: clear sections of %d: %w", sourceID, err)
	}
	for i := range sections {
		sections[i].SourceID = sourceID
		if err := s.InsertSection(ctx, q, &sections[i]); err != nil {
			return err
		}
	}
	return nil
}

// ReplaceUnits deletes all of a section's units (cascading to their progress)
// and inserts the given set, setting each Unit.ID and SectionID. Like
// ReplaceSections this drops progress by design. The caller owns the transaction.
func (s *Store) ReplaceUnits(ctx context.Context, q dbtx, sectionID int64, units []Unit) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM units WHERE section_id = ?`, sectionID); err != nil {
		return fmt.Errorf("store: clear units of %d: %w", sectionID, err)
	}
	for i := range units {
		units[i].SectionID = sectionID
		if err := s.InsertUnit(ctx, q, &units[i]); err != nil {
			return err
		}
	}
	return nil
}

// UnitProgressByUnit returns a unit's progress, or (nil, nil) if the unit is not
// done.
func (s *Store) UnitProgressByUnit(ctx context.Context, q dbtx, unitID int64) (*UnitProgress, error) {
	var p UnitProgress
	err := q.QueryRowContext(ctx, `
		SELECT unit_id, done_at, correct, wrong, blank, completed_source
		FROM unit_progress WHERE unit_id = ?`, unitID,
	).Scan(&p.UnitID, &p.DoneAt, &p.Correct, &p.Wrong, &p.Blank, &p.CompletedSource)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: get unit progress %d: %w", unitID, err)
	}
	return &p, nil
}

// UnitProgressBySource returns the progress rows of every done unit in a source,
// keyed by unit id. It powers the D/Y/B report without a per-unit query.
func (s *Store) UnitProgressBySource(ctx context.Context, q dbtx, sourceID int64) (map[int64]UnitProgress, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT p.unit_id, p.done_at, p.correct, p.wrong, p.blank, p.completed_source
		FROM unit_progress p
		JOIN units u ON u.id = p.unit_id
		JOIN sections sec ON sec.id = u.section_id
		WHERE sec.source_id = ?`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("store: query unit progress of source %d: %w", sourceID, err)
	}
	defer rows.Close()
	out := map[int64]UnitProgress{}
	for rows.Next() {
		var p UnitProgress
		if err := rows.Scan(&p.UnitID, &p.DoneAt, &p.Correct, &p.Wrong, &p.Blank, &p.CompletedSource); err != nil {
			return nil, fmt.Errorf("store: scan unit progress of source %d: %w", sourceID, err)
		}
		out[p.UnitID] = p
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate unit progress of source %d: %w", sourceID, err)
	}
	return out, nil
}

// UpsertUnitProgress writes p, inserting a new row or overwriting the existing
// one for p.UnitID. Marking a unit done is idempotent at the store level.
func (s *Store) UpsertUnitProgress(ctx context.Context, q dbtx, p UnitProgress) error {
	if _, err := q.ExecContext(ctx, `
		INSERT INTO unit_progress (unit_id, done_at, correct, wrong, blank, completed_source)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(unit_id) DO UPDATE SET
			done_at = excluded.done_at,
			correct = excluded.correct,
			wrong = excluded.wrong,
			blank = excluded.blank,
			completed_source = excluded.completed_source`,
		p.UnitID, p.DoneAt, p.Correct, p.Wrong, p.Blank, p.CompletedSource,
	); err != nil {
		return fmt.Errorf("store: upsert unit progress %d: %w", p.UnitID, err)
	}
	return nil
}

// DeleteUnitProgress removes a unit's progress row (marking it not done again).
func (s *Store) DeleteUnitProgress(ctx context.Context, q dbtx, unitID int64) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM unit_progress WHERE unit_id = ?`, unitID); err != nil {
		return fmt.Errorf("store: delete unit progress %d: %w", unitID, err)
	}
	return nil
}

// requireOne turns a zero-rows UPDATE into sql.ErrNoRows so callers can tell a
// missing id from a successful write.
func requireOne(res sql.Result, op string, id int64) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: %s rows %d: %w", op, id, err)
	}
	if n == 0 {
		return fmt.Errorf("store: %s %d: %w", op, id, sql.ErrNoRows)
	}
	return nil
}
