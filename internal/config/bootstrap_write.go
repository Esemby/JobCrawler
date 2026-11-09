package config

import (
	"fmt"
	"os"
	"strings"
)

// WriteBootstrap atomically writes a bootstrap file at path with the given
// engine and DSN. Values are written double-quoted, matching parseKeyValue.
// It is only called from the Database page save; a failing write leaves any
// previous file untouched (write to a temp file in the same dir, then rename).
func WriteBootstrap(path, engine, dsn string) error {
	if strings.Contains(dsn, "\"") || strings.Contains(dsn, "\n") || strings.Contains(engine, "\"") {
		return fmt.Errorf("db.dsn/db.engine: quotes and newlines are not allowed in bootstrap values")
	}
	body := bootstrapBody(engine, dsn)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(body), 0o600); err != nil {
		return fmt.Errorf("write bootstrap %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename bootstrap %s: %w", path, err)
	}
	return nil
}

// bootstrapBody renders the canonical file body. Comments explain that this
// file holds credentials-safe records and that env vars override it.
func bootstrapBody(engine, dsn string) string {
	return fmt.Sprintf(`# jobcrawler bootstrap config (ADR-0002).
# Holds which store engine to use plus its connection string; these values
# cannot live inside the database they describe. Environment variables
# JOBCRAWLER_DB_ENGINE / JOBCRAWLER_DB_DSN override this file.
[db]
engine = %q
dsn = %q
`, engine, dsn)
}
