-- 0002: job lifecycle columns.
-- Adds first-seen/last-seen/disappeared tracking (docs/glossary.md) without
-- touching existing rows. Plain ALTER TABLE statements that run on both
-- SQLite and MariaDB, so no dialect suffix split is needed (ADR-0002).

ALTER TABLE jobs ADD COLUMN first_seen TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN last_seen TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN state TEXT NOT NULL DEFAULT 'active';

CREATE INDEX IF NOT EXISTS idx_jobs_fingerprint ON jobs (fingerprint);
