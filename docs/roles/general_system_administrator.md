# General System Administrator

## Purpose
Provide overall ownership of the company’s IT infrastructure: servers, networking, storage, and services that support day‑to‑day operations.

## Core Responsibilities
- Install, configure, and maintain operating systems and application stacks on physical/virtual hosts.
- Monitor system health, logs, and performance metrics; respond to alerts.
- Apply security patches, manage user accounts, and enforce access controls.
- Plan and execute capacity and uptime strategies (redundancy, backups, DR).
- Document configuration, procedures, and change management.
- Collaborate with DevOps, developers, and security teams.

## Common Tools
- Linux distributions (Ubuntu, CentOS, Debian, etc.)
- `systemd`, `cron`, `logrotate`
- Monitoring: Prometheus, Grafana, Zabbix, Nagios
- Configuration management: Ansible, Puppet, Chef
- Source control: Git

## Key Skills
- Deep understanding of Linux internals and file‑system hierarchies.
- Scripting: Bash, Python, or PowerShell.
- Networking fundamentals (TCP/IP, routing, firewalls).
- Troubleshooting: log analysis, process debugging.
- Documentation and change‑management practices.

## Typical Tasks for AI Reference
- *Create a new user:* `useradd`, set UID, home, shell.
- *Apply a system update:* `apt‑get update && apt‑get upgrade` for Debian‑based systems.
- *Set up an SSH key:* `mkdir ~/.ssh && chmod 700 ~/.ssh && echo "<pubkey>" > ~/.ssh/authorized_keys`.
- *Check disk usage:* `df -h`, `du -sh */`.
- *Restart a service:* `systemctl restart <service>`.
