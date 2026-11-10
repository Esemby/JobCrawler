package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Esemby/JobCrawler/internal/store"
)

// SaveVerdicts upserts verdicts by fingerprint. The CHECK constraint in
// migration 0003 rejects anything but accepted/rejected; the reason lines
// are newline-joined (roundtrip split on read).
func (s *Store) SaveVerdicts(ctx context.Context, verdicts map[string]store.Verdict) error {
	now := time.Now().UTC().Format(time.RFC3339)
	for fp, v := range verdicts {
		if fp == "" {
			return fmt.Errorf("save verdict: empty fingerprint")
		}
		switch v.Result {
		case store.Accepted, store.Rejected:
		default:
			return fmt.Errorf("save verdict %s: result %q must be accepted or rejected", fp, v.Result)
		}
		joined := strings.Join(v.Reasons, "\n")
		scoreJoin := strings.Join(v.ScoreReasons, "\n")
		// Two statements instead of SQLite-only ON CONFLICT upsert (ADR-0002:
		// both engines run the same code path). Settings upserts predate
		// this rule and are a Phase 6 cleanup.
		var exists int
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(1) FROM job_verdicts WHERE fingerprint = ?`, fp).Scan(&exists); err != nil {
			return fmt.Errorf("save verdict %s: lookup: %w", fp, err)
		}
		if exists > 0 {
			if _, err := s.db.ExecContext(ctx,
				`UPDATE job_verdicts
				    SET result = ?, reasons = ?, score = ?, score_reasons = ?, evaluated_at = ?
				  WHERE fingerprint = ?`,
				v.Result, joined, v.Score, scoreJoin, now, fp); err != nil {
				return fmt.Errorf("save verdict %s: update: %w", fp, err)
			}
			continue
		}
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO job_verdicts
				(fingerprint, result, reasons, score, score_reasons, evaluated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			fp, v.Result, joined, v.Score, scoreJoin, now); err != nil {
			return fmt.Errorf("save verdict %s: insert: %w", fp, err)
		}
	}
	return nil
}

// ListVerdicts returns all stored verdicts keyed by fingerprint. Reason and
// score-breakdown lines are split back from their newline-joined columns;
// empty text yields nil.
func (s *Store) ListVerdicts(ctx context.Context) (map[string]store.Verdict, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT fingerprint, result, reasons, score, score_reasons, evaluated_at
		 FROM job_verdicts`)
	if err != nil {
		return nil, fmt.Errorf("list verdicts: %w", err)
	}
	defer rows.Close()
	out := make(map[string]store.Verdict)
	for rows.Next() {
		var (
			v         store.Verdict
			joined    string
			scoreJoin string
			seen      string
		)
		if err := rows.Scan(&v.Fingerprint, &v.Result, &joined, &v.Score, &scoreJoin, &seen); err != nil {
			return nil, fmt.Errorf("list verdicts: scan: %w", err)
		}
		if joined != "" {
			v.Reasons = strings.Split(joined, "\n")
		}
		if scoreJoin != "" {
			v.ScoreReasons = strings.Split(scoreJoin, "\n")
		}
		v.EvaluatedAt = seen
		out[v.Fingerprint] = v
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list verdicts: iterate: %w", err)
	}
	return out, nil
}
