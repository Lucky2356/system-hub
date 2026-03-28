package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showLogsWindow(
	title string,
	header string,
	loadFunc func() (string, error),
) {
	logWindow := fyne.CurrentApp().NewWindow(title)
	logWindow.Resize(fyne.NewSize(900, 600))

	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapOff
	logEntry.Disable()

	infoLabel := widget.NewLabel("Логи ещё не загружены")
	autoRefreshCheck := widget.NewCheck("Auto refresh (2 сек)", nil)

	stopAutoRefresh := make(chan struct{})
	autoRefreshStarted := false

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
				logEntry.SetText(logs)
				infoLabel.SetText("Обновлено: " + time.Now().Format("15:04:05"))
			})
		}()
	}

	refreshButton := widget.NewButton("Обновить", func() {
		loadLogs()
	})

	autoRefreshCheck.OnChanged = func(checked bool) {
		if !checked {
			return
		}

		if autoRefreshStarted {
			return
		}
		autoRefreshStarted = true

		go func() {
			ticker := time.NewTicker(2 * time.Second)
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

	logWindow.SetOnClosed(func() {
		close(stopAutoRefresh)
	})

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel(header),
			widget.NewSeparator(),
			container.NewHBox(refreshButton, autoRefreshCheck),
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