package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func Run() {
	a := app.New()
	w := a.NewWindow("System Hub")

	w.SetContent(widget.NewLabel("System Hub"))
	w.Resize(fyne.NewSize(900, 600))
	w.ShowAndRun()
}