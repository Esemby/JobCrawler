-- 0003: verdict table for the Stage-1 filter (docs/original-brief.md 4.2).
-- One row per fingerprint; re-evaluation UPSERTS the same row (no history
-- table in v1 — the reasons chain is the record). Standard SQL that runs on
-- SQLite and MariaDB alike, so no dialect split (ADR-0002).
-- reasons is newline-joined reason lines, one rejection/observation per line.

CREATE TABLE IF NOT EXISTS job_verdicts (
    fingerprint  TEXT PRIMARY KEY,
    result       TEXT NOT NULL CHECK (result IN ('accepted', 'rejected')),
    reasons      TEXT NOT NULL DEFAULT '',
    evaluated_at TEXT NOT NULL
);
