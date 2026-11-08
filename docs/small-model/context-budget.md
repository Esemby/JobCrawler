# Context budget

*When to read this:* every session, before reading any file larger than 100 lines.
*Result you should have:* you reach the commit with context to spare.

## Rules

1. Read by range, never whole big files:
   - `sed -n '1,120p' internal/filter/score.go`
   - To find first: `grep -rn "Fingerprint" internal/`
2. Never `cat`, `tail -f`, or open raw HTML fixtures — fixtures may be megabytes.
   `grep` them if you must know their shape; `wc -c` first.
3. The read order per session is fixed and short (see `how-to-work.md`):
   AGENTS.md → handoff.md → ONE playbook → grep → targeted ranges.
4. Do not re-read docs you already have in context this session.
5. If context feels > half gone and the task is half done: commit working state
   (even as `wip:` ONLY if `make check` is green), rewrite handoff.md, stop, and
   let the next session resume from handoff.md. Never rush a half-tested commit.
6. One grep with a good pattern beats five greps. Include the path prefix
   (`internal/filter/`) to keep result size small.
7. Tests you write are read twice by you; keep them short, else split the file.
8. Never print a full error dump or a long log into the conversation; pipe to a
   file, grep that file.

## Why

The model's context window is the scarce resource, like the Pi's RAM. Running out
mid-task is exactly as bad as the Heart running out of memory: losing work.
