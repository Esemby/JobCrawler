# ADR-0001: Language and stack

*When to read this:* before proposing any new dependency, framework, or language change.
*Result you should have:* your proposal either fits inside this stack or you write a
new ADR with measurements and ask the owner.

## Status

Accepted. Pre-decided in the original brief (section 3). Do not relitigate without a
benchmark and the owner's approval.

## Context

The service must run 24/7 on a Raspberry Pi 4B (aarch64, scarce RAM and SSD wear).
The builder may become a small local model (gpt-oss:20b) that needs unambiguous,
fast feedback from the compiler and test suite.

## Decision

- **Language: Go (current stable release).**
- Web UI: server-rendered `html/template` + **htmx vendored as a static file**.
  No CDN, no JS build step, plain CSS. Must work without internet.
- HTTP routing: standard library `http.ServeMux` (Go 1.22+ patterns).
- HTML parsing: `golang.org/x/net/html` or `github.com/PuerkitoBio/goquery`.
- Crawler: own thin layer on `net/http` (politeness, robots, per-host limits).
  `colly` allowed only if it genuinely saves code — justify in a new ADR.
- Migrations: plain numbered `.sql` files embedded with `embed`, tiny runner.
- Config: one TOML/YAML bootstrap file + env overrides (ADR-0002 for storage).
- Logging: `log/slog` (JSON or text selectable). Tests: stdlib `testing`,
  table-driven, `httptest` with fixtures in `testdata/`.
- Gate: `gofmt`, `go vet`, `staticcheck` (if installable).

Rejected (do not revive without new measurements): Python (slow on Pi, venv/dep
drift, runtime type errors), Rust (borrow checker friction for small models),
Node/TS (heavier runtime, build chain, dependency churn).

## Consequences

- One static arm64 binary; systemd unit + binary = deployment done.
- Standard library first; every dependency needs its own ADR before code exists.
- Compiler + vet + tests are the small model's teacher; keep them fast and strict.
