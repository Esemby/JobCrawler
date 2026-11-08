# Playbook: add a source

*When to read this:* a task says "add source X" or "add a source adapter".
*Result you should have:* new adapter + fixture + test + UI listable + registered.
STATUS: template — fill in corrections after the first real use (Phase 3).

## Checklist

1. Check `docs/sources.md` register: is the source already listed as `rejected` for
   ToS reasons? If yes: STOP, report to owner, do not implement.
2. Confirm the source is structured-first (public ATS JSON/XML, RSS, JSON-LD).
   Building plain-HTML scraping when a public feed exists is a REJECTED approach.
3. Save a real, minimal non-secret fixture: `testdata/<source>/<case>.json|html`
   (or ask the owner to fetch one; never fetch live in tests). "Minimal" means:
   <= 20 KB, <= 10 job records, and at least 2 records so loops iterate past one.
4. Create `internal/sources/<kind>/<name>.go` implementing `Source` (see
   `docs/architecture.md`). Parser file < 300 lines; one parse function per record.
5. Write table-driven test feeding the fixture; every field you map is asserted
   (see `write-a-test.md`). Minimum mapping bar: assert company, title,
   location, and URL for the FIRST and LAST record in the fixture.
6. Register the source type in the sources index so the UI can list it.
7. Add a row to `docs/sources.md` register: method, ToS note, status.
8. `make check` → green or fix.
9. Commit: `feat(sources): add <name> adapter with fixture test`
10. Update `docs/small-model/handoff.md` and append any surprise to `docs/lessons.md`.

## If this fails

- No public fixture available: ask the owner for one; do NOT scrape production
  endpoints from the dev machine as a workaround without asking.
- Source requires login or violates robots: document as `rejected` in
  `docs/sources.md` and stop.
