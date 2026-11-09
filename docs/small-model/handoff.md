# Handoff

OVERWRITE this file at the end of every session (see `how-to-work.md` step 10).
No history here. Newest state only.

## Session

- Date: 2026-11-09 (day block, second half)
- Model: GLM-5.3-Flash Free (cloud builder agent)
- HEAD: ff816f5 + web commit (see git log); `git status` should be clean after
  the docs commit this session ends with. `make check` green.

## Project state

- Phases 0-2 done and pushed (skeleton, store, settings config UI).
- Phase 3 DONE and pushed in this block, commit by commit:
  1. Greenhouse ToS verdict recorded in docs/sources.md (official docs say the
     job-board GET endpoints are public and unauthenticated; robots.txt of
     boards-api.greenhouse.io disallows only /embed/, our path is permitted).
     Real 4-job fixture saved: testdata/greenhouse/board-propublica.json.
  2. internal/sources/greenhouse: parser + endpoint builder + ToJobs mapper,
     fixture-backed tests incl. junk rejection and invalid tokens.
  3. internal/store: Job + Source domain types, Fingerprint (normalization +
     canonical URL), sqlite jobs/sources CRUD, migration 0002 (lifecycle
     columns), storetest contract suite extended (16 cases).
  4. internal/crawler: robots.txt subset parser (RFC 9309 semantics, pinned by
     15 fixture cases), polite fetcher (per-host ≥2s + ≤1s jitter, robots
     cached ≤24h, robots-unreachable => refuse round, honest UA), and
     CrawlSource orchestration tested end-to-end over httptest.
  5. Web: /jobs page (job table + source table + add/toggle forms +
     "Run crawl now"), POST /crawl/run, POST /sources/add, POST /sources/
     toggle; nav links on all pages; handlers fail loudly without store.
  6. SMOKE TEST against the real world PASSED: ran the binary, added source
     greenhouse-demo/propublica via the UI loop, POST /crawl/run stored 4
     real ProPublica jobs, second run reported "0 new" (dedupe works).
- Training loop: OpenCode provider investigation finished (owner decision
  3.1). Findings: the built-in `ollama` provider mis-binds gpt-oss tool calls
  (args emitted as text, 1/7 useful). Added `ollamalocal` provider (custom
  package @opencode/ai/providers/openai-compatible, baseURL 127.0.0.1:11434,
  toolCall capability declared) — tool BINDING fixed; residual failures are
  model quality (empty replies, hallucinated context). explore-agent read
  tasks remain 3/3. Evidence: docs/small-model/known-pitfalls.md, scorecard.

## Blockers / owner input needed (only ONE)

1. UA contact email: the first real crawl has now HAPPENED (2026-11-09), so
   per owner decision 2.2 the placeholder in crawler.DefaultUA must be
   replaced with a real contact BEFORE any further/scheduled real crawl.
   Until the owner supplies the address: do not run real crawls; tests all
   use fakes, so nothing is blocked in code.

## Next task (ONE)

Phase 4 — filter engine (docs/original-brief.md section 4 via
docs/playbooks/add-a-filter-rule.md): tags, blacklist terms, must-haves and
no-gos per job, scoring with "why" explanations persisted as rows, jobs page
sprouting a per-job verdict display. Start by reading docs/glossary.md
(fingerprint, verdict) and docs/architecture.md seams (filter is inward;
web -> everything is allowed).

## Invariants to keep

- `make check` before every commit; Conventional Commits; never force-push.
- Port 8090 default, never 8080; UI stays LAN-only (rule 6).
- robots subset parser stays stdlib-only (any dependency needs an ADR first).
- Every instruction file < 150 lines target (hard cap 250).
