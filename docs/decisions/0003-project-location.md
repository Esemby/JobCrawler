# ADR-0003: Project location and repo bootstrap

*When to read this:* anyone wonders why the repo lives where it lives.
*Result you should have:* no confusion about moving or renaming the project root.

## Status

Accepted (unblocker, cheap to reverse).

## Context

The brief arrived as `~/Downloads/PROJECT_BRIEF.md` while the agent session ran in
`/home/esemby/Documents/Default Project`, which contains unrelated work (twitch-api).
The brief states the GitHub repo "already exists or is being created by the owner"
but gives no local path. Building without a local root is not possible.

## Decision

Create the project as a self-contained git repo at
`/home/esemby/Documents/Default Project/jobcrawler` — a sibling of the unrelated
work, new `git init`, no remote attached yet. The owner moves/renames the directory
whenever they want; nothing in the tooling depends on the absolute path.

## Consequences

- All docs use relative paths, so a move is zero-cost (this is a requirement).
- Remote GitHub URL is added later by the owner; until then keep committing
  locally (small, often). Never force-push once a remote exists.
- Recorded as an assumption per brief section 7 ("state your assumption, proceed,
  record it"). Supersede this ADR by writing 0004 if the location changes.
