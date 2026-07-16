# Changelog

All notable changes to this project are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.2.0]

The release that makes the app work on Windows, speak English, and run on ARM.

### Added
- **Real Windows support.** Services, logs and diagnostic commands previously
  refused to run anywhere but Linux, so half the app was dead on the platform it
  shipped an installer for. The service layer is now split by build tag:
  systemd/journalctl on Linux, the Windows SCM and the event log on Windows,
  behind one API the UI does not have to know about.
  - Services are enumerated with a single `EnumServicesStatusEx` call — 315
    services in ~1.6ms, instead of one `OpenService` per service.
  - Reads no longer need administrator rights: `mgr.Connect` asks for
    `SC_MANAGER_ALL_ACCESS`, so merely listing services failed with "Access is
    denied". Each action now opens a service with exactly the rights it needs.
  - Windows states are mapped onto the systemd vocabulary the UI already speaks,
    so status colours work unchanged.
  - Logs come from `wevtutil` and commands from `sc` / `wevtutil` / `netstat`.
- **English interface alongside Russian** (`internal/i18n`), with a switcher in
  Settings that applies without a restart. The `language` option accepts `ru`
  (default), `en` or `auto` (follow the OS locale). Configs written before the
  option existed keep the Russian UI.
  - Tests parse the source and fail on a missing, unused, or format-verb-
    mismatched translation, so an untranslated string cannot reach a release.
- **History sparklines** on the dashboard for CPU, RAM and network (~5 minutes).
  Samples come from the light tick, which already has the numbers, so the charts
  cost no extra system calls. Network is plotted as a rate — the raw counters
  are monotonic totals since boot — and the RX/TX labels now show that rate
  beside the total.
- **`arm64` `.deb` / `.rpm`** alongside `amd64` (Raspberry Pi, ARM servers).
  Both build natively: Fyne needs CGO, so cross-compiling would mean shipping a
  foreign C toolchain and X11 headers. The release job verifies the produced
  package reports the architecture it was built for.
- Per-container CPU and memory usage on the Docker tab via
  `docker stats --no-stream`. Hosts where `docker stats` fails but `docker ps`
  works still list their containers, just without the usage columns.
- The activity log is persisted next to the config as JSONL and reloaded at
  startup — it is an audit trail of service actions, so losing it on restart was
  wrong. Corrupt lines are skipped rather than failing the load.
- CI runs on Windows as well as Linux, with `-race`, `golangci-lint` and
  `govulncheck`.

### Changed
- Command execution goes through an injectable runner, so the hardening
  (`--`, name validation, timeouts) is covered by tests asserting the exact argv
  rather than by inspection.
- Interface strings moved into a translation catalog; the source strings are now
  English keys, and Russian is a translation of them.
- Problem notifications are throttled to one per problem per 10 minutes and
  re-arm once the problem clears, instead of firing on every dashboard refresh.

### Fixed
- Service and container states were never visible in list rows: a trailing label
  in a border layout's right slot renders past the visible row width. Both tabs
  now lead with the state badge.
- The Settings tab did not scroll: at the default window size it cut off at the
  auto-refresh checkboxes, leaving the theme, the language and both the Save and
  Reset buttons unreachable.
- The Docker tab opened in Containers mode while showing the image-only actions;
  the container buttons stayed hidden until the mode was toggled.
- `BuildProblems` returned "No problems detected" *as a problem*, so a healthy
  system raised a desktop notification announcing its health.
- The dashboard reported "systemd is unavailable" on Windows, which was never a
  problem to fix — only the absence of a service manager is. The label now names
  whichever manager the platform has.

### Security
- Path traversal in the unit-file lookup: a service name of `../../etc/passwd`
  escaped the systemd unit roots. Lookups are validated and confined to those
  roots.

## [0.1.2]

- Fixed the Windows installer asset name, which v0.1.1 published as
  `system-hub--setup.exe`.

## [0.1.1]

- **Broken release**: the Windows installer was published with an empty version
  in its file name. Use 0.1.2 or later.
- A real Windows installer (Inno Setup) instead of a bare `.exe`: Start Menu
  shortcut and an uninstall entry.
- Fixed console windows flashing constantly on Windows. The uptime lookup
  spawned PowerShell on every 2s tick, and child processes were created without
  `CREATE_NO_WINDOW`. Uptime now comes from gopsutil, and every spawned process
  hides its console.
- Redesigned dashboard and action rows; custom dark-first theme with a single
  palette shared by the theme and the status colours.
- Fixed logging being silently disabled when `log_file` was empty: the path
  resolved to the config directory itself and every write failed with "is a
  directory".

## [0.1.0]

- Initial release: dashboard, systemd services, Docker containers/images,
  processes and ports, file browser, safe commands, unified log viewer,
  activity log, diagnostics report, settings.
- Packaging: `nfpm.yaml` producing both `.deb` and `.rpm`, a Windows `.exe`,
  GitHub Actions for CI and releases, `LICENSE` (MIT) and `CONTRIBUTING.md`.
- `--version` flag; the version is injected at build time via ldflags.
- Hardening: binaries resolved via `exec.LookPath` and invoked by absolute path;
  service/container/image names validated and passed after `--`; the file
  browser requires absolute paths, rejects NUL bytes, and no longer deletes
  directories recursively.
- Performance: polling split into light and heavy ticks, listings cached with a
  short TTL, availability checks cached, per-core widgets updated in place.
