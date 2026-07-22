package ui

import (
	"log"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/i18n"
	"github.com/Lucky2356/system-hub/internal/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// tabKeys names the tabs in the order they are built. The names are also the
// keys the refresh registry uses (Ctrl+R), so they must stay stable across
// languages — only the labels are translated.
//
// This is the reduced set: the four rarely-used screens (System Info, Commands,
// Report, Activity) moved into a single "Tools" section, so the sidebar holds
// the handful of places someone switches between daily.
var tabKeys = []string{
	"Dashboard", "Services", "Docker", "Processes",
	"Logs", "Files", "Tools", "Settings",
}

func NewMainWindow(a fyne.App, cfg config.Config) fyne.Window {
	w := a.NewWindow("System Hub")
	// This now takes effect: the tabs used to force a ~2015px minimum, so any
	// requested size was ignored. A comfortable default that still fits a
	// 1366x768 laptop.
	w.Resize(fyne.NewSize(1100, 720))

	w.SetContent(buildWindowContent(a, w, cfg))

	// Switching the language rebuilds the whole content tree: widget labels are
	// set at construction time, so there is nothing to re-translate in place.
	// RunClosers stops the tickers the old tree owned, otherwise every switch
	// would leak a set of them.
	i18n.OnChange(func() {
		fyne.Do(func() {
			RunClosers()
			w.SetContent(buildWindowContent(a, w, appstate.GetConfig()))
		})
	})

	w.SetCloseIntercept(func() {
		RunClosers()
		w.Close()
	})

	return w
}

func buildWindowContent(a fyne.App, w fyne.Window, cfg config.Config) fyne.CanvasObject {
	// Icons and a leading (side) tab bar: eleven tabs across the top overflowed
	// the window — "Settings" was clipped off the right edge at any reasonable
	// width. A vertical list has room for every entry and its icon.
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon(i18n.T("Dashboard"), theme.HomeIcon(), buildDashboardTab(w, cfg)),
		container.NewTabItemWithIcon(i18n.T("Services"), theme.ListIcon(), buildServicesTab(w, cfg)),
		container.NewTabItemWithIcon("Docker", theme.StorageIcon(), buildDockerTab(w, cfg)),
		container.NewTabItemWithIcon(i18n.T("Processes"), theme.ComputerIcon(), buildProcessesTab(w)),
		container.NewTabItemWithIcon(i18n.T("Logs"), theme.DocumentIcon(), buildLogsTab(cfg)),
		container.NewTabItemWithIcon(i18n.T("Files"), theme.FolderIcon(), buildFilesTab(w)),
		container.NewTabItemWithIcon(i18n.T("Tools"), theme.GridIcon(), buildToolsTab(w)),
		container.NewTabItemWithIcon(i18n.T("Settings"), theme.SettingsIcon(), buildSettingsTab(w, cfg)),
	)

	tabs.SetTabLocation(container.TabLocationLeading)

	isDark := cfg.Theme != "light"

	var themeBtn *widget.Button
	applyThemeButton := func() {
		if isDark {
			themeBtn.SetIcon(theme.NewThemedResource(theme.RadioButtonCheckedIcon()))
			themeBtn.SetText(i18n.T("Light"))
		} else {
			themeBtn.SetIcon(theme.NewThemedResource(theme.RadioButtonIcon()))
			themeBtn.SetText(i18n.T("Dark"))
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
		if selected >= 0 && selected < len(tabKeys) {
			if fn := GetRefresh(tabKeys[selected]); fn != nil {
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

	return container.NewBorder(
		container.NewVBox(container.NewPadded(topBar), widget.NewSeparator()),
		nil, nil, nil,
		tabs,
	)
}
