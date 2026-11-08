# Playbook: add a filter rule

*When to read this:* a task says "add filter rule/law/rule type X".
*Result you should have:* new rule type, test, "why" explanation wired, UI page row,
docs. STATUS: template, refine after Phase 4.

## Checklist

1. Rules live in `internal/filter/`. Stage 1 rules are cheap, deterministic, and
   every rejection MUST produce a human-readable reason string. If your rule needs
   randomness or AI, it is not stage 1; check which stage the task assigns.
2. Implement the rule as a small struct + one `Apply(Job) (Outcome, Reason)` shape
   consistent with existing rules in the package. Copy the shape of the most
   similar existing rule file; do not invent a new pattern.
3. Role of the rule: blacklist / no-go / must-have / tag-weight / location. Set it
   explicitly; the scoring pipeline reads it.
4. Test (see `write-a-test.md`): table of jobs where the rule fires, jobs where it
   does not, and the exact reason string asserted.
5. UI page for this filter category exists already (Phase 4 pages) — add the field
   there; if the page doesn't exist yet, follow `add-a-ui-page.md` first.
6. Update `docs/config.md` and `docs/glossary.md` if the rule introduces a word.
7. `make check` → green.
8. Commit: `feat(filter): add <rule-name> rule with reason accounting`.

## Hard rules

- Every reject reason must be understandable by the owner 6 months later.
  "matched no-gos: 'unpaid internship'" is ok. "failed rule 7" is NOT ok.
- AND/OR groups: implement with the existing group mechanism; do not parse
  free-form strings unless the task says so and the test covers precedence.

## If this fails

- The score breakdown changes: update the docs example in `docs/testing.md` and
  say so in the commit body.
