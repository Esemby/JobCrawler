package config

import (
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{"JOBCRAWLER_BIND_ADDR", "JOBCRAWLER_LOG_LEVEL", "JOBCRAWLER_LOG_FORMAT", "JOBCRAWLER_DB_ENGINE"} {
		t.Setenv(key, "")
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with no env: %v", err)
	}
	if cfg.BindAddr != ":8090" {
		t.Errorf("BindAddr = %q, want \":8090\"", cfg.BindAddr)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want \"info\"", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("LogFormat = %q, want \"text\"", cfg.LogFormat)
	}
	if cfg.DBEngine != "sqlite" {
		t.Errorf("DBEngine = %q, want \"sqlite\"", cfg.DBEngine)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	t.Setenv("JOBCRAWLER_BIND_ADDR", "192.168.1.10:9090")
	t.Setenv("JOBCRAWLER_LOG_LEVEL", "debug")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if cfg.BindAddr != "192.168.1.10:9090" || cfg.LogLevel != "debug" {
		t.Errorf("overrides not applied: %+v", cfg)
	}
}

func TestValidateRejects(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{"port 8080 reserved", func(c *Config) { c.BindAddr = ":8080" }, "reserved for redirect.py"},
		{"bad addr", func(c *Config) { c.BindAddr = "nonsense" }, "must be host:port"},
		{"bad level", func(c *Config) { c.LogLevel = "verbose" }, "log_level"},
		{"bad format", func(c *Config) { c.LogFormat = "yaml" }, "log_format"},
		{"bad engine", func(c *Config) { c.DBEngine = "postgres" }, "db_engine"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{BindAddr: ":8090", LogLevel: "info", LogFormat: "text", DBEngine: "sqlite"}
			tt.mutate(&cfg)
			err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAccepts(t *testing.T) {
	cfg := Config{BindAddr: "192.168.1.10:9090", LogLevel: "warn", LogFormat: "json", DBEngine: "mariadb"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}
