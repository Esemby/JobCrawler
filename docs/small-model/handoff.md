# Handoff

OVERWRITE this file at the end of every session (see `how-to-work.md` step 10).
No history here. Newest state only.

## Session

- Date: 2026-11-08 (overnight block)
- Model: GLM-5.3-Flash Free (cloud bootstrap agent) — with gpt-oss:20b subagent runs
- Commit: see `git log --oneline -6`

## Project state

- Phase 0(exit gate ran): handover test #0 executed twice by gpt-oss:20b via
  opencode subagent — attempt 1 failed (chat-opener response, zero tool use),
  attempt 2 passed the task (lesson appended + committed) with partial loop
  fidelity. Both attempts + fixes logged in `docs/small-model/scorecard.md` and
  `docs/small-model/known-pitfalls.md`.
- Phase 1 skeleton (bind/c/version/config/web+healthz+status/htmx/LICENSE/
  systemd unit) landed on the Flagship in `~/tools/go` (no sudo available),
  symlinked at `~/.local/bin/go`.
- Repo: github.com/Esemby/JobCrawler (SSH remote `origin/main`, no force-push).
- Next in training loop (per scorecard progress):: handover test #1 in a fresh
  clean tree, then a docs-only task variant, then raising difficulty
  (`make check`-gated tasks only).

## Next task (ONE)

Run handover test #1 (gpt-oss:20b via opencode subagent): same loop, task =
add `internal/version` CHANGELOG-style entries? — no: first assess Phase 1
acceptance from a clean tree:
    1. `go build ./...` green
    2. `go test ./...` green
    3. `make check` green

## Blockers / questions for the owner

- None blocking. Owner asleep per ADR-0005; defaults documented there.

## Verify commands

```
git log --oneline -6
make check
curl -s http://127.0.0.1:8090/healthz   # after make run
```
