package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func NewMainWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("System Hub")
	w.Resize(fyne.NewSize(900, 550))

	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", buildDashboardTab()),
		container.NewTabItem("Services", buildServicesTab()),
		container.NewTabItem("Docker", buildDockerTab()),
		container.NewTabItem("Logs", buildLogsTab()),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(tabs)
	return w
}