package web

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/Esemby/JobCrawler/internal/config"
	"github.com/Esemby/JobCrawler/internal/store/sqlite"
)

// dbTpl is parsed once; renders the Database access page.
var dbTpl = template.Must(template.ParseFS(embedded, "templates/database.html"))

// dbPage is the Database access page model. DSN is shown masked; changing it
// requires retyping it, never round-tripping the clear value (brief 4.3).
type dbPage struct {
	// BootstrapPath names the file these values live in (ADR-0002).
	BootstrapPath string
	Engine        string
	// DSMMasked shows the current DSN with a password segment replaced by
	// *** (empty when there is no password). The raw DSN is never rendered.
	DSNMasked string
	Flash     string
}

// databaseHandler shows engine + masked DSN + test-connection button.
func (s *Server) databaseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := s.deps.Bootstrap
		page := dbPage{
			BootstrapPath: b.Path,
			Engine:        b.Engine,
			DSNMasked:     maskDSN(b.DSN),
			Flash:         flashQuery(r),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := dbTpl.ExecuteTemplate(w, "database", page); err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

// saveDatabaseHandler validates and writes a NEW bootstrap file (the DB
// settings cannot live inside the DB they describe, ADR-0002). The current
// process keeps serving the old store; the UI notes "restart required".
func (s *Server) saveDatabaseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form: "+err.Error(), http.StatusBadRequest)
			return
		}
		engine := strings.TrimSpace(r.PostFormValue("engine"))
		dsn := strings.TrimSpace(r.PostFormValue("dsn"))
		b := config.Bootstrap{Path: s.deps.Bootstrap.Path, Engine: engine, DSN: dsn}
		if err := b.Validate(); err != nil { // validates engine + non-empty DSN
			http.Error(w, "invalid: "+err.Error(), http.StatusBadRequest)
			return
		}
		if err := config.WriteBootstrap(b.Path, engine, dsn); err != nil {
			http.Error(w, "write failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		s.savedThenRedirect(w, r, "/database", "saved bootstrap file (restart required)")
	})
}

// testDatabaseHandler opens the requested engine at request time WITHOUT
// saving. sqlite opens+ping+closes; mariadb reports Phase 6 status.
func (s *Server) testDatabaseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form: "+err.Error(), http.StatusBadRequest)
			return
		}
		engine := strings.TrimSpace(r.PostFormValue("engine"))
		dsn := strings.TrimSpace(r.PostFormValue("dsn"))
		switch engine {
		case "mariadb":
			s.savedThenRedirectMessage(w, r, "/database",
				"mariadb engine is not implemented until Phase 6 (ADR-0002); nothing was tested")
		case "sqlite":
			st, err := sqlite.Open(r.Context(), dsn)
			if err != nil {
				s.savedThenRedirectMessage(w, r, "/database", "sqlite test FAILED: "+err.Error())
				return
			}
			if err := st.Close(); err != nil {
				s.savedThenRedirectMessage(w, r, "/database", "sqlite test FAILED to close: "+err.Error())
				return
			}
			s.savedThenRedirectMessage(w, r, "/database", "sqlite test OK (opened, pinged, closed)")
		default:
			http.Error(w, "invalid: engine must be sqlite or mariadb", http.StatusBadRequest)
		}
	})
}

// savedThenRedirectMessage redirects with a flash and skips the settings hook
// (database saves are not runtime settings).
func (s *Server) savedThenRedirectMessage(w http.ResponseWriter, r *http.Request, base, msg string) {
	q := url.Values{"msg": []string{msg}}
	http.Redirect(w, r, base+"?"+q.Encode(), http.StatusSeeOther)
}

// maskDSN hides the password segment of a URL-style DSN. The rule mirrors
// standard URI userinfo: user:password@… If nothing matches, the DSN has no
// password (sqlite files) and is shown as-is.
func maskDSN(dsn string) string {
	if at := strings.LastIndex(dsn, "@"); at >= 0 {
		if colon := strings.IndexByte(dsn[:at], ':'); colon >= 0 {
			return dsn[:colon+1] + "***" + dsn[at:]
		}
	}
	return dsn
}
