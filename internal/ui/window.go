package ui

import (
	"log"

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
		container.NewTabItem("Дашборд", buildDashboardTab(w, cfg)),
		container.NewTabItem("Сервисы", buildServicesTab(w, cfg)),
		container.NewTabItem("Docker", buildDockerTab(w, cfg)),
		container.NewTabItem("Процессы", buildProcessesTab(w)),
		container.NewTabItem("Файлы", buildFilesTab(w)),
		container.NewTabItem("Команды", buildCommandsTab(w)),
		container.NewTabItem("О системе", NewSystemInfoTab(w)),
		container.NewTabItem("Логи", buildLogsTab(cfg)),
		container.NewTabItem("История", buildActivityTab()),
		container.NewTabItem("Отчёт", buildReportTab(w)),
		container.NewTabItem("Настройки", buildSettingsTab(w, cfg)),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	isDark := cfg.Theme != "light"

	var themeBtn *widget.Button
	themeBtn = widget.NewButton("", func() {
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

		if err := config.Save(newCfg); err != nil {
			log.Printf("save theme config: %v", err)
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

	w.SetCloseIntercept(func() {
		RunClosers()
		w.Close()
	})

	return w
}
