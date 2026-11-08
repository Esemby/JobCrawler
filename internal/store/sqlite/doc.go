// Package sqlite implements the store.Store interface on the local SQLite
// engine (default per ADR-0002), driven by modernc.org/sqlite (pure Go, no
// CGO, so arm64 cross-compiling stays trivial). Opens set WAL mode and a
// busy timeout because the Heart is a single, slow disk; migrations run on
// every open and are idempotent.
package sqlite
