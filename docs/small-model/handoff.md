# Handoff

OVERWRITE this file at the end of every session (see `how-to-work.md` step 10).
No history here. Newest state only.

## Session

- Date: 2026-11-09/10 (overnight block while owner slept, 6h window)
- Model: GLM-5.3-Flash Free (cloud builder agent)
- HEAD: 8f48fd5 pushed to origin/main; `make check` green; git clean.

## Project state

- Phases 0-3 done and pushed: skeleton, store, settings UI, crawler with
  politeness, greenhouse source, jobs page, first REAL crawl (propublica,
  4 jobs, dedupe verified), UA contact resolved (owner decision 1.3: public
  repo issue tracker; GitHub relays no mail to noreply addresses).
- Phase 4 stage 1 + stage 2 v1 DONE this block, commit by commit:
  1. `internal/filter` (new package, pure): stage-1 rules — company
     blacklist, no-gos, must-haves with AND semantics and `|` OR alternati-
     ves inside an entry; fixed reason order (blacklist → no-gos → must-
     haves); 10 table-driven tests, exact reason strings asserted.
  2. Migration 0003 `job_verdicts` (fingerprint PK, result CHECKed,
     newline-joined reasons); store methods SaveVerdicts/ListVerdicts;
     two-statement upsert (no SQLite-only ON CONFLICT — both engines).
  3. Config rows filter.blacklist_companies / filter.no_gos /
     filter.must_haves (+validation: single line ≤ 2000 chars); editable on
     the existing System page.
  4. Wiring: crawl run now crawls, then re-filters ALL active jobs and
     upserts verdicts (idempotent by fingerprint; verdicts stay truthful
     when settings change — covered by test); jobs page has Stage 1 badge +
     Why column.
  5. Stage 2 v1: `filter.tag_weights` setting ("name@weight=kw1|kw2", ...),
     score = sum of matched tag weights; informational (no cutoff);
     migration 0004 adds score/score_reasons columns; jobs page shows a
     score badge and the breakdown appended to Why.
  6. Two real-binary smokes passed: verdicts + scores on live propublica
     data (3 accepted / 1 rejected with "no-go match: video editor";
     3× score 4, 1× score 0 suppressed).
- Owner decisions recorded: 1.3 (UA contact = repo issues URL), earlier
  2.2 (placeholder-until-first-crawl — now resolved).

## Blockers / owner input needed

None. Everything needed for stage-3 (AI review) exists in the brief but it
is explicitly disabled by default; no owner input required until they want
it enabled (would need the Ollama IP/port they already mentioned).

## Next task (ONE)

Phase 4 leftovers, in this order (docs/playbooks/reference):
1. Jobs page usability: search/sort-by-score, and owner flags
   seen/applied/hidden (glossary "Seen flags") — add-a-migration 0005 +
   add-a-ui-page scope, small.
2. Stage 2 completion: recency + source-trust score components (needs a
   scheduler to be meaningful — pair with the crawl schedule work) and the
   docs example in `docs/testing.md` if the breakdown format changes.
3. Phase 5: scheduler (poll per source, settings crawl.interval etc.),
   systemd-ready.
Training task option meanwhile: another `explore` read task on a file under
100 lines (rules.go is ~230 → fails; candidate: internal/store/job.go, ~95
lines) — score it into docs/small-model/scorecard.md.

## Invariants to keep

- `make check` before every commit; Conventional Commits; never force-push.
- Port 8090 default, never 8080; UI stays LAN-only (rule 6).
- filter package stays pure (no store/network/time); strings for reasons
  are human-first, escaping-safe (avoid leading "+", template escapes it).
- Every instruction file < 150 lines target (hard cap 250).
