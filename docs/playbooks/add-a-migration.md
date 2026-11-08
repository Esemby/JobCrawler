# Playbook: add a migration

*When to read this:* a task says "add table/column X" or "schema change needed".
*Result you should have:* numbered migration, applied cleanly, both engines ready.
STATUS: template, refine after Phase 2.

## Checklist

1. Files: `internal/store/migrations/NNNN_description.sql` — next free number.
   NEVER edit an applied migration. New file = new truth.
2. Write plain SQL. One concern per file. If it must touch existing rows, write
   UPDATE statements explicitly; no stored procedures, no triggers.
3. Engine dialect: SQLite is the default; write standard SQL that also runs on
   MariaDB where possible. If a difference is unavoidable, split the file into
   `NNNN_desc.sqlite.sql` and `NNNN_desc.mysql.sql` — the runner picks by suffix.
4. Add the migration runner test: fresh DB → all migrations apply in order → apply
   twice is a no-op (idempotence check at the bookkeeping level).
5. Update the Store implementation + its interface test for the new schema.
6. Update `docs/architecture.md` data-flow or `docs/config.md` if schema affects
   documented behaviour.
7. `make check` → green.
8. Commit: `feat(store): add migration NNNN <description>`.

## Hard rules

- Never `DROP TABLE` / destructive change without explicit owner confirmation,
  in writing, in the task text.
- Migrations must run offline: no external fetches, no data files outside repo.
- Never store a password in the DB in clear text. Hash (bcrypt/argon2) if ever
  needed; prefer not storing at all.

## If this fails

- Applied migration has a bug: write NNNN+1 that repairs it; log the incident in
  `docs/lessons.md` ("applied migrations are history, not drafts").
