package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/config"

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
		fmt.Sprintf("Автообновление (%d сек)", cfg.RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.LogViewerAutoRefresh)

	followButton := widget.NewButton("Следить", nil)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск по тексту...")

	levelSelect := widget.NewSelect([]string{"Все", "Ошибки", "Предупреждения", "Инфо"}, nil)
	levelSelect.SetSelected("Все")

	followStarted := false

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
			case "ошибки":
				levelMatch = strings.Contains(lineLower, "error") ||
					strings.Contains(lineLower, "failed") ||
					strings.Contains(lineLower, "fatal")
			case "предупреждения":
				levelMatch = strings.Contains(lineLower, "warn") ||
					strings.Contains(lineLower, "warning")
			case "инфо":
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

	autoRefresh := newAutoRefresher(func() { fyne.Do(loadLogs) })
	autoRefreshCheck.OnChanged = autoRefresh.SetEnabled
	autoRefresh.SetEnabled(cfg.LogViewerAutoRefresh)

	follow := newFixedRefresher(time.Second, func() { fyne.Do(loadLogs) })

	followButton.OnTapped = func() {
		if followStarted {
			follow.Stop()
			followStarted = false
			followButton.SetText("Следить")
			infoLabel.SetText("Слежение остановлено")
			return
		}

		loadLogs()
		followStarted = true
		followButton.SetText("Слежение...")
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
