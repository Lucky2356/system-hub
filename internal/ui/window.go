package ui

import (
	"log"
	"path/filepath"

	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/appstate"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewMainWindow(a fyne.App, cfg config.Config) fyne.Window {
	w := a.NewWindow("System Hub")
	w.Resize(fyne.NewSize(900, 550))

	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", buildDashboardTab(w, cfg)),
		container.NewTabItem("Services", buildServicesTab(w, cfg)),
		container.NewTabItem("Docker", buildDockerTab(w, cfg)),
		container.NewTabItem("Processes", buildProcessesTab(w)),
		container.NewTabItem("Files", buildFilesTab(w)),
		container.NewTabItem("Commands", buildCommandsTab(w)),
		container.NewTabItem("System Info", NewSystemInfoTab(w)),
		container.NewTabItem("Logs", buildLogsTab(cfg)),
		container.NewTabItem("Activity", buildActivityTab()),
		container.NewTabItem("Report", buildReportTab(w)),
		container.NewTabItem("Settings", buildSettingsTab(w, cfg)),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	isDark := cfg.Theme != "light"

	themeBtn := widget.NewButton("", func() {
		isDark = !isDark
		if isDark {
			a.Settings().SetTheme(theme.DarkTheme())
			themeBtn.SetText("☀")
		} else {
			a.Settings().SetTheme(theme.LightTheme())
			themeBtn.SetText("🌙")
		}

		newCfg := appstate.GetConfig()
		newCfg.Theme = "dark"
		if !isDark {
			newCfg.Theme = "light"
		}
		appstate.SetConfig(newCfg)

		cfgPath, err := config.ConfigFilePath()
		if err == nil {
			dir := filepath.Dir(cfgPath)
			savePath := filepath.Join(dir, "config.json")
			if err := config.Save(newCfg); err != nil {
				log.Printf("save theme config: %v", err)
			}
		}
	})
	if isDark {
		themeBtn.SetText("☀")
	} else {
		themeBtn.SetText("🌙")
	}

	ctrlR := &desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlR, func(shortcut fyne.Shortcut) {
		selected := tabs.SelectedIndex()
		tabNames := []string{
			"Dashboard", "Services", "Docker", "Processes",
			"Files", "Commands", "System Info", "Logs",
			"Activity", "Report", "Settings",
		}
		if selected >= 0 && selected < len(tabNames) {
			if fn := GetRefresh(tabNames[selected]); fn != nil {
				fn()
			}
		}
	})

	topBar := container.NewBorder(
		nil, nil, nil, themeBtn,
		container.NewHBox(),
	)

	content := container.NewBorder(topBar, nil, nil, nil, tabs)
	w.SetContent(content)
	return w
}
