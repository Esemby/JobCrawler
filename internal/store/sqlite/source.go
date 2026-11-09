package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/Esemby/JobCrawler/internal/store"
)

// AddSource inserts a configured crawler input; the name is unique per DB
// and returning it signals a duplicate rather than an error.
func (s *Store) AddSource(ctx context.Context, name, kind, endpoint string) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO job_sources (name, kind, url) VALUES (?, ?, ?)`, name, kind, endpoint)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, fmt.Errorf("source %q: already exists", name)
		}
		return 0, fmt.Errorf("add source %q: %w", name, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("add source %q: last id: %w", name, err)
	}
	return id, nil
}

// ListSources returns every configured source, sorted by name.
func (s *Store) ListSources(ctx context.Context) ([]store.Source, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, kind, url, enabled FROM job_sources ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()
	var out []store.Source
	for rows.Next() {
		var so store.Source
		if err := rows.Scan(&so.ID, &so.Name, &so.Kind, &so.Endpoint, &so.Enabled); err != nil {
			return nil, fmt.Errorf("list sources: scan: %w", err)
		}
		out = append(out, so)
	}
	return out, rows.Err()
}

// SetSourceEnabled turns one source on or off by name; unknown names error.
func (s *Store) SetSourceEnabled(ctx context.Context, name string, enabled bool) error {
	flag := 0
	if enabled {
		flag = 1
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE job_sources SET enabled = ? WHERE name = ?`, flag, name)
	if err != nil {
		return fmt.Errorf("source %q enabled=%v: %w", name, enabled, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("source %q enabled=%v: affected: %w", name, enabled, err)
	}
	if n == 0 {
		return fmt.Errorf("source %q: no such source", name)
	}
	return nil
}

// MarkJobState moves one fingerprint's lifecycle state ('active' or
// 'disappeared'); other values are rejected, never stored.
func (s *Store) MarkJobState(ctx context.Context, fingerprint, state string) error {
	switch state {
	case "active", "disappeared":
	default:
		return fmt.Errorf("mark job %s: state %q must be active or disappeared", fingerprint, state)
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE jobs SET state = ? WHERE fingerprint = ?`, state, fingerprint)
	if err != nil {
		return fmt.Errorf("mark job %s: %w", fingerprint, err)
	}
	return nil
}
