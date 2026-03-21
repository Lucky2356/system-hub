package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildDockerTab() fyne.CanvasObject {
	title := widget.NewLabel("Docker")
	title.TextStyle = fyne.TextStyle{Bold: true}

	info := widget.NewLabel(
		"Здесь позже будет управление Docker.\n\n" +
			"План:\n" +
			"- список контейнеров\n" +
			"- статус контейнера\n" +
			"- запуск / остановка",
	)

	return container.NewPadded(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			info,
		),
	)
}