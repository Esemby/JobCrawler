package filter

import (
	"strings"

	"github.com/Esemby/JobCrawler/internal/store"
)

// Result values mirror store.Accepted / store.Rejected (the strings are
// also CHECKed by migration 0003); aliases live here so filter code reads
// naturally without an import cycle.
const (
	Accepted = store.Accepted
	Rejected = store.Rejected
)

// Config is one evaluated rule set. Every field is a lowercase keyword
// list; empty list means "rule inactive, jobs pass it".
type Config struct {
	// BlacklistCompanies: any of these as substring (case-insensitive) in
	// the company -> reject.
	BlacklistCompanies []string
	// NoGos: any of these in title OR company OR location -> reject.
	// (Description matching arrives when sources fetch content; see
	// docs/sources.md.)
	NoGos []string
	// MustHaves: EVERY entry must match somewhere in title OR company OR
	// location. Inside one entry, "|" separates OR alternatives.
	// E.g. "berlin|remote" matches either; "senior, rust" requires both.
	MustHaves []string
}

// Verdict is the stage-1 result for one job (docs/glossary.md: Verdict,
// Reason). Reasons are ordered blacklist, no-gos, must-haves — the order is
// deterministic so UI display and tests stay stable.
type Verdict struct {
	Result  string
	Reasons []string
}

// ConfigFromSettings parses the three comma-separated setting values
// (docs/config.md rows filter.*). Unknown whitespace and empty entries are
// dropped: an empty keyword list never matches anything and never rejects.
func ConfigFromSettings(effective map[string]string) Config {
	return Config{
		BlacklistCompanies: splitList(effective["filter.blacklist_companies"]),
		NoGos:              splitList(effective["filter.no_gos"]),
		MustHaves:          splitList(effective["filter.must_haves"]),
	}
}

// splitList trims entries and lowercases keywords; blank entries vanish.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

// Evaluate runs the stage-1 rules in fixed order and returns the verdict.
// An empty config accepts everything with one plain reason.
func Evaluate(j store.Job, cfg Config) Verdict {
	var reasons []string

	// Haystack: the fields we have today, lowercased once. Matching is
	// substring-based; "exact and pattern" from the brief is approximated by
	// substring for v1 and stays explainable either way.
	haystack := strings.ToLower(strings.Join([]string{
		strings.TrimSpace(j.Title), strings.TrimSpace(j.Company),
		strings.TrimSpace(j.Location),
	}, " "))

	for _, kw := range cfg.BlacklistCompanies {
		if kw != "" && strings.Contains(haystack, kw) {
			reasons = append(reasons, "company blacklist match: "+kw)
		}
	}
	for _, kw := range cfg.NoGos {
		if kw != "" && strings.Contains(haystack, kw) {
			reasons = append(reasons, "no-go match: "+kw)
		}
	}
	for _, entry := range cfg.MustHaves {
		if entry == "" {
			continue
		}
		if !mustEntryMatch(entry, haystack) {
			reasons = append(reasons, "must-have not met: "+entry)
		}
	}

	if len(reasons) > 0 {
		return Verdict{Result: Rejected, Reasons: reasons}
	}
	return Verdict{Result: Accepted, Reasons: []string{"passed stage 1 (no rule fired)"}}
}

// mustEntryMatch: an entry is either a plain keyword or an OR group
// "a|b|c"; the group matches when any alternative appears.
func mustEntryMatch(entry, haystack string) bool {
	for _, alt := range strings.Split(entry, "|") {
		alt = strings.TrimSpace(alt)
		if alt != "" && strings.Contains(haystack, alt) {
			return true
		}
	}
	return false
}

// EvaluateAll maps jobs to verdicts, keyed by fingerprint. It never writes.
func EvaluateAll(jobs []store.Job, cfg Config) map[string]store.Verdict {
	out := make(map[string]store.Verdict, len(jobs))
	for _, j := range jobs {
		v := Evaluate(j, cfg)
		out[j.Fingerprint] = store.Verdict{
			Fingerprint: j.Fingerprint,
			Result:      v.Result,
			Reasons:     v.Reasons,
		}
	}
	return out
}
