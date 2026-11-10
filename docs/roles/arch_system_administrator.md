# Arch System Administrator

## Purpose
Maintain the stability, security, and performance of the company’s Arch Linux deployments, often used for cutting‑edge development or research environments.

## Core Responsibilities
- Keep Arch systems up to date with the rolling‑release model.
- Manage `pacman`, `yay`, and other AUR helpers.
- Resolve dependency conflicts, downgrades, and system rollbacks.
- Harden systems following Arch security guidelines.
- Set up and manage the Xorg / Wayland stack where needed.
- Configure and maintain local repositories or pkgbuild automations.
- Provide documentation and scripts for new setup or migration.

## Common Tools
- `pacman`, `pacman-contrib`, `yay`, `makepkg`.
- `systemd` services, `journalctl`, `systemd‑timesyncd`.
- Security: `sudo`, `pam`, `auditd`.
- Monitoring: `htop`, `glances`, `sysstat`.
- Automation: Ansible, shell scripts.

## Key Skills
- Deep understanding of Arch’s rolling release and package management.
- Experience with creating PKGBUILDs and handling AUR packages.
- System hardening specific to Arch (e.g., `pacman-key` rotation).
- Managing Xorg/Wayland and GPU drivers.
- Familiarity with container runtimes on Arch.

## Typical Tasks for AI Reference
- *Update system:* `pacman -Syu`.
- *Install AUR package:* `yay -S <package>`.
- *Rebuild a specific package:* `makepkg -s -i -f` from PKGBUILD.
- *Disable a service:* `systemctl disable <service>`.
- *Check journal for errors:* `journalctl -p err -b`.
