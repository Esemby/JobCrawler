// Package store defines the Store interface: the single boundary for all
// persistence (jobs, sources, settings, verdicts) per docs/architecture.md.
//
// The interface lives here; per-engine implementations live in subpackages
// (sqlite/ now, mariadb/ in Phase 6). Tests are written once against the
// interface and run for every backend via a constructor seam.
//
// Heart uptime is the design constraint (ADR-0002): the default engine is
// local SQLite in WAL mode, so the Heart keeps working when MariaDB is
// unreachable. Settings are opaque key/value strings here; keys, types,
// defaults and validation live in internal/config (docs/config.md).
package store
