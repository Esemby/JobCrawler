// Package sources parses structured job listings into store.Job records.
//
// One file per source type (docs/sources.md priority order). Parsers are
// pure: they take saved bytes and return jobs, never touch the network and
// never touch the store — the crawler wires fetching + storage. Network
// adapter functions accept an *http.Client so tests use httptest.
package sources
