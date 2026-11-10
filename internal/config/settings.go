package config

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
)

// Setting describes one runtime setting row (docs/config.md "Runtime
// settings"): a key, its type, default, validation and help text, plus
// whether changing it needs a process restart to take effect
// (docs/playbooks/add-a-setting.md step 5).
type Setting struct {
	Key     string
	Type    string
	Default string
	Help    string
	// Restart is true when the running process cannot pick the new value up
	// live; the UI must show "restart required" for such keys.
	Restart bool
}

// Registered returns the runtime settings implemented in Phase 2, in UI
// order. Later phases append their keys here and a docs/config.md row.
func Registered() []Setting {
	return []Setting{
		{
			Key: "system.bind_addr", Type: "string", Default: ":8090", Restart: true,
			Help: "HTTP bind for the LAN UI. Port 8080 is reserved for redirect.py. Restart required.",
		},
		{
			Key: "system.log_level", Type: "string", Default: "info", Restart: false,
			Help: "One of debug, info, warn, error. Applies live.",
		},
		{
			Key: "system.log_format", Type: "string", Default: "text", Restart: true,
			Help: "One of text, json. Restart required.",
		},
		{
			Key: "crawl.concurrency", Type: "int", Default: "2", Restart: false,
			Help: "Parallel fetches, 1-8. Keep low on the Pi. Placeholder until the crawler exists (Phase 4).",
		},
		{
			Key: "filter.blacklist_companies", Type: "string", Default: "", Restart: false,
			Help: "Comma-separated companies (case-insensitive substring) that are always rejected. Applies live.",
		},
		{
			Key: "filter.no_gos", Type: "string", Default: "", Restart: false,
			Help: "Comma-separated keywords/phrases that disqualify a job (title/company/location match). Applies live.",
		},
		{
			Key: "filter.must_haves", Type: "string", Default: "", Restart: false,
			Help: "Comma-separated required terms; ALL must match, '|' separates OR alternatives (e.g. berlin|remote). Applies live.",
		},
		{
			Key: "filter.tag_weights", Type: "string", Default: "", Restart: false,
			Help: "Weighted interest tags: specs 'name@weight=kw1|kw2', comma-separated, weight 1-99. Adds an informational score per job (stage 2, no cutoff). Applies live.",
		},
	}
}

// Effective merges stored overrides over registered defaults. Unknown
// override keys are ignored (they are not settings this build knows).
func Effective(overrides map[string]string) map[string]string {
	out := make(map[string]string, len(overrides)+4)
	for _, s := range Registered() {
		if v, ok := overrides[s.Key]; ok {
			out[s.Key] = v
			continue
		}
		out[s.Key] = s.Default
	}
	return out
}

// ValidateSetting checks value for key and returns an error that names the
// key, per docs/playbooks/add-a-setting.md step 3.
func ValidateSetting(key, value string) error {
	switch key {
	case "system.bind_addr":
		_, port, err := net.SplitHostPort(value)
		if err != nil {
			return fmt.Errorf("system.bind_addr %q: must be host:port", value)
		}
		if port == "8080" {
			return fmt.Errorf("system.bind_addr %q: port 8080 is reserved for redirect.py", value)
		}
	case "system.log_level":
		if _, err := ParseLogLevel(value); err != nil {
			return err
		}
	case "system.log_format":
		if value != "text" && value != "json" {
			return fmt.Errorf("system.log_format %q: must be one of text, json", value)
		}
	case "crawl.concurrency":
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 8 {
			return fmt.Errorf("crawl.concurrency %q: must be an integer from 1 to 8", value)
		}
	case "filter.blacklist_companies":
		return validateFilterList("filter.blacklist_companies", value)
	case "filter.no_gos":
		return validateFilterList("filter.no_gos", value)
	case "filter.must_haves":
		return validateFilterList("filter.must_haves", value)
	case "filter.tag_weights":
		return validateFilterList("filter.tag_weights", value)
	default:
		return fmt.Errorf("unknown setting key %q", key)
	}
	return nil
}

// validateFilterList bounds the owner-typed rule lists: single line, sane
// length, no empty sub-entry sneaking in as match-everything (the filter
// package also defends this, but load-time rejection is cheaper).
func validateFilterList(key, value string) error {
	if len(value) > 2000 {
		return fmt.Errorf("%s: value is %d chars, must be at most 2000", key, len(value))
	}
	if strings.Contains(value, "\n") {
		return fmt.Errorf("%s: must be a single line (comma-separated)", key)
	}
	return nil
}

// ParseLogLevel maps a level name to an slog.Level. Shared by main (startup)
// and the hot-reload hook after each settings save.
func ParseLogLevel(name string) (slog.Level, error) {
	switch name {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("system.log_level %q: must be one of debug, info, warn, error", name)
	}
}
