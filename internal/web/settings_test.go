package web

import (
	"context"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Esemby/JobCrawler/internal/config"
	"github.com/Esemby/JobCrawler/internal/store/sqlite"
)

// newTestServer builds a real SQLite store on a temp file plus a server wired
// like main does. The hook records the last effective settings.
func newTestServer(t *testing.T) (*Server, *[]string) {
	t.Helper()
	st, err := sqlite.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	saved := &[]string{}
	srv, err := New(StatusPage{BindAddr: ":8090"}, Deps{
		Store:     st,
		Bootstrap: config.Bootstrap{Path: filepath.Join(t.TempDir(), "config.bootstrap.toml"), Engine: "sqlite", DSN: "jc.db"},
		SavedHook: func(effective map[string]string) error {
			*saved = append(*saved, effective["system.log_level"])
			return nil
		},
	})
	if err != nil {
		t.Fatalf("web.New: %v", err)
	}
	return srv, saved
}

// postForm sends an application/x-www-form-urlencoded POST with the right
// header (ParseForm requires it); returns the recorder for assertions.
func postForm(srv *Server, path string, form url.Values) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestSystemPageShowsSettings(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/system", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	for _, want := range []string{"system.bind_addr", "system.log_level", "crawl.concurrency", "Save"} {
		if !strings.Contains(string(body), want) {
			t.Errorf("system page missing %q", want)
		}
	}
}

func TestSaveSettingStoresAndHooks(t *testing.T) {
	srv, saved := newTestServer(t)
	form := url.Values{"key": {"system.log_level"}, "value": {"debug"}}
	rec := postForm(srv, "/system/save", form)
	if rec.Code != 303 {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if len(*saved) != 1 || (*saved)[0] != "debug" {
		t.Errorf("SavedHook saw %v, want [debug]", *saved)
	}
}

func TestSaveSettingRejectsInvalid(t *testing.T) {
	srv, _ := newTestServer(t)
	form := url.Values{"key": {"system.bind_addr"}, "value": {":8080"}}
	rec := postForm(srv, "/system/save", form)
	if rec.Code != 400 {
		t.Errorf("status = %d, want 400 (reserved port must not pass)", rec.Code)
	}
}

func TestDatabasePageMasksPassword(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.deps.Bootstrap.DSN = "user:secret@127.0.0.1:3306/jc"
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/database", nil))
	body, _ := io.ReadAll(rec.Body)
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if strings.Contains(string(body), "secret") {
		t.Errorf("clear DSN password leaked to the page: %s", body)
	}
	if !strings.Contains(string(body), "user:***@127.0.0.1:3306/jc") {
		t.Errorf("masked DSN not rendered: %s", body)
	}
}

func TestDatabaseSaveWritesBootstrapAndRereads(t *testing.T) {
	srv, _ := newTestServer(t)
	form := url.Values{"engine": {"sqlite"}, "dsn": {"fresh.db"}}
	rec := postForm(srv, "/database/save", form)
	if rec.Code != 303 {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	data, err := os.ReadFile(srv.deps.Bootstrap.Path)
	if err != nil {
		t.Fatalf("bootstrap file not written: %v", err)
	}
	boot, err := config.LoadBootstrap(srv.deps.Bootstrap.Path)
	if err != nil {
		t.Fatalf("written bootstrap cannot be re-read: %v", err)
	}
	if !strings.Contains(string(data), `engine = "sqlite"`) || boot.DSN != "fresh.db" {
		t.Errorf("roundtrip mismatch: body=%s load=%+v", data, boot)
	}
}

func TestDatabaseTestConnectionSqliteSuffix(t *testing.T) {
	srv, _ := newTestServer(t)
	path := filepath.Join(t.TempDir(), "probe.db")
	form := url.Values{"engine": {"sqlite"}, "dsn": {path}}
	rec := postForm(srv, "/database/test", form)
	if rec.Code != 303 {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Location"), "sqlite+test+OK") &&
		!strings.Contains(url.QueryEscape(rec.Header().Get("Location")), "OK") {
		t.Errorf("flash does not report OK: %q", rec.Header().Get("Location"))
	}
}
