package system

import "strings"

func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "access denied") ||
		strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "interactive authentication required") ||
		strings.Contains(msg, "authentication is required") ||
		strings.Contains(msg, "polkit") ||
		strings.Contains(msg, "root privileges") ||
		strings.Contains(msg, "must be root") ||
		strings.Contains(msg, "sudo")
}

func BuildPermissionHint(target, action, name string) string {
	switch target {
	case "service":
		return "Недостаточно прав для действия " + action + " над сервисом " + name + ".\n\n" +
			"Запусти приложение с повышенными правами или настрой polkit/sudo для systemctl."
	case "docker":
		return "Недостаточно прав для действия " + action + " над контейнером " + name + ".\n\n" +
			"Проверь доступ к Docker daemon.\n" +
			"Обычно помогает запуск от пользователя из группы docker или запуск с повышенными правами."
	default:
		return "Недостаточно прав для выполнения действия."
	}
}