package ui

import (
	"github.com/Lucky2356/system-hub/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
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

	w.SetContent(tabs)
	return w
}
