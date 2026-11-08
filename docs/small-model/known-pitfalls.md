# Known pitfalls (gpt-oss:20b and any small model)

*When to read this:* when you stumble, and at the end of every session you stumbled
(or when you discover a NEW pitfall).
*Result you should have:* one entry appended, with a concrete counter-instruction.
Entries are grouped by pattern; newest first in each group. Keep each <= 3 lines.

Format: `- <symptom> → COUNTERMEASURE: <exact rule>`

## Entries

- 2026-11-08 — Task text alone was treated as a chat opener: model replied
  "Sure! Whenever you're ready" instead of reading files or acting. An
  imperative one-sentence prompt DID work. → COUNTERMEASURE: every task text
  must BEGIN with the fixed execution preamble line "You are executing the task
  below verbatim: select and read the playbook with your file tools, then act
  autonomously until the 'Done when' criteria hold." and END with "Begin now."
- 2026-11-08 — Small model adrift starting in the parent workspace: it answers "I
  have loaded AGENTS.md" while having read the WRONG file (workspace-level, not
  the project's). → COUNTERMEASURE: sessions must start with working directory
  `jobcrawler/`; any cross-directory task text pins the absolute path; a small
  model that answers tasks before reading files is treated as mis-stepped, not
  as evidence it read them.

Watch-list (predicted failure modes; promote to entries with evidence):

- Rebuilding existing helpers from scratch instead of using stdlib/imported
  patterns → COUNTERMEASURE: grep for an existing function before writing one.
- Avoiding/deleting tests when they fail → COUNTERMEASURE: fix code first; editing
  the test needs a written justification line.
- Summarizing code changes in words instead of running them → COUNTERMEASURE:
  every claim must trace to `make check` output pasted in the transcript.
- Editing an applied SQL migration instead of adding a new one → COUNTERMEASURE:
  migrations are immutable once applied; write NNNN+1.
