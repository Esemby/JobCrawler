# Playbook: debug the crawler on the Heart

*When to read this:* the crawler misbehaves on the Pi (no crawls, errors, memory).
*Result you should have:* a diagnosis and a small fix, or a handoff note asking the
owner. STATUS: template; refine after Phase 3 when the crawler exists.

## Checklist

1. Confirm the service state first, always:
   - `systemctl status jobcrawler`
   - `journalctl -u jobcrawler -n 100 --no-pager`
2. Check the UI status page `http://<pi>:8090/` → last crawl time, queue, errors.
   The UI is read-only truth about the crawl state; do not guess.
3. Confirm the port did not get broken: `sudo ss -tlnp | grep -E '8090|8080'`.
   8080 = redirect.py (other service). If 8080 appears in OUR logs, it's a bug in
   our config default — fix our default, never their service.
4. RAM: `systemctl show jobcrawler -p MemoryCurrent -p MemoryMax`.
   Peak should stay < 250 MB. If higher: check concurrency and crawl window
   settings before touching code.
5. DB: is SQLite file growing? Is WAL huge? Is MariaDB unreachable (expected to be
   tolerated — ADR-0002)? Report numbers, don't speculate.
6. Reproduce the smallest case: one source, one crawl, from the UI "test-fetch
   one source" button (when it exists) rather than full runs.
7. Record the diagnosis + fix in `docs/lessons.md` (one line) and handoff.

## If this fails

- No access to the Pi from where you are: hand the exact commands to the owner,
  one step at a time (brief section 7 rule) and wait for output between steps.
- Suspect politeness/robots bugs (crawl hammering a host): STOP the crawler
  (`systemctl stop jobcrawler`) before diagnosing more. Being a nuisance is worse
  than being offline.
