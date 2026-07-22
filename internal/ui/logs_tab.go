package ui

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/i18n"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildLogsTab(cfg config.Config) fyne.CanvasObject {
	// The source labels are both shown and switched on, so they are built once
	// per tab construction rather than being package constants.
	logSourceSystem := i18n.T("System")
	logSourceServices := i18n.T("Services")
	const logSourceDocker = "Docker"

	// Sidebar names the section; keep the one-line caption.
	subtitle := widget.NewLabel(i18n.T("Working with logs: system / services / Docker"))

	sourceSelect := widget.NewSelect(
		[]string{logSourceSystem, logSourceServices, logSourceDocker},
		nil,
	)

	targetSelect := widget.NewSelect([]string{}, nil)
	targetSelect.PlaceHolder = i18n.T("Choose a source")

	linesEntry := widget.NewEntry()
	linesEntry.SetText(strconv.Itoa(appstate.GetConfig().DefaultLogLines))

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(i18n.T("Search within the loaded logs"))

	levelAll := i18n.T("All")
	levelErrors, levelWarnings, levelInfo := i18n.T("Errors"), i18n.T("Warnings"), i18n.T("Info")
	levelSelect := widget.NewSelect([]string{levelAll, levelErrors, levelWarnings, levelInfo}, nil)
	levelSelect.SetSelected(levelAll)

	statusLabel := newDataLabel(i18n.T("Status: waiting"))
	infoLabel := newDataLabel("")

	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapOff
	logEntry.Disable()

	autoRefreshCheck := widget.NewCheck(
		i18n.Tf("Auto-refresh (%d s)", cfg.RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.LogsAutoRefresh)
	refreshButton := widget.NewButton(i18n.T("Refresh"), nil)
	reloadSourcesButton := widget.NewButton(i18n.T("Reload list"), nil)
	copyButton := widget.NewButton(i18n.T("Copy"), nil)
	saveButton := widget.NewButton(i18n.T("Save to file"), nil)
	followButton := widget.NewButton(i18n.T("Follow"), nil)

	var rawLogs string
	var followStarted bool
	var lastSelectedTarget string

	var loadMu sync.Mutex
	isLoading := false

	buildLogFileName := func() string {
		timestamp := time.Now().Format("20060102_150405")

		switch sourceSelect.Selected {
		case logSourceSystem:
			return "system_logs_" + timestamp + ".log"
		case logSourceServices:
			target := strings.TrimSpace(targetSelect.Selected)
			if target == "" {
				target = "service"
			}
			target = strings.ReplaceAll(target, "/", "_")
			target = strings.ReplaceAll(target, "\\", "_")
			target = strings.ReplaceAll(target, " ", "_")
			return target + "_" + timestamp + ".log"
		case logSourceDocker:
			target := strings.TrimSpace(targetSelect.Selected)
			if target == "" {
				target = "container"
			}
			target = strings.ReplaceAll(target, "/", "_")
			target = strings.ReplaceAll(target, "\\", "_")
			target = strings.ReplaceAll(target, " ", "_")
			return target + "_" + timestamp + ".log"
		default:
			return "logs_" + timestamp + ".log"
		}
	}

	parseLines := func() int {
		text := strings.TrimSpace(linesEntry.Text)
		if text == "" {
			return 100
		}

		n, err := strconv.Atoi(text)
		if err != nil || n <= 0 {
			return 100
		}

		return n
	}

	applySearchFilter := func() {
		query := strings.ToLower(strings.TrimSpace(searchEntry.Text))
		level := strings.TrimSpace(levelSelect.Selected)

		if strings.TrimSpace(rawLogs) == "" {
			setTextIfChanged(logEntry, "")
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

		// Follow mode re-runs this every second; only write when the text
		// changed, so the scroll position holds instead of snapping to the top.
		setTextIfChanged(logEntry, strings.Join(filtered, "\n"))
	}

	updateTargetState := func() {
		switch sourceSelect.Selected {
		case logSourceSystem:
			targetSelect.ClearSelected()
			targetSelect.Options = []string{}
			targetSelect.PlaceHolder = i18n.T("No target needed for system logs")
			targetSelect.Disable()
		case logSourceServices:
			targetSelect.Options = []string{}
			targetSelect.PlaceHolder = i18n.T("Choose a service")
			targetSelect.Enable()
		case logSourceDocker:
			targetSelect.Options = []string{}
			targetSelect.PlaceHolder = i18n.T("Choose a container")
			targetSelect.Enable()
		default:
			targetSelect.Options = []string{}
			targetSelect.PlaceHolder = i18n.T("Choose a source")
			targetSelect.Disable()
		}

		targetSelect.Refresh()
	}

	loadTargets := func() {
		switch sourceSelect.Selected {
		case logSourceSystem:
			updateTargetState()
			lastSelectedTarget = ""
			statusLabel.SetText(i18n.T("Status: system logs ready"))
			infoLabel.SetText("")

		case logSourceServices:
			statusLabel.SetText(i18n.T("Status: loading the service list..."))

			go func() {
				names, err := system.ListServiceNames()
				if err != nil {
					fyne.Do(func() {
						targetSelect.Options = []string{}
						targetSelect.ClearSelected()
						targetSelect.Refresh()
						lastSelectedTarget = ""
						statusLabel.SetText(i18n.T("Status: could not load services"))
						infoLabel.SetText(err.Error())
					})
					return
				}

				fyne.Do(func() {
					targetSelect.Options = names
					targetSelect.PlaceHolder = i18n.T("Choose a service")
					targetSelect.Enable()
					targetSelect.Refresh()

					restored := false
					if lastSelectedTarget != "" {
						for _, name := range names {
							if name == lastSelectedTarget {
								targetSelect.SetSelected(lastSelectedTarget)
								restored = true
								break
							}
						}
					}

					if !restored {
						targetSelect.ClearSelected()
					}

					statusLabel.SetText(i18n.Tf("Status: %d services found", len(names)))
					infoLabel.SetText("")
				})
			}()

		case logSourceDocker:
			statusLabel.SetText(i18n.T("Status: loading the container list..."))

			go func() {
				names, err := system.ListDockerContainerNames()
				if err != nil {
					fyne.Do(func() {
						targetSelect.Options = []string{}
						targetSelect.ClearSelected()
						targetSelect.Refresh()
						lastSelectedTarget = ""
						statusLabel.SetText(i18n.T("Status: could not load containers"))
						infoLabel.SetText(err.Error())
					})
					return
				}

				fyne.Do(func() {
					targetSelect.Options = names
					targetSelect.PlaceHolder = i18n.T("Choose a container")
					targetSelect.Enable()
					targetSelect.Refresh()

					restored := false
					if lastSelectedTarget != "" {
						for _, name := range names {
							if name == lastSelectedTarget {
								targetSelect.SetSelected(lastSelectedTarget)
								restored = true
								break
							}
						}
					}

					if !restored {
						targetSelect.ClearSelected()
					}

					statusLabel.SetText(i18n.Tf("Status: %d containers found", len(names)))
					infoLabel.SetText("")
				})
			}()

		default:
			updateTargetState()
			lastSelectedTarget = ""
			statusLabel.SetText(i18n.T("Status: choose a log source"))
		}
	}

	loadLogs := func() {
		source := sourceSelect.Selected
		target := strings.TrimSpace(targetSelect.Selected)
		lines := parseLines()

		switch source {
		case "":
			statusLabel.SetText(i18n.T("Status: choose a log source"))
			return
		case logSourceServices, logSourceDocker:
			if target == "" {
				statusLabel.SetText(i18n.T("Status: choose a log target"))
				return
			}
		}

		loadMu.Lock()
		if isLoading {
			loadMu.Unlock()
			return
		}
		isLoading = true
		loadMu.Unlock()

		statusLabel.SetText(i18n.T("Status: loading logs..."))
		infoLabel.SetText("")

		go func(source, target string, lines int) {
			defer func() {
				loadMu.Lock()
				isLoading = false
				loadMu.Unlock()
			}()

			var (
				logs string
				err  error
			)

			switch source {
			case logSourceSystem:
				logs, err = system.GetSystemLogs(lines)
			case logSourceServices:
				logs, err = system.GetServiceLogs(target, lines)
			case logSourceDocker:
				logs, err = system.GetDockerContainerLogs(target, lines)
			default:
				err = fmt.Errorf("unknown log source")
			}

			if err != nil {
				fyne.Do(func() {
					rawLogs = ""
					logEntry.SetText("")
					statusLabel.SetText(i18n.T("Status: could not load logs"))
					infoLabel.SetText(err.Error())
				})
				return
			}

			fyne.Do(func() {
				rawLogs = logs
				applySearchFilter()

				lineCount := 0
				if strings.TrimSpace(logs) != "" {
					lineCount = len(strings.Split(strings.TrimRight(logs, "\n"), "\n"))
				}

				switch source {
				case logSourceSystem:
					statusLabel.SetText(
						i18n.Tf(
							"Status: system logs | %d lines | updated %s",
							lineCount,
							time.Now().Format("15:04:05"),
						),
					)
				case logSourceServices:
					statusLabel.SetText(
						i18n.Tf(
							"Status: service %s | %d lines | updated %s",
							target,
							lineCount,
							time.Now().Format("15:04:05"),
						),
					)
				case logSourceDocker:
					statusLabel.SetText(
						i18n.Tf(
							"Status: container %s | %d lines | updated %s",
							target,
							lineCount,
							time.Now().Format("15:04:05"),
						),
					)
				}

				infoLabel.SetText(i18n.T("Filters apply to the logs already loaded"))
			})
		}(source, target, lines)
	}

	targetSelect.OnChanged = func(value string) {
		lastSelectedTarget = strings.TrimSpace(value)
	}

	// Follow mode tails the log at a fixed 1s cadence, independent of the
	// configurable refresh interval.
	follow := newFixedRefresher(time.Second, func() { fyne.Do(loadLogs) })
	RegisterCloser(follow.Stop)

	sourceSelect.OnChanged = func(string) {
		rawLogs = ""
		logEntry.SetText("")
		searchEntry.SetText("")
		levelSelect.SetSelected(levelAll)
		updateTargetState()
		loadTargets()

		if followStarted {
			follow.Stop()
			followStarted = false
			followButton.SetText(i18n.T("Follow"))
		}
	}

	searchEntry.OnChanged = func(string) {
		applySearchFilter()
	}

	levelSelect.OnChanged = func(string) {
		applySearchFilter()
	}

	refreshButton.OnTapped = func() {
		loadLogs()
	}

	reloadSourcesButton.OnTapped = func() {
		loadTargets()
	}

	followButton.OnTapped = func() {
		if followStarted {
			follow.Stop()
			followStarted = false
			followButton.SetText(i18n.T("Follow"))
			statusLabel.SetText(i18n.T("Status: follow stopped"))
			return
		}

		loadLogs()
		followStarted = true
		followButton.SetText(i18n.T("Following..."))
		follow.Start()
	}

	copyButton.OnTapped = func() {
		text := logEntry.Text
		if strings.TrimSpace(text) == "" {
			statusLabel.SetText(i18n.T("Status: nothing to copy"))
			return
		}

		fyne.CurrentApp().Clipboard().SetContent(text)
		statusLabel.SetText(i18n.T("Status: logs copied"))
		infoLabel.SetText(i18n.T("Copied") + ": " + time.Now().Format("15:04:05"))
	}

	saveButton.OnTapped = func() {
		text := logEntry.Text
		if strings.TrimSpace(text) == "" {
			statusLabel.SetText(i18n.T("Status: nothing to save"))
			return
		}

		windows := fyne.CurrentApp().Driver().AllWindows()
		if len(windows) == 0 {
			statusLabel.SetText(i18n.T("Status: could not get a window"))
			return
		}

		fileName := buildLogFileName()

		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				statusLabel.SetText(i18n.T("Status: save failed"))
				infoLabel.SetText(err.Error())
				return
			}
			if writer == nil {
				statusLabel.SetText(i18n.T("Status: save cancelled"))
				return
			}
			if _, err := io.WriteString(writer, text); err != nil {
				_ = writer.Close()
				statusLabel.SetText(i18n.T("Status: file write error"))
				infoLabel.SetText(err.Error())
				return
			}

			// Close reports flush errors: ignoring it can silently truncate the
			// saved file.
			if err := writer.Close(); err != nil {
				statusLabel.SetText(i18n.T("Status: file write error"))
				infoLabel.SetText(err.Error())
				return
			}

			statusLabel.SetText(i18n.T("Status: logs saved"))
			infoLabel.SetText(i18n.T("Saved") + ": " + time.Now().Format("15:04:05"))
		}, windows[0])

		saveDialog.SetFileName(fileName)
		saveDialog.Show()
	}

	autoRefresh := newAutoRefresher(func() { fyne.Do(loadLogs) })
	RegisterCloser(autoRefresh.Stop)
	autoRefreshCheck.OnChanged = autoRefresh.SetEnabled
	autoRefresh.SetEnabled(cfg.LogsAutoRefresh)

	header := container.NewVBox(
		subtitle,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			container.NewVBox(
				widget.NewLabel(i18n.T("Source")),
				sourceSelect,
			),
			container.NewVBox(
				widget.NewLabel(i18n.T("Target")),
				targetSelect,
			),
		),
		container.NewGridWithColumns(3,
			container.NewVBox(
				widget.NewLabel(i18n.T("Number of lines")),
				linesEntry,
			),
			container.NewVBox(
				widget.NewLabel(i18n.T("Text search")),
				searchEntry,
			),
			container.NewVBox(
				widget.NewLabel(i18n.T("Level")),
				levelSelect,
			),
		),
		container.NewVBox(
			newToolbarRow(refreshButton, reloadSourcesButton, copyButton, saveButton, followButton),
			autoRefreshCheck,
		),
		statusLabel,
		infoLabel,
		widget.NewSeparator(),
	)

	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		container.NewVScroll(logEntry),
	)

	sourceSelect.SetSelected(logSourceSystem)
	updateTargetState()
	statusLabel.SetText(i18n.T("Status: system logs ready"))
	infoLabel.SetText(i18n.T("Press «Refresh» to load the logs"))

	if cfg.LogsAutoRefresh {
		autoRefreshCheck.OnChanged(true)
	}

	RegisterRefresh("Logs", loadLogs)

	return container.NewPadded(content)
}
