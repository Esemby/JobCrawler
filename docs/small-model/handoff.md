# Handoff

OVERWRITE this file at the end of every session (see `how-to-work.md` step 10).
No history here. Newest state only.

## Session

- Date: 2026-11-08
- Model: GLM-5.3-Flash Free (cloud bootstrap agent)
- Commit: e6dd7c6 pushed to origin/main (see `git log --oneline -4`)

## Project state

- Phase 0 complete except **handover test #0** (in progress with the owner).
- Remote live at github.com/Esemby/JobCrawler (SSH remote, pushed, tracking main).
- Owner answered all six Phase 0 questions → recorded in
  `docs/decisions/0004-owner-answers-phase0.md`. Key facts: Heart 3.9 GB RAM;
  EU + Norway, German + English (configurable filters); jobcrawler, MIT; first
  source Greenhouse; LAN-only UI, no auth yet, multi-user-ready flags; Ollama on
  the Flagship.
- No product code yet, by design. Next code phase is 1 (skeleton + /healthz + MIT
  LICENSE), which starts after handover test #0 passes.

## Next task (ONE)

Handover test #0 — owner runs it with gpt-oss:20b in OpenCode, fresh session, this
repo as workspace, nothing else pasted. Task text is below, paste EXACTLY:

```
TASK: Append one lesson line to docs/lessons.md
Phase: 0
Playbook: docs/playbooks/update-docs-after-change.md
Goal: append, do not edit existing lines, exactly this lesson at the end of "## Lessons":
- 2026-11-08 — Handover test #0: the session loop was exercised end to end with a
  docs-only task. Rule: docs-only commits still follow the full loop, including
  handoff.md update.
Done when:
  1. `grep -c "Handover test #0" docs/lessons.md` prints 1
  2. docs/small-model/handoff.md rewritten per how-to-work.md step 10
  3. git status clean after one commit
STOP if: any step is ambiguous — report instead of guessing.
```

## Blockers / questions for the owner

- Run handover test #0, then paste back the small-model transcript or a summary of
  where it got stuck.

## Verify commands

```
git -C . log --oneline -5
grep -c "Handover test #0" docs/lessons.md   # 1 after the test
make lint                                     # expected: green (vet/gofmt no-op now)
```
