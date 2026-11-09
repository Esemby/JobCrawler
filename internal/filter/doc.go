// Package filter is the Stage 1 rules engine (docs/original-brief.md 4.2):
// cheap, deterministic, explainable. Every rejection produces human-readable
// reason lines; the stage-2 scoring layer (later) builds on the same Verdict
// shape.
//
// The package is pure: no store, no network, no time — inputs in, verdicts
// out, so tests are table-driven and fast. Persistence lives in store
// (Verdict type) and the wiring lives in the caller (web crawl handler).
package filter
