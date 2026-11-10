# Lessons (append-only)

*When to read this:* when you make a mistake, discover a non-obvious fact about the
environment, or break something — before you commit the fix.
*Result you should have:* one line below with date, what was learned, and the rule
that prevents recurrence. Newest LAST. Never delete or edit old lines.

- 2026-11-08 — The repo remote is SSH (`git@github.com:Esemby/JobCrawler.git`); the
  machine has no HTTPS credentials, so HTTPS pushes fail with "could not read
  Username". Rule: push via SSH; if a new environment lacks the key, hand the
  owner one setup step, never store credentials.
- 2026-11-08 — `template.ParseFS(files, "templates/status.html")` names templates
  by FILE NAME, and `ExecuteTemplate(w, "status", ...)` 500s: template not found.
  Rule: `ParseFS` → execute by file name unless templates `{{define}}` a name;
  every handler-template pair asserts rendering in a test (caught at 22:04).
- 2026-11-09 — Two test helpers named `postForm` collided inside one package
  (settings_test vs a new file). Rule: package-wide test helpers get
  check-listed before adding — grep for the name in the package first.
- 2026-11-09 — The first real crawl ran with the placeholder UA contact
  (decision 2.2 said placeholder only UNTIL the first real crawl). Rule:
  real-crawl code paths must assert contact-string != placeholder, or run
  them only after the address is filled in. RESOLVED 2026-11-09 same day:
  owner chose the repo issue tracker (decision 1.3).
- 2026-11-09 — After a session working-directory move, a file *write* through a
  relative path silently landed nowhere (the old-path content persisted), and a
  later build failed on stale symbols. Rule: after any directory change, write
  with absolute paths, or read back the file before compiling against it.
- 2026-11-09 — Twice today, an "append a block above X" edit used an existing
  entry's first line as the anchor and clobbered the rest of that entry —
  then lodged THIS SAME CLOBBER a third time the same day, on lessons.md itself,
  despite logging the rule. Rule: in append-only logs, anchor on the file's
  final line, never on an existing entry's first line; re-read after every
  edit.
- 2026-11-10 — Same clobber class, FOURTH time, on known-pitfalls.md: the
  new 2026-11-10 entry's edit consumed the first line of the 2026-11-09
  entry below it. Rule confirmed empirically: NEVER anchor an edit on an
  existing entry's first line. Anchor on the final line of the target
  block and re-emit it verbatim in newString.

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
