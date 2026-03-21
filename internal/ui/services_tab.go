package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildServicesTab() fyne.CanvasObject {
	title := widget.NewLabel("Services")
	title.TextStyle = fyne.TextStyle{Bold: true}

	info := widget.NewLabel(
		"Здесь позже будет управление systemd-сервисами.\n\n" +
			"План:\n" +
			"- список сервисов\n" +
			"- статус сервиса\n" +
			"- start / stop / restart",
	)

	return container.NewPadded(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			info,
		),
	)
}