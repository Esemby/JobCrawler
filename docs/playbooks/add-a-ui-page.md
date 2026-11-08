# Playbook: add a UI page

*When to read this:* a task says "add page/section Y" for the web UI.
*Result you should have:* a server-rendered page with htmx partials reachable from
nav, tested, documented. STATUS: template, refine after Phase 2.

## Checklist

1. UI lives in `internal/web/`: handlers + `templates/` + `static/` (embedded).
   Server-rendered `html/template` ONLY; htmx vendored in `static/`, no CDN links.
2. Route: add to the mux registry in `internal/web/` using Go 1.22 pattern syntax
   (`GET /jobs/{id}`). One line per route, grouped by page, alphabetical.
3. Handler < 50 lines; all DB access via the `Store` interface; `context.Context`
   passed from the http request. NO business logic in handlers — call into the
   right internal package instead.
4. Template file named after the page (`jobs.html`), one base layout, one partial
   per htmx-updated block. No inline scripts; htmx attributes only.
5. Test: `httptest` request → assert status code, header, and a distinctive string
   in the body. One test per page.
6. Add nav link if the page is a top-level section.
7. `make check` → green.
8. Commit: `feat(web): <page name> page`.

## Rules

- Never bind to 0.0.0.0 or add internet-facing auth bypass routes. LAN only.
- Forms that change settings must validate server-side and show errors inline.
- Password fields: never render the stored value; render `"•••••"` placeholder.
- If page needs new JS beyond htmx: STOP. That's an ADR + owner question.

## If this fails

- Template compile error at runtime: add a template parse test that runs on
  startup paths (parse all templates in TestMain for `internal/web`).
