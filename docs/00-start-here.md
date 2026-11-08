# 00 — Start here

*When to read this:* every model's first read in a fresh session.
*Result you should have:* you know what jobcrawler is, where things live, and which
file to open next.

## What this project is

A self-hosted web crawler that finds job offers sensibly:

1. Pull from structured sources first (ATS feeds: Greenhouse, Lever, ...), then RSS,
   sitemaps, JSON-LD. Scraping plain HTML is the last resort.
2. Filter mechanically first (blacklists, no-gos, must-haves, tags, scoring). Every
   rejection stores a human-readable reason.
3. Optional AI review (local Ollama) only for jobs that survived mechanical filtering.
4. Everything tunable from a browser on the LAN. UI binds LAN-only, default port 8090.

Runs 24/7 on a Raspberry Pi 4B ("the Heart") with very scarce RAM (~100 MB idle target).

## Where things live

- Production layout: `cmd/jobcrawler/`, `internal/…`, `docs/`, `testdata/`.
  Full tree and data flow: `docs/architecture.md`.
- Roadmap and phase gates: `docs/original-brief.md` section 6. Never start a phase
  before the previous one's checks pass.
- Decisions history: `docs/decisions/`. Read ADR-0001 (Go) and ADR-0002 (storage)
  before proposing anything that touches those areas.

## Golden rules (summary)

Full numbered list is in `../AGENTS.md`. The ones that end arguments:

- Standard library first; a dependency needs an ADR before code.
- `make check` green before every commit.
- Polite and legal crawling only; skip sources that forbid it.
- Never touch port 8080 or `redirect.py` on the Heart; default port 8090.
- One way to do things: playbooks over improvisation.

## Next step

Open `docs/small-model/how-to-work.md` and follow it literally.
