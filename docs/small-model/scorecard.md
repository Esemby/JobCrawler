# Scorecard (handover test log)

*When to read this:* before/after running a handover test (brief section 9.4).
*Result you should have:* one row per test run, and the linked fix documented.

Graduation criteria ("lone builder"): small model completes, unaided, `make check`
green: (1) new setting with UI, (2) new source adapter, (3) new filter rule type,
(4) bug fix from a failing test, (5) docs-only update, (6) release build for Pi —
each at least TWICE in a row.

| Date | Task | Model | Passed unaided? | Failure cause | Doc fix made |
|---|---|---|---|---|---|
| 2026-11-08 | handover #0, attempt 1: append lesson | gpt-oss:20b (via opencode provider) | NO | treated structured task text as chat opener; zero tool use | task-template.md: mandatory preamble + "Begin now." |
| 2026-11-08 | handover #0, attempt 2: same task | gpt-oss:20b | PARTIAL | lesson + commit correct; handoff not attributed to itself; empty final answer; tree was pre-polluted by cloud-model WIP (env fault, now a protocol rule) | how-to-work.md step 10: self-attribution checklist |
| 2026-11-08 | handover #1, attempt 1: glossary terms | gpt-oss:20b | NO | chat opener again; vague preamble insufficient | task-template: preamble names ONE concrete file to read |
| 2026-11-08 | handover #1, attempt 2: same task | gpt-oss:20b | NO | answered IN the workspace-AGENTS.md question format => wrong session cwd; wrong AGENTS.md auto-loaded | protocol: session relocated into project dir; project AGENTS.md now auto-loaded |
| 2026-11-08 | handover #1, attempt 3: same task | gpt-oss:20b | NO | zero action despite correct cwd — general-agent tool dispatch failure class (see known-pitfalls) | known-pitfalls entry; read-based tasks via explore proven 3/3 |
| 2026-11-08 | read dry-run of add-a-source.md | gpt-oss:20b | YES | first reported both real ambiguities (fixture minimality, mapping bar) — doc feedback adopted; one provider truncation retry needed | add-a-source.md steps 3+5 now define "minimal" and the minimum asserted field bar |Notes: attempt 2 ran against pre-polluted working tree (Phase 1 WIP present), so
"criterion: git status clean" was unsatisfiable by design — assess future tests
from a clean tree.
