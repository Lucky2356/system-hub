package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildDockerTab(parent fyne.Window, cfg config.Config) fyne.CanvasObject {
	title := widget.NewLabel("Docker")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Просмотр и управление Docker-контейнерами")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск контейнера (например: nginx, postgres...)")

	sortSelect := widget.NewSelect([]string{"Name", "Status"}, nil)
	sortSelect.SetSelected("Name")

	statusLabel := widget.NewLabel("Статус: ожидание")

	var allContainers []system.DockerContainerInfo
	var filteredContainers []system.DockerContainerInfo
	selectedIndex := -1
	lastSelectedContainerName := ""

	autoRefreshCheck := widget.NewCheck(
		fmt.Sprintf("Auto refresh (%d сек)", appstate.Config.RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.DockerAutoRefresh)
	stopAutoRefresh := make(chan struct{})
	autoRefreshStarted := false

	detailsButton := widget.NewButton("Подробнее", nil)
	logsButton := widget.NewButton("Logs", nil)
	startButton := widget.NewButton("Start", nil)
	stopButton := widget.NewButton("Stop", nil)
	restartButton := widget.NewButton("Restart", nil)
	favoriteButton := widget.NewButton("☆", nil)

	detailsButton.Disable()
	logsButton.Disable()
	startButton.Disable()
	stopButton.Disable()
	restartButton.Disable()
	favoriteButton.Disable()

	updateActionButtons := func() {
		hasSelection := selectedIndex >= 0 && selectedIndex < len(filteredContainers)
		if hasSelection {
			detailsButton.Enable()
			logsButton.Enable()
			startButton.Enable()
			stopButton.Enable()
			restartButton.Enable()
			favoriteButton.Enable()

			c := filteredContainers[selectedIndex]
			if isFavoriteContainer(c.Names) {
				favoriteButton.SetText("★")
			} else {
				favoriteButton.SetText("☆")
			}
			return
		}

		detailsButton.Disable()
		logsButton.Disable()
		startButton.Disable()
		stopButton.Disable()
		restartButton.Disable()
		favoriteButton.Disable()
		favoriteButton.SetText("☆")
	}

	getSelectedContainer := func() (*system.DockerContainerInfo, bool) {
		if selectedIndex < 0 || selectedIndex >= len(filteredContainers) {
			return nil, false
		}

		c := filteredContainers[selectedIndex]
		return &c, true
	}

	sortContainers := func(items []system.DockerContainerInfo) {
		switch sortSelect.Selected {
		case "Status":
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].State == items[j].State {
					return items[i].Names < items[j].Names
				}
				return items[i].State < items[j].State
			})
		default:
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].Names < items[j].Names
			})
		}
	}

	filterContainers := func(query string) []system.DockerContainerInfo {
		query = strings.ToLower(strings.TrimSpace(query))

		var result []system.DockerContainerInfo
		for _, c := range allContainers {
			if query == "" ||
				strings.Contains(strings.ToLower(c.Names), query) ||
				strings.Contains(strings.ToLower(c.Image), query) ||
				strings.Contains(strings.ToLower(c.State), query) ||
				strings.Contains(strings.ToLower(c.Status), query) {
				result = append(result, c)
			}
		}

		sortContainers(result)
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

			prefix := ""
			if isFavoriteContainer(c.Names) {
				prefix = "★ "
			}

			nameLabel.SetText(fmt.Sprintf("%s%s (%s)", prefix, c.Names, c.Image))
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
		filteredContainers = filterContainers(searchEntry.Text)

		restoredIndex := -1
		if lastSelectedContainerName != "" {
			for i, c := range filteredContainers {
				if c.Names == lastSelectedContainerName {
					restoredIndex = i
					break
				}
			}
		}

		selectedIndex = restoredIndex
		containerList.Refresh()

		if restoredIndex >= 0 {
			containerList.Select(restoredIndex)
		} else {
			containerList.UnselectAll()
		}

		updateActionButtons()

		if len(filteredContainers) == 0 {
			statusLabel.SetText("Статус: 0 результатов")
			return
		}

		statusLabel.SetText(
			fmt.Sprintf(
				"Статус: %d контейнеров | обновлено %s",
				len(filteredContainers),
				time.Now().Format("15:04:05"),
			),
		)
	}

	showErrorState := func(err error) {
		allContainers = nil
		filteredContainers = nil
		selectedIndex = -1
		lastSelectedContainerName = ""
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
			favoriteButton.Disable()

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
		showLogsWindow(
			"Docker Logs: "+containerName,
			"Logs for container "+containerName,
			func() (string, error) {
				return system.GetDockerContainerLogs(containerName, cfg.DefaultLogLines)
			},
			cfg,
		)
	}

	containerList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		if id >= 0 && id < len(filteredContainers) {
			lastSelectedContainerName = filteredContainers[id].Names
		}
		updateActionButtons()
	}

	containerList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
		lastSelectedContainerName = ""
		updateActionButtons()
	}

	searchEntry.OnChanged = func(string) {
		refreshList()
	}

	sortSelect.OnChanged = func(string) {
		refreshList()
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
			ticker := time.NewTicker(time.Duration(appstate.Config.RefreshIntervalSeconds) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if autoRefreshCheck.Checked {
						go refreshContainers()
					}
				case <-stopAutoRefresh:
					return
				}
			}
		}()
	}

	if cfg.DockerAutoRefresh {
		autoRefreshCheck.OnChanged(true)
	}

	favoriteButton.OnTapped = func() {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText("Выбери контейнер из списка")
			updateActionButtons()
			return
		}

		if err := toggleFavoriteContainer(c.Names); err != nil {
			ShowError(parent, err)
			statusLabel.SetText("Статус: ошибка сохранения избранного")
			return
		}

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
		autoRefreshCheck,
		favoriteButton,
		detailsButton,
		logsButton,
		startButton,
		stopButton,
		restartButton,
	)

	content := container.NewBorder(
		container.NewPadded(
			container.NewVBox(
				title,
				subtitle,
				widget.NewSeparator(),
				searchEntry,
				sortSelect,
				actionsRow,
				statusLabel,
				widget.NewSeparator(),
			),
		),
		nil,
		nil,
		nil,
		container.NewPadded(containerList),
	)

	go refreshContainers()

	return content
}