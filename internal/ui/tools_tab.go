package ui

import (
	"github.com/Lucky2356/system-hub/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// buildToolsTab groups the four screens a user reaches for rarely — system
// info, safe commands, the diagnostic report and the activity log — behind one
// sidebar entry.
//
// They were four top-level tabs competing with the daily ones (Services,
// Docker, Logs). Folding them here keeps the sidebar to the handful of sections
// someone actually switches between, without dropping any feature: each is the
// same screen it was, now an inner tab.
func buildToolsTab(w fyne.Window) fyne.CanvasObject {
	inner := container.NewAppTabs(
		container.NewTabItem(i18n.T("System Info"), NewSystemInfoTab(w)),
		container.NewTabItem(i18n.T("Commands"), buildCommandsTab(w)),
		container.NewTabItem(i18n.T("Report"), buildReportTab(w)),
		container.NewTabItem(i18n.T("Activity"), buildActivityTab()),
	)
	inner.SetTabLocation(container.TabLocationTop)
	return inner
}
