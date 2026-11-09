package crawler

import "testing"

// TestRobotRules covers the robots.txt subset against fixture snippets, so
// politeness behaviour is pinned, not vibes.
func TestRobotRules(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		agent string
		path  string
		want  bool
	}{
		{"empty file allows all", "", "jobcrawler", "/x", true},
		{"garbage allows all", "hello\nworld", "jobcrawler", "/x", true},
		{
			"wildcard disallow blocks",
			"User-agent: *\nDisallow: /embed/\n",
			"jobcrawler", "/embed/jobs", false,
		},
		{
			"wildcard disallow outside prefix allows",
			"User-agent: *\nDisallow: /embed/\n",
			"jobcrawler", "/v1/boards/acme/jobs", true,
		},
		{
			"exact agent group wins over wildcard",
			"User-agent: *\nDisallow: /\n\nUser-agent: jobcrawler\nDisallow: /admin\n",
			"jobcrawler", "/v1/other", true,
		},
		{
			"unknown agent falls back to wildcard",
			"User-agent: *\nDisallow: /\n",
			"somebot", "/anything", false,
		},
		{
			"never-matching agent allows all",
			"User-agent: otherbot\nDisallow: /\n",
			"jobcrawler", "/anything", true,
		},
		{
			"allow beats shorter disallow",
			"User-agent: *\nDisallow: /\nAllow: /v1/\n",
			"jobcrawler", "/v1/boards/acme", true,
		},
		{
			"disallow outside allow prefix still blocks",
			"User-agent: *\nDisallow: /\nAllow: /v1/\n",
			"jobcrawler", "/embed/x", false,
		},
		{
			"tie favours allow",
			"User-agent: *\nDisallow: /job\nAllow: /job\n",
			"jobcrawler", "/jobs", true,
		},
		{
			"empty disallow means everything",
			"User-agent: *\nDisallow:\n",
			"jobcrawler", "/x", false,
		},
		{
			"comments and blank lines ignored",
			"# polite file\n\nUser-agent: *   # all bots\nDisallow: /embed/  # embed\n",
			"jobcrawler", "/embed/1", false,
		},
		{
			"back-to-back agents share a group",
			"User-agent: aBot\nUser-agent: *\nDisallow: /x\n",
			"jobcrawler", "/x", false,
		},
		{
			"agent after rule starts new group",
			"User-agent: *\nDisallow: /x\n\nUser-agent: jobcrawler\nDisallow: /y\n",
			"jobcrawler", "/x", true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := parseRobotRules([]byte(tt.body), tt.agent)
			if got := r.allowed(tt.path); got != tt.want {
				t.Errorf("allowed(%q) = %v, want %v (agents=%v disallow=%v allow=%v)",
					tt.path, got, tt.want, r.agents, r.disallow, r.allow)
			}
		})
	}
}

// TestRealGreenhouseRobots pins the actual decision for the Phase 3 source:
// the live robots.txt (saved in testdata, fetched 2026-11-09) disallows only
// /embed/, so /v1/boards/... is permitted. If Greenhouse ever changes this,
// the crawler must refuse.
func TestRealGreenhouseRobots(t *testing.T) {
	r := parseRobotRules([]byte(
		"User-agent: *\nDisallow: /embed/\n"), "jobcrawler")
	if r.allowed("/v1/boards/propublica/jobs") != true {
		t.Errorf("our listing endpoint must be allowed")
	}
	if r.allowed("/embed/job_list?for=propublica") != false {
		t.Errorf("/embed/ must stay disallowed")
	}
}
