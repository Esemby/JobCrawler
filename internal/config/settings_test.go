package config

import (
	"strings"
	"testing"
)

func TestValidateSetting(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr string
	}{
		{"bind addr accept", "system.bind_addr", "192.168.1.10:9090", ""},
		{"bind addr port 8080 rejected", "system.bind_addr", ":8080", "reserved for redirect.py"},
		{"bind addr malformed", "system.bind_addr", "nonsense", "must be host:port"},
		{"log level accept", "system.log_level", "warn", ""},
		{"log level reject", "system.log_level", "verbose", "system.log_level"},
		{"log format accept", "system.log_format", "json", ""},
		{"log format reject", "system.log_format", "yaml", "system.log_format"},
		{"concurrency accept", "crawl.concurrency", "8", ""},
		{"concurrency too high", "crawl.concurrency", "9", "1 to 8"},
		{"concurrency too low", "crawl.concurrency", "0", "1 to 8"},
		{"concurrency not numeric", "crawl.concurrency", "many", "1 to 8"},
		{"unknown key rejected", "filters.magic", "yes", `unknown setting key "filters.magic"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSetting(tt.key, tt.value)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidateSetting(%q, %q) err = %v, want nil", tt.key, tt.value, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidateSetting(%q, %q) err = %v, want containing %q", tt.key, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestEffectiveMergesDefaults(t *testing.T) {
	got := Effective(map[string]string{
		"system.log_level": "debug",
		"not.a.setting":    "ignored",
	})
	if got["system.log_level"] != "debug" {
		t.Errorf("override not applied: got %q", got["system.log_level"])
	}
	if got["system.log_format"] != "text" {
		t.Errorf("default missing: got %q", got["system.log_format"])
	}
	if got["crawl.concurrency"] != "2" {
		t.Errorf("default crawl.concurrency = %q, want \"2\"", got["crawl.concurrency"])
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"debug", "debug", false},
		{"info", "info", false},
		{"warn", "warn", false},
		{"error", "error", false},
		{"reject", "loud", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseLogLevel(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLogLevel(%q) err = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
