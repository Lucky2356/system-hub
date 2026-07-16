package ui

import (
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showLogsWindow(
	title string,
	header string,
	loadFunc func() (string, error),
	cfg config.Config,
) {
	logWindow := fyne.CurrentApp().NewWindow(title)
	logWindow.Resize(fyne.NewSize(900, 600))

	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapOff
	logEntry.Disable()

	infoLabel := widget.NewLabel(i18n.T("Logs are not loaded yet"))
	autoRefreshCheck := widget.NewCheck(
		i18n.Tf("Auto-refresh (%d s)", cfg.RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.LogViewerAutoRefresh)

	followButton := widget.NewButton(i18n.T("Follow"), nil)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(i18n.T("Text search..."))

	levelAll := i18n.T("All")
	levelErrors, levelWarnings, levelInfo := i18n.T("Errors"), i18n.T("Warnings"), i18n.T("Info")
	levelSelect := widget.NewSelect([]string{levelAll, levelErrors, levelWarnings, levelInfo}, nil)
	levelSelect.SetSelected(levelAll)

	followStarted := false

	var rawLogs string

	applySearchFilter := func() {
		query := strings.ToLower(strings.TrimSpace(searchEntry.Text))
		level := strings.TrimSpace(levelSelect.Selected)

		if strings.TrimSpace(rawLogs) == "" {
			logEntry.SetText("")
			return
		}

		lines := strings.Split(rawLogs, "\n")
		filtered := make([]string, 0, len(lines))

		for _, line := range lines {
			lineLower := strings.ToLower(line)

			// The level is matched against the log text, which is emitted by
			// journald/wevtutil in English regardless of the UI language.
			levelMatch := true
			switch level {
			case levelErrors:
				levelMatch = strings.Contains(lineLower, "error") ||
					strings.Contains(lineLower, "failed") ||
					strings.Contains(lineLower, "fatal")
			case levelWarnings:
				levelMatch = strings.Contains(lineLower, "warn") ||
					strings.Contains(lineLower, "warning")
			case levelInfo:
				levelMatch = strings.Contains(lineLower, "info")
			}

			queryMatch := query == "" || strings.Contains(lineLower, query)

			if levelMatch && queryMatch {
				filtered = append(filtered, line)
			}
		}

		logEntry.SetText(strings.Join(filtered, "\n"))
	}

	loadLogs := func() {
		infoLabel.SetText(i18n.T("Loading logs..."))

		go func() {
			logs, err := loadFunc()
			if err != nil {
				fyne.Do(func() {
					infoLabel.SetText(i18n.T("Could not load logs"))
					dialog.ShowError(err, logWindow)
				})
				return
			}

			fyne.Do(func() {
				rawLogs = logs
				applySearchFilter()
				infoLabel.SetText(i18n.T("Updated") + ": " + time.Now().Format("15:04:05"))
			})
		}()
	}

	refreshButton := widget.NewButton(i18n.T("Refresh"), func() {
		loadLogs()
	})

	searchEntry.OnChanged = func(string) {
		applySearchFilter()
	}

	levelSelect.OnChanged = func(string) {
		applySearchFilter()
	}

	autoRefresh := newAutoRefresher(func() { fyne.Do(loadLogs) })
	autoRefreshCheck.OnChanged = autoRefresh.SetEnabled
	autoRefresh.SetEnabled(cfg.LogViewerAutoRefresh)

	follow := newFixedRefresher(time.Second, func() { fyne.Do(loadLogs) })

	followButton.OnTapped = func() {
		if followStarted {
			follow.Stop()
			followStarted = false
			followButton.SetText(i18n.T("Follow"))
			infoLabel.SetText(i18n.T("Follow stopped"))
			return
		}

		loadLogs()
		followStarted = true
		followButton.SetText(i18n.T("Following..."))
		follow.Start()
	}

	// This is a transient window, so stop pollers on close rather than
	// registering a global closer (which would accumulate per opened window).
	logWindow.SetOnClosed(func() {
		autoRefresh.Stop()
		follow.Stop()
	})

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel(header),
			widget.NewSeparator(),
			container.NewHBox(refreshButton, followButton, autoRefreshCheck),
			container.NewGridWithColumns(2,
				container.NewVBox(
					widget.NewLabel(i18n.T("Search")),
					searchEntry,
				),
				container.NewVBox(
					widget.NewLabel(i18n.T("Level")),
					levelSelect,
				),
			),
			infoLabel,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewVScroll(logEntry),
	)

	logWindow.SetContent(container.NewPadded(content))
	logWindow.Show()

	loadLogs()
}
