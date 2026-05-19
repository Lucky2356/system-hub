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
	"github.com/Lucky2356/system-hub/internal/activity"

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

	modeSelect := widget.NewSelect([]string{"Containers", "Images"}, nil)
	modeSelect.SetSelected("Containers")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск контейнера (например: nginx, postgres...)")

	statusFilter := widget.NewSelect([]string{"All", "Running", "Exited", "Paused"}, nil)
	statusFilter.SetSelected("All")

	sortSelect := widget.NewSelect([]string{"Name", "Status"}, nil)
	sortSelect.SetSelected("Name")

	statusLabel := widget.NewLabel("Статус: ожидание")

	var allContainers []system.DockerContainerInfo
	var filteredContainers []system.DockerContainerInfo
	var allImages []system.DockerImageInfo
	var filteredImages []system.DockerImageInfo
	selectedIndex := -1
	lastSelectedContainerName := ""

	autoRefreshCheck := widget.NewCheck(
		fmt.Sprintf("Auto refresh (%d сек)", appstate.GetConfig().RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.DockerAutoRefresh)
	stopAutoRefresh := make(chan struct{})
	autoRefreshStarted := false

	detailsButton := widget.NewButton("Подробнее", nil)
	inspectButton := widget.NewButton("Inspect", nil)
	logsButton := widget.NewButton("Logs", nil)
	startButton := widget.NewButton("Start", nil)
	stopButton := widget.NewButton("Stop", nil)
	restartButton := widget.NewButton("Restart", nil)
	favoriteButton := widget.NewButton("☆", nil)

	pullButton := widget.NewButton("Pull image", nil)
	removeImageButton := widget.NewButton("Remove image", nil)
	removeImageButton.Disable()

	hideContainerButtons := func() {
		detailsButton.Hide()
		inspectButton.Hide()
		logsButton.Hide()
		startButton.Hide()
		stopButton.Hide()
		restartButton.Hide()
		favoriteButton.Hide()
		pullButton.Show()
		removeImageButton.Show()
	}

	showContainerButtons := func() {
		detailsButton.Show()
		inspectButton.Show()
		logsButton.Show()
		startButton.Show()
		stopButton.Show()
		restartButton.Show()
		favoriteButton.Show()
		pullButton.Hide()
		removeImageButton.Hide()
	}

	isImagesMode := func() bool {
		return modeSelect.Selected == "Images"
	}

	detailsButton.Disable()
	inspectButton.Disable()
	logsButton.Disable()
	startButton.Disable()
	stopButton.Disable()
	restartButton.Disable()
	favoriteButton.Disable()

	updateActionButtons := func() {
		if isImagesMode() {
			hasSelection := selectedIndex >= 0 && selectedIndex < len(filteredImages)
			removeImageButton.Disable()
			if hasSelection {
				removeImageButton.Enable()
			}
			return
		}

		hasSelection := selectedIndex >= 0 && selectedIndex < len(filteredContainers)
		if hasSelection {
			detailsButton.Enable()
			inspectButton.Enable()
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
		inspectButton.Disable()
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
		status := statusFilter.Selected

		var result []system.DockerContainerInfo
		for _, c := range allContainers {
			if status != "All" && !strings.EqualFold(c.State, status) {
				continue
			}

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

	filterImages := func(query string) []system.DockerImageInfo {
		query = strings.ToLower(strings.TrimSpace(query))

		var result []system.DockerImageInfo
		for _, img := range allImages {
			if query == "" ||
				strings.Contains(strings.ToLower(img.Repository), query) ||
				strings.Contains(strings.ToLower(img.Tag), query) ||
				strings.Contains(strings.ToLower(img.ID), query) {
				result = append(result, img)
			}
		}
		return result
	}

	containerList := widget.NewList(
		func() int {
			if isImagesMode() {
				return len(filteredImages)
			}
			return len(filteredContainers)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("")
			stateText := canvas.NewText("", color.White)
			stateText.Alignment = fyne.TextAlignTrailing
			return container.NewBorder(nil, nil, nil, stateText, nameLabel)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 {
				return
			}

			row := obj.(*fyne.Container)
			nameLabel := row.Objects[0].(*widget.Label)
			stateText := row.Objects[1].(*canvas.Text)

			if isImagesMode() {
				if id >= len(filteredImages) {
					return
				}
				img := filteredImages[id]
				nameLabel.SetText(fmt.Sprintf("%s:%s", img.Repository, img.Tag))
				stateText.Text = fmt.Sprintf("%s  %s", img.ID[:12], img.Size)
				stateText.Color = color.RGBA{R: 100, G: 180, B: 255, A: 255}
			} else {
				if id >= len(filteredContainers) {
					return
				}
				c := filteredContainers[id]

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
			}

			stateText.Refresh()
		},
	)

	refreshList := func() {
		if isImagesMode() {
			filteredImages = filterImages(searchEntry.Text)
		} else {
			filteredContainers = filterContainers(searchEntry.Text)
		}

		restoredIndex := -1
		if !isImagesMode() && lastSelectedContainerName != "" {
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

		count := 0
		if isImagesMode() {
			count = len(filteredImages)
		} else {
			count = len(filteredContainers)
		}

		if count == 0 {
			statusLabel.SetText("Статус: 0 результатов")
			return
		}

		label := "образов"
		if !isImagesMode() {
			label = "контейнеров"
		}

		statusLabel.SetText(
			fmt.Sprintf(
				"Статус: %d %s | обновлено %s",
				count, label,
				time.Now().Format("15:04:05"),
			),
		)
	}

	showErrorState := func(err error) {
		if isImagesMode() {
			allImages = nil
			filteredImages = nil
		} else {
			allContainers = nil
			filteredContainers = nil
		}
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

	refreshImages := func() {
		images, err := system.ListDockerImages()
		if err != nil {
			fyne.Do(func() {
				showErrorState(err)
			})
			return
		}

		fyne.Do(func() {
			allImages = images
			refreshList()
		})
	}

	refreshData := func() {
		if isImagesMode() {
			go refreshImages()
		} else {
			go refreshContainers()
		}
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
					activity.Add("docker", action, containerName, "failed", err.Error())

					fyne.Do(func() {
						if system.IsPermissionError(err) {
							ShowErrorMsg(parent, system.BuildPermissionHint("docker", action, containerName))
						} else {
							ShowError(parent, err)
						}

						statusLabel.SetText("Статус: ошибка выполнения")
						updateActionButtons()
					})
					return
				}

				activity.Add("docker", action, containerName, "success", "")

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
		if !isImagesMode() && id >= 0 && id < len(filteredContainers) {
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

	statusFilter.OnChanged = func(string) {
		refreshList()
	}

	sortSelect.OnChanged = func(string) {
		refreshList()
	}

	modeSelect.OnChanged = func(mode string) {
		switch mode {
		case "Images":
			subtitle.SetText("Просмотр и управление Docker-образами")
			searchEntry.SetPlaceHolder("Поиск образа (например: nginx, ubuntu...)")
			statusFilter.Hide()
			sortSelect.Hide()
			hideContainerButtons()
		default:
			subtitle.SetText("Просмотр и управление Docker-контейнерами")
			searchEntry.SetPlaceHolder("Поиск контейнера (например: nginx, postgres...)")
			statusFilter.Show()
			sortSelect.Show()
			showContainerButtons()
		}
		selectedIndex = -1
		containerList.UnselectAll()
		refreshData()
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
						go refreshData()
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

	pullButton.OnTapped = func() {
		imageEntry := widget.NewEntry()
		imageEntry.SetPlaceHolder("nginx:latest, ubuntu:22.04...")

		dialog.ShowCustomConfirm(
			"Pull Docker Image",
			"Pull",
			"Cancel",
			container.NewPadded(container.NewVBox(
				widget.NewLabel("Введите имя образа для загрузки:"),
				imageEntry,
			)),
			func(confirmed bool) {
				if !confirmed || strings.TrimSpace(imageEntry.Text) == "" {
					return
				}

				imageName := strings.TrimSpace(imageEntry.Text)
				statusLabel.SetText(fmt.Sprintf("Статус: загрузка %s...", imageName))
				pullButton.Disable()

				go func() {
					err := system.PullDockerImage(imageName)

					fyne.Do(func() {
						pullButton.Enable()
						if err != nil {
							ShowError(parent, err)
							statusLabel.SetText("Статус: ошибка загрузки")
							return
						}

						activity.Add("docker", "pull", imageName, "success", "")
						dialog.ShowInformation("Готово", "Образ "+imageName+" загружен.", parent)
						refreshImages()
					})
				}()
			},
			parent,
		)
	}

	removeImageButton.OnTapped = func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredImages) {
			return
		}
		img := filteredImages[selectedIndex]
		imageName := img.Repository + ":" + img.Tag

		dialog.ShowConfirm(
			"Remove image",
			fmt.Sprintf("Удалить образ %s?", imageName),
			func(confirmed bool) {
				if !confirmed {
					return
				}

				statusLabel.SetText(fmt.Sprintf("Статус: удаление %s...", imageName))
				removeImageButton.Disable()

				go func() {
					err := system.RemoveDockerImage(imageName)

					fyne.Do(func() {
						removeImageButton.Enable()
						if err != nil {
							ShowError(parent, err)
							statusLabel.SetText("Статус: ошибка удаления")
							return
						}

						activity.Add("docker", "rmi", imageName, "success", "")
						statusLabel.SetText("Статус: образ удалён")
						refreshImages()
					})
				}()
			},
			parent,
		)
	}

	favoriteButton.OnTapped = func() {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText("Выбери контейнер из списка")
			updateActionButtons()
			return
		}

		if err := toggleFavoriteContainer(c.Names); err != nil {
			activity.Add("docker", "favorite", c.Names, "failed", err.Error())
			ShowError(parent, err)
			statusLabel.SetText("Статус: ошибка сохранения избранного")
			return
		}

		if isFavoriteContainer(c.Names) {
			activity.Add("docker", "favorite", c.Names, "success", "added to favorites")
		} else {
			activity.Add("docker", "favorite", c.Names, "success", "removed from favorites")
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

	inspectButton.OnTapped = func() {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText("Выбери контейнер из списка")
			updateActionButtons()
			return
		}

		statusLabel.SetText("Статус: загрузка docker inspect...")

		go func(containerName string) {
			result, err := system.GetDockerContainerInspect(containerName)

			fyne.Do(func() {
				if err != nil {
					if system.IsPermissionError(err) {
						ShowErrorMsg(parent, system.BuildPermissionHint("docker", "inspect", containerName))
					} else {
						ShowError(parent, err)
					}
					statusLabel.SetText("Статус: ошибка docker inspect")
					return
				}

				output := widget.NewMultiLineEntry()
				output.SetText(result)
				output.Wrapping = fyne.TextWrapWord
				output.Disable()

				dialog.ShowCustom(
					"Docker Inspect: "+containerName,
					"Закрыть",
					container.NewPadded(output),
					parent,
				)

				statusLabel.SetText("Статус: docker inspect загружен")
			})
		}(c.Names)
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
		go refreshData()
	})

	actionsRow := container.NewHBox(
		refreshButton,
		autoRefreshCheck,
		favoriteButton,
		detailsButton,
		inspectButton,
		logsButton,
		startButton,
		stopButton,
		restartButton,
		pullButton,
		removeImageButton,
	)

	hideContainerButtons()

	content := container.NewBorder(
		container.NewPadded(
			container.NewVBox(
				title,
				subtitle,
				widget.NewSeparator(),
				modeSelect,
				container.NewGridWithColumns(2, searchEntry, statusFilter),
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
	RegisterRefresh("Docker", refreshData)

	return content
}
