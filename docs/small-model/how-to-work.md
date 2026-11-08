# How to work (the loop)

*When to read this:* the FIRST thing you read every session, after `../AGENTS.md`.
*Result you should have:* one committed task and a fresh handoff note.

## The loop (never skip a step, never do two tasks)

1. Read `../AGENTS.md` (rules + index). Then `handoff.md` (current state).
2. `git status` — if the tree is dirty, STOP and tell the owner in your note;
   unexplained dirt is not yours to clean.
3. Pick ONE task from `handoff.md` → "Next task". Never self-assign bigger work.
4. Read the playbook for that task type (index in `../AGENTS.md`). If there is no
   playbook, follow `../AGENTS.md` rules and write one as part of the task.
5. Plan in ≤ 5 lines (in your head or a scratch note; planning must not consume
   context budget).
6. Implement. Smallest change that completes the task.
7. Run `make check`. Green? Continue. Red? Playbook `fix-a-failing-test.md`.
   3 failed attempts → STOP at step 12 with findings.
8. Update docs: follow `../playbooks/update-docs-after-change.md` checklist (same
   commit as the code).
9. Commit (Conventional Commits). Message = what changed, one line.
10. Rewrite `handoff.md` (see below).
11. If you hit a pitfall this session, add one line to `known-pitfalls.md`.
12. Stop. Report in your final answer: task, commit hash, `make check` result,
    what is next. If blocked, say exactly which step blocked you.

## handoff.md contents (overwrite, never append)

- Date, model, commit hash at end of session
- Project state: which roadmap phase, where it stands (2-3 lines)
- Next task: ONE task, written with `task-template.md`
- Blockers + questions for the owner
- Verify commands: the exact 2-3 commands that prove the tree is healthy

## Stop conditions (any → stop and report, do not improvise)

- Test stays red after 3 attempts
- A needed dependency (ADR + owner needed)
- Dirty tree you didn't create
- Task text contradicts a doc; treat docs as source of truth and report the conflict

## Size limits

- Files < 300 lines, functions < 50 lines. Prefer a new small file over editing a
  big one.
- Read files by range (`sed -n`), never dump big files into context (see
  `context-budget.md`).
