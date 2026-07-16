package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Lucky2356/system-hub/internal/system"
)

func NewSystemInfoTab(w fyne.Window) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("О системе", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	form := widget.NewForm(
		widget.NewFormItem("Имя хоста", widget.NewLabel("")),
		widget.NewFormItem("ОС", widget.NewLabel("")),
		widget.NewFormItem("Платформа", widget.NewLabel("")),
		widget.NewFormItem("Версия платформы", widget.NewLabel("")),
		widget.NewFormItem("Ядро", widget.NewLabel("")),
		widget.NewFormItem("Версия ядра", widget.NewLabel("")),
		widget.NewFormItem("Архитектура", widget.NewLabel("")),
		widget.NewFormItem("Текущий пользователь", widget.NewLabel("")),
		widget.NewFormItem("Аптайм", widget.NewLabel("")),
		widget.NewFormItem("Время загрузки", widget.NewLabel("")),
		widget.NewFormItem("Версия Go", widget.NewLabel("")),
	)

	statusLabel := widget.NewLabel("")

	var lastInfo system.HostInfo

	updateForm := func(info system.HostInfo) {
		lastInfo = info

		setFormValue(form, 0, info.Hostname)
		setFormValue(form, 1, info.OS)
		setFormValue(form, 2, info.Platform)
		setFormValue(form, 3, info.PlatformVersion)
		setFormValue(form, 4, info.Kernel)
		setFormValue(form, 5, info.KernelVersion)
		setFormValue(form, 6, info.Architecture)
		setFormValue(form, 7, info.CurrentUser)
		setFormValue(form, 8, info.Uptime)
		setFormValue(form, 9, info.BootTime)
		setFormValue(form, 10, info.GoVersion)

		statusLabel.SetText("Информация загружена")
	}

	refresh := func() {
		info := system.GetHostInfo()
		updateForm(info)
	}

	refreshBtn := widget.NewButton("Обновить", func() {
		refresh()
	})

	copyBtn := widget.NewButton("Копировать", func() {
		fyne.CurrentApp().Clipboard().SetContent(lastInfo.ToMultilineString())
		statusLabel.SetText("Скопировано в буфер обмена")
	})

	buttons := container.NewHBox(refreshBtn, copyBtn)

	refresh()

	content := container.NewVBox(
		title,
		buttons,
		statusLabel,
		widget.NewSeparator(),
		form,
	)

	return container.NewVScroll(content)
}

func setFormValue(form *widget.Form, index int, value string) {
	if index < 0 || index >= len(form.Items) {
		return
	}

	label, ok := form.Items[index].Widget.(*widget.Label)
	if !ok {
		return
	}

	if value == "" {
		value = "н/д"
	}
	label.SetText(value)
}
