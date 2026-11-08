# ADR-0005: Unattended overnight operation

*When to read this:* reviewing what happened tonight, or proposing future unattended runs.
*Result you should have:* you know what was allowed to run alone, on which defaults,
and what still needs the owner in the morning.

## Status

Accepted (2026-11-08, owner tasking: "training and building process should work while
I sleep, 8 hours").

## Context

The owner went offline for ~8 hours. The brief's handover loop (`scorecard.md`)
depends on a human pacing it, so that loop is restructured for tonight.

## Decisions

1. **Toolchain without sudo:** no passwordless sudo on the Flagship, so Go
   1.27.3 installs to `~/tools/go` with symlinks in `~/.local/bin` (already on
   PATH). Nothing system-wide was touched. `gofmt` comes from the same tree.
2. **Successor model:** Ollama on this machine lacked `gpt-oss:20b`; a download
   was started (~14 GB). Until it lands, published model in `docs/small-model/`
   examples stays `gpt-oss:20b`, and any live smoke test uses the present
   `qwen3.5-tuned:latest` stand-in, clearly labeled in the scorecard.
3. **Building continues without the owner:** Phase 1 runs tonight from the
   explicit owner instruction. Phases beyond Phase 1 proceed only with their
   standard acceptance gates green (`make check`), one commit per phase, never
   force-pushed. Work that fails a gate three times is stopped and documented in
   `handoff.md` for the owner instead of expanding scope.
4. **Assumed defaults (no new owner input):** bind `:8090`; `sqlite` engine; UI
   no auth (ADR-0004); Greenhouse as first source (Phase 3); no MariaDB work
   (Phase 6); no AI review work beyond config stubs (Phase 7); dev runs may bind
   `127.0.0.1` when testing.
5. **Safety envelope (unchanged):** no secrets, no install of new system-wide
   software, no touching other services, no destructive git operations, commit
   discipline: small + Conventional Commits + `make check` green.

## Consequences

- The morning review owns: (a) verdicts on anything the background agents
  produced, (b) the scorecard follow-ups requiring a human-run handover session.
- Nothing here modifies the pre-decided ADRs; conflicts resolved in favor of
  ADR-0001/0002/0004 and the original brief.
