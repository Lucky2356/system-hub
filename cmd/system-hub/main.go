package main

import (
	"log"

	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("load config: %v", err)
		cfg = config.DefaultConfig()
	}

	a := app.New()
	w := ui.NewMainWindow(a, cfg)
	w.ShowAndRun()
}