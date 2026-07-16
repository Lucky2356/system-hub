//go:build linux

package system

// serviceElevationHint tells a Linux user how to get rights over systemd units.
const serviceElevationHint = "Запусти приложение с повышенными правами или настрой polkit/sudo для systemctl."
