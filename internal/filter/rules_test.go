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
	tags := TagRules("rust@5=rust")
	out := EvaluateAll(jobs, cfg, tags)
	if len(out) != 2 {
		t.Fatalf("map size = %d, want 2", len(out))
	}
	if out["fa"].Result != Accepted || out["fb"].Result != Rejected {
		t.Errorf("verdict map wrong: %+v", out)
	}
	if out["fb"].Reasons[0] != "company blacklist match: acme" {
		t.Errorf("reason string changed: %q", out["fb"].Reasons[0])
	}
	if out["fa"].Score != 0 {
		t.Errorf("job with no tag keyword scored %d, want 0", out["fa"].Score)
	}
}

func TestTagRulesParse(t *testing.T) {
	rules := TagRules("rust@5=rust | rustc, lang-de@3=deutsch|german, broken@=x, =y, bad@2=")
	if len(rules) != 3 {
		t.Fatalf("rules = %d (+%v), want 3 (empty name or no keywords dropped)", len(rules), rules)
	}
	if rules[0].Name != "rust" || rules[0].Weight != 5 ||
		rules[0].Keywords[0] != "rust" || rules[0].Keywords[1] != "rustc" {
		t.Errorf("rule 0 = %+v", rules[0])
	}
	if rules[1].Name != "lang-de" || rules[1].Weight != 3 {
		t.Errorf("rule 1 = %+v", rules[1])
	}
	// Missing weight falls back to 1 and keeps the tag usable.
	if rules[2].Name != "broken" || rules[2].Weight != 1 {
		t.Errorf("rule 2 = %+v, want weight fallback 1", rules[2])
	}
	if got := TagRules(""); len(got) != 0 {
		t.Errorf("empty value produced %v", got)
	}
}

func TestScoreTagRules(t *testing.T) {
	rules := TagRules("rust@5=rust, lang-de@3=deutsch|german")
	score, breakdown := ScoreTagRules("backend engineer, deutsch-fließend", rules)
	if score != 3 {
		t.Errorf("score = %d, want 3", score)
	}
	if len(breakdown) != 1 || breakdown[0] != `tag lang-de matched "deutsch" (weight 3)` {
		t.Errorf("breakdown = %v", breakdown)
	}
	// Weights add up.
	score, _ = ScoreTagRules("rust developer (german team)", rules)
	if score != 8 {
		t.Errorf("combined score = %d, want 8", score)
	}
	// No rule matches.
	score, breakdown = ScoreTagRules("python wizard", rules)
	if score != 0 || breakdown != nil {
		t.Errorf("no-match score/breakdown = %d %v", score, breakdown)
	}
}
