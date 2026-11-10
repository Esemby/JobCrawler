# Known pitfalls (gpt-oss:20b and any small model)

*When to read this:* when you stumble, and at the end of every session you stumbled
(or when you discover a NEW pitfall).
*Result you should have:* one entry appended, with a concrete counter-instruction.
Entries are grouped by pattern; newest FIRST overall (newer entries at the top of
"## Entries"). Keep each <= 3 lines. NEVER edit an existing entry's text; append
new entries only.

Format: `- <symptom> → COUNTERMEASURE: <exact rule>`

## Entries

- 2026-11-10 — With toolCall binding fixed via the `ollamalocal` provider,
  gpt-oss still fails READ tasks on files past its comfort size
  (rules.go ~200 lines): attempt 1 dies in output-token truncation loops,
  attempt 2 hallucinates an unrelated program instead of reading. Rule:
  give small-model read tasks files < 100 lines, or pre-chunk into inline
  context; treat file-size, not prompt quality, as the failure lever.
- 2026-11-09 — OpenCode-native `ollama` provider + `general` agent: the model
  emits tool-call hand-rails ("We have tools... Use opencode." + raw JSON args)
  as TEXT instead of tool calls — 1/7 useful. Root evidence: session DB shows
  tool parts exist but later turns degrade to args-as-text.
  → COUNTERMEASURE: for agent-class local sessions declare the model explicitly
  under a custom provider with the `openai-compatible` package and
  "capabilities": {"toolCall": true} (config `ollamalocal` added 2026-11-09);
  tool binding improved immediately. Residual gpt-oss instability (empty replies,
  off-task answers, "CANNOT READ FILE" bailouts) is MODEL behavior at this
  context size, not wiring — do not lower graduation criteria; prefer
  `explore`-agent runs (3/3) for read-based training until write-runs stabilize.
- 2026-11-08 — Structured task briefs were treated as chat openers: model replied
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
