// Package web serves the LAN-only configuration and status UI.
//
// Everything reachable lives under the mux built by New; handlers stay thin and
// call into other internal packages. The UI binds LAN-side by default and must
// never be exposed to the internet (see AGENTS.md rule 6). Templates and static
// assets (incl. vendored htmx) are embedded so the single binary is complete.
package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/Esemby/JobCrawler/internal/version"
)

//go:embed templates static
var embedded embed.FS

// Server holds the routed HTTP handler. Create with New, serve via Handler().
type Server struct {
	mux *http.ServeMux
}

// StatusPage is the data shown on the landing/status page. Fields are filled
// by main from config and version; more fields arrive with later phases.
type StatusPage struct {
	Version   string
	Commit    string
	Date      string
	GoVersion string
	BindAddr  string
	DBEngine  string
}

// New builds the routed handler. Port binding happens in cmd/main, never here,
// so tests can exercise the routes with httptest without listening.
func New(status StatusPage) (*Server, error) {
	s := &Server{mux: http.NewServeMux()}

	st, err := newStatusHandler(status)
	if err != nil {
		return nil, fmt.Errorf("parse status templates: %w", err)
	}
	s.mux.Handle("GET /{$}", st)

	health := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "ok %s\n", version.Version)
	})
	s.mux.Handle("GET /healthz", health)

	staticFS, err := fs.Sub(embedded, "static")
	if err != nil {
		return nil, fmt.Errorf("embed static: %w", err)
	}
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	return s, nil
}

// Handler returns the full routed handler; it is never nil after New succeeds.
func (s *Server) Handler() http.Handler { return s.mux }
