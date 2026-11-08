# Task template

*When to read this:* a task is being written (by you or the owner) for a session.
*Result you should have:* a task small enough to execute in one session.

## Rules for tasks

- Preamble and closing are mandatory and always byte-identical:
  START: "You are executing the task below verbatim: load the playbook with your
  file tools first, then act autonomously until the 'Done when' criteria hold."
  END: "Begin now."
- ONE verb, ONE area. "Add an endpoint" ok; "add endpoint + redesign store" not ok.
- Written imperative, includes: goal, exact files/paths, done criteria, and the
  playbook to follow.
- Explicit stop condition: when to stop, when to ask the human, when to revert.
- No vague words: "appropriate", "clean", "robust", "as needed" forbidden. Numbers
  or examples instead.

## Format

```
TASK: <one imperative line, verb first>
Phase: <roadmap phase>
Playbook: docs/playbooks/<name>.md
Goal: <1-2 lines, concrete outcome>
Files to touch: <exact paths; "~" means "new file">
Done when:
  1. <observable criterion>
  2. `make check` green
  3. docs rows updated per update-docs-after-change.md
STOP if: <boundary condition>
```

## Examples

```
TASK: Add GET /healthz returning 200 "ok" and version string
Phase: 1
Playbook: docs/playbooks/add-a-ui-page.md
Goal: healthz handler + test; wire only this endpoint.
Files to touch: internal/web/healthz.go (~), internal/web/healthz_test.go (~),
  cmd/jobcrawler/main.go (edit: register route)
Done when:
  1. curl http://127.0.0.1:8090/healthz returns 200 and body "ok vX.Y.Z"
  2. make check green
  3. handoff.md updated
STOP if: it requires touching internal/ packages other than web/config/version.
```

```
TASK: Add setting crawl.concurrency (runtime, default 2, validated 1-8)
Phase: 2
Playbook: docs/playbooks/add-a-setting.md
Done when: UI field visible with help text; invalid value rejected with message
  naming the key; docs/config.md row exists; make check green.
STOP if: hot reload impossible → mark "restart required" in UI, note in handoff.
```
