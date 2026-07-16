package system

import (
	"strings"
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
		return "Недостаточно прав для действия " + action + " над службой " + name + ".\n\n" +
			serviceElevationHint
	case "docker":
		return "Недостаточно прав для действия " + action + " над контейнером " + name + ".\n\n" +
			"Проверь доступ к Docker daemon.\n" +
			"Обычно помогает запуск от пользователя из группы docker или запуск с повышенными правами."
	default:
		return "Недостаточно прав для выполнения действия."
	}
}
