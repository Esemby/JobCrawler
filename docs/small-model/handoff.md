# Handoff

OVERWRITE this file at the end of every session (see `how-to-work.md` step 10).
No history here. Newest state only.

## Session

- Date: 2026-11-09 (overnight block, closed ~11:35)
- Model: GLM-5.3-Flash Free (cloud bootstrap agent) + background Phase-2 agent
- Commit: 0fd07c9 pushed to origin/main; `git status` clean

## Project state

- Phase 1 done, pushed. Phase 2 done, pushed: Store interface + embedded
  migration 0001 (settings, job_sources, jobs, users, user_flags skeleton per
  ADR-0004), SQLite engine (WAL, modernc.org/sqlite, no CGO), bootstrap TOML
  with env precedence, settings CRUD, System + Database pages (validation,
  hot-reload hook, masked DSN, sqlite test-connection), wired in main.
- Training loop: handover #0 passed (2nd attempt), #1 passed partially via
  explore-agent reads; write-tasks through `general` blocked by ollama provider
  tool-dispatch failures — evidence in docs/small-model/scorecard.md and
  known-pitfalls.md. gpt-oss:20b IS now pulled into local ollama (was missing).
- Apply the morning-review rules after reading the docs above.

## Next task (ONE)

Phase 3 — first real crawl (Greenhouse public JSON, per docs/sources.md
register + docs/playbooks/add-a-source.md). Needs the owner for: ToS sign-off
for using boards.greenhouse.io public JSON at crawl volume, and a UA contact
email for `crawl.user_agent`. Before code: save a real non-secret fixture of a
Greenhouse board into testdata/greenhouse/.

## Blockers / questions for the owner

1. Greenhouse ToS check + polite volume: confirm allowed request rate.
2. Contact email for User-Agent (politeness contract requires it).
3. opencode↔ollama provider tool-dispatch — decide whether to debug locally or
   run small-model tests from another runner.

## Verify commands

```
git log --oneline -8 ; git status --short   # expect clean
make check                                   # expect green
./bin/jobcrawler & curl -s http://127.0.0.1:8090/healthz   # "ok dev"
curl -s http://127.0.0.1:8090/system         # settings table
```
