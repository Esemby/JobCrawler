package web

import (
	"html/template"
	"net/http"
)

// newStatusHandler parses the status template once and returns a handler that
// always renders the same StatusPage it was constructed with. It errors when
// the template cannot be parsed (a build-blocking bug, caught by tests too).
func newStatusHandler(status StatusPage) (http.Handler, error) {
	t, err := template.ParseFS(embedded, "templates/status.html")
	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := t.ExecuteTemplate(w, "status.html", status); err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	}), nil
}
