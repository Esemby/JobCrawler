# Sources

*When to read this:* adding or reviewing a crawl source, or checking what a source is
allowed to do.
*Result you should have:* your source is listed here with its method and ToS note,
or you know it must NOT be implemented.

## Priority order (from the brief, section 4.1)

1. Public ATS endpoints/feeds: Greenhouse, Lever, Ashby, Workable, Personio,
   SmartRecruiters (public JSON/XML).
2. RSS/Atom feeds, sitemaps, `schema.org/JobPosting` JSON-LD in pages.
3. Free official APIs of job boards.
4. Plain HTML parsing of career pages — last resort.

Never: login-walled content, CAPTCHA/bot-protection bypass, or sources whose terms
forbid automated access.

## Source Register

| Source | Method | Access notes / ToS | Status |
|---|---|---|---|
| Greenhouse | Public JSON endpoint `boards-api.greenhouse.io` | Public, unauthenticated board JSON; verify ToS before implementing (ADR-0004: first source) | planned (Phase 3) |
| Lever | Public JSON postings API | Public, unauthenticated; same ToS check rule | planned, after Greenhouse |

## Politeness contract (every source, no exceptions)

- Respect robots.txt (cache it, re-read at most daily).
- Honest User-Agent with contact field (see `config.md`: crawl.user_agent).
- Per-host rate limit with jitter; global politeness ceiling.
- Conditional requests: ETag / If-Modified-Since; exponential backoff on errors.
- Deduplicate via fingerprint (normalized company + title + location + canonical URL).

After adding a working source, add a row above and commit together with the adapter.
If a source's terms forbid automated access, write the row with status `rejected` and
the reason — a documented "no" prevents repeated relitigating.
