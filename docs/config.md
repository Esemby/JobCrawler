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

Loaded in `internal/config`: defaults first, then `config.bootstrap.toml`,
then environment overrides. The file may be absent on a fresh install
(sqlite defaults apply); an explicitly configured missing path is an error.

| Env key | Type | Default | Validation | Help |
|---|---|---|---|---|
| JOBCRAWLER_BIND_ADDR | string | `:8090` | host:port; port 8080 FORBIDDEN | HTTP bind for the web UI. LAN interface only. Port 8080 is reserved for redirect.py on the Heart. |
| JOBCRAWLER_LOG_LEVEL | string | `info` | debug, info, warn, error | Log verbosity. |
| JOBCRAWLER_LOG_FORMAT | string | `text` | text, json | slog output format (stdout only; no chatty disk logs). |
| JOBCRAWLER_CONFIG | string | `config.bootstrap.toml` | file must exist | Path of the bootstrap file; adapter, not a setting in the file itself. |
| JOBCRAWLER_DB_ENGINE | string | `sqlite` | sqlite, mariadb | Overrides the engine from the bootstrap file (ADR-0002). MariaDB not implemented until Phase 6. |
| JOBCRAWLER_DB_DSN | string | `jobcrawler.db` (sqlite) | non-empty | Overrides the DSN from the bootstrap file. Never logged, never shown clear. |

Bootstrap-file settings (`config.bootstrap.toml`, keys live under `[db]`;
ADR-0002, loaded in `internal/config`):

| Key | Type | Default | Validation | Help |
|---|---|---|---|---|
| db.engine | string | `sqlite` | sqlite, mariadb | Which Store implementation to open. |
| db.dsn | string | `jobcrawler.db` (sqlite) | non-empty; non-empty required for mariadb | Connection string for the chosen engine. Passwords go here; never logged, never shown clear, never committed. |

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
