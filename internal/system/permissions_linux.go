//go:build linux

package system

import "github.com/Lucky2356/system-hub/internal/i18n"

// serviceElevationHint tells a Linux user how to get rights over systemd units.
// It is a function, not a constant, because the language can change at runtime.
func serviceElevationHint() string {
	return i18n.T("Run the application with elevated privileges, or configure polkit/sudo for systemctl.")
}
