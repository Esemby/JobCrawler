package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// isNoRows reports err for "no such row" through any driver wrapping.
func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// GetSetting returns the value for key; found=false, err=nil when absent.
func (s *Store) GetSetting(ctx context.Context, key string) (string, bool, error) {
	var value string
	const q = `SELECT value FROM settings WHERE key = ?`
	err := s.db.QueryRowContext(ctx, q, key).Scan(&value)
	switch {
	case err == nil:
		return value, true, nil
	case isNoRows(err):
		return "", false, nil
	default:
		return "", false, fmt.Errorf("get setting %q: %w", key, err)
	}
}

// SetSetting upserts key=value, covering both create and update.
func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	const q = `INSERT INTO settings (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value`
	if _, err := s.db.ExecContext(ctx, q, key, value); err != nil {
		return fmt.Errorf("set setting %q: %w", key, err)
	}
	return nil
}

// ListSettings returns all stored key/value overrides; an empty map means
// "no runtime overrides, use defaults from internal/config".
func (s *Store) ListSettings(ctx context.Context) (map[string]string, error) {
	const q = `SELECT key, value FROM settings`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan settings row: %w", err)
		}
		out[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate settings: %w", err)
	}
	return out, nil
}

// DeleteSetting removes a key; deleting an absent key succeeds.
func (s *Store) DeleteSetting(ctx context.Context, key string) error {
	const q = `DELETE FROM settings WHERE key = ?`
	if _, err := s.db.ExecContext(ctx, q, key); err != nil {
		return fmt.Errorf("delete setting %q: %w", key, err)
	}
	return nil
}
