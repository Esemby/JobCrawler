# Glossary

*When to read this:* a doc or code comment uses a word you are not 100% sure of.
*Result you should have:* the exact meaning; if you had to guess, the word goes here
with a definition after your change.

| Term | Definition |
|---|---|
| Heart | The Raspberry Pi 4B production host (Esemby-Pi). |
| Flagship | The development machine that cross-compiles for the Heart. |
| Job | One parsed job offer with fields: company, title, location, url, description, tags, timestamps. |
| Source | A configured crawl origin with a type (ats, rss, jsonld, html), schedule, and credentials if any. |
| Adapted source | A Go `Source` implementation fetching listings from one source type. |
| Fingerprint | Stable dedupe key: normalized company + title + location + canonical URL. |
| Canonical URL | The job's deduplicated-down URL (strips tracking params, prefers absolute). |
| Verdict | Output of the filter pipeline: accepted/rejected + human-readable reason chain. |
| Reason | One line of explanation stored for every accept/reject decision. |
| Stage 1 | Hard mechanical rules: blacklist, no-gos, must-haves, tags, location. Deterministic. |
| Stage 2 | Score = weighted tag match + recency + source trust; produces breakdown. |
| Stage 3 | Optional AI review via `Reviewer` interface (Ollama first). Off by default. |
| Verdict cache | Stored AI verdicts reused for identical fingerprints to cap API calls. |
| Seen flags | Owner state on a job: new, seen, applied, hidden. UI-controlled. |
| Disappeared | A job no longer returned by its source but tracked (first-seen/last-seen kept). |
| Hot reload | Settings change taking effect without process restart. |
| Bootstrap config | The file that must exist before the store does: engine + credentials. |
| Green | `make check` passes. The only definition of "working". |
| Handover test | A small-model task run unaided, tracked in `docs/small-model/scorecard.md`. |
