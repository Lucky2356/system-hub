# System Hub

[![CI](https://github.com/Lucky2356/system-hub/actions/workflows/ci.yml/badge.svg)](https://github.com/Lucky2356/system-hub/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Lucky2356/system-hub?include_prereleases)](https://github.com/Lucky2356/system-hub/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](go.mod)

System Hub is a cross-platform GUI application for monitoring and managing Linux systems. Built with Go and the [Fyne](https://fyne.io/) toolkit, it provides a unified interface for systemd services, Docker containers, system metrics, logs, processes, and diagnostics.

## Features

- **Dashboard** — CPU, RAM, disk usage; per-core CPU load; network RX/TX; system temperature (lm-sensors, Nvidia GPU); uptime; top processes by CPU
- **Services** — Browse systemd services, search/sort, start/stop/restart/enable/disable, view logs, open unit files, mark favorites
- **Docker** — Browse containers (running/exited/paused), start/stop/restart, inspect, logs, manage images (list/pull/remove), mark favorites
- **Processes** — Top processes by CPU/memory with search/sort, listening TCP/UDP ports viewer, kill processes
- **Files** — File browser and text editor for system configs and logs; view/edit/create/rename/delete, preset paths for common directories (destructive actions require confirmation and run with the current user's privileges)
- **Commands** — Run predefined safe diagnostic commands (systemctl, docker, journalctl, ss)
- **System Info** — Hostname, OS, kernel, uptime, user info with copy-to-clipboard
- **Logs** — Unified log viewer: system logs (journalctl), service logs, container logs; search, level filter, copy, save, auto-refresh, Follow mode
- **Activity** — Chronological log of all service/container actions with search filter
- **Report** — Generate diagnostics report and bundle with system stats, problems, top processes, activity
- **Settings** — Configurable refresh interval, default log lines, per-tab auto-refresh toggles

## Install

Download from the [latest release](https://github.com/Lucky2356/system-hub/releases/latest):

| Platform | File | Install |
| --- | --- | --- |
| Windows | `system-hub-<version>-setup.exe` | Run the installer (adds Start Menu shortcut and an uninstall entry) |
| Debian/Ubuntu | `system-hub_<version>_amd64.deb` | `sudo apt install ./system-hub_<version>_amd64.deb` |
| Fedora/RHEL | `system-hub-<version>-1.x86_64.rpm` | `sudo dnf install ./system-hub-<version>-1.x86_64.rpm` |

A portable `system-hub.exe` is also attached if you prefer no installation.

## Requirements

- Go 1.26+
- Linux (primary target; partial functionality on Windows/macOS)
- systemd (for service management)
- Docker CLI (for container management)
- lm-sensors (optional, for temperature monitoring)
- nvidia-smi (optional, for GPU temperature)

## Build & Run

```bash
go run ./cmd/system-hub
```

Build a standalone binary (with version stamped in):

```bash
make build VERSION=v0.1.0        # or: go build -o system-hub ./cmd/system-hub
./system-hub --version
```

## Packaging

Build artifacts are produced by CI (`.github/workflows/release.yml`) on a `vX.Y.Z`
tag: a Windows `.exe`, plus `.deb` and `.rpm` generated from a single
[`nfpm.yaml`](nfpm.yaml) and attached to the GitHub Release.

Local builds:

```bash
# Windows .exe (run on Windows, needs a C compiler for CGO/Fyne)
make windows-amd64 VERSION=v0.1.0

# .deb / .rpm (run on Linux; needs nfpm and Fyne dev headers)
make packages VERSION=0.1.0
```

Distro RPM specs live in [`packaging/`](packaging/) (Fedora and ALT Linux).

## Configuration

Config file location:
- Linux: `~/.config/system-hub/config.json`
- Windows: `%APPDATA%/system-hub/config.json`

### Language

The interface ships in Russian and English. Set `"language"` in the config, or
use the switcher in the Settings tab — it applies immediately, without a
restart:

| Value  | Meaning                          |
| ------ | -------------------------------- |
| `ru`   | Russian (default)                |
| `en`   | English                          |
| `auto` | Follow the operating system locale |

Adding a language means dropping a `<code>.json` file into
[`internal/i18n/translation/`](internal/i18n/translation/) and listing the code
in `i18n.Supported()`. The keys are the English source strings, so an untranslated
entry renders in English rather than breaking the layout. A test parses the
source and fails if a translation is missing, unused, or drops a format verb.

## Project Structure

```
cmd/system-hub/       — Application entry point
internal/
  activity/           — Activity logger, persisted as JSONL
  appstate/           — Global config state
  config/             — JSON config management
  i18n/               — Translations (ru/en), keys are the English strings
  logger/             — File-based logger
  system/             — System interaction (gopsutil, CLI tools)
    info.go           — CPU, RAM, disk, uptime, network, per-CPU
    processes.go      — Ports, top processes, kill
    services.go       — systemd service management
    docker.go         — Docker containers, images
    commands.go       — Safe diagnostic commands
    files.go          — File browser operations
    health.go         — Problem detection
    sensors.go        — Temperature monitoring
    report.go         — Diagnostics report
    permissions.go    — Permission error handling
  ui/                 — Fyne widgets and tabs
    dashboard_tab.go
    processes_tab.go
    services_tab.go
    docker_tab.go
    files_tab.go
    commands_tab.go
    system_info_tab.go
    logs_tab.go
    log_viewer.go
    activity_tab.go
    report_tab.go
    settings_tab.go
    favorites.go
    files_bridge.go
    notify.go
    components.go
    window.go
```

## Dependencies

- [Fyne](https://fyne.io/) v2.7.3 — GUI toolkit
- [gopsutil](https://github.com/shirou/gopsutil) v4 — System metrics

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).
Notable changes are tracked in [CHANGELOG.md](CHANGELOG.md).

## License

[MIT](LICENSE)
