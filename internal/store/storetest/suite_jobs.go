package storetest

import (
	"context"
	"strings"
	"testing"

	"github.com/Esemby/JobCrawler/internal/store"
)

// runJobs extends the contract suite with source + job lifecycle tests. It
// runs through the same open seam so every backend must pass identically.
func runJobs(t *testing.T, open func(t *testing.T) store.Store) {
	t.Helper()
	cases := []struct {
		name string
		run  func(t *testing.T, s store.Store)
	}{
		{"add and list source", addAndListSource},
		{"duplicate source errors", duplicateSourceErrors},
		{"toggle source", toggleSource},
		{"upsert inserts then dedupes", upsertThenDedupe},
		{"fingerprint stable despite querystring", fingerprintStableSURL},
		{"list filters by state", listFiltersByState},
		{"mark state rejects junk", markStateRejectsJunk},
		{"empty title rejected", emptyTitleRejected},
		{"save and list verdicts", saveAndListVerdicts},
		{"verdict upsert overwrites", verdictUpsertOverwrites},
		{"verdict junk result rejected", verdictJunkResultRejected},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := open(t)
			t.Cleanup(func() { _ = s.Close() })
			tc.run(t, s)
		})
	}
}

func addAndListSource(t *testing.T, s store.Store) {
	ctx := context.Background()
	if _, err := s.AddSource(ctx, "greenhouse-acme", "greenhouse", "acme"); err != nil {
		t.Fatalf("AddSource: %v", err)
	}
	srcs, err := s.ListSources(ctx)
	if err != nil {
		t.Fatalf("ListSources: %v", err)
	}
	if len(srcs) != 1 || !srcs[0].Enabled || srcs[0].Kind != "greenhouse" {
		t.Fatalf("sources = %+v, want one enabled greenhouse source", srcs)
	}
}

func duplicateSourceErrors(t *testing.T, s store.Store) {
	ctx := context.Background()
	if _, err := s.AddSource(ctx, "greenhouse-acme", "greenhouse", "acme"); err != nil {
		t.Fatalf("first AddSource: %v", err)
	}
	_, err := s.AddSource(ctx, "greenhouse-acme", "greenhouse", "acme")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second AddSource err = %v, want duplicate error", err)
	}
}

func toggleSource(t *testing.T, s store.Store) {
	ctx := context.Background()
	if _, err := s.AddSource(ctx, "greenhouse-acme", "greenhouse", "acme"); err != nil {
		t.Fatalf("AddSource: %v", err)
	}
	if err := s.SetSourceEnabled(ctx, "greenhouse-acme", false); err != nil {
		t.Fatalf("SetSourceEnabled(false): %v", err)
	}
	srcs, _ := s.ListSources(ctx)
	if srcs[0].Enabled {
		t.Errorf("source still enabled")
	}
	if err := s.SetSourceEnabled(ctx, "nosuch", false); err == nil {
		t.Errorf("unknown source must error")
	}
}

func upsertThenDedupe(t *testing.T, s store.Store) {
	ctx := context.Background()
	id, _ := s.AddSource(ctx, "greenhouse-acme", "greenhouse", "acme")
	j := store.Job{SourceID: id, Title: "Backend Engineer", Company: "ACME GmbH",
		URL: "https://job-boards.greenhouse.io/acme/jobs/1"}
	n1, err := s.UpsertJobs(ctx, []store.Job{j})
	if err != nil {
		t.Fatalf("UpsertJobs #1: %v", err)
	}
	n2, err := s.UpsertJobs(ctx, []store.Job{j})
	if err != nil {
		t.Fatalf("UpsertJobs #2: %v", err)
	}
	if n1 != 1 || n2 != 0 {
		t.Errorf("inserted = %d then %d, want 1 then 0", n1, n2)
	}
}

func fingerprintStableSURL(t *testing.T, s store.Store) {
	ctx := context.Background()
	id, _ := s.AddSource(ctx, "greenhouse-acme", "greenhouse", "acme")
	one := store.Job{SourceID: id, Title: "Backend Engineer", Company: "ACME",
		URL: "https://boards.greenhouse.io/acme/jobs/1?gh_src=abc"}
	two := store.Job{SourceID: id, Title: "backend  engineer", Company: "acme",
		URL: "https://boards.greenhouse.io/acme/jobs/1/"}
	j1 := store.Job{SourceID: id, Title: one.Title, URL: one.URL}
	if _, err := s.UpsertJobs(ctx, []store.Job{one}); err != nil {
		t.Fatalf("UpsertJobs: %v", err)
	}
	jobs, err := s.ListJobs(ctx, "")
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("got %d jobs, want 1 deduped", len(jobs))
	}
	// The second job's fingerprint matches the stored one (normalization):
	// verify via pure function equality instead of storing a near-duplicate.
	if store.Fingerprint(two) != jobs[0].Fingerprint {
		t.Errorf("normalization differs: %q vs %q", store.Fingerprint(two), jobs[0].Fingerprint)
	}
	_ = j1
}

func listFiltersByState(t *testing.T, s store.Store) {
	ctx := context.Background()
	id, _ := s.AddSource(ctx, "greenhouse-acme", "greenhouse", "acme")
	a := store.Job{SourceID: id, Title: "A role", URL: "https://x.test/a"}
	b := store.Job{SourceID: id, Title: "B role", URL: "https://x.test/b"}
	if _, err := s.UpsertJobs(ctx, []store.Job{a, b}); err != nil {
		t.Fatalf("UpsertJobs: %v", err)
	}
	afp := store.Fingerprint(a)
	if err := s.MarkJobState(ctx, afp, "disappeared"); err != nil {
		t.Fatalf("MarkJobState: %v", err)
	}
	active, _ := s.ListJobs(ctx, "active")
	if len(active) != 1 || active[0].Title != "B role" {
		t.Fatalf("active = %+v, want only B", active)
	}
	all, _ := s.ListJobs(ctx, "*")
	if len(all) != 2 {
		t.Fatalf("*all returned %d, want 2", len(all))
	}
}

func markStateRejectsJunk(t *testing.T, s store.Store) {
	ctx := context.Background()
	if err := s.MarkJobState(ctx, "fp", "weird"); err == nil {
		t.Errorf("junk state accepted")
	}
}

func emptyTitleRejected(t *testing.T, s store.Store) {
	ctx := context.Background()
	_, err := s.UpsertJobs(ctx, []store.Job{{URL: "https://x.test/x"}})
	if err == nil {
		t.Errorf("empty title accepted")
	}
}

func saveAndListVerdicts(t *testing.T, s store.Store) {
	ctx := context.Background()
	vs := map[string]store.Verdict{
		"fp-acc": {Result: store.Accepted,
			Reasons: []string{"passed stage 1 (no rule fired)"},
			Score:   3, ScoreReasons: []string{`tag rust matched "rust" (weight 3)`}},
		"fp-rej": {Result: store.Rejected,
			Reasons: []string{"no-go match: unpaid", "must-have not met: senior"}},
	}
	if err := s.SaveVerdicts(ctx, vs); err != nil {
		t.Fatalf("SaveVerdicts: %v", err)
	}
	got, err := s.ListVerdicts(ctx)
	if err != nil {
		t.Fatalf("ListVerdicts: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("verdicts = %d, want 2", len(got))
	}
	acc := got["fp-acc"]
	if acc.Result != store.Accepted || acc.Score != 3 ||
		len(acc.ScoreReasons) != 1 || acc.ScoreReasons[0] != `tag rust matched "rust" (weight 3)` {
		t.Errorf("score roundtrip lost: %+v", acc)
	}
	rej := got["fp-rej"]
	if rej.Result != store.Rejected || len(rej.Reasons) != 2 ||
		rej.Reasons[1] != "must-have not met: senior" {
		t.Errorf("roundtrip lost reason lines: %+v", rej)
	}
	if rej.EvaluatedAt == "" {
		t.Errorf("evaluated_at missing")
	}
}

func verdictUpsertOverwrites(t *testing.T, s store.Store) {
	ctx := context.Background()
	if err := s.SaveVerdicts(ctx, map[string]store.Verdict{
		"fp": {Result: store.Rejected, Reasons: []string{"old"}, Score: 2}}); err != nil {
		t.Fatalf("first SaveVerdicts: %v", err)
	}
	if err := s.SaveVerdicts(ctx, map[string]store.Verdict{
		"fp": {Result: store.Accepted, Reasons: []string{"new"}, Score: 0}}); err != nil {
		t.Fatalf("second SaveVerdicts: %v", err)
	}
	got, _ := s.ListVerdicts(ctx)
	if got["fp"].Result != store.Accepted || got["fp"].Reasons[0] != "new" || got["fp"].Score != 0 {
		t.Errorf("upsert did not overwrite: %+v", got["fp"])
	}
}

func verdictJunkResultRejected(t *testing.T, s store.Store) {
	ctx := context.Background()
	err := s.SaveVerdicts(ctx, map[string]store.Verdict{
		"fp": {Result: "weird"}})
	if err == nil {
		t.Errorf("junk result accepted by store")
	}
}
