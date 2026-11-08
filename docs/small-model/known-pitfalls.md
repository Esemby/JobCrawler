# Known pitfalls (gpt-oss:20b and any small model)

*When to read this:* when you stumble, and at the end of every session you stumbled
(or when you discover a NEW pitfall).
*Result you should have:* one entry appended, with a concrete counter-instruction.
Entries are grouped by pattern; newest first in each group. Keep each <= 3 lines.

Format: `- <symptom> → COUNTERMEASURE: <exact rule>`

## Entries

(open — will fill during handover tests; see 9.4 of the original brief)

Watch-list (predicted failure modes; promote to entries with evidence):

- Rebuilding existing helpers from scratch instead of using stdlib/imported
  patterns → COUNTERMEASURE: grep for an existing function before writing one.
- Avoiding/deleting tests when they fail → COUNTERMEASURE: fix code first; editing
  the test needs a written justification line.
- Summarizing code changes in words instead of running them → COUNTERMEASURE:
  every claim must trace to `make check` output pasted in the transcript.
- Editing an applied SQL migration instead of adding a new one → COUNTERMEASURE:
  migrations are immutable once applied; write NNNN+1.
