package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/i18n"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Sort keys for filterProcs/filterPorts. The Select shows translated labels, so
// the sort criterion is carried as one of these stable keys instead.
const (
	sortKeyCPU     = "cpu"
	sortKeyMemory  = "memory"
	sortKeyPID     = "pid"
	sortKeyName    = "name"
	sortKeyPort    = "port"
	sortKeyProcess = "process"
)

func buildProcessesTab(parent fyne.Window) fyne.CanvasObject {
	// Sidebar names the section; the caption flips with the processes/ports mode.
	subtitle := widget.NewLabel(i18n.T("View processes and listening ports"))

	modeProcesses, modePorts := i18n.T("Top processes"), i18n.T("Listening ports")
	modeSelect := widget.NewSelect([]string{modeProcesses, modePorts}, nil)
	modeSelect.SetSelected(modeProcesses)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(i18n.T("Search by name, PID, port..."))

	// Translated label -> stable sort key.
	sortKeys := map[string]string{
		"CPU":             sortKeyCPU,
		i18n.T("Memory"):  sortKeyMemory,
		"PID":             sortKeyPID,
		i18n.T("Name"):    sortKeyName,
		i18n.T("Port"):    sortKeyPort,
		i18n.T("Process"): sortKeyProcess,
	}
	procSortLabels := []string{"CPU", i18n.T("Memory"), "PID", i18n.T("Name")}
	portSortLabels := []string{i18n.T("Port"), i18n.T("Process"), "PID"}

	sortSelect := widget.NewSelect(procSortLabels, nil)
	sortSelect.SetSelected(procSortLabels[0])

	selectedSortKey := func() string { return sortKeys[sortSelect.Selected] }

	statusLabel := newDataLabel(i18n.T("Status: waiting"))

	autoRefreshCheck := widget.NewCheck(
		i18n.Tf("Auto-refresh (%d s)", appstate.GetConfig().RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(false)

	var allProcs []system.ProcessUsageInfo
	var filteredProcs []system.ProcessUsageInfo
	var allPorts []system.PortProcessInfo
	var filteredPorts []system.PortProcessInfo
	selectedIndex := -1

	detailsButton := widget.NewButton(i18n.T("Details"), nil)
	detailsButton.Disable()
	killButton := widget.NewButton(i18n.T("Kill"), nil)
	killButton.Disable()

	isPortsMode := func() bool {
		return modeSelect.Selected == modePorts
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

	// Assigned once the actions below exist; row callbacks read them at click
	// time.
	var onRowMenu func(id widget.ListItemID, e *fyne.PointEvent)
	var onRowActivate func(id widget.ListItemID)

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
			return newTappableRow(container.NewBorder(nil, nil, nil, right, left))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			tr := obj.(*tappableRow)
			row := tr.content.(*fyne.Container)
			left := row.Objects[0].(*widget.Label)
			right := row.Objects[1].(*canvas.Text)

			rowID := id
			tr.onSecondary = func(e *fyne.PointEvent) { onRowMenu(rowID, e) }
			tr.onDouble = func() { onRowActivate(rowID) }

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
			filteredPorts = filterPorts(searchEntry.Text, allPorts, selectedSortKey())
		} else {
			filteredProcs = filterProcs(searchEntry.Text, allProcs, selectedSortKey())
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
			i18n.Tf(
				"Status: %d entries | updated %s",
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
					statusLabel.SetText(i18n.Tf("Status: load error — %s", err.Error()))
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
					statusLabel.SetText(i18n.Tf("Status: load error — %s", err.Error()))
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
		case modeProcesses:
			sortSelect.Options = procSortLabels
			sortSelect.SetSelected(procSortLabels[0])
			subtitle.SetText(i18n.T("View processes by CPU load"))
		case modePorts:
			sortSelect.Options = portSortLabels
			sortSelect.SetSelected(portSortLabels[0])
			subtitle.SetText(i18n.T("View listening ports and their processes"))
		}
		selectedIndex = -1
		list.UnselectAll()
		refreshData()
	}

	autoRefresh := newAutoRefresher(refreshData)
	RegisterCloser(autoRefresh.Stop)
	autoRefreshCheck.OnChanged = autoRefresh.SetEnabled

	detailsButton.OnTapped = func() {
		if isPortsMode() {
			if selectedIndex < 0 || selectedIndex >= len(filteredPorts) {
				statusLabel.SetText(i18n.T("Select an entry from the list"))
				updateButtons()
				return
			}
			item := filteredPorts[selectedIndex]
			details := system.GetPortProcessDetails(item)
			dialog.ShowCustom(
				i18n.T("Port and process details"),
				i18n.T("Close"),
				container.NewPadded(widget.NewLabel(details)),
				parent,
			)
		} else {
			if selectedIndex < 0 || selectedIndex >= len(filteredProcs) {
				statusLabel.SetText(i18n.T("Select a process from the list"))
				updateButtons()
				return
			}
			p := filteredProcs[selectedIndex]
			details := i18n.Tf(
				"Process: %s\nPID: %d\nCPU: %.1f%%\nMemory: %s (%.1f%%)",
				p.ProcessName, p.PID, p.CPUPercent,
				system.FormatBytes(p.MemoryBytes), p.MemoryPercent,
			)
			dialog.ShowCustom(
				i18n.T("Process details"),
				i18n.T("Close"),
				container.NewPadded(widget.NewLabel(details)),
				parent,
			)
		}
	}

	killButton.OnTapped = func() {
		pid := getSelectedPID()
		name := getSelectedName()
		if pid <= 0 {
			statusLabel.SetText(i18n.T("This process cannot be killed"))
			return
		}

		dialog.ShowConfirm(
			i18n.T("Confirmation"),
			i18n.Tf("Kill process %s (PID %d)?", name, pid),
			func(confirm bool) {
				if !confirm {
					return
				}
				go func() {
					err := system.KillProcess(pid)
					fyne.Do(func() {
						if err != nil {
							switch {
							case system.IsSelfKill(err):
								ShowErrorMsg(parent, i18n.T("That is System Hub itself — closing it would just quit the app."))
							case system.IsPermissionError(err):
								ShowErrorMsg(parent, i18n.T("Not enough permissions to kill the process"))
							default:
								ShowError(parent, err)
							}
							return
						}
						statusLabel.SetText(i18n.T("Process killed"))
						refreshData()
					})
				}()
			},
			parent,
		)
	}

	onRowMenu = func(id widget.ListItemID, e *fyne.PointEvent) {
		list.Select(id)
		showRowMenu(parent, e, []rowAction{
			{label: i18n.T("Details"), do: func() { detailsButton.OnTapped() }},
			{label: i18n.T("Kill"), do: func() { killButton.OnTapped() }},
		})
	}

	onRowActivate = func(id widget.ListItemID) {
		list.Select(id)
		detailsButton.OnTapped()
	}

	refreshButton := widget.NewButton(i18n.T("Refresh"), func() {
		startInitialLoad(refreshData)
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

	modeSelect.SetSelected(modeProcesses)
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
	case sortKeyProcess:
		sort.SliceStable(result, func(i, j int) bool {
			if result[i].ProcessName == result[j].ProcessName {
				return result[i].LocalPort < result[j].LocalPort
			}
			return result[i].ProcessName < result[j].ProcessName
		})
	case sortKeyPID:
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
	case sortKeyMemory:
		sort.SliceStable(result, func(i, j int) bool {
			if result[i].MemoryBytes == result[j].MemoryBytes {
				return result[i].ProcessName < result[j].ProcessName
			}
			return result[i].MemoryBytes > result[j].MemoryBytes
		})
	case sortKeyPID:
		sort.SliceStable(result, func(i, j int) bool {
			if result[i].PID == result[j].PID {
				return result[i].ProcessName < result[j].ProcessName
			}
			return result[i].PID < result[j].PID
		})
	case sortKeyName:
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
