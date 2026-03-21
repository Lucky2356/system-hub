package ui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildServicesTab(parent fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Services")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Просмотр и управление systemd-сервисами")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск (например: ssh, docker...)")

	statusLabel := widget.NewLabel("Статус: ожидание")

	var allServices []system.ServiceInfo
	var filteredServices []system.ServiceInfo
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
		hasSelection := selectedIndex >= 0 && selectedIndex < len(filteredServices)
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

	getSelectedService := func() (*system.ServiceInfo, bool) {
		if selectedIndex < 0 || selectedIndex >= len(filteredServices) {
			return nil, false
		}

		svc := filteredServices[selectedIndex]
		return &svc, true
	}

	filterServices := func(query string) []system.ServiceInfo {
		query = strings.ToLower(strings.TrimSpace(query))
		if query == "" {
			return allServices
		}

		var result []system.ServiceInfo
		for _, svc := range allServices {
			if strings.Contains(strings.ToLower(svc.Name), query) ||
				strings.Contains(strings.ToLower(svc.Description), query) {
				result = append(result, svc)
			}
		}

		return result
	}

	serviceList := widget.NewList(
		func() int {
			return len(filteredServices)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("service")

			stateText := canvas.NewText("state", color.White)
			stateText.Alignment = fyne.TextAlignTrailing

			return container.NewBorder(nil, nil, nil, stateText, nameLabel)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(filteredServices) {
				return
			}

			svc := filteredServices[id]

			border := obj.(*fyne.Container)
			nameLabel := border.Objects[0].(*widget.Label)
			stateText := border.Objects[1].(*canvas.Text)

			nameLabel.SetText(svc.Name)
			stateText.Text = svc.ActiveState

			switch svc.ActiveState {
			case "active":
				stateText.Color = color.RGBA{0, 200, 0, 255}
			case "failed":
				stateText.Color = color.RGBA{200, 0, 0, 255}
			case "activating":
				stateText.Color = color.RGBA{200, 200, 0, 255}
			default:
				stateText.Color = color.RGBA{150, 150, 150, 255}
			}

			stateText.Refresh()
		},
	)

	refreshList := func() {
		selectedIndex = -1
		filteredServices = filterServices(searchEntry.Text)
		serviceList.UnselectAll()
		serviceList.Refresh()
		updateActionButtons()

		if len(filteredServices) == 0 {
			statusLabel.SetText("Статус: 0 результатов")
			return
		}

		statusLabel.SetText(fmt.Sprintf("Статус: %d сервисов", len(filteredServices)))
	}

	showErrorState := func(err error) {
		allServices = nil
		filteredServices = nil
		selectedIndex = -1
		serviceList.UnselectAll()
		serviceList.Refresh()
		updateActionButtons()
		statusLabel.SetText("Статус: ошибка загрузки — " + err.Error())
	}

	refreshServices := func() {
		services, err := system.ListServices()
		if err != nil {
			fyne.Do(func() {
				showErrorState(err)
			})
			return
		}

		fyne.Do(func() {
			allServices = services
			refreshList()
		})
	}

	showServiceDetails := func(svc system.ServiceInfo) {
		description := widget.NewLabel(svc.Description)
		description.Wrapping = fyne.TextWrapWord

		content := container.NewVBox(
			widget.NewLabel("Name: "+svc.Name),
			widget.NewLabel("Load: "+svc.LoadState),
			widget.NewLabel("Active: "+svc.ActiveState),
			widget.NewLabel("Sub: "+svc.SubState),
			widget.NewSeparator(),
			widget.NewLabel("Description:"),
			description,
		)

		dialog.ShowCustom(
			"Service Details",
			"Закрыть",
			container.NewPadded(content),
			parent,
		)
	}

	runServiceAction := func(action string) {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText("Выбери сервис из списка")
			updateActionButtons()
			return
		}

		message := fmt.Sprintf("Выполнить %s для %s?", strings.ToUpper(action), svc.Name)

		dialog.ShowConfirm("Подтверждение", message, func(confirmed bool) {
			if !confirmed {
				return
			}

			statusLabel.SetText(fmt.Sprintf("Статус: выполняется %s для %s...", action, svc.Name))
			detailsButton.Disable()
			logsButton.Disable()
			startButton.Disable()
			stopButton.Disable()
			restartButton.Disable()

			go func(serviceName string) {
				err := system.ControlService(action, serviceName)
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
						fmt.Sprintf("Команда %s для %s выполнена.", strings.ToUpper(action), serviceName),
						parent,
					)
				})

				refreshServices()
			}(svc.Name)
		}, parent)
	}

	serviceList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		updateActionButtons()
	}

	serviceList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
		updateActionButtons()
	}

	searchEntry.OnChanged = func(text string) {
		refreshList()
	}

	detailsButton.OnTapped = func() {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText("Выбери сервис из списка")
			updateActionButtons()
			return
		}

		showServiceDetails(*svc)
	}

	logsButton.OnTapped = func() {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText("Выбери сервис из списка")
			updateActionButtons()
			return
		}

		statusLabel.SetText("Загрузка логов...")

		go func(serviceName string) {
			logs, err := system.GetServiceLogs(serviceName, 100)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, parent)
					statusLabel.SetText("Ошибка загрузки логов")
				})
				return
			}

			fyne.Do(func() {
				logEntry := widget.NewMultiLineEntry()
				logEntry.SetText(logs)
				logEntry.Wrapping = fyne.TextWrapOff

				w := fyne.CurrentApp().NewWindow("Logs: " + serviceName)
				w.SetContent(container.NewPadded(
					container.NewVBox(
						widget.NewLabel("Logs for " + serviceName),
						widget.NewSeparator(),
						container.NewVScroll(logEntry),
					),
				))
				w.Resize(fyne.NewSize(800, 500))
				w.Show()

				statusLabel.SetText("Логи загружены")
			})
		}(svc.Name)
	}

	startButton.OnTapped = func() {
		runServiceAction("start")
	}

	stopButton.OnTapped = func() {
		runServiceAction("stop")
	}

	restartButton.OnTapped = func() {
		runServiceAction("restart")
	}

	refreshButton := widget.NewButton("Обновить", func() {
		go refreshServices()
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
			container.NewVScroll(serviceList),
		),
	)

	go refreshServices()

	return content
}