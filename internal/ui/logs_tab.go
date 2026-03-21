package ui

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
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
	copyButton := widget.NewButton("Copy logs", nil)

	var rawLogs string
	var autoRefreshStarted bool
	var lastSelectedTarget string
	stopAutoRefresh := make(chan struct{})

	var loadMu sync.Mutex
	isLoading := false

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
			lastSelectedTarget = ""
			statusLabel.SetText("Статус: system logs готовы")
			infoLabel.SetText("")

		case logSourceServices:
			statusLabel.SetText("Статус: загрузка списка сервисов...")

			go func() {
				names, err := system.ListServiceNames()
				if err != nil {
					fyne.Do(func() {
						targetSelect.Options = []string{}
						targetSelect.ClearSelected()
						targetSelect.Refresh()
						lastSelectedTarget = ""
						statusLabel.SetText("Статус: ошибка загрузки сервисов")
						infoLabel.SetText(err.Error())
					})
					return
				}

				fyne.Do(func() {
					targetSelect.Options = names
					targetSelect.PlaceHolder = "Выбери service"
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
						lastSelectedTarget = ""
						statusLabel.SetText("Статус: ошибка загрузки контейнеров")
						infoLabel.SetText(err.Error())
					})
					return
				}

				fyne.Do(func() {
					targetSelect.Options = names
					targetSelect.PlaceHolder = "Выбери container"
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

					statusLabel.SetText(fmt.Sprintf("Статус: контейнеров найдено %d", len(names)))
					infoLabel.SetText("")
				})
			}()

		default:
			updateTargetState()
			lastSelectedTarget = ""
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

		loadMu.Lock()
		if isLoading {
			loadMu.Unlock()
			return
		}
		isLoading = true
		loadMu.Unlock()

		statusLabel.SetText("Статус: загрузка логов...")
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
					statusLabel.SetText("Статус: ошибка загрузки логов")
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
						fmt.Sprintf(
							"Статус: system logs | %d строк | обновлено %s",
							lineCount,
							time.Now().Format("15:04:05"),
						),
					)
				case logSourceServices:
					statusLabel.SetText(
						fmt.Sprintf(
							"Статус: service %s | %d строк | обновлено %s",
							target,
							lineCount,
							time.Now().Format("15:04:05"),
						),
					)
				case logSourceDocker:
					statusLabel.SetText(
						fmt.Sprintf(
							"Статус: container %s | %d строк | обновлено %s",
							target,
							lineCount,
							time.Now().Format("15:04:05"),
						),
					)
				}

				infoLabel.SetText("Поиск применяется к уже загруженным логам")
			})
		}(source, target, lines)
	}

	targetSelect.OnChanged = func(value string) {
		lastSelectedTarget = strings.TrimSpace(value)
	}
	
		sourceSelect.OnChanged = func(string) {
		rawLogs = ""
		logEntry.SetText("")
		searchEntry.SetText("")
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

	copyButton.OnTapped = func() {
		text := logEntry.Text
		if strings.TrimSpace(text) == "" {
			statusLabel.SetText("Статус: нечего копировать")
			return
		}

		if w := fyne.CurrentApp().Driver().AllWindows(); len(w) > 0 {
			w[0].Clipboard().SetContent(text)
			statusLabel.SetText("Статус: логи скопированы")
			infoLabel.SetText("Скопировано: " + time.Now().Format("15:04:05"))
			return
		}

		statusLabel.SetText("Статус: не удалось получить окно для clipboard")
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
		container.NewHBox(refreshButton, reloadSourcesButton, copyButton, autoRefreshCheck),
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
	statusLabel.SetText("Статус: system logs готовы")
	infoLabel.SetText("Нажми «Обновить», чтобы загрузить логи")

	return container.NewPadded(content)
}