package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Esemby/JobCrawler/internal/store"
	"github.com/Esemby/JobCrawler/internal/store/sqlite"
)

// jobsFixture bundles the server plus its store so tests can seed rows via
// the documented store seam (full-URL endpoints are a crawler test feature,
// never produced by UI validation).
type jobsFixture struct {
	http  *httptest.Server
	store store.Store
}

func setupJobs(t *testing.T) jobsFixture {
	t.Helper()
	path := filepath.Join(t.TempDir(), "webtest.db")
	st, err := sqlite.Open(context.Background(), path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	srv, err := New(StatusPage{}, Deps{Store: st})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	return jobsFixture{http: httptest.NewServer(srv.Handler()), store: st}
}

// postJobs sends a form; returns status + body (redirects followed).
func postJobs(t *testing.T, client *http.Client, urlStr, form string) (int, string) {
	t.Helper()
	resp, err := client.Post(urlStr, "application/x-www-form-urlencoded", strings.NewReader(form))
	if err != nil {
		t.Fatalf("POST %s: %v", urlStr, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func getJobs(t *testing.T, client *http.Client, urlStr string) (int, string) {
	t.Helper()
	resp, err := client.Get(urlStr)
	if err != nil {
		t.Fatalf("GET %s: %v", urlStr, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestJobsPageEmpty(t *testing.T) {
	fx := setupJobs(t)
	code, body := getJobs(t, fx.http.Client(), fx.http.URL+"/jobs")
	if code != 200 {
		t.Fatalf("status = %d, want 200", code)
	}
	for _, want := range []string{"No jobs stored yet", "No sources configured yet", "Jobs"} {
		if !strings.Contains(body, want) {
			t.Errorf("page missing %q", want)
		}
	}
}

func TestSourceAddValidateToggle(t *testing.T) {
	fx := setupJobs(t)
	c := fx.http.Client()

	// Bad names/tokens rejected before touching the store.
	if code, _ := postJobs(t, c, fx.http.URL+"/sources/add", "name=A B&token=x"); code != 400 {
		t.Errorf("bad name status = %d, want 400", code)
	}
	if code, _ := postJobs(t, c, fx.http.URL+"/sources/add", "name=greenhouse-acme&token="); code != 400 {
		t.Errorf("bad token status = %d, want 400", code)
	}

	if code, body := postJobs(t, c, fx.http.URL+"/sources/add", "name=greenhouse-acme&token=acme"); code != 200 || !strings.Contains(body, "added source greenhouse-acme") {
		t.Errorf("valid add code=%d body=%.200s, want 200 + flash", code, body)
	}
	if code, _ := postJobs(t, c, fx.http.URL+"/sources/add", "name=greenhouse-acme&token=acme"); code != 400 {
		t.Errorf("duplicate status = %d, want 400", code)
	}

	code, body := getJobs(t, c, fx.http.URL+"/jobs")
	if code != 200 || !strings.Contains(body, "greenhouse-acme") || !strings.Contains(body, "enabled") {
		t.Errorf("page should list the enabled added source (code=%d)", code)
	}

	if code, body := postJobs(t, c, fx.http.URL+"/sources/toggle", "name=greenhouse-acme&enabled=off"); code != 200 || !strings.Contains(body, "greenhouse-acme disabled") {
		t.Errorf("disable code=%d body=%.200s, want 200 + flash", code, body)
	}
	_, body = getJobs(t, c, fx.http.URL+"/jobs")
	if !strings.Contains(body, "disabled") {
		t.Errorf("page should show disabled badge")
	}

	if code, _ := postJobs(t, c, fx.http.URL+"/sources/toggle", "name=ghost&enabled=on"); code != 400 {
		t.Errorf("unknown toggle status = %d, want 400", code)
	}
}

// greenhouseFixture reads the real saved board response.
func greenhouseFixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "greenhouse", "board-propublica.json"))
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	return b
}

// TestCrawlRunEndToEnd pushes jobs through the FULL UI loop: seed one
// full-URL source row (crawler test seam), run /crawl/run, then read the
// rendered list — politeness, robots, parse, dedupe and template in one go.
func TestCrawlRunEndToEnd(t *testing.T) {
	board := greenhouseFixture(t)
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "robots.txt"):
			w.Write([]byte("User-agent: *\nDisallow: /embed/\n"))
		case strings.HasSuffix(r.URL.Path, "/jobs"):
			w.Write(board)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(origin.Close)

	f := setupJobs(t)
	c := f.http.Client()
	ctx := context.Background()
	id, err := f.store.AddSource(ctx, "greenhouse-test", "greenhouse", origin.URL+"/v1/boards/propublica/jobs")
	if err != nil {
		t.Fatalf("seed source: %v", err)
	}
	_ = id

	code, body := postJobs(t, c, f.http.URL+"/crawl/run", "")
	if code != 200 {
		t.Fatalf("run status = %d, body: %s", code, body)
	}
	if !strings.Contains(body, "greenhouse-test: 4 new") {
		t.Errorf("flash missing \"greenhouse-test: 4 new\"; body: %.300s", body)
	}

	code, body = getJobs(t, c, f.http.URL+"/jobs")
	if code != 200 || !strings.Contains(body, "Freelance Pitch Form") {
		t.Fatalf("jobs page should list the crawled job (code=%d)", code)
	}
	if strings.Count(body, "job-boards.greenhouse.io") == 0 {
		t.Errorf("jobs rows missing")
	}

	// Second run dedupes: flash must say 0 new.
	code, body = postJobs(t, c, f.http.URL+"/crawl/run", "")
	if code != 200 || !strings.Contains(body, "greenhouse-test: 0 new") {
		t.Errorf("second run flash = %.200s (want 0 new), code=%d", body, code)
	}
}

func TestJobsPageNeedsStore(t *testing.T) {
	// Zero deps: the jobs handlers must fail loudly, never render lies.
	srv, err := New(StatusPage{}, Deps{})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	hs := httptest.NewServer(srv.Handler())
	t.Cleanup(hs.Close)
	for _, path := range []string{"/jobs"} {
		resp, err := hs.Client().Get(hs.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != 500 {
			t.Errorf("%s without store: status %d, want 500", path, resp.StatusCode)
		}
	}
}
