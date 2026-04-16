package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildProcessesTab(parent fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Processes / Ports")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Просмотр listening ports и процессов")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск: 80, 443, nginx, 5432, PID...")

	sortSelect := widget.NewSelect([]string{"Port", "Process", "PID"}, nil)
	sortSelect.SetSelected("Port")

	statusLabel := widget.NewLabel("Статус: ожидание")

	autoRefreshCheck := widget.NewCheck(
		fmt.Sprintf("Auto refresh (%d сек)", appstate.Config.RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(false)

	var allItems []system.PortProcessInfo
	var filteredItems []system.PortProcessInfo
	selectedIndex := -1
	autoRefreshStarted := false
	stopAutoRefresh := make(chan struct{})

	detailsButton := widget.NewButton("Details", nil)
	detailsButton.Disable()

	getSelectedItem := func() (*system.PortProcessInfo, bool) {
		if selectedIndex < 0 || selectedIndex >= len(filteredItems) {
			return nil, false
		}
		item := filteredItems[selectedIndex]
		return &item, true
	}

	updateButtons := func() {
		if selectedIndex >= 0 && selectedIndex < len(filteredItems) {
			detailsButton.Enable()
			return
		}
		detailsButton.Disable()
	}

	filterItems := func(query string) []system.PortProcessInfo {
		query = strings.ToLower(strings.TrimSpace(query))

		result := make([]system.PortProcessInfo, 0)

		for _, item := range allItems {
			portText := strconv.Itoa(int(item.LocalPort))
			pidText := strconv.Itoa(int(item.PID))

			if query == "" ||
				strings.Contains(strings.ToLower(item.ProcessName), query) ||
				strings.Contains(strings.ToLower(item.LocalAddr), query) ||
				strings.Contains(strings.ToLower(item.Proto), query) ||
				strings.Contains(strings.ToLower(item.Status), query) ||
				strings.Contains(portText, query) ||
				strings.Contains(pidText, query) {
				result = append(result, item)
			}
		}

		switch sortSelect.Selected {
		case "Process":
			sort.SliceStable(result, func(i, j int) bool {
				if result[i].ProcessName == result[j].ProcessName {
					return result[i].LocalPort < result[j].LocalPort
				}
				return result[i].ProcessName < result[j].ProcessName
			})
		case "PID":
			sort.SliceStable(result, func(i, j int) bool {
				if result[i].PID == result[j].PID {
					return result[i].LocalPort < result[j].LocalPort
				}
				return result[i].PID < result[j].PID
			})
		default:
			sort.SliceStable(result, func(i, j int) bool {
				if result[i].LocalPort == result[j].LocalPort {
					return result[i].ProcessName < result[j].ProcessName
				}
				return result[i].LocalPort < result[j].LocalPort
			})
		}

		return result
	}

	list := widget.NewList(
		func() int {
			return len(filteredItems)
		},
		func() fyne.CanvasObject {
			left := widget.NewLabel("tcp :80")
			right := canvas.NewText("pid=123 nginx", color.White)
			right.Alignment = fyne.TextAlignTrailing
			return container.NewBorder(nil, nil, nil, right, left)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(filteredItems) {
				return
			}

			item := filteredItems[id]

			row := obj.(*fyne.Container)
			left := row.Objects[0].(*widget.Label)
			right := row.Objects[1].(*canvas.Text)

			left.SetText(fmt.Sprintf("%s %s:%d", item.Proto, item.LocalAddr, item.LocalPort))
			right.Text = fmt.Sprintf("pid=%d  %s", item.PID, item.ProcessName)

			switch strings.ToLower(item.Proto) {
			case "tcp":
				right.Color = color.RGBA{R: 80, G: 160, B: 255, A: 255}
			case "udp":
				right.Color = color.RGBA{R: 180, G: 120, B: 255, A: 255}
			default:
				right.Color = color.RGBA{R: 180, G: 180, B: 180, A: 255}
			}

			right.Refresh()
		},
	)

	refreshList := func() {
		filteredItems = filterItems(searchEntry.Text)
		list.Refresh()
		updateButtons()

		statusLabel.SetText(
			fmt.Sprintf(
				"Статус: %d записей | обновлено %s",
				len(filteredItems),
				time.Now().Format("15:04:05"),
			),
		)
	}

	refreshData := func() {
		items, err := system.ListListeningPorts()
		if err != nil {
			fyne.Do(func() {
				statusLabel.SetText("Статус: ошибка загрузки — " + err.Error())
				allItems = nil
				filteredItems = nil
				selectedIndex = -1
				list.Refresh()
				updateButtons()
			})
			return
		}

		fyne.Do(func() {
			allItems = items
			refreshList()
		})
	}

	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		updateButtons()
	}

	list.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
		updateButtons()
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
						go refreshData()
					}
				case <-stopAutoRefresh:
					return
				}
			}
		}()
	}

	detailsButton.OnTapped = func() {
		item, ok := getSelectedItem()
		if !ok {
			statusLabel.SetText("Выбери запись из списка")
			updateButtons()
			return
		}

		details := system.GetPortProcessDetails(*item)

		dialog.ShowCustom(
			"Port / Process Details",
			"Закрыть",
			container.NewPadded(widget.NewLabel(details)),
			parent,
		)
	}

	refreshButton := widget.NewButton("Обновить", func() {
		go refreshData()
	})

	toolbar := container.NewHBox(
		refreshButton,
		autoRefreshCheck,
		detailsButton,
	)

	content := container.NewBorder(
		container.NewPadded(
			container.NewVBox(
				title,
				subtitle,
				widget.NewSeparator(),
				searchEntry,
				sortSelect,
				toolbar,
				statusLabel,
				widget.NewSeparator(),
			),
		),
		nil,
		nil,
		nil,
		container.NewPadded(list),
	)

	go refreshData()

	return content
}