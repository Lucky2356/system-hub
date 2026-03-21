package ui

import (
	"fmt"
	"strings"

	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildServicesTab() fyne.CanvasObject {
	title := widget.NewLabel("Services")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Просмотр systemd-сервисов")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск (например: ssh, docker...)")

	statusLabel := widget.NewLabel("Статус: ожидание")

	serviceListContainer := container.NewVBox(
		widget.NewLabel("Список сервисов пока не загружен."),
	)

	scroll := container.NewVScroll(serviceListContainer)
	scroll.SetMinSize(fyne.NewSize(800, 380))

	var allServices []system.ServiceInfo

	filterServices := func(query string) []system.ServiceInfo {
		if query == "" {
			return allServices
		}

		query = strings.ToLower(query)

		var filtered []system.ServiceInfo

		for _, svc := range allServices {
			if strings.Contains(strings.ToLower(svc.Name), query) ||
				strings.Contains(strings.ToLower(svc.Description), query) {
				filtered = append(filtered, svc)
			}
		}

		return filtered
	}

	updateServicesUI := func(services []system.ServiceInfo) {
		serviceListContainer.Objects = nil

		if len(services) == 0 {
			serviceListContainer.Add(widget.NewLabel("Ничего не найдено."))
			serviceListContainer.Refresh()
			statusLabel.SetText("Статус: 0 результатов")
			return
		}

		for _, svc := range services {
			serviceTitle := fmt.Sprintf("%s [%s]", svc.Name, svc.ActiveState)
			serviceDetails := fmt.Sprintf(
				"Load: %s | Active: %s | Sub: %s\n%s",
				svc.LoadState,
				svc.ActiveState,
				svc.SubState,
				svc.Description,
			)

			card := widget.NewCard(
				serviceTitle,
				"",
				widget.NewLabel(serviceDetails),
			)

			serviceListContainer.Add(card)
		}

		serviceListContainer.Refresh()
		statusLabel.SetText(fmt.Sprintf("Статус: %d сервисов", len(services)))
	}

	showError := func(err error) {
		serviceListContainer.Objects = []fyne.CanvasObject{
			widget.NewCard(
				"Ошибка",
				"",
				widget.NewLabel(err.Error()),
			),
		}
		serviceListContainer.Refresh()
		statusLabel.SetText("Статус: ошибка загрузки")
	}

	refreshServices := func() {
		services, err := system.ListServices()
		if err != nil {
			fyne.Do(func() {
				showError(err)
			})
			return
		}

		allServices = services

		fyne.Do(func() {
			updateServicesUI(allServices)
		})
	}

	// 🔥 Фильтр при вводе
	searchEntry.OnChanged = func(text string) {
		filtered := filterServices(text)
		updateServicesUI(filtered)
	}

	refreshButton := widget.NewButton("Обновить", func() {
		go refreshServices()
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
			scroll,
		),
	)

	go refreshServices()

	return content
}