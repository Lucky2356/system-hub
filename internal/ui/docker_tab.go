package ui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildDockerTab() fyne.CanvasObject {
	title := widget.NewLabel("Docker")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Просмотр Docker-контейнеров")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск контейнера (например: nginx, postgres...)")

	statusLabel := widget.NewLabel("Статус: ожидание")

	var allContainers []system.DockerContainerInfo
	var filteredContainers []system.DockerContainerInfo

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
		filteredContainers = filterContainers(searchEntry.Text)
		containerList.Refresh()

		if len(filteredContainers) == 0 {
			statusLabel.SetText("Статус: 0 результатов")
			return
		}

		statusLabel.SetText(fmt.Sprintf("Статус: %d контейнеров", len(filteredContainers)))
	}

	showErrorState := func(err error) {
		allContainers = nil
		filteredContainers = nil
		containerList.Refresh()
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

	searchEntry.OnChanged = func(text string) {
		refreshList()
	}

	refreshButton := widget.NewButton("Обновить", func() {
		go refreshContainers()
	})

	content := container.NewPadded(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
			searchEntry,
			refreshButton,
			statusLabel,
			widget.NewSeparator(),
			container.NewVScroll(containerList),
		),
	)

	go refreshContainers()

	return content
}