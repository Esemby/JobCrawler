# Config reference

*When to read this:* before adding, changing, or validating any setting, or when a
user asks what a setting does.
*Result you should have:* the table you add here is already updated for your change.

## Rules

- Every setting has: key, type, default, validation, help text, and a row below.
- Settings live in two layers:
  1. **Bootstrap** (`config.bootstrap.toml` + env overrides): storage engine choice
     and credentials, because this cannot live inside the DB it describes.
  2. **Runtime** (rows in the active store): everything else, editable in the UI,
     hot-reloaded where practical; if hot reload is impossible the UI shows
     "restart required".

## Bootstrap settings

| Key | Type | Default | Validation | Help |
|---|---|---|---|---|
| bootstrap.settings_path | string | `config.bootstrap.toml` | path exists | Where bootstrap config lives; overridden by env `JOBCRAWLER_CONFIG`. |
| bootstrap.db_engine | string | `sqlite` | one of: sqlite, mariadb | Which Store implementation to open. |
| bootstrap.db.dsn | string | `jobcrawler.db` (sqlite) | non-empty | Connection string for the chosen engine. Passwords go here; never logged, never shown clear. |

## Runtime settings

| Key | Type | Default | Validation | Help |
|---|---|---|---|---|
| system.bind_addr | string | `:8090` | valid host:port | HTTP bind for the web UI. LAN interface only. Port 8080 is FORBIDDEN (redirect.py lives there). |
| system.log_level | string | `info` | debug, info, warn, error | Log verbosity. |
| system.log_format | string | `text` | text, json | slog output format. |
| crawl.concurrency | int | `2` | 1-8 | Parallel fetches. Keep low on the Pi. |
| crawl.rate_per_host_ms | int | `2000` | ≥ 250ms | Min delay between requests to one host, plus jitter. |
| crawl.user_agent | string | jobcrawler (+contact) | must contain "/" | UA string; include a contact email per politeness. |
| sources.* | list | - | per-source | One entry per configured source; see `sources.md`. |
| filters.* | lists | empty | see filter playbook | Blacklist, no-gos, must-haves, tags, location rules. |
| review.enabled | bool | `false` | - | Stage 3 gate. System must work fully with this off. |
| review.endpoint | string | Flagship Ollama URL (http://<flagship-ip>:11434) | URL | Ollama runs on the Flagship (ADR-0004); Heart reaches it by IP, never localhost. |
| review.model | string | (gpt-oss:20b) | - | Model name at the endpoint. |
| review.daily_cap | int | `50` | ≥ 0 | Max AI review calls per day. 0 = unlimited. |

Placeholder note: until Phase 2 lands, values above are targets, not implemented
behaviour. Do not quote this file as fact in commit messages before then.
