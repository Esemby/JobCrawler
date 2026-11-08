# Handoff

OVERWRITE this file at the end of every session (see `how-to-work.md` step 10).
No history here. Newest state only.

## Session

- Date: 2026-11-08
- Model: GLM-5.3-Flash Free (cloud bootstrap agent)
- Commit at end: 0000000 (uncommitted — Phase 0 created the repo fresh)

## Project state

- Roadmap: Phase 0 (bootstrap instruction system) — docs/AGENTS.md/Makefile built.
- No product code yet, by design (brief says Phase 0 before product code).
- Phase 0 remaining: owner answers (brief section 10), first handover test, phase 0
  gate check.

## Next task (ONE)

Owner answers six open questions (brief section 10) — then FIRST code task is
Phase 1 skeleton: `go build`, `go test ./...`, `make run`, `/healthz` + status page.
Task template: docs/small-model/task-template.md.

## Blockers / questions for the owner

- The 6 open questions in brief section 10 (RAM, region/languages, name+license,
  first source, UI login day-one?, Ollama endpoint location).

## Verify commands

```
make lint        # will pass now (no Go files yet → vet is a no-op, gofmt clean)
test -f AGENTS.md && test -f docs/original-brief.md && echo "baseline files OK"
git log --oneline | head -5
```
