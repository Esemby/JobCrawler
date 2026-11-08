package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Esemby/JobCrawler/internal/store/migrations"
)

// Migrate brings the database up to the latest embedded schema version for
// the given dialect ("sqlite" today, "mysql" from Phase 6). It is idempotent:
// applied versions are recorded in schema_migrations and skipped on re-run
// (docs/playbooks/add-a-migration.md step 4), so callers may run Migrate on
// every open.
func Migrate(ctx context.Context, db *sql.DB, dialect string) error {
	if err := ensureBookkeeping(ctx, db); err != nil {
		return err
	}
	current, err := currentVersion(ctx, db)
	if err != nil {
		return err
	}
	toRun, err := pickFiles(dialect)
	if err != nil {
		return err
	}
	for _, f := range toRun {
		if f.version <= current {
			continue
		}
		if err := applyOne(ctx, db, f); err != nil {
			return err
		}
		current = f.version
	}
	return nil
}

type migrationFile struct {
	version int
	name    string
}

// ensureBookkeeping creates the version table. CURRENT_TIMESTAMP is valid in
// both SQLite and MariaDB, so the bookkeeping stays engine-less.
func ensureBookkeeping(ctx context.Context, db *sql.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS schema_migrations (
	version    INTEGER PRIMARY KEY,
	applied_at TEXT    NOT NULL DEFAULT (CURRENT_TIMESTAMP)
)`
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	return nil
}

func currentVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	const q = `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`
	if err := db.QueryRowContext(ctx, q).Scan(&version); err != nil {
		return 0, fmt.Errorf("read applied schema version: %w", err)
	}
	return version, nil
}

// pickFiles selects one migration file per version: the dialect file if one
// exists ("NNNN_desc.sqlite.sql"), otherwise the engine-less shared file.
// Dialect files of other engines are skipped. Duplicates are an error: a
// version may be defined at most once for a given dialect.
func pickFiles(dialect string) ([]migrationFile, error) {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("list embedded migrations: %w", err)
	}
	byVersion := map[int]string{}
	for _, e := range entries {
		name := e.Name()
		version, fileDialect := versionAndDialect(name)
		if version == 0 {
			return nil, fmt.Errorf("migration %q: name must start with NNNN_", name)
		}
		if fileDialect != "" && fileDialect != dialect {
			continue // other engine's file; picked when that dialect runs
		}
		if _, exists := byVersion[version]; exists && fileDialect != "" {
			return nil, fmt.Errorf("migration %d: both shared file and %s file exist", version, dialect)
		}
		if _, exists := byVersion[version]; !exists {
			byVersion[version] = name
		}
	}
	out := make([]migrationFile, 0, len(byVersion))
	for v, name := range byVersion {
		out = append(out, migrationFile{version: v, name: name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

// versionAndDialect extracts the leading version number and, if the file is
// dialect-specific, its dialect name: "0001_init.sql" -> (1, "").
// "0001_init.sqlite.sql" -> (1, "sqlite"). Unknown shapes are (0, "").
func versionAndDialect(name string) (version int, dialect string) {
	dot := strings.LastIndex(name, ".sql")
	if dot < 1 {
		return 0, ""
	}
	base := name[:dot]
	dialectSuffix := strings.LastIndex(base, ".")
	if dialectSuffix > 0 {
		dialect = base[dialectSuffix+1:]
		base = base[:dialectSuffix]
	}
	underscore := strings.IndexByte(base, '_')
	if underscore < 1 {
		return 0, ""
	}
	v, err := strconv.Atoi(base[:underscore])
	if err != nil {
		return 0, ""
	}
	return v, dialect
}

// applyOne runs one migration file inside its own transaction and records
// the version. Full-line "--" comments are stripped first, because comment
// prose may contain ";" (this bit the first runner version); statements are
// then split on ";". The migration playbook forbids triggers and stored
// procedures, so a semicolon can never be part of a string literal that
// splitting would break.
func applyOne(ctx context.Context, db *sql.DB, f migrationFile) (err error) {
	sqlBytes, err := migrations.FS.ReadFile(f.name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", f.name, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", f.name, err)
	}
	// Rollback only when the body failed; after Commit the transaction is
	// done, and Rollback would report sql.ErrTxDone, which is not an error.
	defer func() {
		if err == nil {
			return
		}
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			// The original error already names the migration; the rollback
			// failure is appended as context, it must not replace it.
			err = fmt.Errorf("%w (rollback also failed: %v)", err, rbErr)
		}
	}()

	for _, stmt := range splitStatements(string(sqlBytes)) {
		if _, err = tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply migration %s: %w", f.name, err)
		}
	}
	const q = `INSERT INTO schema_migrations (version) VALUES (?)`
	if _, err = tx.ExecContext(ctx, q, f.version); err != nil {
		return fmt.Errorf("record migration %s: %w", f.name, err)
	}
	return tx.Commit()
}

// splitStatements turns raw migration SQL into executable statements:
// full-line comments dropped, remainder split on ";" and trimmed. Inline
// trailing comments stay — drivers handle them.
func splitStatements(sqlText string) []string {
	var lines []string
	for _, line := range strings.Split(sqlText, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		lines = append(lines, line)
	}
	var out []string
	for _, part := range strings.Split(strings.Join(lines, "\n"), ";") {
		stmt := strings.TrimSpace(part)
		if stmt != "" {
			out = append(out, stmt)
		}
	}
	return out
}
