package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildLogsTab() fyne.CanvasObject {
	title := widget.NewLabel("Logs")
	title.TextStyle = fyne.TextStyle{Bold: true}

	info := widget.NewLabel(
		"Здесь позже будет просмотр логов.\n\n" +
			"План:\n" +
			"- выбор источника логов\n" +
			"- текстовый просмотр\n" +
			"- обновление",
	)

	return container.NewPadded(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			info,
		),
	)
}