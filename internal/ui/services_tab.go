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

func buildServicesTab(parent fyne.Window, cfg config.Config) fyne.CanvasObject {
	title := widget.NewLabel("Сервисы")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Просмотр и управление systemd-сервисами")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск (например: ssh, docker...)")

	// Русские подписи фильтра сопоставляются со значениями ActiveState systemd.
	statusFilterValues := map[string]string{
		"Все":         "",
		"Активные":    "active",
		"Неактивные":  "inactive",
		"Сбойные":     "failed",
		"Запускаются": "activating",
	}
	statusFilter := widget.NewSelect([]string{"Все", "Активные", "Неактивные", "Сбойные", "Запускаются"}, nil)
	statusFilter.SetSelected("Все")

	sortSelect := widget.NewSelect([]string{"Имя", "Статус"}, nil)
	sortSelect.SetSelected("Имя")

	statusLabel := widget.NewLabel("Статус: ожидание")

	var allServices []system.ServiceInfo
	var filteredServices []system.ServiceInfo
	selectedIndex := -1
	lastSelectedServiceName := ""

	autoRefreshCheck := widget.NewCheck(
		fmt.Sprintf("Автообновление (%d сек)", appstate.GetConfig().RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.ServicesAutoRefresh)

	detailsButton := widget.NewButton("Подробнее", nil)
	openUnitButton := widget.NewButton("Юнит-файл", nil)
	logsButton := widget.NewButton("Логи", nil)
	startButton := widget.NewButton("Старт", nil)
	stopButton := widget.NewButton("Стоп", nil)
	restartButton := widget.NewButton("Рестарт", nil)
	enableButton := widget.NewButton("Включить", nil)
	disableButton := widget.NewButton("Отключить", nil)
	favoriteButton := widget.NewButton("☆", nil)

	detailsButton.Disable()
	openUnitButton.Disable()
	logsButton.Disable()
	startButton.Disable()
	stopButton.Disable()
	restartButton.Disable()
	enableButton.Disable()
	disableButton.Disable()
	favoriteButton.Disable()

	updateActionButtons := func() {
		hasSelection := selectedIndex >= 0 && selectedIndex < len(filteredServices)
		if hasSelection {
			detailsButton.Enable()
			openUnitButton.Enable()
			logsButton.Enable()
			startButton.Enable()
			stopButton.Enable()
			restartButton.Enable()
			enableButton.Enable()
			disableButton.Enable()
			favoriteButton.Enable()

			svc := filteredServices[selectedIndex]
			if isFavoriteService(svc.Name) {
				favoriteButton.SetText("★")
			} else {
				favoriteButton.SetText("☆")
			}
			return
		}

		detailsButton.Disable()
		openUnitButton.Disable()
		logsButton.Disable()
		startButton.Disable()
		stopButton.Disable()
		restartButton.Disable()
		enableButton.Disable()
		disableButton.Disable()
		favoriteButton.Disable()
		favoriteButton.SetText("☆")
	}

	getSelectedService := func() (*system.ServiceInfo, bool) {
		if selectedIndex < 0 || selectedIndex >= len(filteredServices) {
			return nil, false
		}

		svc := filteredServices[selectedIndex]
		return &svc, true
	}

	sortServices := func(items []system.ServiceInfo) {
		switch sortSelect.Selected {
		case "Статус":
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].ActiveState == items[j].ActiveState {
					return items[i].Name < items[j].Name
				}
				return items[i].ActiveState < items[j].ActiveState
			})
		default:
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].Name < items[j].Name
			})
		}
	}

	filterServices := func(query string) []system.ServiceInfo {
		query = strings.ToLower(strings.TrimSpace(query))
		status := statusFilterValues[statusFilter.Selected]

		var result []system.ServiceInfo
		for _, svc := range allServices {
			if status != "" && !strings.EqualFold(svc.ActiveState, status) {
				continue
			}

			if query == "" ||
				strings.Contains(strings.ToLower(svc.Name), query) ||
				strings.Contains(strings.ToLower(svc.Description), query) {
				result = append(result, svc)
			}
		}

		sortServices(result)
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

			prefix := ""
			if isFavoriteService(svc.Name) {
				prefix = "★ "
			}

			nameLabel.SetText(prefix + svc.Name)
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
		filteredServices = filterServices(searchEntry.Text)

		restoredIndex := -1
		if lastSelectedServiceName != "" {
			for i, svc := range filteredServices {
				if svc.Name == lastSelectedServiceName {
					restoredIndex = i
					break
				}
			}
		}

		selectedIndex = restoredIndex
		serviceList.Refresh()

		if restoredIndex >= 0 {
			serviceList.Select(restoredIndex)
		} else {
			serviceList.UnselectAll()
		}

		updateActionButtons()

		if len(filteredServices) == 0 {
			statusLabel.SetText("Статус: 0 результатов")
			return
		}

		statusLabel.SetText(
			fmt.Sprintf(
				"Статус: %d сервисов | обновлено %s",
				len(filteredServices),
				time.Now().Format("15:04:05"),
			),
		)
	}

	showErrorState := func(err error) {
		allServices = nil
		filteredServices = nil
		selectedIndex = -1
		lastSelectedServiceName = ""
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
			widget.NewLabel("Имя: "+svc.Name),
			widget.NewLabel("Load-состояние: "+svc.LoadState),
			widget.NewLabel("Active-состояние: "+svc.ActiveState),
			widget.NewLabel("Sub-состояние: "+svc.SubState),
			widget.NewSeparator(),
			widget.NewLabel("Описание:"),
			description,
		)

		dialog.ShowCustom(
			"Сведения о сервисе",
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
			enableButton.Disable()
			disableButton.Disable()
			favoriteButton.Disable()

			go func(serviceName string) {
				err := system.ControlService(action, serviceName)
				if err != nil {
					activity.Add("service", action, serviceName, "failed", err.Error())

					fyne.Do(func() {
						if system.IsPermissionError(err) {
							ShowErrorMsg(parent, system.BuildPermissionHint("service", action, serviceName))
						} else {
							ShowError(parent, err)
						}

						statusLabel.SetText("Статус: ошибка выполнения")
						updateActionButtons()
					})
					return
				}
				activity.Add("service", action, serviceName, "success", "")

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
		if id >= 0 && id < len(filteredServices) {
			lastSelectedServiceName = filteredServices[id].Name
		}
		updateActionButtons()
	}

	serviceList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
		lastSelectedServiceName = ""
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

	autoRefresh := newAutoRefresher(refreshServices)
	RegisterCloser(autoRefresh.Stop)
	autoRefreshCheck.OnChanged = autoRefresh.SetEnabled
	autoRefresh.SetEnabled(cfg.ServicesAutoRefresh)

	favoriteButton.OnTapped = func() {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText("Выбери сервис из списка")
			updateActionButtons()
			return
		}

		if err := toggleFavoriteService(svc.Name); err != nil {
			activity.Add("service", "favorite", svc.Name, "failed", err.Error())
			ShowError(parent, err)
			statusLabel.SetText("Статус: ошибка сохранения избранного")
			return
		}

		if isFavoriteService(svc.Name) {
			activity.Add("service", "favorite", svc.Name, "success", "added to favorites")
		} else {
			activity.Add("service", "favorite", svc.Name, "success", "removed from favorites")
		}

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

	openUnitButton.OnTapped = func() {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText("Выбери сервис из списка")
			updateActionButtons()
			return
		}

		unitPath, err := system.FindServiceUnitFile(svc.Name)
		if err != nil {
			ShowError(parent, err)
			statusLabel.SetText("Статус: юнит-файл не найден")
			return
		}

		if err := OpenFileInFiles(unitPath); err != nil {
			ShowError(parent, err)
			statusLabel.SetText("Статус: не удалось открыть юнит-файл")
			return
		}

		dialog.ShowInformation(
			"Юнит-файл открыт",
			"Файл открыт во вкладке «Файлы»:\n"+unitPath,
			parent,
		)
		statusLabel.SetText("Статус: юнит-файл открыт")
	}

	logsButton.OnTapped = func() {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText("Выбери сервис из списка")
			updateActionButtons()
			return
		}

		statusLabel.SetText("Открытие окна логов...")

		showLogsWindow(
			"Логи сервиса: "+svc.Name,
			"Логи сервиса "+svc.Name,
			func() (string, error) {
				return system.GetServiceLogs(svc.Name, cfg.DefaultLogLines)
			},
			cfg,
		)

		statusLabel.SetText("Окно логов открыто")
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

	enableButton.OnTapped = func() {
		runServiceAction("enable")
	}

	disableButton.OnTapped = func() {
		runServiceAction("disable")
	}

	refreshButton := widget.NewButton("Обновить", func() {
		go refreshServices()
	})

	actionsRow := container.NewHBox(
		refreshButton,
		autoRefreshCheck,
		favoriteButton,
		detailsButton,
		openUnitButton,
		logsButton,
		startButton,
		stopButton,
		restartButton,
		enableButton,
		disableButton,
	)

	content := container.NewBorder(
		container.NewPadded(
			container.NewVBox(
				title,
				subtitle,
				widget.NewSeparator(),
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
		container.NewPadded(serviceList),
	)

	go refreshServices()
	RegisterRefresh("Services", refreshServices)

	return content
}
