# Playbook: release a build / install on the Heart

*When to read this:* a task says "build for the Pi", "release", "deploy", "upgrade".
*Result you should have:* new arm64 binary running on the Heart, previous one kept.
STATUS: template; fill exact Pi paths after first real release (Phase 1 checks).

## On the Flagship (dev machine, Linux)

1. `make check` must be green. Release tag only from a green tree.
2. `make build-arm64` → expect `bin/jobcrawler-arm64`; verify with `file` (aarch64).
3. Commit changelog line + tag: `git tag vX.Y.Z && git push --tags` (owner confirms
   remote exists; never force-push).

## On the Heart (Raspberry Pi, via ssh)

4. Copy binary to staging path (fill after first release; suggest
   `/opt/jobcrawler/staging/jobcrawler`).
5. Stop service: `sudo systemctl stop jobcrawler`.
6. Swap: move staged binary to live path, `chmod +x`, keep the old binary at
   `<live>.prev` for rollback.
7. `sudo systemctl start jobcrawler`.
8. Verify in order (stop at first failure):
   - `systemctl status jobcrawler` → active (running)
   - `journalctl -u jobcrawler -n 50 --no-pager` → startup banner, no errors
   - `curl http://<pi>:8090/healthz` → 200 OK
   - RAM: `systemctl show jobcrawler -p MemoryCurrent` → report the number.

## If this fails

- Service fails to start: ROLLBACK (keep `.prev`), then debug offline via
  `debug-crawler-on-the-heart.md`. Never leave the Heart down.
- Port conflict/`bind: address already in use`: another service took 8090 — check
  `sudo ss -tlnp | grep 8090`. NEVER touch 8080 (redirect.py) as a fix.
- Disk full: say so to the owner; do not delete other services' files.

## Stop conditions

Any step that would modify redirect.py, port 8080, or services other than
jobcrawler → STOP and ask the owner.
