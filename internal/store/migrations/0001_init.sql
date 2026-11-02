-- 0001_init: full v2 schema (source catalog).
-- Dates are TEXT 'YYYY-MM-DD', times TEXT 'HH:MM', timestamps TEXT ISO-8601 UTC.
-- Enums and range ordering are enforced by CHECK constraints; the store layer
-- holds no business rules. The catalog (sources → sections → units) is generic:
-- no exam, book code or unit label is hard-coded in the schema.

CREATE TABLE phases (
    id         INTEGER PRIMARY KEY,
    key        TEXT    NOT NULL UNIQUE,
    ordinal    INTEGER NOT NULL,
    name       TEXT    NOT NULL,
    start_date TEXT    NOT NULL,
    end_date   TEXT    NOT NULL,
    CHECK (start_date <= end_date)
);

CREATE TABLE topic_groups (
    id           INTEGER PRIMARY KEY,
    key          TEXT    NOT NULL UNIQUE,
    name         TEXT    NOT NULL,
    ordinal      INTEGER NOT NULL,
    question_min INTEGER,
    question_max INTEGER,
    CHECK (question_min IS NULL OR question_max IS NULL OR question_min <= question_max)
);

-- sources: one book / video / course / other the user studies. key is a short
-- user-given code (no fixed list); location_unit names how ranges are counted.
CREATE TABLE sources (
    id            INTEGER PRIMARY KEY,
    key           TEXT    NOT NULL UNIQUE,
    name          TEXT    NOT NULL,
    kind          TEXT    NOT NULL CHECK (kind IN ('book', 'video', 'course', 'other')),
    location_unit TEXT    NOT NULL CHECK (location_unit IN ('page', 'minute', 'lesson', 'other')),
    ordinal       INTEGER NOT NULL,
    created_at    TEXT    NOT NULL,
    updated_at    TEXT    NOT NULL
);

-- sections: an ordered part of a source (chapter / module), with an optional
-- location range in the source's location_unit.
CREATE TABLE sections (
    id        INTEGER PRIMARY KEY,
    source_id INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    ordinal   INTEGER NOT NULL,
    name      TEXT    NOT NULL,
    loc_from  INTEGER,
    loc_to    INTEGER,
    CHECK (loc_from IS NULL OR loc_to IS NULL OR loc_from <= loc_to),
    UNIQUE (source_id, ordinal)
);

CREATE INDEX idx_sections_source ON sections(source_id);

-- units: an ordered study piece inside a section. label is the source's own
-- word ("Antrenman" / "Test" / "Föy" / "Ders"); the code never enumerates them.
CREATE TABLE units (
    id         INTEGER PRIMARY KEY,
    section_id INTEGER NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    ordinal    INTEGER NOT NULL,
    label      TEXT    NOT NULL,
    number     INTEGER,
    loc_from   INTEGER,
    loc_to     INTEGER,
    CHECK (loc_from IS NULL OR loc_to IS NULL OR loc_from <= loc_to),
    UNIQUE (section_id, ordinal)
);

CREATE INDEX idx_units_section ON units(section_id);

-- unit_progress: a unit's study result, separate from the catalog definition so
-- re-writing a source (save_source) never wipes progress — one row per completed
-- unit, keyed by unit id; absence means not done. D/Y/B (doğru/yanlış/boş) are
-- optional (video units leave them empty).
CREATE TABLE unit_progress (
    unit_id          INTEGER PRIMARY KEY REFERENCES units(id) ON DELETE CASCADE,
    done_at          TEXT,
    correct          INTEGER,
    wrong            INTEGER,
    blank            INTEGER,
    completed_source TEXT    NOT NULL CHECK (completed_source IN ('mcp', 'api', 'web')),
    CHECK (correct IS NULL OR correct >= 0),
    CHECK (wrong   IS NULL OR wrong   >= 0),
    CHECK (blank   IS NULL OR blank   >= 0)
);

CREATE TABLE sessions (
    id               INTEGER PRIMARY KEY,
    date             TEXT    NOT NULL,
    start_time       TEXT    NOT NULL,
    duration_min     INTEGER NOT NULL,
    kind             TEXT    NOT NULL,
    topic_group_id   INTEGER REFERENCES topic_groups(id),
    makeup_of        INTEGER REFERENCES sessions(id) ON DELETE SET NULL,
    completed_at     TEXT,
    completed_source TEXT    CHECK (completed_source IS NULL OR completed_source IN ('mcp', 'api', 'web')),
    created_at       TEXT    NOT NULL,
    updated_at       TEXT    NOT NULL,
    UNIQUE (date, start_time)
);

CREATE INDEX idx_sessions_date ON sessions(date);

-- session_items: an ordered reference inside a session to a source + optional
-- section + a unit range and/or a location range. Never topic text or content.
CREATE TABLE session_items (
    id         INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    ordinal    INTEGER NOT NULL,
    source_id  INTEGER NOT NULL REFERENCES sources(id),
    section_id INTEGER REFERENCES sections(id),
    unit_from  INTEGER,
    unit_to    INTEGER,
    loc_from   INTEGER,
    loc_to     INTEGER,
    CHECK (unit_from IS NULL OR unit_to IS NULL OR unit_from <= unit_to),
    CHECK (loc_from  IS NULL OR loc_to  IS NULL OR loc_from  <= loc_to),
    UNIQUE (session_id, ordinal)
);

CREATE INDEX idx_session_items_session ON session_items(session_id);

CREATE TABLE audit_log (
    id           INTEGER PRIMARY KEY,
    at_utc       TEXT    NOT NULL,
    source       TEXT    NOT NULL CHECK (source IN ('mcp', 'api', 'web')),
    action       TEXT    NOT NULL,
    entity       TEXT    NOT NULL,
    entity_id    INTEGER,
    before_json  TEXT,
    after_json   TEXT,
    request_hash TEXT
);

CREATE TABLE idempotency_keys (
    key          TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,
    response_json TEXT NOT NULL,
    created_at   TEXT NOT NULL
);
