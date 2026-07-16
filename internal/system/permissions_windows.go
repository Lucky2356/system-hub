//go:build windows

package system

import "github.com/Lucky2356/system-hub/internal/i18n"

// serviceElevationHint tells a Windows user how to get rights over the SCM.
// Listing services works for everyone; starting and stopping them does not.
// It is a function, not a constant, because the language can change at runtime.
func serviceElevationHint() string {
	return i18n.T("Managing Windows services requires administrator rights.\nClose the application and start it again using «Run as administrator».")
}
