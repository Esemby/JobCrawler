package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Esemby/JobCrawler/internal/store"
)

// UpsertJobs stores jobs by fingerprint (create or refresh last_seen) and
// returns how many were genuinely new. An empty Title or URL is an error,
// never silently stored.
func (s *Store) UpsertJobs(ctx context.Context, jobs []store.Job) (int, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	inserted := 0
	for _, j := range jobs {
		if j.Title == "" || j.URL == "" {
			return inserted, fmt.Errorf("upsert job: empty title or url in %q", j.URL)
		}
		fp := j.Fingerprint
		if fp == "" {
			fp = store.Fingerprint(j)
		}
		n, err := s.upsertJob(ctx, j, fp, now)
		if err != nil {
			return inserted, fmt.Errorf("upsert job %s: %w", fp, err)
		}
		inserted += n
	}
	return inserted, nil
}

// upsertJob: fingerprint exists -> touch last_seen, else insert. Two plain
// statements instead of engine-specific upsert SQL (ADR-0002: both engines).
func (s *Store) upsertJob(ctx context.Context, j store.Job, fp, now string) (int, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM jobs WHERE fingerprint = ?`, fp).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("lookup fingerprint: %w", err)
	}
	if exists > 0 {
		_, err = s.db.ExecContext(ctx,
			`UPDATE jobs SET last_seen = ?, state = 'active' WHERE fingerprint = ?`, now, fp)
		if err != nil {
			return 0, fmt.Errorf("touch existing: %w", err)
		}
		return 0, nil
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO jobs
			(fingerprint, source_id, title, company, url, location, published_at, first_seen, last_seen, state)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'active')`,
		fp, j.SourceID, j.Title, j.Company, j.URL, j.Location,
		j.PublishedAt, now, now)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}
	return 1, nil
}

// ListJobs returns stored jobs, newest last_seen first. state filters
// ('active', 'disappeared'); pass "*" for everything. SourceName is filled
// from job_sources when the row still exists.
func (s *Store) ListJobs(ctx context.Context, state string) ([]store.Job, error) {
	q := `SELECT j.id, j.fingerprint, j.title, COALESCE(j.company,''), j.url,
	                 COALESCE(j.location,''), COALESCE(j.published_at,''),
	                 COALESCE(j.first_seen,''), COALESCE(j.last_seen,''), j.state, j.source_id,
	                 COALESCE(src.name, '')
	          FROM jobs j LEFT JOIN job_sources src ON src.id = j.source_id`
	var args []any
	if state != "" && state != "*" {
		q += ` WHERE j.state = ?`
		args = append(args, state)
	}
	q += ` ORDER BY j.last_seen DESC`

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	return collectJobs(rows)
}

// collectJobs scans a jobs result set; rows is closed here.
func collectJobs(rows *sql.Rows) ([]store.Job, error) {
	defer rows.Close()
	var out []store.Job
	for rows.Next() {
		var j store.Job
		if err := rows.Scan(&j.SourceID, &j.Fingerprint, &j.Title, &j.Company,
			&j.URL, &j.Location, &j.PublishedAt, &j.FirstSeen, &j.LastSeen,
			&j.State, &j.SourceID, &j.SourceName); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		out = append(out, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate jobs: %w", err)
	}
	return out, nil
}
