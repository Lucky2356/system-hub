# System Hub

[![CI](https://github.com/Lucky2356/system-hub/actions/workflows/ci.yml/badge.svg)](https://github.com/Lucky2356/system-hub/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Lucky2356/system-hub?include_prereleases)](https://github.com/Lucky2356/system-hub/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](go.mod)

System Hub is a desktop GUI for monitoring and managing a machine's services,
containers, processes and logs. Built with Go and the [Fyne](https://fyne.io/)
toolkit, it runs natively on **Linux and Windows**: the same interface drives
systemd and journalctl on Linux, and the Service Control Manager and the Event
Log on Windows.

The interface is available in Russian and English.

## Features

- **Dashboard** — CPU, RAM, disk usage with ~5 minutes of history; per-core CPU load; network RX/TX with live rate; system temperature (lm-sensors, Nvidia GPU); uptime; top processes by CPU
- **Services** — Browse services (systemd on Linux, the SCM on Windows), search/sort, start/stop/restart/enable/disable, view logs, open unit files, mark favorites
- **Docker** — Browse containers (running/exited/paused) with per-container CPU and memory, start/stop/restart, inspect, logs, manage images (list/pull/remove), mark favorites
- **Processes** — Top processes by CPU/memory with search/sort, listening TCP/UDP ports viewer, kill processes
- **Files** — File browser and text editor for system configs and logs; view/edit/create/rename/delete, preset paths for common directories (destructive actions require confirmation and run with the current user's privileges)
- **Commands** — Run predefined safe diagnostic commands (systemctl/journalctl/ss on Linux, sc/wevtutil/netstat on Windows, docker on both)
- **System Info** — Hostname, OS, kernel, uptime, user info with copy-to-clipboard
- **Logs** — Unified log viewer: system logs (journalctl / Windows Event Log), service logs, container logs; search, level filter, copy, save, auto-refresh, Follow mode
- **Activity** — Chronological, persisted log of all service/container actions with search filter
- **Report** — Generate diagnostics report and bundle with system stats, problems, top processes, activity
- **Settings** — Language, theme, refresh interval, default log lines, per-tab auto-refresh toggles

## Install

Download from the [latest release](https://github.com/Lucky2356/system-hub/releases/latest):

| Platform | File | Install |
| --- | --- | --- |
| Windows | `system-hub-<version>-setup.exe` | Run the installer (adds Start Menu shortcut and an uninstall entry) |
| Debian/Ubuntu | `system-hub_<version>_amd64.deb` | `sudo apt install ./system-hub_<version>_amd64.deb` |
| Fedora/RHEL | `system-hub-<version>-1.x86_64.rpm` | `sudo dnf install ./system-hub-<version>-1.x86_64.rpm` |

`arm64` / `aarch64` packages are published too (Raspberry Pi, ARM servers) —
swap the architecture in the file name.

A portable `system-hub.exe` is also attached if you prefer no installation.

## Requirements

Running a released build needs nothing beyond the OS itself; the rest is
optional and only unlocks the matching tab.

| | Linux | Windows |
| --- | --- | --- |
| Services | systemd | built in (managing services needs administrator rights) |
| Logs | journalctl | built in (Event Log) |
| Containers | Docker CLI | Docker CLI |
| Temperature | lm-sensors, nvidia-smi (optional) | nvidia-smi (optional) |

Building from source needs Go 1.26+ and, on Linux, the Fyne C dependencies
(see [Build & Run](#build--run)). macOS is not tested.

## Build & Run

```bash
go run ./cmd/system-hub
```

Build a standalone binary (with version stamped in):

```bash
make build VERSION=v0.2.0        # or: go build -o system-hub ./cmd/system-hub
./system-hub --version
```

## Packaging

Build artifacts are produced by CI (`.github/workflows/release.yml`) on a `vX.Y.Z`
tag and attached to the GitHub Release: a Windows installer and portable `.exe`,
plus `.deb` and `.rpm` for `amd64` and `arm64`, all generated from a single
[`nfpm.yaml`](nfpm.yaml).

The Linux packages build on native runners per architecture rather than
cross-compiling — Fyne needs CGO, so a cross-build would have to carry a foreign
C toolchain and X11 headers.

Local builds:

```bash
# Windows .exe (run on Windows, needs a C compiler for CGO/Fyne)
make windows-amd64 VERSION=v0.2.0

# .deb / .rpm (run on Linux; needs nfpm and Fyne dev headers)
make packages VERSION=0.2.0
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
    history.go        — Ring buffer of recent samples for the dashboard charts
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
