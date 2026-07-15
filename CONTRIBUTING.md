# Contributing to System Hub

Thanks for your interest in contributing!

## Getting started

1. Fork the repository and clone your fork.
2. Install Go 1.26+ and the Fyne build dependencies:
   - **Debian/Ubuntu:** `sudo apt install gcc pkg-config libgl1-mesa-dev libxrandr-dev libxcursor-dev libxi-dev libxinerama-dev libxxf86vm-dev libxkbcommon-dev`
   - **Fedora:** `sudo dnf install gcc pkgconf-pkg-config mesa-libGL-devel libXrandr-devel libXcursor-devel libXi-devel libXinerama-devel libXxf86vm-devel libxkbcommon-devel`
   - **Windows:** a working GCC (e.g. via MSYS2/TDM-GCC) for CGO.
3. Build and run:

```bash
make build
./system-hub
```

## Before submitting a PR

```bash
go vet ./...
go test ./...
```

Both must pass — CI runs them on every pull request.

## Guidelines

- Keep the layering: `internal/system` must not import `internal/ui`.
- External commands go through `runCmd`/`runCmdContext` in `internal/system/exec.go` (absolute-path resolution, timeouts). Never build shell strings.
- User-supplied names passed to `systemctl`/`docker` must go through `validateName` and be preceded by `--`.
- Add table-driven tests for new pure functions (parsers, formatters, validators).
- One logical change per pull request.

## Reporting bugs

Open an issue with your OS/distro, Go version, steps to reproduce, and relevant output from the log file (`~/.config/system-hub/system-hub.log`).
