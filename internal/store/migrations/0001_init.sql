-- 0001_init: full v1 schema.
-- Dates are TEXT 'YYYY-MM-DD', times TEXT 'HH:MM', timestamps TEXT ISO-8601 UTC.
-- Enums and range ordering are enforced by CHECK constraints; the store layer
-- holds no business rules.

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
    id               INTEGER PRIMARY KEY,
    key              TEXT    NOT NULL UNIQUE,
    name             TEXT    NOT NULL,
    ordinal          INTEGER NOT NULL,
    dgs_question_min INTEGER,
    dgs_question_max INTEGER,
    CHECK (dgs_question_min IS NULL OR dgs_question_max IS NULL OR dgs_question_min <= dgs_question_max)
);

CREATE TABLE sessions (
    id               INTEGER PRIMARY KEY,
    date             TEXT    NOT NULL,
    start_time       TEXT    NOT NULL,
    duration_min     INTEGER NOT NULL,
    kind             TEXT    NOT NULL CHECK (kind IN ('aksam', 'uzun_blok', 'deneme')),
    topic_group_id   INTEGER REFERENCES topic_groups(id),
    makeup_of        INTEGER REFERENCES sessions(id) ON DELETE SET NULL,
    completed_at     TEXT,
    completed_source TEXT    CHECK (completed_source IS NULL OR completed_source IN ('mcp', 'api', 'web')),
    created_at       TEXT    NOT NULL,
    updated_at       TEXT    NOT NULL,
    UNIQUE (date, start_time)
);

CREATE INDEX idx_sessions_date ON sessions(date);

CREATE TABLE session_items (
    id             INTEGER PRIMARY KEY,
    session_id     INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    ordinal        INTEGER NOT NULL,
    book           TEXT    NOT NULL CHECK (book IN ('M1', 'M2', 'M3', 'G1', 'G2')),
    chapter_no     INTEGER NOT NULL,
    chapter_title  TEXT    NOT NULL,
    page_from      INTEGER NOT NULL,
    page_to        INTEGER NOT NULL,
    konu_from      INTEGER,
    konu_to        INTEGER,
    soru_from      INTEGER,
    soru_to        INTEGER,
    antrenman_from INTEGER,
    antrenman_to   INTEGER,
    CHECK (page_from <= page_to),
    CHECK (konu_from IS NULL OR konu_to IS NULL OR konu_from <= konu_to),
    CHECK (soru_from IS NULL OR soru_to IS NULL OR soru_from <= soru_to),
    CHECK (antrenman_from IS NULL OR antrenman_to IS NULL OR antrenman_from <= antrenman_to),
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
