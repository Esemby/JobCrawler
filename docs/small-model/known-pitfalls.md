# Known pitfalls (gpt-oss:20b and any small model)

*When to read this:* when you stumble, and at the end of every session you stumbled
(or when you discover a NEW pitfall).
*Result you should have:* one entry appended, with a concrete counter-instruction.
Entries are grouped by pattern; newest first in each group. Keep each <= 3 lines.

Format: `- <symptom> → COUNTERMEASURE: <exact rule>`

## Entries

- 2026-11-08 — Task text alone was treated as a chat opener: model replied
  "Sure! Whenever you're ready" instead of reading files or acting. An
  imperative one-sentence prompt DID work. → COUNTERMEASURE: every task text must
  BEGIN by naming ONE concrete first file to read ("One mandatory first action:
  read docs/playbooks/<name>.md with your file tools now") and END with
  "Begin now." Vague preambles ("load the playbook") did NOT work reliably;
  naming the path did. Retries are expected: never assume a silent run means the
  task ran — verify acceptance criteria outputs yourself.
- 2026-11-08 — Evidence across 8 small-model runs: with the `explore` agent the
  model acts and answers correctly (3/3); with the `general` agent it mostly
  chat-opens or hallucinates tool output (1/5, and that one was partial).
  Symptom class: tool dispatch via the ollama provider — NOT instruction-file
  quality, since identical prompts through a different provider work.
  → COUNTERMEASURE (provisional): read-based training tasks via `explore` are
  reliable; small-model write/commit tasks wait until the provider behavior is
  verified with the owner. Graduation criteria still require small-model write
  tasks; do not lower them to escape this issue.
  Second half: the small model says it "has loaded AGENTS.md" while having read
  the WRONG file (workspace-level, not the project's). → COUNTERMEASURE: sessions
  must start with working directory `jobcrawler/` (done: session moved into the
  project dir so the project AGENTS.md is the auto-loaded one); a small model
  that answers before reading is treated as mis-stepped, not as having read.

Watch-list (predicted failure modes; promote to entries with evidence):

- Rebuilding existing helpers from scratch instead of using stdlib/imported
  patterns → COUNTERMEASURE: grep for an existing function before writing one.
- Avoiding/deleting tests when they fail → COUNTERMEASURE: fix code first; editing
  the test needs a written justification line.
- Summarizing code changes in words instead of running them → COUNTERMEASURE:
  every claim must trace to `make check` output pasted in the transcript.
- Editing an applied SQL migration instead of adding a new one → COUNTERMEASURE:
  migrations are immutable once applied; write NNNN+1.
