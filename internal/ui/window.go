package ui

import (
	"github.com/Lucky2356/system-hub/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func NewMainWindow(a fyne.App, cfg config.Config) fyne.Window {
	w := a.NewWindow("System Hub")
	w.Resize(fyne.NewSize(900, 550))

	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", buildDashboardTab(w, cfg)),
		container.NewTabItem("Services", buildServicesTab(w, cfg)),
		container.NewTabItem("Docker", buildDockerTab(w, cfg)),
		container.NewTabItem("System Info", NewSystemInfoTab(w)),
		container.NewTabItem("Logs", buildLogsTab(cfg)),
		container.NewTabItem("Activity", buildActivityTab()),
		container.NewTabItem("Report", buildReportTab(w)),
		container.NewTabItem("Settings", buildSettingsTab(w, cfg)),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(tabs)
	return w
}