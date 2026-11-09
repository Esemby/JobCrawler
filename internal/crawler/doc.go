// Package crawler fetches source listings politely (docs/sources.md
// politeness contract): per-host rate limiting with jitter, robots.txt subset
// matching, an honest User-Agent, and plain stdlib HTTP. It orchestrates one
// source end-to-end: fetch -> parse -> fingerprint -> store upsert.
//
// Everything is injectable (http.Client, delays) so tests run against
// httptest servers with zero network. Keep the rate-limit maps on the struct,
// never in globals.
package crawler
