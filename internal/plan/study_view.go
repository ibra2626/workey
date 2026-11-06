package plan

import "github.com/ibra2626/workey/internal/store"

// StudyView is additive read-only context, independent of SessionState. Times
// record lifecycle events, never measured study time. No row means no record,
// not proof that the user never studied.
type StudyView struct {
	Status         string        `json:"status" jsonschema:"Study lifecycle: active, paused or closed; independent of session state."`
	FirstStartedAt string        `json:"first_started_at" jsonschema:"UTC timestamp of the first start; not measured study time."`
	LastStartedAt  string        `json:"last_started_at"`
	UpdatedAt      string        `json:"updated_at"`
	EntryMode      string        `json:"entry_mode" jsonschema:"normal or ten_minutes. A starting goal; does not change planned duration or progress."`
	Version        int64         `json:"version" jsonschema:"Read-only study revision. Existing REST/MCP completion writes do not check this version."`
	Bookmark       *BookmarkView `json:"bookmark,omitempty"`
}

// BookmarkView resolves the stable zero-based item ordinal to a catalog
// reference. ResumePage is a location value in the source's location unit.
type BookmarkView struct {
	ItemOrdinal int      `json:"item_ordinal" jsonschema:"Zero-based ordinal in the session's ordered items."`
	Stage       string   `json:"stage" jsonschema:"konu, soru, yanlis_defteri for aksam; calisma for other kinds."`
	ResumePage  int      `json:"resume_page" jsonschema:"Next location to work on in the source's location unit, including a partly done one. Never automatically incremented."`
	Item        ItemView `json:"item"`
	Notation    string   `json:"notation" jsonschema:"Resolved canonical source/section/unit reference for the selected item."`
}

func attachStudy(sv *SessionView, cat *catalog, study *store.Study, items []store.SessionItem) {
	sv.StudyVersion = studyVersion(study)
	if study == nil {
		return
	}
	v := &StudyView{Status: study.Status, FirstStartedAt: study.FirstStartedAt,
		LastStartedAt: study.LastStartedAt, UpdatedAt: study.UpdatedAt,
		EntryMode: study.EntryMode, Version: study.Version}
	if study.ItemOrdinal != nil && study.Stage != nil && study.ResumePage != nil {
		for _, item := range items {
			if item.Ordinal != *study.ItemOrdinal {
				continue
			}
			v.Bookmark = &BookmarkView{ItemOrdinal: item.Ordinal, Stage: *study.Stage,
				ResumePage: *study.ResumePage, Item: cat.itemView(item),
				Notation: FormatRef(cat.resolve(item))}
			break
		}
	}
	sv.Study = v
}
