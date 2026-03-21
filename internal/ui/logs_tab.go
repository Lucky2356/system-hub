package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	logSourceSystem   = "System"
	logSourceServices = "Services"
	logSourceDocker   = "Docker"
)

func buildLogsTab() fyne.CanvasObject {
	title := widget.NewLabel("Logs")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Общая работа с логами: system / services / docker")

	sourceSelect := widget.NewSelect(
		[]string{logSourceSystem, logSourceServices, logSourceDocker},
		nil,
	)

	targetSelect := widget.NewSelect([]string{}, nil)
	targetSelect.PlaceHolder = "Выбери источник"

	linesEntry := widget.NewEntry()
	linesEntry.SetText("100")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск по уже загруженным логам")

	statusLabel := widget.NewLabel("Статус: ожидание")
	infoLabel := widget.NewLabel("")

	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapOff
	logEntry.Disable()

	autoRefreshCheck := widget.NewCheck("Auto refresh (2 сек)", nil)
	refreshButton := widget.NewButton("Обновить", nil)
	reloadSourcesButton := widget.NewButton("Обновить список", nil)

	var rawLogs string
	var autoRefreshStarted bool
	stopAutoRefresh := make(chan struct{})

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
		if query == "" {
			logEntry.SetText(rawLogs)
			return
		}

		lines := strings.Split(rawLogs, "\n")
		filtered := make([]string, 0, len(lines))

		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), query) {
				filtered = append(filtered, line)
			}
		}

		logEntry.SetText(strings.Join(filtered, "\n"))
	}

	updateTargetState := func() {
		switch sourceSelect.Selected {
		case logSourceSystem:
			targetSelect.ClearSelected()
			targetSelect.Options = []string{}
			targetSelect.PlaceHolder = "Для System выбор не нужен"
			targetSelect.Disable()
		case logSourceServices:
			targetSelect.Options = []string{}
			targetSelect.PlaceHolder = "Выбери service"
			targetSelect.Enable()
		case logSourceDocker:
			targetSelect.Options = []string{}
			targetSelect.PlaceHolder = "Выбери container"
			targetSelect.Enable()
		default:
			targetSelect.Options = []string{}
			targetSelect.PlaceHolder = "Выбери источник"
			targetSelect.Disable()
		}

		targetSelect.Refresh()
	}

	loadTargets := func() {
		switch sourceSelect.Selected {
		case logSourceSystem:
			updateTargetState()
			statusLabel.SetText("Статус: system logs готовы")
		case logSourceServices:
			statusLabel.SetText("Статус: загрузка списка сервисов...")

			go func() {
				names, err := system.ListServiceNames()
				if err != nil {
					fyne.Do(func() {
						targetSelect.Options = []string{}
						targetSelect.ClearSelected()
						targetSelect.Refresh()
						statusLabel.SetText("Статус: ошибка загрузки сервисов")
						infoLabel.SetText(err.Error())
					})
					return
				}

				fyne.Do(func() {
					targetSelect.Options = names
					targetSelect.ClearSelected()
					targetSelect.PlaceHolder = "Выбери service"
					targetSelect.Enable()
					targetSelect.Refresh()
					statusLabel.SetText(fmt.Sprintf("Статус: сервисов найдено %d", len(names)))
					infoLabel.SetText("")
				})
			}()
		case logSourceDocker:
			statusLabel.SetText("Статус: загрузка списка контейнеров...")

			go func() {
				names, err := system.ListDockerContainerNames()
				if err != nil {
					fyne.Do(func() {
						targetSelect.Options = []string{}
						targetSelect.ClearSelected()
						targetSelect.Refresh()
						statusLabel.SetText("Статус: ошибка загрузки контейнеров")
						infoLabel.SetText(err.Error())
					})
					return
				}

				fyne.Do(func() {
					targetSelect.Options = names
					targetSelect.ClearSelected()
					targetSelect.PlaceHolder = "Выбери container"
					targetSelect.Enable()
					targetSelect.Refresh()
					statusLabel.SetText(fmt.Sprintf("Статус: контейнеров найдено %d", len(names)))
					infoLabel.SetText("")
				})
			}()
		default:
			updateTargetState()
			statusLabel.SetText("Статус: выбери источник логов")
		}
	}

	loadLogs := func() {
		source := sourceSelect.Selected
		target := strings.TrimSpace(targetSelect.Selected)
		lines := parseLines()

		switch source {
		case "":
			statusLabel.SetText("Статус: выбери источник логов")
			return
		case logSourceServices, logSourceDocker:
			if target == "" {
				statusLabel.SetText("Статус: выбери цель для логов")
				return
			}
		}

		statusLabel.SetText("Статус: загрузка логов...")
		infoLabel.SetText("")

		go func() {
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
					statusLabel.SetText("Статус: ошибка загрузки логов")
					infoLabel.SetText(err.Error())
				})
				return
			}

			fyne.Do(func() {
				rawLogs = logs
				applySearchFilter()
				statusLabel.SetText("Статус: логи загружены")
				infoLabel.SetText("Обновлено: " + time.Now().Format("15:04:05"))
			})
		}()
	}

	sourceSelect.OnChanged = func(string) {
		updateTargetState()
		loadTargets()
	}

	searchEntry.OnChanged = func(string) {
		applySearchFilter()
	}

	refreshButton.OnTapped = func() {
		loadLogs()
	}

	reloadSourcesButton.OnTapped = func() {
		loadTargets()
	}

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
						loadLogs()
					}
				case <-stopAutoRefresh:
					return
				}
			}
		}()
	}

	header := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			container.NewVBox(
				widget.NewLabel("Источник"),
				sourceSelect,
			),
			container.NewVBox(
				widget.NewLabel("Цель"),
				targetSelect,
			),
		),
		container.NewGridWithColumns(2,
			container.NewVBox(
				widget.NewLabel("Количество строк"),
				linesEntry,
			),
			container.NewVBox(
				widget.NewLabel("Поиск по тексту"),
				searchEntry,
			),
		),
		container.NewHBox(refreshButton, reloadSourcesButton, autoRefreshCheck),
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

	return container.NewPadded(content)
}