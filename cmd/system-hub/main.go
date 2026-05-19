package main

import (
	"log"
	"path/filepath"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/logger"
	"github.com/Lucky2356/system-hub/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("load config: %v", err)
		cfg = config.DefaultConfig()
	}
	appstate.SetConfig(cfg)
	logPath, _ := config.ConfigFilePath()
	logPath = filepath.Join(filepath.Dir(logPath), cfg.LogFile)

	if err := logger.Init(logPath); err != nil {
		log.Printf("warning: logger init: %v", err)
	}

	a := app.New()
	if cfg.Theme == "light" {
		a.Settings().SetTheme(theme.LightTheme())
	}
	w := ui.NewMainWindow(a, cfg)
	w.ShowAndRun()
}