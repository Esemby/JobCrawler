# Playbook: fix a failing test

*When to read this:* `make check` is red for you or for a task.
*Result you should have:* `make check` green, root cause understood, lesson recorded
if it was a surprise.

## Checklist

1. Reproduce alone: `go test ./internal/<pkg>/ -run <TestName> -v`. Read the diff
   in the output before reading any code. "want / got" is usually the whole bug.
2. Classify: (a) code bug, (b) test expectation stale, (c) fixture stale,
   (d) flaky test.
3. (a) Fix the code. (b)/(c): confirm against the task description AND docs before
   editing the expectation. Never edit a test to make red go green without a
   one-line justification in the commit message.
4. (d) Flaky: find the nondeterminism (time, map order, port, goroutine race).
   Run `go test -race -count=3 .` on the package. Never add `time.Sleep`.
5. If the failing behaviour could recur in the field, save its input as a fixture
   in `testdata/` before fixing.
6. `make check` → green.
7. Commit: `fix(<pkg>): <one line cause>`.
8. If the failure surprised you, append one line to `docs/lessons.md`.

## Stop conditions

- 3 fix attempts, still red → STOP, write the findings in the handoff note, ask
  the owner. Do not commit red code and run away.
- The fix needs a new dependency → ADR first (rule 4 in AGENTS.md).
- The fix touches a pre-decided ADR area (storage, language) → STOP, it's an
  owner decision.

## worked example of thinking (keep)

```
FAIL TestNormalizeURL …
want  https://esemby.org/jobs/12
got   https://esemby.org/jobs/12?utm=x
→ normalize query-stripping missing whitelist for tracking params; add utm_*.
```
