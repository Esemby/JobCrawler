# Architecture

*When to read this:* before adding or changing a package, an interface, or a DB table.
*Result you should have:* you know which package your change belongs in and which
direction dependencies may point.

## Package layout

```
cmd/jobcrawler/        main.go: wiring only, no logic
internal/config/       bootstrap config + runtime settings (typed, validated)
internal/store/        Store interface + sqlite/ and mariadb/ implementations
internal/store/migrations/   numbered .sql files
internal/crawler/      scheduler, fetcher, politeness, robots
internal/sources/      one package/file per source type (ats/, rss/, jsonld/, html/)
internal/filter/       rules engine, scoring, explain output
internal/review/       Reviewer interface + ollama implementation + cache
internal/web/          handlers, templates/, static/ (embedded, incl. vendored htmx)
internal/version/      build-time version info
docs/                  instruction + decision + reference docs
testdata/              saved HTML/JSON fixtures per source
```

## Core interfaces (boundaries)

Only three interfaces cross packages. Do not add more without an ADR.

```go
type Store interface { /* save/load jobs, sources, settings, verdicts */ }
type Source interface { // one method set: fetch listings -> []Job
    Name() string
    Fetch(ctx context.Context) ([]Job, error)
}
type Reviewer interface { // optional stage 3
    Review(ctx context.Context, job Job) (Verdict, error)
}
```

## Data flow (one direction, left to right)

```
scheduler -> Source.Fetch -> fetcher (politeness: robots, rates, ETag)
          -> dedupe (fingerprint) -> Store
          -> filter stage 1 hard rules -> stage 2 scoring -> (.reason per step)
          -> [optional] review.Reviewer -> Store (verdict + cache)
          -> web UI reads Store
```

Dependency direction: always inward toward `internal/filter` and the domain types in
`internal/store`. Allowed: `web -> everything`, `crawler -> sources -> filter`.
Forbidden: `filter -> crawler`, any import of `internal/web` from non-web packages.

## Rules you must not break

- No global state, no package-level mutable variables.
- `context.Context` as first parameter on I/O calls.
- Errors wrapped with context: `fmt.Errorf("fetch %s: %w", url, err)`.
- No panics in library code. No silent `_ = err`.
- Files < 300 lines, functions < 50 lines. Prefer a new small file.
- Every package has `doc.go` with 3-6 lines of purpose and boundaries.
- The Heart is poor: batch DB writes, WAL mode, no raw HTML by default,
  graceful SIGTERM shutdown.

## Job flags are per-user (ADR-0004)

`Seen flags` (new / seen / applied / hidden) MUST be modeled as per-user rows from
the first schema revision, even though Phase 1 ships with one implied user. The
owner plans multi-user (friends) later; retrofitting identity onto shared flags
would be expensive. Jobs themselves stay shared; only the flags are per user.

## Storage (ADR-0002, pre-decided)

Default engine is local SQLite (WAL), so the Heart survives VPN/MariaDB outages.
MariaDB is the optional second engine behind the same `Store` interface. The
bootstrap file (`config.bootstrap.toml`) holds *which* engine + credentials, because
those settings cannot live inside the database they describe. Runtime settings live
in the active store; export/import migrates between engines.
