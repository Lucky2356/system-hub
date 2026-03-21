package main

import (
	"github.com/Lucky2356/system-hub/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	w := ui.NewMainWindow(a)
	w.ShowAndRun()
}