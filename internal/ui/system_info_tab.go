package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Lucky2356/system-hub/internal/i18n"
	"github.com/Lucky2356/system-hub/internal/system"
)

func NewSystemInfoTab(w fyne.Window) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(i18n.T("System Info"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	form := widget.NewForm(
		widget.NewFormItem(i18n.T("Hostname"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("OS"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("Platform"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("Platform version"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("Kernel"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("Kernel version"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("Architecture"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("Current user"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("Uptime"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("Boot time"), widget.NewLabel("")),
		widget.NewFormItem(i18n.T("Go version"), widget.NewLabel("")),
	)

	statusLabel := newDataLabel("")

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

		statusLabel.SetText(i18n.T("Information loaded"))
	}

	refresh := func() {
		info := system.GetHostInfo()
		updateForm(info)
	}

	refreshBtn := widget.NewButton(i18n.T("Refresh"), func() {
		refresh()
	})

	copyBtn := widget.NewButton(i18n.T("Copy"), func() {
		fyne.CurrentApp().Clipboard().SetContent(lastInfo.ToMultilineString())
		statusLabel.SetText(i18n.T("Copied to the clipboard"))
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
		value = i18n.T("n/a")
	}
	label.SetText(value)
}
