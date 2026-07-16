package ui

import (
	"fmt"

	"github.com/Lucky2356/system-hub/internal/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func ShowError(parent fyne.Window, err error) {
	if err == nil {
		return
	}
	if logger.Log != nil {
		logger.Log.Println(err)
	}

	dialog.ShowError(err, parent)
}

func ShowErrorMsg(parent fyne.Window, msg string) {
	dialog.ShowError(fmt.Errorf("%s", msg), parent)
}
