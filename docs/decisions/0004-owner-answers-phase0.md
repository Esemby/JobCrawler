# ADR-0004: Owner answers from Phase 0

*When to read this:* before Phase 1, Phase 3, or anything touching filter, auth, or
review; whenever a phase-start doubt needs an environment fact.
*Result you should have:* these six answers applied, never re-asked.

## Status

Accepted (owner-confirmed, 2026-11-08).

## Context

Brief section 10 listed six open questions gating Phase 1+. Owner answered all six.

## Decisions

1. **Heart RAM: ~3.9 GB** (`MemTotal: 3871632 kB`), `MemAvailable: 2156364 kB`
   observed at Phase 0. Budget unchanged: < 100 MB idle, < 250 MB crawl peak. The
   Heart shares itself with other services — do not tune beyond those limits.
2. **Region/languages:** owner is job-hunting in the **EU + Norway**, languages
   **German + English**. The filter layer must treat country/region/language/
   must-have keywords as *configurable runtime settings*, never hard-coded. No
   region-specific logic may assume a single country.
3. **Name & license:** project name stays `jobcrawler`; license **MIT** (add
   LICENSE file in Phase 1).
4. **First source: Greenhouse** public ATS feed (free JSON, no auth). See
   `docs/sources.md` register.
5. **UI:** LAN-only from day one, **no login in Phase 1**; auth arrives per brief
   Phase 8, but the design must not hard-code "single user": the owner wants
   **multiple user accounts later (friends)**. Consequence: user-scoped state
   (seen/applied/hidden flags) must be designed as per-user from the start, since
   retrofitting identity onto shared job flags is expensive.
6. **Ollama endpoint:** runs on the Flagship, reachable from the Heart by IP.
   Configure review.endpoint as URL, not localhost; document LAN firewall assumptions
   in `docs/deploy.md` when Phase 7 lands.

## Consequences

- `docs/config.md`, `docs/sources.md`, `docs/deploy.md` updated the same day.
- Phase 1 adds a MIT `LICENSE` file alongside the skeleton.
- Phase 8 auth task must support N users on LAN, not a single password.
- All six answers are settled; reopening any needs the owner and a new ADR.
