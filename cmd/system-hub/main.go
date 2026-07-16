package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Lucky2356/system-hub/internal/activity"
	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/i18n"
	"github.com/Lucky2356/system-hub/internal/logger"
	"github.com/Lucky2356/system-hub/internal/ui"
	"github.com/Lucky2356/system-hub/internal/version"

	"fyne.io/fyne/v2/app"
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

	// Set the language before any UI is built: widget labels are read at
	// construction time.
	i18n.SetLanguage(cfg.Language)

	if logPath, err := config.LogFilePath(cfg); err != nil {
		log.Printf("warning: resolve log path: %v", err)
	} else if err := logger.Init(logPath); err != nil {
		log.Printf("warning: logger init: %v", err)
	}

	// The activity log survives restarts; failing to load it must not stop the
	// app from managing services.
	if activityPath, err := config.ActivityFilePath(); err != nil {
		log.Printf("warning: resolve activity path: %v", err)
	} else if err := activity.Init(activityPath); err != nil {
		log.Printf("warning: activity log init: %v", err)
	}

	// NewWithID (not New): the Preferences API Fyne uses internally requires a
	// unique app ID, otherwise it logs an error at startup.
	a := app.NewWithID(config.AppID)
	ui.ApplyTheme(a, cfg.Theme)
	w := ui.NewMainWindow(a, cfg)
	w.ShowAndRun()
}
