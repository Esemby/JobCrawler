package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/Esemby/JobCrawler/internal/store"
	"github.com/Esemby/JobCrawler/internal/store/migrations"
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

// maxMigratedVersion reads the highest embedded migration number for the
// engine so the expected schema version is never hard-coded (adding 0003+
// must not break this test).
func maxMigratedVersion(t *testing.T, engine string) int {
	t.Helper()
	highest := 0
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}
	for _, e := range entries {
		var v int
		if n, _ := fmt.Sscanf(e.Name(), "%04d", &v); n == 1 && v > highest {
			highest = v
		}
	}
	return highest
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
	if version < 1 {
		t.Errorf("schema version = %d, want >= 1 after migrations", version)
	}
	if want := maxMigratedVersion(t, "sqlite"); version != want {
		t.Errorf("schema version = %d, want %d (embedded migrations are authoritative)",
			version, want)
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
