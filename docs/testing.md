# Testing

*When to read this:* writing or running any test, or deciding if work is "done".
*Result you should have:* you know how to be "green" and how fixtures work.

## Commands

```
make test          # go test ./...
make check         # gofmt + go vet + staticcheck (if present) + go test ./...
```

A task is not done until `make check` passes. That is the definition of green.

## Rules

- Table-driven tests only. One table, one loop, subtests named `t.Run`.
- HTTP interactions use `httptest`; HTML/JSON fixtures are real saved files in
  `testdata/<source-name>/…` — hand-built fake HTML is wrong by construction.
- Every new behaviour gets (a) a test, (b) a docs line, (c) a changelog line.
- No network calls in unit tests. If a test wants a server, it starts a local
  `httptest.Server` on 127.0.0.1.
- Fix a failing test by fixing the code OR the test's expectation — never by
  deleting the test. Deleting a test needs a written reason in the commit message.
- Store tests are written once, against the `Store` interface, and run for every
  backend (SQLite now, MariaDB in Phase 6) using a constructor seam.

## Fixture conventions

- `testdata/<topic>/<case-name>.html|json` — small, minimal, but structurally real.
- Name tests after the fixture: `TestParseGreenhouseBoard` with fixture
  `testdata/greenhouse/board-01.json`.
- A failing behaviour that reached the field gets its fixture saved here plus a
  regression test before the fix (see `playbooks/fix-a-failing-test.md`).
