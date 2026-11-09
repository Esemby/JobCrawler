package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// testFetcher: zero-delay fetcher for fast httptest rounds.
func testFetcher() *Fetcher {
	f := NewFetcher("jobcrawler/test (+contact: placeholder@example.invalid)")
	f.MinDelay = 0
	f.MaxJitter = 0
	return f
}

func TestFetchHonestUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		if strings.HasSuffix(r.URL.Path, "robots.txt") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte("body"))
	}))
	t.Cleanup(srv.Close)

	f := testFetcher()
	body, err := f.Fetch(context.Background(), srv.URL+"/jobs")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if string(body) != "body" {
		t.Errorf("body = %q", body)
	}
	if !strings.HasPrefix(gotUA, "jobcrawler/") || !strings.Contains(gotUA, "contact:") {
		t.Errorf("User-Agent %q must be honest and carry a contact field", gotUA)
	}
}

func TestFetchForbiddenPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "robots.txt") {
			w.Write([]byte("User-agent: *\nDisallow: /private/\n"))
			return
		}
		w.Write([]byte("should never be reached"))
	}))
	t.Cleanup(srv.Close)

	_, err := testFetcher().Fetch(context.Background(), srv.URL+"/private/list")
	if err == nil || !strings.Contains(err.Error(), "disallows") {
		t.Fatalf("err = %v, want robots disallow error", err)
	}
}

func TestFetchRobotsUnreachableRefuses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500) // robots fetch AND target fail
	}))
	t.Cleanup(srv.Close)

	_, err := testFetcher().Fetch(context.Background(), srv.URL+"/jobs")
	if err == nil || !strings.Contains(err.Error(), "robots.txt") {
		t.Fatalf("err = %v, want conservative robots refusal", err)
	}
}

func TestFetchRobotsMissingMeansAllowed(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "robots.txt") {
			w.WriteHeader(404)
			return
		}
		hits++
		w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	f := testFetcher()
	for i := 0; i < 2; i++ {
		if _, err := f.Fetch(context.Background(), srv.URL+"/jobs"); err != nil {
			t.Fatalf("round %d: %v", i, err)
		}
	}
	// robots probe happens once, then is cached within the TTL.
	if hits != 2 {
		t.Errorf("target hits = %d, want 2", hits)
	}
}

func TestFetchHostPacing(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "robots.txt") {
			w.WriteHeader(404)
			return
		}
		w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	f := testFetcher()
	f.MinDelay = 50 * time.Millisecond
	f.MaxJitter = 10 * time.Millisecond
	f.Client = srv.Client() // self-signed cert

	start := time.Now()
	for i := 0; i < 3; i++ {
		if _, err := f.Fetch(context.Background(), srv.URL+"/x"); err != nil {
			t.Fatalf("round %d: %v", i, err)
		}
	}
	elapsed := time.Since(start)
	// Two pacing gaps of >= 50ms each, jitter only adds.芷
	if elapsed < 2*f.MinDelay {
		t.Errorf("3 fetches took %v, want >= %v (pacing missing)", elapsed, 2*f.MinDelay)
	}
}

func TestFetchNonHTTPURL(t *testing.T) {
	for _, u := range []string{"ftp://x/y", "/relative/only", ""} {
		if _, err := testFetcher().Fetch(context.Background(), u); err == nil {
			t.Errorf("Fetch(%q) must error", u)
		}
	}
}
