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
	title := widget.NewLabel("Процессы")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Просмотр процессов и прослушиваемых портов")

	modeSelect := widget.NewSelect([]string{"Топ процессов", "Прослушиваемые порты"}, nil)
	modeSelect.SetSelected("Топ процессов")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск по имени, PID, порту...")

	sortSelect := widget.NewSelect([]string{"CPU", "Память", "PID", "Имя"}, nil)
	sortSelect.SetSelected("CPU")

	statusLabel := widget.NewLabel("Статус: ожидание")

	autoRefreshCheck := widget.NewCheck(
		fmt.Sprintf("Автообновление (%d сек)", appstate.GetConfig().RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(false)

	var allProcs []system.ProcessUsageInfo
	var filteredProcs []system.ProcessUsageInfo
	var allPorts []system.PortProcessInfo
	var filteredPorts []system.PortProcessInfo
	selectedIndex := -1
	autoRefreshStarted := false
	stopAutoRefresh := make(chan struct{})

	detailsButton := widget.NewButton("Подробнее", nil)
	detailsButton.Disable()
	killButton := widget.NewButton("Завершить", nil)
	killButton.Disable()

	isPortsMode := func() bool {
		return modeSelect.Selected == "Прослушиваемые порты"
	}

	updateButtons := func() {
		if isPortsMode() {
			if selectedIndex >= 0 && selectedIndex < len(filteredPorts) {
				detailsButton.Enable()
				killButton.Enable()
				return
			}
		} else {
			if selectedIndex >= 0 && selectedIndex < len(filteredProcs) {
				detailsButton.Enable()
				killButton.Enable()
				return
			}
		}
		detailsButton.Disable()
		killButton.Disable()
	}

	getSelectedPID := func() int32 {
		if selectedIndex < 0 {
			return 0
		}
		if isPortsMode() {
			if selectedIndex < len(filteredPorts) {
				return filteredPorts[selectedIndex].PID
			}
		} else {
			if selectedIndex < len(filteredProcs) {
				return filteredProcs[selectedIndex].PID
			}
		}
		return 0
	}

	getSelectedName := func() string {
		if selectedIndex < 0 {
			return ""
		}
		if isPortsMode() {
			if selectedIndex < len(filteredPorts) {
				return filteredPorts[selectedIndex].ProcessName
			}
		} else {
			if selectedIndex < len(filteredProcs) {
				return filteredProcs[selectedIndex].ProcessName
			}
		}
		return ""
	}

	list := widget.NewList(
		func() int {
			if isPortsMode() {
				return len(filteredPorts)
			}
			return len(filteredProcs)
		},
		func() fyne.CanvasObject {
			left := widget.NewLabel("")
			right := canvas.NewText("", color.White)
			right.Alignment = fyne.TextAlignTrailing
			return container.NewBorder(nil, nil, nil, right, left)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			row := obj.(*fyne.Container)
			left := row.Objects[0].(*widget.Label)
			right := row.Objects[1].(*canvas.Text)

			if isPortsMode() {
				if id < 0 || id >= len(filteredPorts) {
					return
				}
				item := filteredPorts[id]
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
			} else {
				if id < 0 || id >= len(filteredProcs) {
					return
				}
				p := filteredProcs[id]
				cpuStr := fmt.Sprintf("%.1f%%", p.CPUPercent)
				memStr := system.FormatBytes(p.MemoryBytes)
				left.SetText(fmt.Sprintf("%s  |  CPU: %s  Mem: %s", p.ProcessName, cpuStr, memStr))
				right.Text = fmt.Sprintf("pid=%d", p.PID)
				right.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255}
			}
			right.Refresh()
		},
	)

	refreshList := func() {
		if isPortsMode() {
			filteredPorts = filterPorts(searchEntry.Text, allPorts, sortSelect.Selected)
		} else {
			filteredProcs = filterProcs(searchEntry.Text, allProcs, sortSelect.Selected)
		}
		list.Refresh()
		updateButtons()

		count := 0
		if isPortsMode() {
			count = len(filteredPorts)
		} else {
			count = len(filteredProcs)
		}

		statusLabel.SetText(
			fmt.Sprintf(
				"Статус: %d записей | обновлено %s",
				count,
				time.Now().Format("15:04:05"),
			),
		)
	}

	refreshData := func() {
		if isPortsMode() {
			items, err := system.ListListeningPorts()
			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("Статус: ошибка загрузки — " + err.Error())
					allPorts = nil
					filteredPorts = nil
					selectedIndex = -1
					list.Refresh()
					updateButtons()
				})
				return
			}
			fyne.Do(func() {
				allPorts = items
				refreshList()
			})
		} else {
			procs, err := system.ListTopProcesses(200)
			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("Статус: ошибка загрузки — " + err.Error())
					allProcs = nil
					filteredProcs = nil
					selectedIndex = -1
					list.Refresh()
					updateButtons()
				})
				return
			}
			fyne.Do(func() {
				allProcs = procs
				refreshList()
			})
		}
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

	modeSelect.OnChanged = func(mode string) {
		switch mode {
		case "Топ процессов":
			sortSelect.Options = []string{"CPU", "Память", "PID", "Имя"}
			sortSelect.SetSelected("CPU")
			subtitle.SetText("Просмотр процессов по нагрузке на CPU")
		case "Прослушиваемые порты":
			sortSelect.Options = []string{"Порт", "Процесс", "PID"}
			sortSelect.SetSelected("Порт")
			subtitle.SetText("Просмотр прослушиваемых портов и процессов")
		}
		selectedIndex = -1
		list.UnselectAll()
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

	detailsButton.OnTapped = func() {
		if isPortsMode() {
			if selectedIndex < 0 || selectedIndex >= len(filteredPorts) {
				statusLabel.SetText("Выбери запись из списка")
				updateButtons()
				return
			}
			item := filteredPorts[selectedIndex]
			details := system.GetPortProcessDetails(item)
			dialog.ShowCustom(
				"Сведения о порте и процессе",
				"Закрыть",
				container.NewPadded(widget.NewLabel(details)),
				parent,
			)
		} else {
			if selectedIndex < 0 || selectedIndex >= len(filteredProcs) {
				statusLabel.SetText("Выбери процесс из списка")
				updateButtons()
				return
			}
			p := filteredProcs[selectedIndex]
			details := fmt.Sprintf(
				"Процесс: %s\nPID: %d\nCPU: %.1f%%\nПамять: %s (%.1f%%)",
				p.ProcessName, p.PID, p.CPUPercent,
				system.FormatBytes(p.MemoryBytes), p.MemoryPercent,
			)
			dialog.ShowCustom(
				"Сведения о процессе",
				"Закрыть",
				container.NewPadded(widget.NewLabel(details)),
				parent,
			)
		}
	}

	killButton.OnTapped = func() {
		pid := getSelectedPID()
		name := getSelectedName()
		if pid <= 0 {
			statusLabel.SetText("Невозможно завершить процесс")
			return
		}

		dialog.ShowConfirm(
			"Подтверждение",
			fmt.Sprintf("Завершить процесс %s (PID %d)?", name, pid),
			func(confirm bool) {
				if !confirm {
					return
				}
				go func() {
					err := system.KillProcess(pid)
					fyne.Do(func() {
						if err != nil {
							if system.IsPermissionError(err) {
								ShowErrorMsg(parent, "Недостаточно прав для завершения процесса")
							} else {
								ShowError(parent, err)
							}
							return
						}
						statusLabel.SetText("Процесс завершён")
						refreshData()
					})
				}()
			},
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
		killButton,
	)

	content := container.NewBorder(
		container.NewPadded(
			container.NewVBox(
				title,
				subtitle,
				widget.NewSeparator(),
				modeSelect,
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

	modeSelect.SetSelected("Топ процессов")
	RegisterRefresh("Processes", refreshData)

	return content
}

func filterPorts(query string, items []system.PortProcessInfo, sortBy string) []system.PortProcessInfo {
	query = strings.ToLower(strings.TrimSpace(query))

	result := make([]system.PortProcessInfo, 0)

	for _, item := range items {
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

	switch sortBy {
	case "Процесс":
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

func filterProcs(query string, items []system.ProcessUsageInfo, sortBy string) []system.ProcessUsageInfo {
	query = strings.ToLower(strings.TrimSpace(query))

	result := make([]system.ProcessUsageInfo, 0)

	for _, p := range items {
		pidText := strconv.Itoa(int(p.PID))

		if query == "" ||
			strings.Contains(strings.ToLower(p.ProcessName), query) ||
			strings.Contains(pidText, query) {
			result = append(result, p)
		}
	}

	switch sortBy {
	case "Память":
		sort.SliceStable(result, func(i, j int) bool {
			if result[i].MemoryBytes == result[j].MemoryBytes {
				return result[i].ProcessName < result[j].ProcessName
			}
			return result[i].MemoryBytes > result[j].MemoryBytes
		})
	case "PID":
		sort.SliceStable(result, func(i, j int) bool {
			if result[i].PID == result[j].PID {
				return result[i].ProcessName < result[j].ProcessName
			}
			return result[i].PID < result[j].PID
		})
	case "Имя":
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].ProcessName < result[j].ProcessName
		})
	default:
		sort.SliceStable(result, func(i, j int) bool {
			if result[i].CPUPercent == result[j].CPUPercent {
				return result[i].ProcessName < result[j].ProcessName
			}
			return result[i].CPUPercent > result[j].CPUPercent
		})
	}

	return result
}
