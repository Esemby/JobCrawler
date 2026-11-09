package filter

import (
	"testing"

	"github.com/Esemby/JobCrawler/internal/store"
)

func TestEvaluateBlacklist(t *testing.T) {
	j := store.Job{Title: "Backend Engineer", Company: "ACME GmbH",
		URL: "https://x.test/a", Fingerprint: "fp1"}
	cfg := Config{BlacklistCompanies: []string{"acme"}}
	v := Evaluate(j, cfg)
	if v.Result != Rejected {
		t.Errorf("result = %q, want rejected", v.Result)
	}
	if len(v.Reasons) != 1 || v.Reasons[0] != "company blacklist match: acme" {
		t.Errorf("reasons = %v, want exact blacklist reason", v.Reasons)
	}
}

func TestEvaluateNoGo(t *testing.T) {
	j := store.Job{Title: "Unpaid internship: backend", Company: "Nice Co",
		URL: "https://x.test/b"}
	cfg := Config{NoGos: []string{"unpaid"}}
	v := Evaluate(j, cfg)
	if v.Result != Rejected {
		t.Fatalf("result = %q, want rejected", v.Result)
	}
	// No-gos also scan company + location, not just the title.
	j2 := store.Job{Title: "Engineer", Location: "Bonn unpaid experiment"}
	if v2 := Evaluate(j2, cfg); v2.Result != Rejected {
		t.Errorf("no-go missed in location")
	}
}

func TestEvaluateMustHavesAND(t *testing.T) {
	cfg := Config{MustHaves: []string{"senior", "rust"}}
	both := store.Job{Title: "Senior Rust Developer"}
	if v := Evaluate(both, cfg); v.Result != Accepted ||
		v.Reasons[0] != "passed stage 1 (no rule fired)" {
		t.Errorf("AND-complete job rejected: %+v", v)
	}
	one := store.Job{Title: "Rust Developer"}
	v := Evaluate(one, cfg)
	if v.Result != Rejected {
		t.Fatalf("AND- incomplete job accepted")
	}
	if len(v.Reasons) != 1 || v.Reasons[0] != "must-have not met: senior" {
		t.Errorf("reasons = %v, want exactly the missing entry", v.Reasons)
	}
}

func TestEvaluateMustHavesORGroup(t *testing.T) {
	cfg := Config{MustHaves: []string{"berlin|remote"}}
	m := store.Job{Title: "Backend Engineer", Location: "Remote, United States"}
	if v := Evaluate(m, cfg); v.Result != Accepted {
		t.Errorf("OR group should match remote: %+v", v)
	}
	n := store.Job{Title: "Backen Engineer", Location: "München"}
	if v := Evaluate(n, cfg); v.Result != Rejected {
		t.Errorf("missed OR group should reject: %+v", v)
	}
}

func TestEvaluateReasonOrderIsDeterministic(t *testing.T) {
	j := store.Job{Title: "senior unpaid thing", Company: "acme corp",
		Fingerprint: "fp"}
	cfg := Config{
		BlacklistCompanies: []string{"acme"},
		NoGos:              []string{"unpaid"},
		MustHaves:          []string{"impossible-word"},
	}
	for i := 0; i < 5; i++ {
		v := Evaluate(j, cfg)
		want := []string{
			"company blacklist match: acme",
			"no-go match: unpaid",
			"must-have not met: impossible-word",
		}
		if len(v.Reasons) != 3 || v.Reasons[0] != want[0] || v.Reasons[1] != want[1] || v.Reasons[2] != want[2] {
			t.Fatalf("round %d reasons = %v, want %v", i, v.Reasons, want)
		}
	}
}

func TestEvaluateEmptyConfigAcceptsAll(t *testing.T) {
	j := store.Job{Title: "Whatever", Fingerprint: "fp"}
	v := Evaluate(j, Config{})
	if v.Result != Accepted || v.Reasons[0] != "passed stage 1 (no rule fired)" {
		t.Errorf("empty config verdict = %+v", v)
	}
}

func TestConfigFromSettings(t *testing.T) {
	cfg := ConfigFromSettings(map[string]string{
		"filter.blacklist_companies": " Acme Corp ,  test  , ,",
		"filter.no_gos":              "UNPAID, unpaid internship",
		"filter.must_haves":          "berlin| Remote ,rust",
		"some.other.key":             "ignored",
	})
	// Settings keywords are lowercased for matching but reason strings keep
	// the lowcased form (docs claim: reasons must still be understandable —
	// values are owner-typed rule text, lowcasing whole keywords is fine).
	wantBL := []string{"acme corp", "test"}
	if cfg.BlacklistCompanies[0] != wantBL[0] || cfg.BlacklistCompanies[1] != wantBL[1] {
		t.Errorf("blacklist = %v", cfg.BlacklistCompanies)
	}
	if cfg.NoGos[0] != "unpaid" || cfg.NoGos[1] != "unpaid internship" {
		t.Errorf("no-gos = %v", cfg.NoGos)
	}
	if cfg.MustHaves[0] != "berlin| remote" || cfg.MustHaves[1] != "rust" {
		t.Errorf("must-haves = %v", cfg.MustHaves)
	}
}

func TestConfigFromSettingsEmptyLeavesRulesInactive(t *testing.T) {
	cfg := ConfigFromSettings(map[string]string{})
	if len(cfg.BlacklistCompanies)+len(cfg.NoGos)+len(cfg.MustHaves) != 0 {
		t.Errorf("empty settings produced rules: %+v", cfg)
	}
	if v := Evaluate(store.Job{Title: "x"}, cfg); v.Result != Accepted {
		t.Errorf("empty config must accept all")
	}
}

func TestEvaluateAll(t *testing.T) {
	jobs := []store.Job{
		{Title: "A", Fingerprint: "fa"},
		{Title: "B", Company: "acme", Fingerprint: "fb"},
	}
	cfg := Config{BlacklistCompanies: []string{"acme"}}
	out := EvaluateAll(jobs, cfg)
	if len(out) != 2 {
		t.Fatalf("map size = %d, want 2", len(out))
	}
	if out["fa"].Result != Accepted || out["fb"].Result != Rejected {
		t.Errorf("verdict map wrong: %+v", out)
	}
	if out["fb"].Reasons[0] != "company blacklist match: acme" {
		t.Errorf("reason string changed: %q", out["fb"].Reasons[0])
	}
}
