package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func ShowError(parent fyne.Window, err error) {
	if err == nil {
		return
	}
	dialog.ShowError(err, parent)
}

func ShowErrorMsg(parent fyne.Window, msg string) {
	dialog.ShowError(fmt.Errorf("%s", msg), parent)
}

func ShowInfo(parent fyne.Window, title, message string) {
	dialog.ShowInformation(title, message, parent)
}