# AGENTS.md

jobcrawler: a polite, self-hosted job crawler. Mechanical filtering first; AI review
optional on top. Target host: Raspberry Pi 4B (Ubuntu, arm64). Dev host: big Linux/Win box.

## The 10 golden rules

1. Follow the loop in `docs/small-model/how-to-work.md` every session. First action:
   read `docs/small-model/handoff.md`.
2. One small task per session. Finish it: test, docs, handoff, commit.
3. `make check` must be green before every commit. If a test fails, fix it or revert
   your last change. Sees `docs/playbooks/fix-a-failing-test.md`.
4. Go standard library first. Adding any dependency requires a written ADR in
   `docs/decisions/` BEFORE the code. No exceptions.
5. Never touch port 8080 or `redirect.py` on the Heart. This service defaults to 8090.
6. The web UI binds LAN-only. Never add a way to expose it to the internet.
7. Never put secrets in git. Never print secrets. Never force-push. Never run
   destructive commands without explicit owner confirmation.
8. Legal, polite crawling only: robots.txt, honest User-Agent, rate limits,
   no login-walls, no CAPTCHA bypass. Sources that forbid crawling are not implemented.
9. Every instruction file: < 150 lines target, hard cap 250. Imperative, numbered
   steps, exact commands. One way to do each thing.
10. Append mistakes and their prevention rules to `docs/lessons.md`. Fix docs, not
    memories.

## Doc index

Where to look, by need:

| Need | File |
|---|---|
| Start of session loop | `docs/small-model/how-to-work.md`, `docs/small-model/handoff.md` |
| What the project is + golden rules rationale | `docs/00-start-here.md` |
| Packages, interfaces, data flow | `docs/architecture.md` |
| Settings (key, type, default) | `docs/config.md` |
| Sources and ToS notes | `docs/sources.md` |
| Running and writing tests | `docs/testing.md` |
| arm64 build, Pi install, systemd | `docs/deploy.md` |
| Domain words (fingerprint, verdict...) | `docs/glossary.md` |
| Mistakes log (append-only) | `docs/lessons.md` |
| Why decisions were made | `docs/decisions/NNNN-title.md` |
| Original requirements | `docs/original-brief.md` (read-only history) |

Playbooks (follow them; do not improvise):

| Task | Playbook |
|---|---|
| New source type/adapter | `docs/playbooks/add-a-source.md` |
| New setting | `docs/playbooks/add-a-setting.md` |
| New filter rule | `docs/playbooks/add-a-filter-rule.md` |
| New UI page | `docs/playbooks/add-a-ui-page.md` |
| New DB migration | `docs/playbooks/add-a-migration.md` |
| Write a test | `docs/playbooks/write-a-test.md` |
| Fix failing test | `docs/playbooks/fix-a-failing-test.md` |
| Build + install on Pi | `docs/playbooks/release-a-build.md` |
| Debug on the Heart | `docs/playbooks/debug-crawler-on-the-heart.md` |
| Update docs after change | `docs/playbooks/update-docs-after-change.md` |

Small-model material: `docs/small-model/` (how-to-work, context-budget, handoff,
known-pitfalls, task-template, scorecard).

When in doubt: smaller step, test, commit, document.
