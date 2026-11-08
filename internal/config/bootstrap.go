// bootstrap.go: bootstrap-file (config.bootstrap.toml) parsing, engine + DSN
// per ADR-0002.
//
// The bootstrap TOML is a 5-line file, so parsing is deliberately hand-rolled
// (no third-party TOML dependency, per the "Go standard library first" rule).
// Accepted shape, best effort line by line:
//
//	# comment
//	[db]
//	engine = "sqlite"
//	dsn    = "jobcrawler.db"

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// DefaultBootstrapPath is where main looks for the bootstrap file unless
// JOBCRAWLER_CONFIG points elsewhere.
const DefaultBootstrapPath = "config.bootstrap.toml"

// Bootstrap is the storage-engine layer of the configuration (ADR-0002).
// These settings cannot live inside the database they describe, and they are
// never stored in the managed DB either.
type Bootstrap struct {
	// Path names the file the values came from (error messages, UI).
	Path string
	// Engine is the storage engine: sqlite or mariadb.
	Engine string
	// DSN is the connection string of the active engine. For sqlite this is
	// a file path (no secret); for mariadb it carries credentials and must
	// never be logged or round-tripped in clear.
	DSN string
}

// LoadBootstrap reads the bootstrap file at path, then applies environment
// overrides (env wins over file, file wins over defaults):
//
//	JOBCRAWLER_DB_ENGINE   overrides engine
//	JOBCRAWLER_DB_DSN      overrides DSN
//
// A missing file at DefaultBootstrapPath is fine on a fresh install (defaults
// are used); a missing file at any explicitly configured path is an error.
func LoadBootstrap(path string) (Bootstrap, error) {
	b := Bootstrap{Path: path, Engine: "sqlite"}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if parseErr := parseBootstrap(string(data), &b); parseErr != nil {
			return Bootstrap{}, fmt.Errorf("%s: %w", path, parseErr)
		}
	case errors.Is(err, fs.ErrNotExist) && path == DefaultBootstrapPath:
		b.DSN = "jobcrawler.db" // fresh install: defaults + env below
	default:
		return Bootstrap{}, fmt.Errorf("read bootstrap config %s: %w", path, err)
	}
	if v, ok := os.LookupEnv("JOBCRAWLER_DB_ENGINE"); ok && v != "" {
		b.Engine = v
	}
	if v, ok := os.LookupEnv("JOBCRAWLER_DB_DSN"); ok && v != "" {
		b.DSN = v
	}
	if err := b.Validate(); err != nil {
		return Bootstrap{}, err
	}
	return b, nil
}

// Validate checks the bootstrap values; errors always name the key.
func (b Bootstrap) Validate() error {
	switch b.Engine {
	case "sqlite":
	case "mariadb":
		if b.DSN == "" {
			return fmt.Errorf("db.dsn: non-empty DSN required when engine is mariadb")
		}
	default:
		return fmt.Errorf("db.engine %q: must be one of sqlite, mariadb", b.Engine)
	}
	if b.DSN == "" {
		return fmt.Errorf("db.dsn: must not be empty")
	}
	return nil
}

// parseBootstrap fills b from the file body. Line numbers appear in errors:
// the file is operator-owned and must be easy to fix by hand.
func parseBootstrap(body string, b *Bootstrap) error {
	inDB := false
	for n, line := range strings.Split(body, "\n") {
		lineNo := n + 1
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			table, err := parseTableHeader(trimmed)
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			inDB = table == "db"
			if !inDB {
				return fmt.Errorf("line %d: unknown table [%s]: only [db] is supported", lineNo, table)
			}
			continue
		}
		if !inDB {
			return fmt.Errorf("line %d: %q must appear under the [db] table", lineNo, trimmed)
		}
		key, quoted, err := parseKeyValue(trimmed)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNo, err)
		}
		switch key {
		case "engine":
			b.Engine = quoted
		case "dsn":
			b.DSN = quoted
		default:
			return fmt.Errorf("line %d: unknown key %q: supported keys are engine, dsn", lineNo, key)
		}
	}
	return nil
}

// parseTableHeader accepts exactly "[db]" (trimmed).
func parseTableHeader(line string) (string, error) {
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return "", fmt.Errorf("malformed table header")
	}
	return strings.TrimSpace(line[1 : len(line)-1]), nil
}

// parseKeyValue splits "key = \"value\"" into key and the unquoted value.
// Only double-quoted string values are accepted; anything else is a hard
// error so a mistyped bootstrap file never starts a half-configured service.
func parseKeyValue(line string) (key, value string, err error) {
	eq := strings.IndexByte(line, '=')
	if eq < 1 {
		return "", "", fmt.Errorf("missing key = \"value\" pair")
	}
	key = strings.TrimSpace(line[:eq])
	raw := strings.TrimSpace(line[eq+1:])
	if !strings.HasPrefix(raw, `"`) || !strings.HasSuffix(raw, `"`) || len(raw) < 2 {
		return "", "", fmt.Errorf("value for %q must be a double-quoted string", key)
	}
	return key, raw[1 : len(raw)-1], nil
}
