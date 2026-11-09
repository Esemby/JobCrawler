package crawler

import (
	"context"
	"fmt"
	"strings"

	"github.com/Esemby/JobCrawler/internal/sources/greenhouse"
	"github.com/Esemby/JobCrawler/internal/store"
)

// CrawlSource runs one configured source end-to-end: build the listing URL,
// fetch it politely, parse, and upsert by fingerprint. It returns the number
// of NEW jobs stored. Sequential by design — politeness over parallelism.
func CrawlSource(ctx context.Context, f *Fetcher, st store.Store, src store.Source) (int, error) {
	boardURL, err := listingURL(src)
	if err != nil {
		return 0, fmt.Errorf("crawl %s: %w", src.Name, err)
	}
	body, err := f.Fetch(ctx, boardURL)
	if err != nil {
		return 0, fmt.Errorf("crawl %s: %w", src.Name, err)
	}
	layout, err := greenhouse.Parse(body, boardToken(src))
	if err != nil {
		return 0, fmt.Errorf("crawl %s: %w", src.Name, err)
	}
	jobs := greenhouse.ToJobs(layout, src.ID, company(src))
	return st.UpsertJobs(ctx, jobs)
}

// listingURL builds the fetch address for a source. A full http(s) URL stored
// in the row overrides the canonical builder — that is a test seam for
// httptest hosts, never produced by the UI validation.
func listingURL(src store.Source) (string, error) {
	ep := strings.TrimSpace(src.Endpoint)
	if strings.HasPrefix(ep, "http://") || strings.HasPrefix(ep, "https://") {
		return ep, nil
	}
	switch src.Kind {
	case "greenhouse":
		if ep == "" {
			return "", fmt.Errorf("empty board token")
		}
		if u := greenhouse.Endpoint(ep); u == "" {
			return "", fmt.Errorf("invalid board token %q", ep)
		}
		return greenhouse.Endpoint(ep), nil
	default:
		return "", fmt.Errorf("source kind %q is not implemented", src.Kind)
	}
}

// boardToken is what the greenhouse parser wants from the row: the bare
// token, or — when the row holds an overriding full URL (test seam) — the
// path segment right after "/boards/". Bare tokens pass through untouched;
// query strings are dropped.
func boardToken(src store.Source) string {
	ep := strings.TrimSpace(src.Endpoint)
	if q := strings.IndexByte(ep, '?'); q >= 0 {
		ep = ep[:q]
	}
	ep = strings.TrimSuffix(ep, "/")
	if !strings.Contains(ep, "://") {
		return ep
	}
	const marker = "/boards/"
	if i := strings.Index(ep, marker); i >= 0 {
		rest := ep[i+len(marker):]
		if slash := strings.IndexByte(rest, '/'); slash >= 0 {
			rest = rest[:slash]
		}
		return rest
	}
	// Fallback: last non-empty path segment.
	if i := strings.LastIndexByte(ep, '/'); i >= 0 {
		return ep[i+1:]
	}
	return ""
}

// company: a board token does not name the legal company (docs/sources.md);
// for now the token itself stands in as the display name. Pretier naming is
// a filter-phase concern, not a crawl-time guess.
func company(src store.Source) string {
	return boardToken(src)
}
