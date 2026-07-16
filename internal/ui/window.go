package ui

import (
	"log"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/version"

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
	applyThemeButton := func() {
		if isDark {
			themeBtn.SetIcon(theme.NewThemedResource(theme.RadioButtonCheckedIcon()))
			themeBtn.SetText("Светлая")
		} else {
			themeBtn.SetIcon(theme.NewThemedResource(theme.RadioButtonIcon()))
			themeBtn.SetText("Тёмная")
		}
	}

	themeBtn = widget.NewButton("", func() {
		isDark = !isDark

		newCfg := appstate.GetConfig()
		newCfg.Theme = "dark"
		if !isDark {
			newCfg.Theme = "light"
		}

		ApplyTheme(a, newCfg.Theme)
		applyThemeButton()
		appstate.SetConfig(newCfg)

		if err := config.Save(newCfg); err != nil {
			log.Printf("save theme config: %v", err)
		}
	})
	themeBtn.Importance = widget.LowImportance
	applyThemeButton()

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

	appTitle := widget.NewLabelWithStyle("System Hub", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	appVersion := widget.NewLabel(version.Version)
	appVersion.Importance = widget.LowImportance

	topBar := container.NewBorder(
		nil, nil,
		container.NewHBox(appTitle, appVersion),
		themeBtn,
		nil,
	)

	content := container.NewBorder(
		container.NewVBox(container.NewPadded(topBar), widget.NewSeparator()),
		nil, nil, nil,
		tabs,
	)
	w.SetContent(content)

	w.SetCloseIntercept(func() {
		RunClosers()
		w.Close()
	})

	return w
}
