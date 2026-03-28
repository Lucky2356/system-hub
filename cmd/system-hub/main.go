package main

import (
	"log"
	"path/filepath"

	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/logger"
	"github.com/Lucky2356/system-hub/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("load config: %v", err)
		cfg = config.DefaultConfig()
	}
	logPath, _ := config.ConfigFilePath()
	logPath = filepath.Join(filepath.Dir(logPath), cfg.LogFile)

	_ = logger.Init(logPath)

	a := app.New()
	w := ui.NewMainWindow(a, cfg)
	w.ShowAndRun()
}