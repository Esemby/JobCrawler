# Lessons (append-only)

*When to read this:* when you make a mistake, discover a non-obvious fact about the
environment, or break something — before you commit the fix.
*Result you should have:* one line below with date, what was learned, and the rule
that prevents recurrence. Newest LAST. Never delete or edit old lines.

- 2026-11-08 — The repo remote is SSH (`git@github.com:Esemby/JobCrawler.git`); the
  machine has no HTTPS credentials, so HTTPS pushes fail with "could not read
  Username". Rule: push via SSH; if a new environment lacks the key, hand the
  owner one setup step, never store credentials.

Rules: one lesson = one insight + one rule. If a lesson needs context, link the file.

## Lessons

- 2026-11-08 — The brief lives outside the repo (`~/Downloads/PROJECT_BRIEF.md`); the
  canonical copy inside the repo is `docs/original-brief.md`. Rule: never edit the
  original; treat it as read-only history and reflect changes in living docs.
- 2026-11-08 — Pre-decided facts live in ADRs, not in commit messages or heads.
  Rule: cite the ADR number (e.g., "per ADR-0002") instead of re-arguing in code
  review.

- 2026-11-08 — Handover test #0: the session loop was exercised end to end with a
  docs-only task. Rule: docs-only commits still follow the full loop, including handoff.md update.
