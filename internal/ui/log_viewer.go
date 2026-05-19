package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/appstate"

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

	infoLabel := widget.NewLabel("Логи ещё не загружены")
	autoRefreshCheck := widget.NewCheck(
		fmt.Sprintf("Auto refresh (%d сек)", cfg.RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.LogViewerAutoRefresh)

	followButton := widget.NewButton("Follow", nil)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск по тексту...")

	levelSelect := widget.NewSelect([]string{"All", "Error", "Warn", "Info"}, nil)
	levelSelect.SetSelected("All")

	stopAutoRefresh := make(chan struct{})
	autoRefreshStarted := false
	followStarted := false
	stopFollow := make(chan struct{})

	var rawLogs string

	applySearchFilter := func() {
		query := strings.ToLower(strings.TrimSpace(searchEntry.Text))
		level := strings.ToLower(strings.TrimSpace(levelSelect.Selected))

		if strings.TrimSpace(rawLogs) == "" {
			logEntry.SetText("")
			return
		}

		lines := strings.Split(rawLogs, "\n")
		filtered := make([]string, 0, len(lines))

		for _, line := range lines {
			lineLower := strings.ToLower(line)

			levelMatch := true
			switch level {
			case "error":
				levelMatch = strings.Contains(lineLower, "error") ||
					strings.Contains(lineLower, "failed") ||
					strings.Contains(lineLower, "fatal")
			case "warn":
				levelMatch = strings.Contains(lineLower, "warn") ||
					strings.Contains(lineLower, "warning")
			case "info":
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
		infoLabel.SetText("Загрузка логов...")

		go func() {
			logs, err := loadFunc()
			if err != nil {
				fyne.Do(func() {
					infoLabel.SetText("Ошибка загрузки логов")
					dialog.ShowError(err, logWindow)
				})
				return
			}

			fyne.Do(func() {
				rawLogs = logs
				applySearchFilter()
				infoLabel.SetText("Обновлено: " + time.Now().Format("15:04:05"))
			})
		}()
	}

	refreshButton := widget.NewButton("Обновить", func() {
		loadLogs()
	})

	searchEntry.OnChanged = func(string) {
		applySearchFilter()
	}

	levelSelect.OnChanged = func(string) {
		applySearchFilter()
	}

	autoRefreshCheck.OnChanged = func(checked bool) {
		if !checked {
			if autoRefreshStarted {
				close(stopAutoRefresh)
				autoRefreshStarted = false
			}
			return
		}

		if autoRefreshStarted {
			return
		}
		autoRefreshStarted = true
		stopAutoRefresh = make(chan struct{})

		go func() {
			ticker := time.NewTicker(time.Duration(appstate.GetConfig().RefreshIntervalSeconds) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if autoRefreshCheck.Checked {
						fyne.Do(func() {
							loadLogs()
						})
					}
				case <-stopAutoRefresh:
					return
				}
			}
		}()
	}

	if cfg.LogViewerAutoRefresh {
		autoRefreshCheck.OnChanged(true)
	}

	followButton.OnTapped = func() {
		if followStarted {
			close(stopFollow)
			followStarted = false
			followButton.SetText("Follow")
			infoLabel.SetText("Follow остановлен")
			return
		}

		loadLogs()
		followStarted = true
		followButton.SetText("Following...")
		stopFollow = make(chan struct{})

		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					fyne.Do(func() {
						loadLogs()
					})
				case <-stopFollow:
					return
				}
			}
		}()
	}

	logWindow.SetOnClosed(func() {
		close(stopAutoRefresh)
		if followStarted {
			close(stopFollow)
		}
	})

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel(header),
			widget.NewSeparator(),
			container.NewHBox(refreshButton, followButton, autoRefreshCheck),
			container.NewGridWithColumns(2,
				container.NewVBox(
					widget.NewLabel("Поиск"),
					searchEntry,
				),
				container.NewVBox(
					widget.NewLabel("Уровень"),
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
