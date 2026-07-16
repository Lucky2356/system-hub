//go:build windows

package system

// serviceElevationHint tells a Windows user how to get rights over the SCM.
// Listing services works for everyone; starting and stopping them does not.
const serviceElevationHint = "Управление службами Windows требует прав администратора.\n" +
	"Закрой приложение и запусти его через «Запуск от имени администратора»."
