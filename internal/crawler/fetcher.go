package crawler

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Esemby/JobCrawler/internal/version"
)

// DefaultUA is the honest User-Agent sent on every request. The contact part
// is a DELIBERATE placeholder (owner decision 2026-11-09: fill in a real
// address before the first crawl against an external host; never leave the
// placeholder in place for production crawls — see docs/config.md).
func DefaultUA() string {
	return "jobcrawler/" + version.Version + " (+https://github.com/Esemby/JobCrawler; contact: PLACEHOLDER)"
}

const (
	// minHostDelay is the politeness floor between two requests to the same
	// host (docs/sources.md). Fixed constant for now; becomes the
	// crawl.rate_limit_ms setting when the scheduler phase lands.
	minHostDelay = 2 * time.Second
	// maxJitter is added randomly on top of the floor to avoid thundering
	// alignment with other crawlers.
	maxJitter = time.Second
	// robotsTTL: robots.txt is re-read at most daily (docs/sources.md).
	robotsTTL = 24 * time.Hour
)

// Fetcher makes polite GET requests. All state lives on the struct; never
// share one across goroutines without thinking — the crawl handler runs
// sources sequentially, one mutex still guards slow sloppiness.
type Fetcher struct {
	Client *http.Client
	// UA is sent as User-Agent on every request, including robots.txt.
	UA string
	// MinDelay / MaxJitter are injectable for fast tests; zero means 0 wait.
	MinDelay  time.Duration
	MaxJitter time.Duration

	mu      sync.Mutex           // guards both maps below
	lastHit map[string]time.Time // host -> last completed request
	robots  map[string]robotEntry
}

type robotEntry struct {
	rules robotRules
	fetch time.Time
}

// NewFetcher builds a fetcher with production delays and robots caching.
func NewFetcher(ua string) *Fetcher {
	return &Fetcher{
		Client:    http.DefaultClient,
		UA:        ua,
		MinDelay:  minHostDelay,
		MaxJitter: maxJitter,
		lastHit:   map[string]time.Time{},
		robots:    map[string]robotEntry{},
	}
}

// Fetch performs one robots-checked GET and returns the body. The contract:
//   - robots.txt consulted first (cached <= robotsTTL); a robots response of
//     404/410 means "no robots file, allowed"; a network error or 5xx while
//     fetching robots refuses the crawl this round (conservative);
//   - the path is matched against the rules; a denied path errors;
//   - host-level pacing: wait >= MinDelay + jitter since the last hit;
//   - non-200 on the actual request is an error.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" || u.Host == "" {
		return nil, fmt.Errorf("fetch %s: only absolute http(s) URLs are fetchable", rawURL)
	}

	if err := f.checkRobots(ctx, u); err != nil {
		return nil, err
	}
	if err := f.waitTurn(ctx, u.Host); err != nil {
		return nil, err
	}

	body, err := f.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	f.noteHit(u.Host)
	return body, nil
}

// checkRobots fetches (or reuses) host robots.txt and errors when the path
// is disallowed or the file cannot be read conservatively.
func (f *Fetcher) checkRobots(ctx context.Context, u *url.URL) error {
	f.mu.Lock()
	entry, ok := f.robots[u.Host]
	f.mu.Unlock()
	if ok && time.Since(entry.fetch) < robotsTTL {
		if !entry.rules.allowed(u.Path) {
			return fmt.Errorf("robots.txt of %s disallows %s", u.Host, u.Path)
		}
		return nil
	}

	robotsURL := u.Scheme + "://" + u.Host + "/robots.txt"
	if err := f.waitTurn(ctx, u.Host); err != nil {
		return err
	}
	body, err := f.get(ctx, robotsURL)
	f.noteHit(u.Host)
	if err != nil {
		respErr, isResp := err.(respError)
		if isResp && (respErr.status == http.StatusNotFound || respErr.status == http.StatusGone) {
			// No robots file: nothing is off limits by robots; record that
			// so every later fetch skips the probe until the TTL passes.
			f.mu.Lock()
			f.robots[u.Host] = robotEntry{rules: robotRules{agentToken: f.UA}, fetch: time.Now()}
			f.mu.Unlock()
			return nil
		}
		// Network trouble or server errors: refuse rather than crawl blind.
		return fmt.Errorf("robots.txt of %s unreadable (%v); refusing this round, try later", u.Host, err)
	}

	entry = robotEntry{rules: parseRobotRules(body, "jobcrawler"), fetch: time.Now()}
	f.mu.Lock()
	f.robots[u.Host] = entry
	f.mu.Unlock()
	if !entry.rules.allowed(u.Path) {
		return fmt.Errorf("robots.txt of %s disallows %s", u.Host, u.Path)
	}
	return nil
}

// waitTurn sleeps until this host's politeness floor has passed.
func (f *Fetcher) waitTurn(ctx context.Context, host string) error {
	f.mu.Lock()
	last := f.lastHit[host]
	minWait := f.MinDelay
	jitter := time.Duration(0)
	if f.MaxJitter > 0 {
		jitter = time.Duration(rand.Int64N(int64(f.MaxJitter)))
	}
	f.mu.Unlock()

	wait := minWait - time.Since(last) + jitter
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (f *Fetcher) noteHit(host string) {
	f.mu.Lock()
	f.lastHit[host] = time.Now()
	f.mu.Unlock()
}

// respError marks an HTTP-level failure with its status code.
type respError struct {
	status int
}

func (e respError) Error() string {
	return fmt.Sprintf("http status %d", e.status)
}

// get does the plain GET with the honest UA and a sanity-capped body.
func (f *Fetcher) get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", rawURL, err)
	}
	if f.UA != "" {
		req.Header.Set("User-Agent", f.UA)
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, respError{status: resp.StatusCode}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", rawURL, err)
	}
	return body, nil
}

// maxBodyBytes: a job list is small; anything bigger is a bug or an attack,
// and politeness includes not slurping hostile payloads.
const maxBodyBytes = 8 << 20 // 8 MiB
