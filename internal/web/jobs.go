package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Esemby/JobCrawler/internal/config"
	"github.com/Esemby/JobCrawler/internal/crawler"
	"github.com/Esemby/JobCrawler/internal/filter"
	"github.com/Esemby/JobCrawler/internal/sources/greenhouse"
	"github.com/Esemby/JobCrawler/internal/store"
)

// jobsTpl is parsed once; renders the Jobs page (list + sources + manual run).
var jobsTpl = template.Must(template.ParseFS(embedded, "templates/jobs.html"))

// jobRow is one listing row for the Jobs page.
type jobRow struct {
	Title, Company, Location, Source, State string
	Link                                    string
	FirstSeen, LastSeen                     string
	// Verdict/Why/Score come from the stage 1+2 tables (empty = not evaluated).
	Verdict string
	Why     string
	Score   int
}

// srcRow is one configured source row for the form table.
type srcRow struct {
	Name, Kind, Endpoint string
	Enabled              bool
}

type jobsPage struct {
	Jobs    []jobRow
	Sources []srcRow
	Flash   string
}

// jobsHandler lists every stored job (state badge included) plus the source
// table. No store -> 500: the page is useless without one.
func (s *Server) jobsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.deps.Store == nil {
			http.Error(w, "no store configured", http.StatusInternalServerError)
			return
		}
		jobs, srcs, err := s.jobsPageData(r.Context())
		if err != nil {
			http.Error(w, "load jobs: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page := jobsPage{Jobs: jobs, Sources: srcs, Flash: flashQuery(r)}
		if err := jobsTpl.ExecuteTemplate(w, "jobs", page); err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

// jobsPageData loads both lists plus the verdict map (stage-1 results).
func (s *Server) jobsPageData(ctx context.Context) ([]jobRow, []srcRow, error) {
	jobs, err := s.deps.Store.ListJobs(ctx, "*")
	if err != nil {
		return nil, nil, err
	}
	sources, err := s.deps.Store.ListSources(ctx)
	if err != nil {
		return nil, nil, err
	}
	verdicts, err := s.deps.Store.ListVerdicts(ctx)
	if err != nil {
		return nil, nil, err
	}
	jobRows := make([]jobRow, 0, len(jobs))
	for _, j := range jobs {
		state := j.State
		if state == "" {
			state = "active"
		}
		row := jobRow{
			Title: j.Title, Company: j.Company, Location: j.Location,
			Source: j.SourceName, State: state, Link: j.URL,
			FirstSeen: j.FirstSeen, LastSeen: j.LastSeen,
		}
		if v, ok := verdicts[j.Fingerprint]; ok {
			row.Verdict = v.Result
			row.Score = v.Score
			row.Why = strings.Join(v.Reasons, "; ")
			if len(v.ScoreReasons) > 0 {
				row.Why += " ‖ score: " + strings.Join(v.ScoreReasons, "; ")
			}
		}
		jobRows = append(jobRows, row)
	}
	srcRows := make([]srcRow, 0, len(sources))
	for _, so := range sources {
		srcRows = append(srcRows, srcRow{
			Name: so.Name, Kind: so.Kind, Endpoint: so.Endpoint, Enabled: so.Enabled,
		})
	}
	return jobRows, srcRows, nil
}

// sourceNameRx keeps source names boring: lowercase + dashes, short. They
// appear in URLs and logs; nothing clever allowed.
var sourceNameRx = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,39}$`)

// addSourceHandler registers one source row. Only kind greenhouse exists in
// Phase 3; the token must pass the parser's own endpoint validation.
func (s *Server) addSourceHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.deps.Store == nil {
			http.Error(w, "no store configured", http.StatusInternalServerError)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form: "+err.Error(), http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(r.PostFormValue("name"))
		token := strings.TrimSpace(r.PostFormValue("token"))
		if !sourceNameRx.MatchString(name) {
			http.Error(w, "invalid: name must be 2-40 chars of a-z, 0-9, dashes", http.StatusBadRequest)
			return
		}
		if greenhouse.Endpoint(token) == "" {
			http.Error(w, "invalid: board token must be 1+ chars of letters, digits, dash, dot, underscore",
				http.StatusBadRequest)
			return
		}
		if _, err := s.deps.Store.AddSource(r.Context(), name, "greenhouse", token); err != nil {
			http.Error(w, "add failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		s.redirect(w, r, "/jobs", "added source "+name)
	})
}

// toggleSourceHandler flips one source by name.
func (s *Server) toggleSourceHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.deps.Store == nil {
			http.Error(w, "no store configured", http.StatusInternalServerError)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form: "+err.Error(), http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(r.PostFormValue("name"))
		enabled := r.PostFormValue("enabled") == "on"
		if err := s.deps.Store.SetSourceEnabled(r.Context(), name, enabled); err != nil {
			http.Error(w, "toggle failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		state := "disabled"
		if enabled {
			state = "enabled"
		}
		s.redirect(w, r, "/jobs", name+" "+state)
	})
}

// crawlRunHandler crawls every ENABLED source, sequentially, each with its
// own politeness budget, then re-runs the stage-1 filter over all active
// jobs, saves verdicts, and reports one summary line per source as flash.
// A fresh Fetcher per run keeps politeness state scoped and predictable.
func (s *Server) crawlRunHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.deps.Store == nil {
			http.Error(w, "no store configured", http.StatusInternalServerError)
			return
		}
		ctx := r.Context()
		sources, err := s.deps.Store.ListSources(ctx)
		if err != nil {
			http.Error(w, "list sources: "+err.Error(), http.StatusInternalServerError)
			return
		}
		f := crawler.NewFetcher(crawler.DefaultUA())
		var lines []string
		for _, src := range sources {
			if !src.Enabled {
				continue
			}
			n, err := crawler.CrawlSource(ctx, f, s.deps.Store, src)
			switch {
			case err != nil:
				lines = append(lines, fmt.Sprintf("%s: FAILED %v", src.Name, err))
				continue
			default:
				lines = append(lines, fmt.Sprintf("%s: %d new", src.Name, n))
			}
		}
		rej, acc, err := s.evaluateVerdicts(ctx)
		if err != nil {
			http.Error(w, "evaluate verdicts: "+err.Error(), http.StatusInternalServerError)
			return
		}
		s.redirect(w, r, "/jobs",
			"crawl run — "+strings.Join(lines, " | ")+fmt.Sprintf(
				" | stage 1: %d accepted, %d rejected", acc, rej))
	})
}

// evaluateVerdicts re-runs stage 1 over every active job and upserts the
// results (migration 0003). Re-evaluating everything each run is idempotent
// by fingerprint and keeps verdicts truthful when filter settings change.
// Returns (rejected, accepted) counts.
func (s *Server) evaluateVerdicts(ctx context.Context) (int, int, error) {
	overrides, err := s.deps.Store.ListSettings(ctx)
	if err != nil {
		return 0, 0, err
	}
	effective := config.Effective(overrides)
	cfg := filter.ConfigFromSettings(effective)
	tags := filter.TagRules(effective["filter.tag_weights"])
	jobs, err := s.deps.Store.ListJobs(ctx, "active")
	if err != nil {
		return 0, 0, err
	}
	verdicts := filter.EvaluateAll(jobs, cfg, tags)
	rej, acc := 0, 0
	for _, v := range verdicts {
		if v.Result == store.Rejected {
			rej++
		} else {
			acc++
		}
	}
	if err := s.deps.Store.SaveVerdicts(ctx, verdicts); err != nil {
		return 0, 0, err
	}
	return rej, acc, nil
}

// redirect sends the standard flash redirect (settings.go pattern, without
// the settings hook: jobs actions are not runtime settings).
func (s *Server) redirect(w http.ResponseWriter, r *http.Request, base, msg string) {
	q := url.Values{"msg": []string{msg}}
	http.Redirect(w, r, base+"?"+q.Encode(), http.StatusSeeOther)
}
