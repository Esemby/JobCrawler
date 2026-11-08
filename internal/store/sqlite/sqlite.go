package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver (pure Go)

	"github.com/Esemby/JobCrawler/internal/store"
)

// Store is the SQLite-backed store.Store implementation. Create with Open.
// It is safe for concurrent HTTP-handler use; the driver serializes writes.
type Store struct {
	db *sql.DB
}

// static check: *sqlite.Store must satisfy store.Store.
var _ store.Store = (*Store)(nil)

// Open opens (creating if needed) the SQLite database at file, enables WAL
// mode plus a busy timeout, applies all embedded migrations, and verifies
// connectivity. WAL is set explicitly here rather than via DSN pragmas so
// the code documents the policy in one place (ADR-0002).
func Open(ctx context.Context, file string) (*Store, error) {
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", file, err)
	}
	if err := configure(ctx, db, file); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.Migrate(ctx, db, "sqlite"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate sqlite %q: %w", file, err)
	}
	return &Store{db: db}, nil
}

// configure applies connection pragmas and pings. Each pragma is one round
// trip; the Heart is slow, but this costs milliseconds only at open.
func configure(ctx context.Context, db *sql.DB, file string) error {
	for _, pragma := range []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA busy_timeout = 5000`,
		`PRAGMA foreign_keys = ON`,
		`PRAGMA synchronous = NORMAL`,
	} {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("sqlite %q: %s: %w", file, pragma, err)
		}
	}
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping sqlite %q: %w", file, err)
	}
	return nil
}

// Close releases the database handle. It is idempotent (database/sql
// guarantees a second Close is a no-op error-wise).
func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close sqlite: %w", err)
	}
	return nil
}
