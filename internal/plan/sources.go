package plan

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ibra2626/workey/internal/store"
)

// --- catalog read views (get_sources) ------------------------------------

// UnitView is one unit in a catalog read: its id (so a caller can address it
// with set_unit_result) plus its definition.
type UnitView struct {
	ID      int64   `json:"id"`
	Ordinal int     `json:"ordinal"`
	Label   string  `json:"label"`
	Number  *int    `json:"number,omitempty"`
	LocFrom *int    `json:"loc_from,omitempty"`
	LocTo   *int    `json:"loc_to,omitempty"`
	Title   *string `json:"title,omitempty"`
	URL     *string `json:"url,omitempty"`
}

// SectionView is one section in a catalog read, with its ordered units.
type SectionView struct {
	ID      int64      `json:"id"`
	Ordinal int        `json:"ordinal"`
	Name    string     `json:"name"`
	LocFrom *int       `json:"loc_from,omitempty"`
	LocTo   *int       `json:"loc_to,omitempty"`
	Units   []UnitView `json:"units"`
}

// SourceView is one source in a catalog read, with its ordered sections.
type SourceView struct {
	ID           int64         `json:"id"`
	Key          string        `json:"key"`
	Name         string        `json:"name"`
	Kind         string        `json:"kind"`
	LocationUnit string        `json:"location_unit"`
	Ordinal      int           `json:"ordinal"`
	Sections     []SectionView `json:"sections"`
}

// CatalogView is the whole source catalog for get_sources: sources ordered by
// ordinal, each with its sections and units.
type CatalogView struct {
	Sources []SourceView `json:"sources"`
}

// SaveSourceResult reports what a SaveSource call did: the resulting source (with
// the final ids a caller needs for set_unit_result) and whether it was newly
// created or left unchanged.
type SaveSourceResult struct {
	Source    SourceView `json:"source"`
	Created   bool       `json:"created"`
	Unchanged bool       `json:"unchanged"`
}

// Sources returns the whole catalog ordered by source then section then unit
// ordinal. It backs get_sources / GET /api/v1/sources.
func (s *Service) Sources(ctx context.Context) (*CatalogView, error) {
	db := s.st.DB()
	sources, err := s.st.ListSources(ctx, db)
	if err != nil {
		return nil, err
	}
	view := &CatalogView{Sources: make([]SourceView, 0, len(sources))}
	for _, src := range sources {
		secs, err := s.st.SectionsBySource(ctx, db, src.ID)
		if err != nil {
			return nil, err
		}
		sv := SourceView{
			ID: src.ID, Key: src.Key, Name: src.Name, Kind: src.Kind,
			LocationUnit: src.LocationUnit, Ordinal: src.Ordinal,
			Sections: make([]SectionView, 0, len(secs)),
		}
		for _, sec := range secs {
			units, err := s.st.UnitsBySection(ctx, db, sec.ID)
			if err != nil {
				return nil, err
			}
			sv.Sections = append(sv.Sections, sectionView(sec, units))
		}
		view.Sources = append(view.Sources, sv)
	}
	return view, nil
}

// sectionView assembles one section's read view from its store rows.
func sectionView(sec store.Section, units []store.Unit) SectionView {
	out := SectionView{
		ID: sec.ID, Ordinal: sec.Ordinal, Name: sec.Name,
		LocFrom: sec.LocFrom, LocTo: sec.LocTo,
		Units: make([]UnitView, 0, len(units)),
	}
	for _, u := range units {
		out.Units = append(out.Units, UnitView{
			ID: u.ID, Ordinal: u.Ordinal, Label: u.Label, Number: u.Number,
			LocFrom: u.LocFrom, LocTo: u.LocTo, Title: u.Title, URL: u.URL,
		})
	}
	return out
}

// SaveSource writes one source and its full set of sections and units,
// replacing the structure keyed by the source's Key. It reconciles by ordinal so
// an existing section (matched on its ordinal) and an existing unit (matched on
// its section ordinal + unit ordinal) keep their ids — and so a re-write never
// touches unit_progress for units whose place is preserved. Sections and units
// absent from the request are deleted, cascading to their progress.
//
// The call is validated, idempotent and audited: an identical structure writes
// nothing and adds no audit row; a real change records one audit row for the
// source (entity "source", before/after JSON). With req.IdempotencyKey the stored
// response is replayed and a different payload under the same key is rejected.
func (s *Service) SaveSource(ctx context.Context, source Source, req SaveSourceRequest) (*SaveSourceResult, error) {
	if !source.valid() {
		return nil, fmt.Errorf("plan: invalid source %q", source)
	}
	normalizeUnitLinks(req.Sections)
	if err := validateSaveSource(req); err != nil {
		return nil, err
	}

	hash, err := hashPayload(source, "save_source", saveSourcePayload{Source: req.Source, Sections: req.Sections})
	if err != nil {
		return nil, err
	}

	var result *SaveSourceResult
	err = s.st.WithTx(ctx, func(tx *sql.Tx) error {
		if replay, done, err := s.replay(ctx, tx, req.IdempotencyKey, hash); err != nil {
			return err
		} else if done {
			result = &SaveSourceResult{}
			if err := json.Unmarshal([]byte(replay), result); err != nil {
				return fmt.Errorf("plan: decode stored response: %w", err)
			}
			return nil
		}

		res, err := s.applySource(ctx, tx, source, hash, req)
		if err != nil {
			return err
		}
		result = res
		return s.remember(ctx, tx, req.IdempotencyKey, hash, res)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// applySource performs the reconcile against the store inside tx.
func (s *Service) applySource(ctx context.Context, tx *sql.Tx, source Source, hash string, req SaveSourceRequest) (*SaveSourceResult, error) {
	now := s.nowUTC()

	existing, err := s.st.SourceByKey(ctx, tx, req.Source.Key)
	if err != nil {
		return nil, err
	}

	// A structural snapshot ignoring ids/timestamps drives the no-op check and
	// the audit before/after.
	desiredSnap, err := requestSnapshotJSON(req)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		before, err := s.sourceSnapshotJSON(ctx, tx, *existing)
		if err != nil {
			return nil, err
		}
		if before == desiredSnap {
			// Identical structure already stored: no write, no audit row.
			sv, err := s.sourceViewByID(ctx, tx, existing.ID)
			if err != nil {
				return nil, err
			}
			return &SaveSourceResult{Source: *sv, Unchanged: true}, nil
		}
		if err := s.reconcileSource(ctx, tx, existing, req); err != nil {
			return nil, err
		}
		if err := s.auditSource(ctx, tx, source, actionUpdate, existing.ID, &before, &desiredSnap, hash, now); err != nil {
			return nil, err
		}
		sv, err := s.sourceViewByID(ctx, tx, existing.ID)
		if err != nil {
			return nil, err
		}
		return &SaveSourceResult{Source: *sv}, nil
	}

	src := store.Source{
		Key: req.Source.Key, Name: req.Source.Name, Kind: req.Source.Kind,
		LocationUnit: req.Source.LocationUnit, Ordinal: req.Source.Ordinal,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.st.InsertSource(ctx, tx, &src); err != nil {
		return nil, err
	}
	if err := s.insertSections(ctx, tx, src.ID, req.Sections); err != nil {
		return nil, err
	}
	if err := s.auditSource(ctx, tx, source, actionCreate, src.ID, nil, &desiredSnap, hash, now); err != nil {
		return nil, err
	}
	sv, err := s.sourceViewByID(ctx, tx, src.ID)
	if err != nil {
		return nil, err
	}
	return &SaveSourceResult{Source: *sv, Created: true}, nil
}

// reconcileSource updates the existing source's fields and reconciles its
// sections and units by ordinal, preserving ids (and so unit progress).
func (s *Service) reconcileSource(ctx context.Context, tx *sql.Tx, existing *store.Source, req SaveSourceRequest) error {
	now := s.nowUTC()
	updated := store.Source{
		Key: existing.Key, Name: req.Source.Name, Kind: req.Source.Kind,
		LocationUnit: req.Source.LocationUnit, Ordinal: req.Source.Ordinal,
		UpdatedAt: now,
	}
	if err := s.st.UpdateSourceByKey(ctx, tx, updated); err != nil {
		return err
	}

	existingSecs, err := s.st.SectionsBySource(ctx, tx, existing.ID)
	if err != nil {
		return err
	}
	secByOrdinal := make(map[int]store.Section, len(existingSecs))
	for _, sec := range existingSecs {
		secByOrdinal[sec.Ordinal] = sec
	}
	wantSec := make(map[int]bool, len(req.Sections))
	for _, in := range req.Sections {
		wantSec[in.Ordinal] = true
	}
	// Delete sections the request dropped (cascades to their units/progress).
	for _, sec := range existingSecs {
		if !wantSec[sec.Ordinal] {
			if err := s.st.DeleteSection(ctx, tx, sec.ID); err != nil {
				return err
			}
		}
	}
	// Create or update each incoming section, then reconcile its units.
	for _, in := range req.Sections {
		cur, ok := secByOrdinal[in.Ordinal]
		if !ok {
			sec := store.Section{SourceID: existing.ID, Ordinal: in.Ordinal, Name: in.Name, LocFrom: in.LocFrom, LocTo: in.LocTo}
			if err := s.st.InsertSection(ctx, tx, &sec); err != nil {
				return err
			}
			if err := s.insertUnits(ctx, tx, sec.ID, in.Units); err != nil {
				return err
			}
			continue
		}
		cur.Name = in.Name
		cur.LocFrom = in.LocFrom
		cur.LocTo = in.LocTo
		if err := s.st.UpdateSection(ctx, tx, cur); err != nil {
			return err
		}
		if err := s.reconcileUnits(ctx, tx, cur.ID, in.Units); err != nil {
			return err
		}
	}
	return nil
}

// reconcileUnits reconciles one section's units by ordinal: a matching unit is
// updated in place (keeping its id and progress), a new one is inserted, and a
// unit the request dropped is deleted (cascading its progress away).
func (s *Service) reconcileUnits(ctx context.Context, tx *sql.Tx, sectionID int64, units []UnitInput) error {
	existing, err := s.st.UnitsBySection(ctx, tx, sectionID)
	if err != nil {
		return err
	}
	byOrdinal := make(map[int]store.Unit, len(existing))
	for _, u := range existing {
		byOrdinal[u.Ordinal] = u
	}
	want := make(map[int]bool, len(units))
	for _, in := range units {
		want[in.Ordinal] = true
	}
	for _, u := range existing {
		if !want[u.Ordinal] {
			if err := s.st.DeleteUnit(ctx, tx, u.ID); err != nil {
				return err
			}
		}
	}
	for _, in := range units {
		cur, ok := byOrdinal[in.Ordinal]
		if !ok {
			u := store.Unit{SectionID: sectionID, Ordinal: in.Ordinal, Label: in.Label, Number: in.Number, LocFrom: in.LocFrom, LocTo: in.LocTo, Title: in.Title, URL: in.URL}
			if err := s.st.InsertUnit(ctx, tx, &u); err != nil {
				return err
			}
			continue
		}
		cur.Label = in.Label
		cur.Number = in.Number
		cur.LocFrom = in.LocFrom
		cur.LocTo = in.LocTo
		cur.Title = in.Title
		cur.URL = in.URL
		if err := s.st.UpdateUnit(ctx, tx, cur); err != nil {
			return err
		}
	}
	return nil
}

// insertSections inserts a new source's sections and their units.
func (s *Service) insertSections(ctx context.Context, tx *sql.Tx, sourceID int64, sections []SectionInput) error {
	for _, in := range sections {
		sec := store.Section{SourceID: sourceID, Ordinal: in.Ordinal, Name: in.Name, LocFrom: in.LocFrom, LocTo: in.LocTo}
		if err := s.st.InsertSection(ctx, tx, &sec); err != nil {
			return err
		}
		if err := s.insertUnits(ctx, tx, sec.ID, in.Units); err != nil {
			return err
		}
	}
	return nil
}

// insertUnits inserts a section's units.
func (s *Service) insertUnits(ctx context.Context, tx *sql.Tx, sectionID int64, units []UnitInput) error {
	for _, in := range units {
		u := store.Unit{SectionID: sectionID, Ordinal: in.Ordinal, Label: in.Label, Number: in.Number, LocFrom: in.LocFrom, LocTo: in.LocTo, Title: in.Title, URL: in.URL}
		if err := s.st.InsertUnit(ctx, tx, &u); err != nil {
			return err
		}
	}
	return nil
}

// sourceViewByID builds the read view of one source (with sections and units)
// within tx, for the SaveSource result.
func (s *Service) sourceViewByID(ctx context.Context, tx *sql.Tx, sourceID int64) (*SourceView, error) {
	// ListSources is cheap for a single-user install; find the one we need so the
	// view carries the stored key/name/kind without a dedicated by-id query.
	sources, err := s.st.ListSources(ctx, tx)
	if err != nil {
		return nil, err
	}
	var src *store.Source
	for i := range sources {
		if sources[i].ID == sourceID {
			src = &sources[i]
			break
		}
	}
	if src == nil {
		return nil, fmt.Errorf("plan: source %d not found after write", sourceID)
	}
	secs, err := s.st.SectionsBySource(ctx, tx, sourceID)
	if err != nil {
		return nil, err
	}
	sv := &SourceView{
		ID: src.ID, Key: src.Key, Name: src.Name, Kind: src.Kind,
		LocationUnit: src.LocationUnit, Ordinal: src.Ordinal,
		Sections: make([]SectionView, 0, len(secs)),
	}
	for _, sec := range secs {
		units, err := s.st.UnitsBySection(ctx, tx, sec.ID)
		if err != nil {
			return nil, err
		}
		sv.Sections = append(sv.Sections, sectionView(sec, units))
	}
	return sv, nil
}

// auditSource appends one audit row for a source structure change. Like plan
// writes it records the request hash.
func (s *Service) auditSource(ctx context.Context, tx *sql.Tx, source Source, action string, sourceID int64, before, after *string, hash, now string) error {
	src := string(source)
	entry := store.AuditEntry{
		AtUTC:       now,
		Source:      src,
		Action:      action,
		Entity:      "source",
		EntityID:    &sourceID,
		BeforeJSON:  before,
		AfterJSON:   after,
		RequestHash: &hash,
	}
	return s.st.InsertAudit(ctx, tx, &entry)
}

// --- structural snapshots (ids/timestamps excluded) ----------------------

type snapUnit struct {
	Ordinal int     `json:"ordinal"`
	Label   string  `json:"label"`
	Number  *int    `json:"number,omitempty"`
	LocFrom *int    `json:"loc_from,omitempty"`
	LocTo   *int    `json:"loc_to,omitempty"`
	Title   *string `json:"title,omitempty"`
	URL     *string `json:"url,omitempty"`
}

type snapSection struct {
	Ordinal int        `json:"ordinal"`
	Name    string     `json:"name"`
	LocFrom *int       `json:"loc_from,omitempty"`
	LocTo   *int       `json:"loc_to,omitempty"`
	Units   []snapUnit `json:"units"`
}

type snapSource struct {
	Key          string        `json:"key"`
	Name         string        `json:"name"`
	Kind         string        `json:"kind"`
	LocationUnit string        `json:"location_unit"`
	Ordinal      int           `json:"ordinal"`
	Sections     []snapSection `json:"sections"`
}

// requestSnapshotJSON renders the requested structure as the canonical snapshot
// used for the no-op check and the audit after-image.
func requestSnapshotJSON(req SaveSourceRequest) (string, error) {
	sn := snapSource{
		Key: req.Source.Key, Name: req.Source.Name, Kind: req.Source.Kind,
		LocationUnit: req.Source.LocationUnit, Ordinal: req.Source.Ordinal,
	}
	for _, sec := range req.Sections {
		ss := snapSection{Ordinal: sec.Ordinal, Name: sec.Name, LocFrom: sec.LocFrom, LocTo: sec.LocTo}
		for _, u := range sec.Units {
			ss.Units = append(ss.Units, snapUnit{Ordinal: u.Ordinal, Label: u.Label, Number: u.Number, LocFrom: u.LocFrom, LocTo: u.LocTo, Title: u.Title, URL: u.URL})
		}
		sn.Sections = append(sn.Sections, ss)
	}
	return marshalSnap(sn)
}

// sourceSnapshotJSON renders the stored structure of src in the same canonical
// shape, so it compares byte-for-byte with a request snapshot.
func (s *Service) sourceSnapshotJSON(ctx context.Context, tx *sql.Tx, src store.Source) (string, error) {
	sn := snapSource{
		Key: src.Key, Name: src.Name, Kind: src.Kind,
		LocationUnit: src.LocationUnit, Ordinal: src.Ordinal,
	}
	secs, err := s.st.SectionsBySource(ctx, tx, src.ID)
	if err != nil {
		return "", err
	}
	for _, sec := range secs {
		ss := snapSection{Ordinal: sec.Ordinal, Name: sec.Name, LocFrom: sec.LocFrom, LocTo: sec.LocTo}
		units, err := s.st.UnitsBySection(ctx, tx, sec.ID)
		if err != nil {
			return "", err
		}
		for _, u := range units {
			ss.Units = append(ss.Units, snapUnit{Ordinal: u.Ordinal, Label: u.Label, Number: u.Number, LocFrom: u.LocFrom, LocTo: u.LocTo, Title: u.Title, URL: u.URL})
		}
		sn.Sections = append(sn.Sections, ss)
	}
	return marshalSnap(sn)
}

func marshalSnap(sn snapSource) (string, error) {
	buf, err := json.Marshal(sn)
	if err != nil {
		return "", fmt.Errorf("plan: encode source snapshot: %w", err)
	}
	return string(buf), nil
}

// saveSourcePayload is the hashed view of a SaveSource request (the idempotency
// key itself is excluded, so a repeat under the same key is recognised).
type saveSourcePayload struct {
	Source   SourceInput    `json:"source"`
	Sections []SectionInput `json:"sections"`
}
