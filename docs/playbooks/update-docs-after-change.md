# Playbook: update docs after a change

*When to read this:* the end of EVERY task, before committing. Also when the task
says "docs only".
*Result you should have:* docs, handoff note, and ledger rows consistent with the
code in the same commit.

## Checklist

1. Does the change add/alter a **setting**? → `docs/config.md` row (key, type,
   default, validation, help).
2. Does it add/alter a **source**? → `docs/sources.md` register row + ToS note.
3. Does it add/alter a **package boundary or interface**? → `docs/architecture.md`.
4. Does it add a **domain word**? → `docs/glossary.md`.
5. Does it add a **new recurring task**? → new playbook in `docs/playbooks/` and an
   index row in `AGENTS.md`.
6. Did you make a **decision** (even a small assumption)? → numbered ADR in
   `docs/decisions/`, title on its own line.
7. Did you (or the owner, or the small model) hit a **mistake** today? → one line in
   `docs/lessons.md` (append-only, newest last).
8. Update `docs/small-model/handoff.md`: current state, next task, blockers,
   verify commands.
9. `make check` → green (docs-only changes skip the build, not the habit).
10. Commit everything in ONE commit: e.g. `docs: <what changed>` or combined
    `feat(x): ... ; docs: ...` — never leave docs for "later"; later never comes.

## Rules

- Docs use exact commands and expected output; prose is a failure mode.
- New instruction files: < 150 lines target, hard cap 250, numbered steps.
- Add index rows in `AGENTS.md` so no file is orphaned.

## If this fails

- Unsure whether a doc needs updating: update it; over-recorded docs beat
  under-recorded docs. Trim next session if bloated.
