package plan

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ibra2626/workey/internal/store"
)

// catalog is a read snapshot of the whole source catalog, loaded once per
// transaction. Read models resolve a session item's references (source key,
// section ordinal, unit range) for display, and SavePlan resolves a write
// item's references to ids, without a query per item. A single-user install has
// a handful of sources, so loading the whole catalog is cheap.
type catalog struct {
	byKey    map[string]store.Source   // source key -> source
	byID     map[int64]store.Source    // source id -> source
	sections map[int64][]store.Section // source id -> sections, ordered by ordinal
	secByID  map[int64]store.Section   // section id -> section
	units    map[int64][]store.Unit    // section id -> units, ordered by ordinal
}

// loadCatalog reads the whole catalog within tx.
func (s *Service) loadCatalog(ctx context.Context, tx *sql.Tx) (*catalog, error) {
	sources, err := s.st.ListSources(ctx, tx)
	if err != nil {
		return nil, err
	}
	c := &catalog{
		byKey:    make(map[string]store.Source, len(sources)),
		byID:     make(map[int64]store.Source, len(sources)),
		sections: map[int64][]store.Section{},
		secByID:  map[int64]store.Section{},
		units:    map[int64][]store.Unit{},
	}
	for _, src := range sources {
		c.byKey[src.Key] = src
		c.byID[src.ID] = src
		secs, err := s.st.SectionsBySource(ctx, tx, src.ID)
		if err != nil {
			return nil, err
		}
		c.sections[src.ID] = secs
		for _, sec := range secs {
			c.secByID[sec.ID] = sec
			units, err := s.st.UnitsBySection(ctx, tx, sec.ID)
			if err != nil {
				return nil, err
			}
			c.units[sec.ID] = units
		}
	}
	return c, nil
}

// resolve turns a stored item into the ItemRef the notation formatter consumes.
// Unknown ids (a catalog deleted out from under a session) resolve to empty
// fields rather than an error: a read model still renders what it can.
func (c *catalog) resolve(it store.SessionItem) ItemRef {
	ref := ItemRef{LocFrom: it.LocFrom, LocTo: it.LocTo}
	if src, ok := c.byID[it.SourceID]; ok {
		ref.SourceKey = src.Key
		ref.LocationUnit = src.LocationUnit
	}
	if it.SectionID != nil {
		if sec, ok := c.secByID[*it.SectionID]; ok {
			ord := sec.Ordinal
			ref.SectionOrdinal = &ord
			if it.UnitFrom != nil && it.UnitTo != nil {
				ref.Units = c.unitsInRange(sec.ID, *it.UnitFrom, *it.UnitTo)
			}
		}
	}
	return ref
}

// resolveAll resolves a session's items in order.
func (c *catalog) resolveAll(items []store.SessionItem) []ItemRef {
	out := make([]ItemRef, 0, len(items))
	for _, it := range items {
		out = append(out, c.resolve(it))
	}
	return out
}

// unitsInRange returns the units of a section whose ordinal falls in [from, to],
// each with the number the reader sees (the unit's own number, or its ordinal
// when it has none).
func (c *catalog) unitsInRange(sectionID int64, from, to int) []RefUnit {
	var out []RefUnit
	for _, u := range c.units[sectionID] {
		if u.Ordinal < from || u.Ordinal > to {
			continue
		}
		num := u.Ordinal
		if u.Number != nil {
			num = *u.Number
		}
		out = append(out, RefUnit{Label: u.Label, Number: num})
	}
	return out
}

// coveredUnits returns the catalog units an item view covers, for per-unit
// marking: the units inside its explicit unit range, or every unit of its
// section when it names a section without a unit range. An item with no section
// (or an unresolvable one) covers no units.
func (c *catalog) coveredUnits(it ItemView) []store.Unit {
	if it.SectionOrdinal == nil {
		return nil
	}
	src, ok := c.byKey[it.SourceKey]
	if !ok {
		return nil
	}
	sec, ok := c.sectionByOrdinal(src.ID, *it.SectionOrdinal)
	if !ok {
		return nil
	}
	units := c.units[sec.ID]
	if it.UnitFrom == nil || it.UnitTo == nil {
		return units
	}
	var out []store.Unit
	for _, u := range units {
		if u.Ordinal >= *it.UnitFrom && u.Ordinal <= *it.UnitTo {
			out = append(out, u)
		}
	}
	return out
}

// sourceKind returns a source's kind by key, or "" when the key is unknown.
func (c *catalog) sourceKind(key string) string {
	if src, ok := c.byKey[key]; ok {
		return src.Kind
	}
	return ""
}

// itemView builds the read-model view of a stored item: its resolved source,
// section and units plus the raw ranges, for screens and read responses.
func (c *catalog) itemView(it store.SessionItem) ItemView {
	v := ItemView{
		UnitFrom: it.UnitFrom, UnitTo: it.UnitTo,
		LocFrom: it.LocFrom, LocTo: it.LocTo,
	}
	ref := c.resolve(it)
	v.SourceKey = ref.SourceKey
	v.SectionOrdinal = ref.SectionOrdinal
	v.Units = ref.Units
	if src, ok := c.byID[it.SourceID]; ok {
		v.SourceName = src.Name
	}
	if it.SectionID != nil {
		if sec, ok := c.secByID[*it.SectionID]; ok {
			v.SectionName = sec.Name
		}
	}
	return v
}

// itemViews builds the views of a session's items in order.
func (c *catalog) itemViews(items []store.SessionItem) []ItemView {
	out := make([]ItemView, 0, len(items))
	for _, it := range items {
		out = append(out, c.itemView(it))
	}
	return out
}

// resolveWrite turns a write item into a persistable store item, resolving the
// source key and section ordinal to ids and checking the unit range against the
// section's units. Reference failures are recorded as typed field errors on b
// (the caller aborts before any write); the returned item is partial on error.
func (c *catalog) resolveWrite(b *errBuilder, path string, it Item) store.SessionItem {
	out := store.SessionItem{
		UnitFrom: it.UnitFrom, UnitTo: it.UnitTo,
		LocFrom: it.LocFrom, LocTo: it.LocTo,
	}
	src, ok := c.byKey[it.SourceKey]
	if !ok {
		b.add(path+".source_key", CodeUnknownSource, fmt.Sprintf("source %q does not exist", it.SourceKey))
		return out
	}
	out.SourceID = src.ID
	if it.SectionOrdinal == nil {
		return out
	}
	sec, ok := c.sectionByOrdinal(src.ID, *it.SectionOrdinal)
	if !ok {
		b.add(path+".section_ordinal", CodeUnknownSection,
			fmt.Sprintf("source %q has no section with ordinal %d", it.SourceKey, *it.SectionOrdinal))
		return out
	}
	out.SectionID = &sec.ID
	if it.UnitFrom != nil && it.UnitTo != nil {
		lo, hi, n := c.unitBounds(sec.ID)
		if n == 0 || *it.UnitFrom < lo || *it.UnitTo > hi {
			b.add(path+".unit_to", CodeUnitOutOfRange,
				fmt.Sprintf("section %d has units %d–%d; unit range %d–%d is out of range",
					*it.SectionOrdinal, lo, hi, *it.UnitFrom, *it.UnitTo))
		}
	}
	return out
}

// sectionByOrdinal returns a source's section with the given ordinal.
func (c *catalog) sectionByOrdinal(sourceID int64, ordinal int) (store.Section, bool) {
	for _, sec := range c.sections[sourceID] {
		if sec.Ordinal == ordinal {
			return sec, true
		}
	}
	return store.Section{}, false
}

// unitBounds returns the lowest and highest unit ordinal in a section and the
// unit count. A section with no units returns (0, 0, 0).
func (c *catalog) unitBounds(sectionID int64) (lo, hi, n int) {
	units := c.units[sectionID]
	if len(units) == 0 {
		return 0, 0, 0
	}
	lo, hi = units[0].Ordinal, units[0].Ordinal
	for _, u := range units {
		if u.Ordinal < lo {
			lo = u.Ordinal
		}
		if u.Ordinal > hi {
			hi = u.Ordinal
		}
	}
	return lo, hi, len(units)
}
