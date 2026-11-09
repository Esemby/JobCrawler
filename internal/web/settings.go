package web

import (
	"context"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/Esemby/JobCrawler/internal/config"
)

// systemTpl is parsed once at server construction (pattern from
// internal/web/status.go: parse-once, render-per-request).
var systemTpl = template.Must(template.ParseFS(embedded, "templates/system.html"))

// systemRow is one settings-table row for the System page.
type systemRow struct {
	Key      string
	Type     string
	Value    string
	Default  string
	Help     string
	Restart  bool
	Override bool // stored value differs from the default
}

// systemHandler shows all registered runtime settings with their effective
// values (docs/config.md). Never mutates state.
func (s *Server) systemHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := s.systemRows(r.Context())
		if err != nil {
			http.Error(w, "load settings: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := systemTpl.ExecuteTemplate(w, "system", struct {
			Rows  []systemRow
			Flash string
		}{Rows: rows, Flash: flashQuery(r)}); err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

// systemRows builds the display rows; errors mean the store is broken.
func (s *Server) systemRows(ctx context.Context) ([]systemRow, error) {
	overrides, err := s.deps.Store.ListSettings(ctx)
	if err != nil {
		return nil, err
	}
	effective := config.Effective(overrides)
	rows := make([]systemRow, 0, len(config.Registered()))
	for _, reg := range config.Registered() {
		v := effective[reg.Key]
		rows = append(rows, systemRow{
			Key: reg.Key, Type: reg.Type, Value: v, Default: reg.Default,
			Help: reg.Help, Restart: reg.Restart, Override: v != reg.Default,
		})
	}
	return rows, nil
}

// saveSettingHandler validates and persists ONE setting, then redirects with
// a flash message. Validation errors never touch the store.
func (s *Server) saveSettingHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form: "+err.Error(), http.StatusBadRequest)
			return
		}
		key := strings.TrimSpace(r.PostFormValue("key"))
		value := strings.TrimSpace(r.PostFormValue("value"))
		if err := config.ValidateSetting(key, value); err != nil {
			http.Error(w, "invalid: "+err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.deps.Store.SetSetting(r.Context(), key, value); err != nil {
			http.Error(w, "save failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		s.savedThenRedirect(w, r, "/system", "saved "+key)
	})
}

// savedThenRedirect runs the hot-reload hook, then redirects back to base
// with a flash message carried in the query string.
func (s *Server) savedThenRedirect(w http.ResponseWriter, r *http.Request, base, msg string) {
	if s.deps.SavedHook != nil {
		overrides, err := s.deps.Store.ListSettings(r.Context())
		if err == nil {
			if err := s.deps.SavedHook(config.Effective(overrides)); err != nil {
				http.Error(w, "hot reload failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	q := url.Values{"msg": []string{msg}}
	http.Redirect(w, r, base+"?"+q.Encode(), http.StatusSeeOther)
}

// flashQuery returns the flash message carried by this package's own
// redirects; html/template escapes it again at render time.
func flashQuery(r *http.Request) string { return r.URL.Query().Get("msg") }
