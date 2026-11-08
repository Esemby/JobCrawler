# PROJECT BRIEF: Job Crawler (working title: `jobcrawler`)

> **Read this first.** You are an OpenCode agent. This file is your starting brief.
> Your FIRST task is *not* to write product code. Your first task is described in
> section 9 ("Phase 0: Bootstrap the instruction system"). Follow the phases in order.

---

## 1. Mission

Build a self-hosted web crawler that finds job offers sensibly. Mechanical filtering
comes first; AI review is optional and layered on top. Everything the owner needs to
tune must be configurable from a browser inside the LAN.

There is a second, nested mission, equally important:

> **You (a capable cloud model, usually GLM-5.3-Flash Free) must write instruction files
> good enough that a small locally hosted model (gpt-oss:20b via Ollama) can become the
> sole builder and maintainer of this project, using nothing but those files.**

The project is the product. The instruction files are also a product. Judge your own
work by whether the small model can continue without you.

## 2. People and environment

- **Owner:** hobbyist/professional developer, runs everything on Linux unless said otherwise.
  Owner will be a *user* of this tool (job hunting).
- **Target host, "the Heart" (Esemby-Pi):** Raspberry Pi 4B, Ubuntu Server, aarch64,
  128 GB SSD (upgrade to 512 GB planned). Runs 24/7 (99.999% intent) and hosts many
  other things. **Treat CPU, RAM and SSD wear as scarce.** Do not assume more than
  ~200 MB RAM for this service. Confirm real RAM with the owner in Phase 0.
- **Dev machine, "the Flagship":** Ryzen 9 9950X, RX 7900 XT, 64 GB RAM, Windows 11 / Ubuntu
  Studio. This is where you and gpt-oss:20b run. Cross-compile for the Heart from here.
- **Database:** a MariaDB exists on a server inside the VPN on the owner's father's LAN.
  It is *remote and may be unreachable*. See ADR-0002 below.
- **Already on the Heart:** a Python `redirect.py` (http.server, **port 8080**) serving
  esemby.org redirects. **Never modify it, never reuse port 8080.** Default this project
  to port **8090** (configurable).
- Git hosting: GitHub. The repository already exists or is being created by the owner.
  Never force-push. Never commit secrets.

## 3. Decision: language and stack (ADR-0001, pre-decided)

**Language: Go (current stable release).**

Why, in the order that matters for this project:

1. **Service speed and footprint on a Pi 4:** compiled, small resident memory, fast start,
   goroutines make polite concurrent crawling trivial.
2. **Deployment:** `GOOS=linux GOARCH=arm64 go build` yields ONE static binary. No runtime,
   no virtualenv, no node_modules on the Heart. systemd unit + binary = done.
3. **Batteries included:** `net/http`, `html/template`, `embed`, `database/sql`, `log/slog`,
   `context`, `testing` are all in the standard library. Fewer dependencies means fewer
   things for a small model to get wrong and fewer supply-chain risks.
4. **Small-model friendliness (decisive):** Go is small, explicit, and has one idiomatic
   way to do most things. The compiler, `go vet` and `go test` give fast, unambiguous
   feedback that a 20B model can act on. Python's dynamic typing and Rust's borrow
   checker both produce more failure modes for weak models.
5. **Development speed:** fast compile times, excellent tooling, easy refactors.

Rejected alternatives (do not relitigate without a benchmark and a written ADR):

| Option | Why not |
|---|---|
| Python | Slowest on a Pi, dependency/venv drift, runtime type errors |
| Rust | Best raw speed, but slow iteration and hard for small models to repair |
| Node/TS | Heavier runtime, build chain, dependency churn |

**Stack defaults** (change only via ADR):

- Web UI: server-rendered `html/template` + **htmx** (vendored as a static file, no CDN,
  no JS build step). Plain CSS. Must work without internet.
- HTTP routing: standard library `http.ServeMux` (Go 1.22+ patterns).
- HTML parsing: `golang.org/x/net/html` or `github.com/PuerkitoBio/goquery`.
- Crawler: own thin layer on `net/http` (politeness, robots.txt, per-host rate limits).
  `colly` is allowed if it genuinely saves code; justify in an ADR.
- Storage: see ADR-0002. Pure-Go SQLite driver `modernc.org/sqlite` (no CGO, so
  cross-compiling stays trivial). MariaDB via `github.com/go-sql-driver/mysql`.
- Migrations: plain numbered `.sql` files embedded with `embed`, applied by a tiny runner.
- Config bootstrap: one small TOML/YAML file (`config.bootstrap.*`) + env var overrides.
- Logging: `log/slog`, JSON or text selectable.
- Tests: standard `testing`, table-driven, `httptest` for crawler fixtures (saved HTML).
- Lint/format gate: `gofmt`, `go vet`, `staticcheck` (if installable).

## 4. Product requirements

### 4.1 Crawler behaviour ("sensibly")

- **Prefer structured sources over scraping HTML**, in this order:
  1. Public ATS endpoints and feeds (Greenhouse, Lever, Ashby, Workable, Personio,
     SmartRecruiters and similar public JSON/XML), RSS/Atom feeds, sitemaps,
     `schema.org/JobPosting` JSON-LD embedded in pages.
  2. Official APIs of job boards where free access exists.
  3. Plain HTML parsing of company career pages.
  4. Never log-in-walled content. Never bypass bot protection or CAPTCHAs.
- Respect `robots.txt`, send an honest `User-Agent` with a contact field (configurable),
  per-host rate limiting with jitter, conditional requests (`ETag`/`If-Modified-Since`),
  exponential backoff, and a global politeness ceiling.
- Check each source's terms of use; maintain a `docs/sources.md` listing each source,
  its access method, and legal/ToS notes. Sources that forbid automated access are
  not implemented.
- Deduplicate jobs (stable fingerprint: normalized company + title + location +
  canonical URL). Track first-seen, last-seen, and "disappeared" state.
- Crawl schedule is configurable per source.

### 4.2 Filtering pipeline (mechanical first)

Stage 1, hard rules (cheap, deterministic, explainable). Every rejection stores a *reason*:

- **Company blacklist** (exact and pattern).
- **No-gos:** keywords/phrases that disqualify (title or description).
- **Must-haves:** keywords/phrases required (supports AND / OR groups).
- **Job tags:** include tags with weights, used for scoring and filtering.
- Location / remote / language / salary range where data exists.

Stage 2, scoring: weighted tag match, recency, source trust. Produces a score + a
human-readable breakdown shown in the UI.

Stage 3, **optional AI review**: disabled by default. Provider-agnostic interface
(local Ollama endpoint first; others later). It receives only jobs that survived stages
1-2, returns a structured verdict + short reasoning, results are cached, and every call
is logged. The system must work fully without it.

### 4.3 Web configuration UI (LAN only)

Everything configurable from the browser, grouped in categories:

- **System:** bind address/port, crawl concurrency, rate limits, User-Agent contact,
  log level, data retention, schedule.
- **Database access:** engine (SQLite default / MariaDB), host, port, name, user,
  password, "test connection" button. Password never shown back in clear text.
- **Sources:** add/edit/disable sources, test-fetch one source, see last result.
- **Filters:** job tags, company blacklist, must-haves, no-gos, location rules.
- **AI review:** on/off, endpoint, model, prompt template, daily call cap.
- **Jobs:** browse, search, mark seen/applied/hidden, see why a job was accepted/rejected.
- **Status:** last crawl, errors, queue, DB health, version.

Rules for the UI layer:

- Binds to the LAN interface only by default; **never** exposes itself to the internet.
  Optional single-user login (argon2/bcrypt) with CSRF protection. Document that
  reverse-proxying to the internet is unsupported.
- Settings changes apply **without restart** wherever practical (hot reload), otherwise
  the UI says clearly "restart required".
- Every setting has: default, validation, help text, and an entry in `docs/config.md`.
- Import/export of filter lists (JSON/YAML) for backup and sharing.

### 4.4 Non-functional

- Idle RAM target on Pi: < 100 MB. Crawl peak: < 250 MB. Report actual measurements.
- Minimise SSD writes: batch writes, no chatty logging to disk, WAL mode for SQLite,
  retention policy for old raw pages (default: don't store raw HTML).
- Survive power loss and DB outages: crash-safe, retry with backoff, never lose config.
- Graceful shutdown on SIGTERM. Ship a systemd unit file and install docs.
- Zero telemetry.

## 5. Architecture (starting point, evolve via ADRs)

```
cmd/jobcrawler/        main.go: wiring only, no logic
internal/config/       bootstrap config + runtime settings (typed, validated)
internal/store/        Store interface + sqlite/ and mariadb/ implementations
internal/store/migrations/   numbered .sql files
internal/crawler/      scheduler, fetcher, politeness, robots
internal/sources/      one package/file per source type (ats/, rss/, jsonld/, html/)
internal/filter/       rules engine, scoring, explain output
internal/review/       Reviewer interface + ollama implementation + cache
internal/web/          handlers, templates/, static/ (embedded), htmx partials
internal/version/      build-time version info
docs/                  instruction + decision + reference docs (see section 8)
testdata/              saved HTML/JSON fixtures per source
```

Principles: small packages, interfaces at boundaries (`Store`, `Source`, `Reviewer`),
dependency direction always inward toward `filter`/domain types, no global state, no
package-level mutable variables, `context.Context` everywhere.

### ADR-0002: storage (pre-decided)

The Heart must stay up even when the MariaDB behind the VPN is down. Therefore:

- **Default engine: local SQLite** (WAL mode) on the Heart. Zero dependencies, survives
  VPN outages.
- **MariaDB is a supported, optional engine** behind the same `Store` interface, selectable
  in the UI ("Database access"). Needed when the owner wants a shared DB.
- The bootstrap file (or a tiny local SQLite `bootstrap.db`) must always hold *which*
  engine to use and its credentials, because the DB settings cannot live inside the
  database they describe.
- Runtime settings (filters, sources, etc.) live in the active store. Provide an
  export/import so they can migrate between engines.

## 6. Roadmap with acceptance criteria

Do not start a phase before the previous phase's checks pass. After each phase: update
docs, commit, and run the **handover test** (section 9.4).

| Phase | Goal | Done when |
|---|---|---|
| 0 | Bootstrap instruction system | `AGENTS.md` + `docs/` exist, handover test #0 passes |
| 1 | Skeleton | `go build`, `go test ./...`, `make run` work; `/healthz` and a status page render; systemd unit + arm64 build documented |
| 2 | Config + store | SQLite store, migrations, settings CRUD, web UI for System + Database access pages, hot reload |
| 3 | First crawl | One structured source (e.g. Greenhouse/Lever) end to end: fetch, parse, dedupe, store, list in UI. Politeness + robots tested with fixtures |
| 4 | Filter engine | Tags, blacklist, must-haves, no-gos, scoring, "why" explanations, UI pages for each |
| 5 | More sources | RSS, JSON-LD, generic HTML adapter; `docs/adding-a-source.md` proven by the small model adding a source alone |
| 6 | MariaDB | Second `Store` implementation passing the same shared test suite |
| 7 | Optional AI review | Ollama reviewer, cache, caps, UI toggles; project works identically with it off |
| 8 | Hardening | Auth, CSRF, backups, retention, resource measurements on the Pi, README for outsiders |

Each task inside a phase must be small enough to fit in one short session of the small
model (see 9.2).

## 7. Working agreements with the human owner

- The owner's preference: **troubleshooting and setup instructions one step at a time**,
  not as a full dump. Wait for the result of a step before giving the next.
- Ask **at most one question at a time**, and only when blocked. Otherwise state your
  assumption, proceed, and record it in `docs/decisions/`.
- Assume Linux. Name Windows explicitly where it differs (the Flagship dual-boots).
- Commit small, commit often. Conventional Commits (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`).
- Never run destructive commands (`rm -rf`, `git reset --hard`, force-push, DB drops)
  without explicit owner confirmation.
- Never touch other services on the Heart. Never read or print secrets.

## 8. Required documentation layout (you create and maintain this)

```
AGENTS.md                  <= 100 lines. Entry point. Rules + index. Auto-loaded by OpenCode.
docs/
  00-start-here.md         What the project is, where things are, the 10 golden rules
  architecture.md          Packages, interfaces, data flow (one diagram in ASCII)
  config.md                Every setting: key, type, default, validation, help text
  sources.md               Each source: method, ToS notes, status
  testing.md               How to run tests, how fixtures work, what "green" means
  deploy.md                Build for arm64, install, systemd, upgrade, rollback
  glossary.md              Domain terms (job, source, fingerprint, verdict, ...)
  decisions/               ADRs: NNNN-title.md (context, decision, consequences)
  playbooks/               One file per recurring task (see below)
  small-model/             Material specific to gpt-oss:20b (see 9)
  lessons.md               Append-only log of mistakes and the rule that prevents them
```

Required playbooks (each <= 80 lines, numbered checklist, exact commands, expected output,
"if this fails, do X"):

- `add-a-source.md`, `add-a-setting.md`, `add-a-filter-rule.md`, `add-a-ui-page.md`,
  `add-a-migration.md`, `write-a-test.md`, `fix-a-failing-test.md`, `release-a-build.md`,
  `debug-crawler-on-the-heart.md`, `update-docs-after-change.md`.

## 9. THE NESTED CHALLENGE: training the successor purely with instruction files

### 9.1 Phase 0 tasks (do these before any product code)

1. Split this brief into the layout from section 8. Keep *this file* as
   `docs/original-brief.md` (read-only history).
2. Write `AGENTS.md` as a short index and rules list. It must fit comfortably in the
   small model's context alongside a task.
3. Write the playbooks as templates, then **fill them with real content as you build**.
   A playbook is only finished after it was actually used to do the task once.
4. Create `docs/small-model/` containing:
   - `how-to-work.md`: the loop the small model must follow every session
     (read AGENTS.md → pick ONE task → read its playbook → plan in <= 5 lines → implement →
     test → update docs → commit → write handoff note).
   - `context-budget.md`: how to stay inside the context window (read files by range,
     never `cat` big files, prefer `grep`/`rg` + targeted views, finish and commit before
     context fills, resume from `handoff.md`).
   - `handoff.md`: overwritten at the end of every session. Current state, next task,
     blockers, commands to verify.
   - `known-pitfalls.md`: model-specific failure patterns found during handover tests,
     each with a concrete counter-instruction.
   - `task-template.md`: how tasks are written so the small model can execute them.
5. Create a `Makefile` (or `justfile`) with the same few verbs forever:
   `make build`, `make test`, `make lint`, `make run`, `make build-arm64`, `make check`
   (= format + vet + lint + test). The small model should only need these verbs.

### 9.2 Writing rules for every instruction file (they are for a 20B model)

- **Short.** One topic per file. Target < 150 lines, hard cap 250. Split instead of growing.
- **Imperative and concrete.** "Run `make check`." not "Make sure the code is good."
- **Exact paths, exact commands, exact expected output.**
- **Checklists over prose.** Numbered steps, one action each.
- **Show, don't describe:** include a minimal worked example for every pattern
  (e.g. a complete tiny `Source` implementation and its test).
- **One way to do each thing.** Do not offer options unless a decision rule is given.
- **Explicit stop conditions:** when to stop, when to ask the human, when to revert.
- **Negative rules paired with the correct alternative** ("Do not add a dependency.
  Instead use the standard library; if impossible, write an ADR and ask the owner").
- No vague words (appropriate, robust, clean, as needed) without a measurable definition.
- Every doc starts with a 2-line "When to read this" and "Result you should have".
- Link by relative path; keep an index in `AGENTS.md` so nothing is orphaned.

### 9.3 Code rules that make the project maintainable by a small model

- Boring, idiomatic Go. No generics unless they remove real duplication. No reflection tricks.
  No code generation frameworks. No clever abstractions.
- Files < 300 lines, functions < 50 lines where reasonable. Descriptive names.
- Every package has a `doc.go` with 3-6 lines of purpose and boundaries.
- Every exported function has a one-line comment stating *what* and *when it errors*.
- Table-driven tests with fixtures in `testdata/`. A change is not done until
  `make check` is green.
- Errors wrapped with context (`fmt.Errorf("fetch %s: %w", url, err)`); no panics in
  library code; no silent `_ = err`.
- Prefer adding a new small file over editing a large one.
- Every new behaviour gets (a) a test, (b) a docs line, (c) a changelog line.

### 9.4 The handover test (the training loop)

You cannot judge the instruction files by reading them yourself. Test them:

1. Pick a small, real task (see `task-template.md`) such as "add a setting",
   "add a source adapter from a fixture", "fix this seeded failing test".
2. Ask the owner to run it with **gpt-oss:20b in OpenCode** from a fresh session with
   *only the repository* (no extra prompting from you or the owner beyond the task text).
3. The owner pastes back the transcript or a short summary: where did it get lost?
4. **Fix the docs, not the model.** For each failure, make one change:
   a missing example, a clearer step, a split file, a new pitfall entry, or a new rule in
   `lessons.md`. Record it in `docs/small-model/known-pitfalls.md`.
5. Re-run the *same* task. Repeat until it passes without human help.
6. Then raise the difficulty: bigger task → multi-file task → whole-feature task.

Track progress in `docs/small-model/scorecard.md`:

| Date | Task | Model | Passed unaided? | Failure cause | Doc fix made |
|---|---|---|---|---|---|

**Graduation criteria for "lone builder":** the small model completes, unaided and with
`make check` green, (1) a new setting with UI, (2) a new source adapter, (3) a new
filter rule type, (4) a bug fix from a failing test, (5) a docs-only update, and (6) a
release build for the Pi, each at least twice in a row.

### 9.5 After graduation

- The small model owns routine work; you (the cloud model) are only for ADR-level
  decisions and for repairing instruction files when the scorecard shows regressions.
- Keep `lessons.md` and `known-pitfalls.md` growing. Prune and merge them monthly so
  they stay short.
- When the small model hits something outside the playbooks, it must write a new
  playbook as part of solving it. That is how the system keeps training itself.

## 10. Open questions for the owner (ask ONE at a time, in Phase 0)

1. What is the exact RAM of the Heart (2/4/8 GB) and what else runs on it?
2. Which country/region/languages and job fields (title keywords) matter first?
3. Final project name and license for the GitHub repo?
4. Which first source should we build against (a company career page or ATS the owner
   actually uses)?
5. Should the web UI have login from day one, or LAN-only without auth at first?
6. Is the Ollama endpoint reachable from the Heart, or only on the Flagship?

Record answers in `docs/decisions/` as they come. If unanswered, pick the stated default
and note the assumption.

## 11. Guardrails recap

- Legal and polite crawling only. No CAPTCHA bypass, no login-wall scraping, honour `robots.txt`.
- LAN-only exposure by default. No telemetry. No secrets in git.
- Do not touch `redirect.py` or port 8080.
- Do not add dependencies casually. Standard library first.
- Do not relitigate ADR-0001 (Go) or ADR-0002 (SQLite default, MariaDB optional) without
  measurements and the owner's approval.
- If in doubt: smaller step, test, commit, document.

**Start now with Phase 0, step 1.**
