# ADR-0002: Storage engines

*When to read this:* before touching `internal/store`, migrations, or DB settings.
*Result you should have:* you know why SQLite is default and what the bootstrap file
must contain.

## Status

Accepted. Pre-decided in the original brief (section 5). Same relitigation rule as
ADR-0001.

## Context

The Heart must stay up even when the remote MariaDB behind the VPN (on the owner's
father's LAN) is unreachable. Also, the database connection settings cannot live
inside the database they describe — chicken and egg.

## Decision

- **Default engine: local SQLite** (WAL mode) on the Heart via
  `modernc.org/sqlite` — pure Go, no CGO, so `GOARCH=arm64` cross-compiling stays
  trivial.
- **MariaDB is the optional second engine** behind the same `Store` interface
  (`github.com/go-sql-driver/mysql`), selectable from the UI ("Database access")
  with a "test connection" button. Password never shown back in clear text.
- The **bootstrap file** (`config.bootstrap.toml`) always holds *which* engine plus
  credentials; env overrides allowed. It must never be stored inside the managed DB.
- Runtime settings (filters, sources, …) live in the active store. Export/import
  lets the owner migrate settings between engines.
- Store tests are written once against the interface and must pass for every backend.

## Consequences

- The Heart survives VPN outages with zero data path degradation.
- Pruning writes matters: WAL mode, batch writes, retention policy for old data;
  raw HTML is not stored by default.
- Never delete or "fix" the bootstrap file during debugging; losing it means the
  service cannot open its store on next start.
