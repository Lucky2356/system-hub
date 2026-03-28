package ui

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	logSourceSystem   = "System"
	logSourceServices = "Services"
	logSourceDocker   = "Docker"
)

func buildLogsTab(cfg config.Config) fyne.CanvasObject {
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
	linesEntry.SetText(strconv.Itoa(cfg.DefaultLogLines))

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск по уже загруженным логам")

	levelSelect := widget.NewSelect([]string{"All", "Error", "Warn", "Info"}, nil)
	levelSelect.SetSelected("All")

	statusLabel := widget.NewLabel("Статус: ожидание")
	infoLabel := widget.NewLabel("")

	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapOff
	logEntry.Disable()

	autoRefreshCheck := widget.NewCheck(
		fmt.Sprintf("Auto refresh (%d сек)", cfg.RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.LogsAutoRefresh)
	refreshButton := widget.NewButton("Обновить", nil)
	reloadSourcesButton := widget.NewButton("Обновить список", nil)
	copyButton := widget.NewButton("Copy logs", nil)
	saveButton := widget.NewButton("Save to file", nil)

	var rawLogs string
	var autoRefreshStarted bool
	var lastSelectedTarget string
	stopAutoRefresh := make(chan struct{})

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
			default:
				levelMatch = true
			}

			queryMatch := query == "" || strings.Contains(lineLower, query)

			if levelMatch && queryMatch {
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

				infoLabel.SetText("Фильтры применяются к уже загруженным логам")
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
		levelSelect.SetSelected("All")
		updateTargetState()
		loadTargets()
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

	saveButton.OnTapped = func() {
		text := logEntry.Text
		if strings.TrimSpace(text) == "" {
			statusLabel.SetText("Статус: нечего сохранять")
			return
		}

		windows := fyne.CurrentApp().Driver().AllWindows()
		if len(windows) == 0 {
			statusLabel.SetText("Статус: не удалось получить окно")
			return
		}

		fileName := buildLogFileName()

		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				statusLabel.SetText("Статус: ошибка сохранения")
				infoLabel.SetText(err.Error())
				return
			}
			if writer == nil {
				statusLabel.SetText("Статус: сохранение отменено")
				return
			}
			defer writer.Close()

			if _, err := io.WriteString(writer, text); err != nil {
				statusLabel.SetText("Статус: ошибка записи файла")
				infoLabel.SetText(err.Error())
				return
			}

			statusLabel.SetText("Статус: логи сохранены")
			infoLabel.SetText("Сохранено: " + time.Now().Format("15:04:05"))
		}, windows[0])

		saveDialog.SetFileName(fileName)
		saveDialog.Show()
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
			ticker := time.NewTicker(time.Duration(cfg.RefreshIntervalSeconds) * time.Second)
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
		container.NewGridWithColumns(3,
			container.NewVBox(
				widget.NewLabel("Количество строк"),
				linesEntry,
			),
			container.NewVBox(
				widget.NewLabel("Поиск по тексту"),
				searchEntry,
			),
			container.NewVBox(
				widget.NewLabel("Уровень"),
				levelSelect,
			),
		),
		container.NewHBox(refreshButton, reloadSourcesButton, copyButton, saveButton, autoRefreshCheck),
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
	
	if cfg.LogsAutoRefresh {
		autoRefreshCheck.OnChanged(true)
	}
	
	return container.NewPadded(content)
}