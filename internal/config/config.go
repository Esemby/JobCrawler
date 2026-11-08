package config

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Config is the Phase 1 bootstrap configuration. New settings must follow
// docs/playbooks/add-a-setting.md and get an entry in docs/config.md.
type Config struct {
	// BindAddr is the host:port the LAN web UI binds to. Default ":8090".
	BindAddr string
	// LogLevel is one of debug, info, warn, error. Default "info".
	LogLevel string
	// LogFormat is one of text, json. Default "text".
	LogFormat string
	// DBEngine is the storage engine (ADR-0002). Default "sqlite".
	DBEngine string
}

// Load builds Config from defaults and environment overrides.
// Environment keys used: JOBCRAWLER_BIND_ADDR, JOBCRAWLER_LOG_LEVEL,
// JOBCRAWLER_LOG_FORMAT, JOBCRAWLER_DB_ENGINE. It errors when any value is
// invalid so main never starts on a broken configuration.
func Load() (Config, error) {
	cfg := Config{
		BindAddr:  getenv("JOBCRAWLER_BIND_ADDR", ":8090"),
		LogLevel:  getenv("JOBCRAWLER_LOG_LEVEL", "info"),
		LogFormat: getenv("JOBCRAWLER_LOG_FORMAT", "text"),
		DBEngine:  getenv("JOBCRAWLER_DB_ENGINE", "sqlite"),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks every field and returns the first error found. It is called
// by Load and by tests; port 8080 is rejected because redirect.py owns it.
func (c Config) Validate() error {
	if _, port, err := net.SplitHostPort(c.BindAddr); err != nil {
		return fmt.Errorf("bind_addr %q: must be host:port: %w", c.BindAddr, err)
	} else if port == "8080" {
		return fmt.Errorf("bind_addr %q: port 8080 is reserved for redirect.py, use a different port", c.BindAddr)
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log_level %q: must be one of debug, info, warn, error", c.LogLevel)
	}
	switch c.LogFormat {
	case "text", "json":
	default:
		return fmt.Errorf("log_format %q: must be one of text, json", c.LogFormat)
	}
	switch c.DBEngine {
	case "sqlite", "mariadb":
	default:
		return fmt.Errorf("db_engine %q: must be one of sqlite, mariadb", c.DBEngine)
	}
	return nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}
