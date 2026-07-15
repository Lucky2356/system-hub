package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/logger"
	"github.com/Lucky2356/system-hub/internal/ui"
	"github.com/Lucky2356/system-hub/internal/version"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("system-hub %s\n", version.Version)
		os.Exit(0)
	}

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