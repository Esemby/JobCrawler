package store

import (
	"strings"
	"unicode"
)

// Job is one parsed job offer, the domain record shared by sources, crawler
// and web (docs/glossary.md). Timestamps are RFC3339 strings so both engines
// store them identically. State: 'active' or 'disappeared'; empty means active.
type Job struct {
	SourceID    int64
	SourceName  string
	Title       string
	Company     string
	Location    string
	URL         string
	PublishedAt string
	Fingerprint string
	FirstSeen   string
	LastSeen    string
	State       string
}

// Source is one configured crawler input (docs/sources.md register).
type Source struct {
	ID       int64
	Name     string
	Kind     string
	Endpoint string
	Enabled  bool
}

// Fingerprint builds the stable dedupe key: normalized company + title +
// location + canonical URL (docs/sources.md politeness contract). Changing
// the formula requires migrate-everything; treat it as frozen.
func Fingerprint(j Job) string {
	return strings.Join([]string{
		normalize(j.Company),
		normalize(j.Title),
		normalize(j.Location),
		CanonicalURL(j.URL),
	}, "|")
}

// normalize lowercases, collapses whitespace to single spaces, and trims.
// Punctuation stays: "C++ developer" keeps the language visible.
func normalize(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.TrimSpace(s) {
		lr := unicode.ToLower(r)
		if unicode.IsSpace(lr) {
			space = true
			continue
		}
		if space {
			b.WriteByte(' ')
			space = false
		}
		b.WriteRune(lr)
	}
	return b.String()
}

// CanonicalURL strips empty query pairs and a fixed set of tracking
// parameters (utm_*, gh_src, ref, refsrc) then trailing slashes, so the same
// job survives dedupe across differently tagged links. Everything else is
// kept byte-for-byte.
func CanonicalURL(u string) string {
	qi := strings.IndexByte(u, '?')
	if qi < 0 {
		return strings.TrimRight(u, "/")
	}
	base := strings.TrimRight(u[:qi], "/")
	var kept []string
	for _, p := range strings.Split(u[qi+1:], "&") {
		if p == "" || isTracking(p) {
			continue
		}
		kept = append(kept, p)
	}
	if len(kept) == 0 {
		return base
	}
	return base + "?" + strings.Join(kept, "&")
}

func isTracking(pair string) bool {
	key := pair
	if eq := strings.IndexByte(pair, '='); eq >= 0 {
		key = pair[:eq]
	}
	switch strings.ToLower(key) {
	case "utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "gh_src", "ref", "refsrc":
		return true
	}
	return false
}
