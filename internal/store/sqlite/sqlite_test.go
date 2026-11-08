package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Esemby/JobCrawler/internal/store"
	"github.com/Esemby/JobCrawler/internal/store/storetest"
)

// TestStoreSuite runs the shared interface suite against SQLite on temp
// files. No network, no port binding.
func TestStoreSuite(t *testing.T) {
	storetest.Run(t, func(t *testing.T) store.Store {
		return newTempStore(t)
	})
}

func newTempStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()
	file := filepath.Join(t.TempDir(), "jobcrawler-test.db")
	s, err := Open(ctx, file)
	if err != nil {
		t.Fatalf("Open(%q): %v", file, err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func openRaw(t *testing.T) (*sql.DB, string) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "raw.db")
	db, err := sql.Open("sqlite", file)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, file
}

func TestMigrateIsIdempotent(t *testing.T) {
	ctx := context.Background()
	db, _ := openRaw(t)

	if err := store.Migrate(ctx, db, "sqlite"); err != nil {
		t.Fatalf("first Migrate(): %v", err)
	}
	if err := store.Migrate(ctx, db, "sqlite"); err != nil {
		t.Fatalf("second Migrate(): %v, want nil (idempotent)", err)
	}

	var version int
	const q = `SELECT MAX(version) FROM schema_migrations`
	if err := db.QueryRowContext(ctx, q).Scan(&version); err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	if version != 1 {
		t.Errorf("schema version = %d, want 1 after migration 0001", version)
	}

	// The skeleton tables exist and are writable.
	if _, err := db.ExecContext(ctx, `INSERT INTO users (name) VALUES ('owner')`); err != nil {
		t.Errorf("insert into users: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO settings (key, value) VALUES ('a', 'b')`); err != nil {
		t.Errorf("insert into settings: %v", err)
	}
}

func TestOpenEnablesWAL(t *testing.T) {
	ctx := context.Background()
	s := newTempStore(t) // same code path main uses
	var mode string
	if err := s.db.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatalf("read journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want \"wal\" (ADR-0002)", mode)
	}
}
