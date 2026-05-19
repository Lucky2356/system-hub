# System Hub

System Hub is a cross-platform GUI application for monitoring and managing Linux systems. Built with Go and the [Fyne](https://fyne.io/) toolkit, it provides a unified interface for systemd services, Docker containers, system metrics, logs, processes, and diagnostics.

## Features

- **Dashboard** — CPU, RAM, disk usage; per-core CPU load; network RX/TX; system temperature (lm-sensors, Nvidia GPU); uptime; top processes by CPU
- **Services** — Browse systemd services, search/sort, start/stop/restart/enable/disable, view logs, open unit files, mark favorites
- **Docker** — Browse containers (running/exited/paused), start/stop/restart, inspect, logs, manage images (list/pull/remove), mark favorites
- **Processes** — Top processes by CPU/memory with search/sort, listening TCP/UDP ports viewer, kill processes
- **Files** — Read-only file browser for system configs and logs, view/edit text files, preset paths for common directories
- **Commands** — Run predefined safe diagnostic commands (systemctl, docker, journalctl, ss)
- **System Info** — Hostname, OS, kernel, uptime, user info with copy-to-clipboard
- **Logs** — Unified log viewer: system logs (journalctl), service logs, container logs; search, level filter, copy, save, auto-refresh, Follow mode
- **Activity** — Chronological log of all service/container actions with search filter
- **Report** — Generate diagnostics report and bundle with system stats, problems, top processes, activity
- **Settings** — Configurable refresh interval, default log lines, per-tab auto-refresh toggles

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

Build a standalone binary:

```bash
go build -o system-hub ./cmd/system-hub
```

## Configuration

Config file location:
- Linux: `~/.config/system-hub/config.json`
- Windows: `%APPDATA%/system-hub/config.json`

## Project Structure

```
cmd/system-hub/       — Application entry point
internal/
  activity/           — In-memory activity logger
  appstate/           — Global config state
  config/             — JSON config management
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
