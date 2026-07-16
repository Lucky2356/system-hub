package system

import (
	"strings"

	"github.com/Lucky2356/system-hub/internal/i18n"
)

// IsPermissionError recognises the many ways systemd, polkit, Docker and the
// Windows SCM say "you may not do that".
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "access denied") ||
		strings.Contains(msg, "access is denied") || // Windows SCM
		strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "interactive authentication required") ||
		strings.Contains(msg, "authentication is required") ||
		strings.Contains(msg, "polkit") ||
		strings.Contains(msg, "root privileges") ||
		strings.Contains(msg, "must be root") ||
		strings.Contains(msg, "sudo")
}

// BuildPermissionHint explains how to get the missing rights. The service hint
// is platform-specific: telling a Windows user to configure polkit is useless.
func BuildPermissionHint(target, action, name string) string {
	switch target {
	case "service":
		return i18n.Tf("Not enough permissions for the %s action on service %s.", action, name) +
			"\n\n" + serviceElevationHint()
	case "docker":
		return i18n.Tf("Not enough permissions for the %s action on container %s.", action, name) +
			"\n\n" + i18n.T("Check access to the Docker daemon.\nRunning as a member of the docker group, or with elevated privileges, usually helps.")
	default:
		return i18n.T("Not enough permissions to perform this action.")
	}
}
