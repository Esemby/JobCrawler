// Package config loads and validates bootstrap configuration for jobcrawler.
//
// Phase 1 scope: bind address and log settings from env vars with defaults.
// Storage engine selection (ADR-0002) will grow here in Phase 2. Validation is
// the boundary: Load fails before main starts serving if anything is invalid.
package config
