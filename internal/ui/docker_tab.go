package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildDockerTab(parent fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Docker")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Просмотр и управление Docker-контейнерами")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск контейнера (например: nginx, postgres...)")

	statusLabel := widget.NewLabel("Статус: ожидание")

	var allContainers []system.DockerContainerInfo
	var filteredContainers []system.DockerContainerInfo
	selectedIndex := -1

	detailsButton := widget.NewButton("Подробнее", nil)
	logsButton := widget.NewButton("Logs", nil)
	startButton := widget.NewButton("Start", nil)
	stopButton := widget.NewButton("Stop", nil)
	restartButton := widget.NewButton("Restart", nil)

	detailsButton.Disable()
	logsButton.Disable()
	startButton.Disable()
	stopButton.Disable()
	restartButton.Disable()

	updateActionButtons := func() {
		hasSelection := selectedIndex >= 0 && selectedIndex < len(filteredContainers)
		if hasSelection {
			detailsButton.Enable()
			logsButton.Enable()
			startButton.Enable()
			stopButton.Enable()
			restartButton.Enable()
			return
		}

		detailsButton.Disable()
		logsButton.Disable()
		startButton.Disable()
		stopButton.Disable()
		restartButton.Disable()
	}

	getSelectedContainer := func() (*system.DockerContainerInfo, bool) {
		if selectedIndex < 0 || selectedIndex >= len(filteredContainers) {
			return nil, false
		}

		c := filteredContainers[selectedIndex]
		return &c, true
	}

	filterContainers := func(query string) []system.DockerContainerInfo {
		query = strings.ToLower(strings.TrimSpace(query))
		if query == "" {
			return allContainers
		}

		var result []system.DockerContainerInfo
		for _, c := range allContainers {
			if strings.Contains(strings.ToLower(c.Names), query) ||
				strings.Contains(strings.ToLower(c.Image), query) ||
				strings.Contains(strings.ToLower(c.State), query) ||
				strings.Contains(strings.ToLower(c.Status), query) {
				result = append(result, c)
			}
		}

		return result
	}

	containerList := widget.NewList(
		func() int {
			return len(filteredContainers)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("container-name")

			stateText := canvas.NewText("state", color.White)
			stateText.Alignment = fyne.TextAlignTrailing

			return container.NewBorder(nil, nil, nil, stateText, nameLabel)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(filteredContainers) {
				return
			}

			c := filteredContainers[id]

			border := obj.(*fyne.Container)
			nameLabel := border.Objects[0].(*widget.Label)
			stateText := border.Objects[1].(*canvas.Text)

			nameLabel.SetText(fmt.Sprintf("%s (%s)", c.Names, c.Image))
			stateText.Text = c.State

			switch strings.ToLower(c.State) {
			case "running":
				stateText.Color = color.RGBA{0, 200, 0, 255}
			case "exited":
				stateText.Color = color.RGBA{150, 150, 150, 255}
			case "paused":
				stateText.Color = color.RGBA{200, 200, 0, 255}
			default:
				stateText.Color = color.RGBA{200, 120, 0, 255}
			}

			stateText.Refresh()
		},
	)

	refreshList := func() {
		selectedIndex = -1
		filteredContainers = filterContainers(searchEntry.Text)
		containerList.UnselectAll()
		containerList.Refresh()
		updateActionButtons()

		if len(filteredContainers) == 0 {
			statusLabel.SetText("Статус: 0 результатов")
			return
		}

		statusLabel.SetText(fmt.Sprintf("Статус: %d контейнеров", len(filteredContainers)))
	}

	showErrorState := func(err error) {
		allContainers = nil
		filteredContainers = nil
		selectedIndex = -1
		containerList.UnselectAll()
		containerList.Refresh()
		updateActionButtons()
		statusLabel.SetText("Статус: ошибка загрузки — " + err.Error())
	}

	refreshContainers := func() {
		containers, err := system.ListDockerContainers()
		if err != nil {
			fyne.Do(func() {
				showErrorState(err)
			})
			return
		}

		fyne.Do(func() {
			allContainers = containers
			refreshList()
		})
	}

	showContainerDetails := func(c system.DockerContainerInfo) {
		content := container.NewVBox(
			widget.NewLabel("Name: "+c.Names),
			widget.NewLabel("Image: "+c.Image),
			widget.NewLabel("State: "+c.State),
			widget.NewLabel("Status: "+c.Status),
			widget.NewLabel("ID: "+c.ID),
		)

		dialog.ShowCustom(
			"Container Details",
			"Закрыть",
			container.NewPadded(content),
			parent,
		)
	}

	runContainerAction := func(action string) {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText("Выбери контейнер из списка")
			updateActionButtons()
			return
		}

		message := fmt.Sprintf("Выполнить %s для контейнера %s?", strings.ToUpper(action), c.Names)

		dialog.ShowConfirm("Подтверждение", message, func(confirmed bool) {
			if !confirmed {
				return
			}

			statusLabel.SetText(fmt.Sprintf("Статус: выполняется %s для %s...", action, c.Names))
			detailsButton.Disable()
			logsButton.Disable()
			startButton.Disable()
			stopButton.Disable()
			restartButton.Disable()

			go func(containerName string) {
				err := system.ControlDockerContainer(action, containerName)
				if err != nil {
					fyne.Do(func() {
						dialog.ShowError(err, parent)
						statusLabel.SetText("Статус: ошибка выполнения")
						updateActionButtons()
					})
					return
				}

				fyne.Do(func() {
					dialog.ShowInformation(
						"Готово",
						fmt.Sprintf("Команда %s для %s выполнена.", strings.ToUpper(action), containerName),
						parent,
					)
				})

				refreshContainers()
			}(c.Names)
		}, parent)
	}

	showContainerLogs := func(containerName string) {
		logWindow := fyne.CurrentApp().NewWindow("Docker Logs: " + containerName)
		logWindow.Resize(fyne.NewSize(900, 600))

		logEntry := widget.NewMultiLineEntry()
		logEntry.Wrapping = fyne.TextWrapOff
		logEntry.Disable()

		infoLabel := widget.NewLabel("Логи ещё не загружены")
		autoRefreshCheck := widget.NewCheck("Auto refresh (2 сек)", nil)

		stopAutoRefresh := make(chan struct{})

		loadLogs := func() {
			infoLabel.SetText("Загрузка логов...")

			go func() {
				logs, err := system.GetDockerContainerLogs(containerName, 200)
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

		logWindow.SetOnClosed(func() {
			close(stopAutoRefresh)
		})

		content := container.NewBorder(
			container.NewVBox(
				widget.NewLabel("Logs for " + containerName),
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

	containerList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		updateActionButtons()
	}

	containerList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
		updateActionButtons()
	}

	searchEntry.OnChanged = func(text string) {
		refreshList()
	}

	detailsButton.OnTapped = func() {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText("Выбери контейнер из списка")
			updateActionButtons()
			return
		}

		showContainerDetails(*c)
	}

	logsButton.OnTapped = func() {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText("Выбери контейнер из списка")
			updateActionButtons()
			return
		}

		showContainerLogs(c.Names)
	}

	startButton.OnTapped = func() {
		runContainerAction("start")
	}

	stopButton.OnTapped = func() {
		runContainerAction("stop")
	}

	restartButton.OnTapped = func() {
		runContainerAction("restart")
	}

	refreshButton := widget.NewButton("Обновить", func() {
		go refreshContainers()
	})

	actionsRow := container.NewHBox(
		refreshButton,
		detailsButton,
		logsButton,
		startButton,
		stopButton,
		restartButton,
	)

	content := container.NewPadded(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
			searchEntry,
			actionsRow,
			statusLabel,
			widget.NewSeparator(),
			container.NewVScroll(containerList),
		),
	)

	go refreshContainers()

	return content
}