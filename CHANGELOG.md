# Changelog

All notable changes to this project are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- `--version` flag; version is injected at build time via ldflags (`internal/version`).
- Unit tests for parsers, formatters and validators (`internal/system`, `internal/config`).
- GitHub Actions: CI (vet/build/test/govulncheck) and Release (builds `.exe`, `.deb`, `.rpm` on `v*` tags).
- `nfpm.yaml` — single config producing both `.deb` and `.rpm` packages.
- `LICENSE` (MIT), `CONTRIBUTING.md`.

### Changed
- Dashboard polling split into light (CPU/RAM/disk/network, every tick) and heavy
  (service/container listings, process scan, sensors — at most every 5s).
- Service/container listings cached with a short TTL and invalidated after
  start/stop/restart actions.
- systemd/docker availability checks cached (15s TTL).
- `docker pull` no longer subject to the 30-second command timeout.
- Per-core CPU widgets are updated in place instead of being rebuilt every tick.
- Dashboard is scrollable on small windows.
- Interface language unified to Russian.
- RPM specs: added missing Fyne build dependencies (GL/X11/xkbcommon), CGO and
  version injection.
- Makefile: version injection, CGO-correct cross targets, `make packages`
  (deb+rpm via nfpm).

### Fixed
- Compilation error in `internal/ui/window.go` (self-referencing button variable).
- Goroutine leak: dashboard auto-refresh now stops when the window closes.
- Theme selection in Settings now applies dark theme without restart.

### Security
- External binaries are resolved via `exec.LookPath` and invoked by absolute
  path (prevents PATH hijacking).
- Service/container/image names are validated and passed after `--`
  (prevents argument injection).
- File browser requires absolute paths and rejects NUL bytes; recursive
  directory deletion removed (only empty directories can be deleted).

## [0.1.0]

- Initial release: dashboard, systemd services, Docker containers/images,
  processes and ports, file browser, safe commands, unified log viewer,
  activity log, diagnostics report, settings.
