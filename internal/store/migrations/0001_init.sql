-- 0001_init: initial schema.
-- One concern: create the skeleton tables of the domain model. Nothing here
-- is destructive; tables use IF NOT EXISTS so re-applying a fresh database
-- state is a no-op.
-- Dialect note: this file uses SQLite syntax (default engine, ADR-0002);
-- the MariaDB twin lives in 0001_init.mysql.sql in Phase 6.

CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Sources are the configured crawler inputs (greenhouse, rss, ...).
CREATE TABLE IF NOT EXISTS job_sources (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    name   TEXT NOT NULL UNIQUE,                -- short unique name, e.g. "greenhouse-acme"
    kind   TEXT NOT NULL,                       -- source type per docs/sources.md
    url    TEXT NOT NULL,                       -- listing endpoint or board URL
    enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1))
);

-- Jobs are shared across users (ADR-0004); fingerprints dedupe them.
CREATE TABLE IF NOT EXISTS jobs (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    fingerprint  TEXT NOT NULL UNIQUE,           -- stable dedupe key (docs/glossary.md)
    source_id    INTEGER REFERENCES job_sources (id),
    title        TEXT NOT NULL,
    company      TEXT,
    url          TEXT NOT NULL,
    location     TEXT,
    published_at TEXT,
    fetched_at   TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

-- Multi-user contract (ADR-0004): flags are per-user rows from revision 0001,
-- even though Phase 2 ships one implied user. No UI yet, by design.
CREATE TABLE IF NOT EXISTS users (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_flags (
    user_id    INTEGER NOT NULL REFERENCES users (id),
    job_id     INTEGER NOT NULL REFERENCES jobs (id),
    flag       TEXT NOT NULL CHECK (flag IN ('new', 'seen', 'applied', 'hidden')),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    PRIMARY KEY (user_id, job_id, flag)
);
