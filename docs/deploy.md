# Deploy (the Heart / Raspberry Pi)

*When to read this:* cutting a release build or installing/upgrading on the Pi.
*Result you should have:* a working arm64 upgrade with a rollback path.
Follow `playbooks/release-a-build.md` for the exact steps; this page holds the
background and the facts that rarely change.

## Facts

- Cross-compile from the Flagship. One static binary, no runtime, no venv:
  ```
  make build-arm64   # GOOS=linux GOARCH=arm64 go build
  ```
- Target host "Esemby-Pi": Raspberry Pi 4B, Ubuntu Server aarch64, big SSD, runs
  many other services 24/7. Treat CPU, RAM, SSD wear as scarce.
- **Port 8080 belongs to `redirect.py` (http.server). Never modify it. Never bind to
  8080. This project uses 8090 by default.**
- systemd unit: type=simple, Restart=on-failure, MemoryMax tuned to Pi budget,
  `RestartSec=5`, stop with SIGTERM (graceful shutdown must work).
- Zero telemetry. No secrets on disk in clear outside the config file the owner owns.

## RAM budget

- Heart has ~3.9 GB physical (`MemTotal: 3871632 kB`), shared with other services;
  ~2.15 GB was available at Phase 0 measurement (ADR-0004).
- Idle target: < 100 MB. Crawl peak: < 250 MB. These are the owner-approved caps.
- After any performance-related change, measure real RSS from systemd or
  `/proc/<pid>/status` and record it in `docs/lessons.md` or the release notes.

## Upgrade procedure (summary)

1. `make build-arm64`.
2. scp binary to the Pi (staging path, NOT the live path).
3. `systemctl stop jobcrawler` → swap binary → `systemctl start jobcrawler`.
4. Watch `journalctl -u jobcrawler -n 50` — confirm healthy log line.
5. If broken: restore previous binary from backup path, start again. Never debug
   live; roll back first, investigate later via `debug-crawler-on-the-heart.md`.

Reverse-proxying the UI to the internet is unsupported by design (LAN-only, ADR TBD).
