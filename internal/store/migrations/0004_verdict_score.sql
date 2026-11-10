-- 0004: stage-2 tag score columns on the verdict table (brief 4.2 stage 2,
-- v1: informational weighted-tag score, no cutoff). Plain ALTER TABLE that
-- runs on both engines, so no dialect split (ADR-0002).

ALTER TABLE job_verdicts ADD COLUMN score INTEGER NOT NULL DEFAULT 0;
ALTER TABLE job_verdicts ADD COLUMN score_reasons TEXT NOT NULL DEFAULT '';
