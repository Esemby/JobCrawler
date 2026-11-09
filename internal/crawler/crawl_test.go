package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Esemby/JobCrawler/internal/store"
	"github.com/Esemby/JobCrawler/internal/store/sqlite"
)

// newTestStore gives CrawlSource a real, migrated store on a temp file.
func newTestStore(t *testing.T) store.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "crawl.db")
	st, err := sqlite.Open(context.Background(), path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

// greenhouseFixture is the real saved ProPublica response.
func greenhouseFixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "greenhouse", "board-propublica.json"))
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	return b
}

func TestCrawlSourceEndToEnd(t *testing.T) {
	fx := greenhouseFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "robots.txt"):
			w.Write([]byte("User-agent: *\nDisallow: /embed/\n"))
		case strings.HasSuffix(r.URL.Path, "/jobs"):
			w.Write(fx)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	st := newTestStore(t)
	id, err := st.AddSource(context.Background(), "greenhouse-test", "greenhouse",
		srv.URL+"/v1/boards/propublica/jobs")
	if err != nil {
		t.Fatalf("AddSource: %v", err)
	}
	src := store.Source{ID: id, Name: "greenhouse-test", Kind: "greenhouse",
		Endpoint: srv.URL + "/v1/boards/propublica/jobs"}

	f := testFetcher()
	n, err := CrawlSource(context.Background(), f, st, src)
	if err != nil {
		t.Fatalf("CrawlSource: %v", err)
	}
	if n != 4 {
		t.Errorf("new jobs = %d, want 4", n)
	}

	// Second run dedupes everything.
	n2, err := CrawlSource(context.Background(), f, st, src)
	if err != nil {
		t.Fatalf("second CrawlSource: %v", err)
	}
	if n2 != 0 {
		t.Errorf("second run stored %d new, want 0 (dedupe)", n2)
	}

	jobs, err := st.ListJobs(context.Background(), "active")
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 4 {
		t.Fatalf("stored %d jobs, want 4", len(jobs))
	}
	for _, j := range jobs {
		if j.SourceName != "greenhouse-test" {
			t.Errorf("job %q source name = %q", j.Title, j.SourceName)
		}
		if j.Company != "propublica" {
			t.Errorf("job %q company = %q, want board token as stand-in", j.Title, j.Company)
		}
	}
}

func TestCrawlSourceUnknownKind(t *testing.T) {
	st := newTestStore(t)
	_, err := CrawlSource(context.Background(), testFetcher(), st,
		store.Source{Name: "x", Kind: "mystery", Endpoint: "whatever"})
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("err = %v, want unknown-kind error", err)
	}
}

func TestCrawlSourceDisabledEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "robots.txt") {
			w.WriteHeader(404)
			return
		}
		w.Write([]byte(`{"jobs":[],"meta":{"total":0}}`))
	}))
	t.Cleanup(srv.Close)

	st := newTestStore(t)
	src := store.Source{Name: "greenhouse-empty", Kind: "greenhouse",
		Endpoint: srv.URL + "/v1/boards/empty/jobs"}
	_, err := CrawlSource(context.Background(), testFetcher(), st, src)
	if err == nil || !strings.Contains(err.Error(), "zero job records") {
		t.Fatalf("err = %v, want zero-records error, not silent success", err)
	}
}
