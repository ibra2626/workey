-- Study lifecycle and bookmarks. Source/page validation belongs to plan.
CREATE TABLE session_study (
    session_id       INTEGER PRIMARY KEY REFERENCES sessions(id) ON DELETE RESTRICT,
    status           TEXT NOT NULL CHECK (status IN ('active', 'paused', 'closed')),
    first_started_at TEXT NOT NULL,
    last_started_at  TEXT NOT NULL,
    updated_at       TEXT NOT NULL,
    entry_mode       TEXT NOT NULL CHECK (entry_mode IN ('normal', 'ten_minutes')),
    item_ordinal     INTEGER CHECK (item_ordinal >= 0),
    stage            TEXT CHECK (stage IN ('konu', 'soru', 'yanlis_defteri', 'calisma')),
    resume_page      INTEGER CHECK (resume_page > 0),
    version          INTEGER NOT NULL CHECK (version >= 1),
    CHECK (
        (item_ordinal IS NULL AND stage IS NULL AND resume_page IS NULL) OR
        (item_ordinal IS NOT NULL AND stage IS NOT NULL AND resume_page IS NOT NULL)
    )
);

CREATE UNIQUE INDEX idx_session_study_one_active ON session_study(status)
    WHERE status = 'active';
