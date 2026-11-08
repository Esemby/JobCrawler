package web

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	srv, err := New(StatusPage{})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.HasPrefix(string(body), "ok ") {
		t.Errorf("body = %q, want prefix \"ok \"", string(body))
	}
}

func TestStatusPage(t *testing.T) {
	srv, err := New(StatusPage{Version: "v1.2.3", Commit: "abc", Date: "2026-11-08", BindAddr: ":8090", DBEngine: "sqlite"})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	want := []string{"v1.2.3", ":8090", "sqlite", "LAN-only"}
	for _, w := range want {
		if !strings.Contains(string(body), w) {
			t.Errorf("status page missing %q; html snippet: %s", w, body)
		}
	}
}

func TestStaticHtmxServed(t *testing.T) {
	srv, err := New(StatusPage{})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/static/htmx.min.js", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if len(body) < 10000 {
		t.Errorf("htmx.min.js size = %d, want >= 10000 (vendored file embedded?)", len(body))
	}
}
