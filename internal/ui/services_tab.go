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

	var allServices []system.ServiceInfo
	var filteredServices []system.ServiceInfo
	var selectedIndex int = -1

	serviceList := widget.NewList(
		func() int {
			return len(filteredServices)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("service")
			stateLabel := widget.NewLabel("state")
			stateLabel.Alignment = fyne.TextAlignTrailing

			return container.NewBorder(nil, nil, nil, stateLabel, nameLabel)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(filteredServices) {
				return
			}

			svc := filteredServices[id]

			border := obj.(*fyne.Container)
			nameLabel := border.Objects[0].(*widget.Label)
			stateLabel := border.Objects[1].(*widget.Label)

			nameLabel.SetText(svc.Name)
			stateLabel.SetText(svc.ActiveState)
		},
	)

	serviceList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
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

	refreshList := func() {
		filteredServices = filterServices(searchEntry.Text)
		serviceList.Refresh()

		if len(filteredServices) == 0 {
			statusLabel.SetText("Статус: 0 результатов")
			return
		}

		statusLabel.SetText(fmt.Sprintf("Статус: %d сервисов", len(filteredServices)))
	}

	showError := func(err error) {
		allServices = nil
		filteredServices = nil
		serviceList.Refresh()
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

		fyne.Do(func() {
			allServices = services
			refreshList()
		})
	}

	// 🔍 поиск
	searchEntry.OnChanged = func(text string) {
		refreshList()
	}

	// 📄 окно деталей
	openDetailsWindow := func(svc system.ServiceInfo) {
		w := fyne.CurrentApp().NewWindow("Service Details")

		content := container.NewVBox(
			widget.NewLabelWithStyle("Service Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),

			widget.NewLabel("Name: " + svc.Name),
			widget.NewLabel("Load: " + svc.LoadState),
			widget.NewLabel("Active: " + svc.ActiveState),
			widget.NewLabel("Sub: " + svc.SubState),

			widget.NewSeparator(),
			widget.NewLabel("Description:"),
			widget.NewLabel(svc.Description),
		)

		w.SetContent(container.NewPadded(content))
		w.Resize(fyne.NewSize(400, 300))
		w.Show()
	}

	// 🔘 кнопка "Подробнее"
	detailsButton := widget.NewButton("Подробнее", func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredServices) {
			statusLabel.SetText("Выбери сервис из списка")
			return
		}

		openDetailsWindow(filteredServices[selectedIndex])
	})

	refreshButton := widget.NewButton("Обновить", func() {
		go refreshServices()
	})

	content := container.NewPadded(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),

			searchEntry,
			container.NewHBox(refreshButton, detailsButton),

			statusLabel,
			widget.NewSeparator(),

			container.NewVScroll(serviceList),
		),
	)

	go refreshServices()

	return content
}