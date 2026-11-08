# Playbook: add a setting

*When to read this:* a task says "add setting X" or "make Y configurable".
*Result you should have:* setting exists in config, defaults, validation, UI row,
docs row, test. STATUS: template, refine after Phase 2.

## Checklist

1. Decide the layer: bootstrap (`config.bootstrap.toml`, storage engine only) or
   runtime (rows in store, everything else). Wrong layer breaks ADR-0002.
2. Add the key, type, default, validation rule, and help text to
   `internal/config/`. Validation must reject bad values at load time, not deep in
   the code.
3. Add the test: defaults load, valid value loads, invalid value fails with a
   readable error naming the key.
4. UI: add the field to the right settings page with (a) default filled,
   (b) validation error display, (c) help text tooltip/line. Add a "test
   connection" button ONLY for DB settings.
5. Hot reload: does it apply without restart? If it cannot (bind address), UI must
   show "restart required". Mark which case applies in code comment.
6. Record the setting in `docs/config.md` (key, type, default, validation, help).
7. `make check` → green.
8. Commit: `feat(config): add <key> setting with validation and UI`.

## Worked micro-example (shape only)

```
Key:      system.bind_addr
Type:     string
Default:  ":8090"
Validate: parses as host:port; value ":8080" or port 8080 is REJECTED by rule.
Help:     "HTTP bind for the LAN UI. Port 8080 is reserved for redirect.py."
```

## If this fails

- UI tests missing in Phase 2: state that in the handoff note, do not skip the docs row.
- Setting needs a dependency to implement: STOP, write ADR, ask owner.
