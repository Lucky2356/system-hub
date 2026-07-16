package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/i18n"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func buildDashboardTab(parent fyne.Window, cfg config.Config) fyne.CanvasObject {
	// The app name lives in the window header; the tab only needs a caption.
	subtitle := widget.NewLabel(i18n.T("System overview: CPU / RAM / disk / network / services / Docker"))

	cpuValueLabel := widget.NewLabel("CPU: ...")
	cpuBar := widget.NewProgressBar()

	ramValueLabel := widget.NewLabel("RAM: ...")
	ramDetailsLabel := widget.NewLabel("...")
	ramBar := widget.NewProgressBar()

	diskValueLabel := widget.NewLabel(i18n.T("Disk") + ": ...")
	diskDetailsLabel := widget.NewLabel("...")
	diskBar := widget.NewProgressBar()

	uptimeLabel := widget.NewLabel(i18n.T("Uptime") + ": ...")
	systemdLabel := widget.NewLabel(system.ServiceManagerName() + ": ...")
	dockerLabel := widget.NewLabel("Docker: ...")

	serviceCountLabel := widget.NewLabel(i18n.T("Services") + ": ...")
	dockerCountLabel := widget.NewLabel(i18n.T("Containers") + ": ...")
	dockerRunningLabel := widget.NewLabel(i18n.T("Running") + ": ...")

	favoriteServicesBox := container.NewVBox(widget.NewLabel(i18n.T("No favorite services")))
	favoriteContainersBox := container.NewVBox(widget.NewLabel(i18n.T("No favorite containers")))
	problemsLabel := widget.NewLabel(i18n.T("Checking for problems..."))
	topProcessesLabel := widget.NewLabel(i18n.T("Loading the process list..."))
	topProcessesLabel.Wrapping = fyne.TextWrapWord

	recvLabel := widget.NewLabel("RX: ...")
	sentLabel := widget.NewLabel("TX: ...")

	// CPU and RAM are percentages, so they are pinned to a 0-100 scale: a flat
	// 5% load must look flat. Network has no ceiling, so it scales to its own
	// peak.
	cpuChart := newSparkline(AccentColor, 100)
	ramChart := newSparkline(SuccessColor, 100)
	netRecvChart := newSparkline(AccentColor, 0)
	netSentChart := newSparkline(WarningColor, 0)

	perCPUBars := container.NewVBox(widget.NewLabel(i18n.T("Loading per-core data...")))

	tempLabel := widget.NewLabel(i18n.T("Loading temperatures..."))
	tempLabel.Wrapping = fyne.TextWrapWord
	fanLabel := widget.NewLabel("")
	fanLabel.Wrapping = fyne.TextWrapWord
	voltageLabel := widget.NewLabel("")
	voltageLabel.Wrapping = fyne.TextWrapWord

	problemsLabel.Wrapping = fyne.TextWrapWord
	topProcessesLabel.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel(i18n.T("Status: waiting"))

	makeStatusText := func(status string) *canvas.Text {
		text := canvas.NewText(status, StatusColor(status))
		text.TextSize = 12
		text.TextStyle = fyne.TextStyle{Bold: true}
		return text
	}

	// refreshStats is assigned below; the closures here capture it by reference.
	var refreshStats func(heavy bool)

	refreshButtonTapped := func() {
		go refreshStats(true)
	}

	buildFavoriteServiceRows := func() []fyne.CanvasObject {
		if len(appstate.GetConfig().FavoriteServices) == 0 {
			return []fyne.CanvasObject{widget.NewLabel(i18n.T("No favorite services"))}
		}

		services, err := system.ListServices()
		if err != nil {
			return []fyne.CanvasObject{widget.NewLabel(i18n.T("Could not load favorite services"))}
		}

		serviceMap := make(map[string]system.ServiceInfo, len(services))
		for _, svc := range services {
			serviceMap[svc.Name] = svc
		}

		runServiceAction := func(action, serviceName string) {
			dialog.ShowConfirm(
				i18n.T("Confirmation"),
				i18n.Tf("Run %s for service %s?", strings.ToUpper(action), serviceName),
				func(ok bool) {
					if !ok {
						return
					}

					go func() {
						err := system.ControlService(action, serviceName)

						fyne.Do(func() {
							if err != nil {
								if system.IsPermissionError(err) {
									ShowErrorMsg(parent, system.BuildPermissionHint("service", action, serviceName))
								} else {
									ShowError(parent, err)
								}
								return
							}

							refreshButtonTapped()
						})
					}()
				},
				parent,
			)
		}

		rows := make([]fyne.CanvasObject, 0, len(appstate.GetConfig().FavoriteServices))

		for _, name := range appstate.GetConfig().FavoriteServices {
			svc, ok := serviceMap[name]

			statusText := i18n.T("unavailable")
			if ok {
				statusText = svc.ActiveState
			}

			nameLabel := widget.NewLabel(name)
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}

			statusTextObj := makeStatusText(statusText)

			infoBox := container.NewVBox(
				nameLabel,
				statusTextObj,
			)

			logsButton := widget.NewButton(i18n.T("Logs"), func(serviceName string) func() {
				return func() {
					showLogsWindow(
						i18n.Tf("Service logs: %s", serviceName),
						i18n.Tf("Logs for service %s", serviceName),
						func() (string, error) {
							return system.GetServiceLogs(serviceName, appstate.GetConfig().DefaultLogLines)
						},
						appstate.GetConfig(),
					)
				}
			}(name))

			unitButton := widget.NewButton(i18n.T("Unit"), func(serviceName string) func() {
				return func() {
					unitPath, err := system.FindServiceUnitFile(serviceName)
					if err != nil {
						ShowError(parent, err)
						return
					}

					if err := OpenFileInFiles(unitPath); err != nil {
						ShowError(parent, err)
						return
					}

					dialog.ShowInformation(
						i18n.T("Unit file opened"),
						i18n.Tf("The file is open in the «Files» tab:\n%s", unitPath),
						parent,
					)
				}
			}(name))

			startButton := widget.NewButton(i18n.T("Start"), func(serviceName string) func() {
				return func() {
					runServiceAction("start", serviceName)
				}
			}(name))

			stopButton := widget.NewButton(i18n.T("Stop"), func(serviceName string) func() {
				return func() {
					runServiceAction("stop", serviceName)
				}
			}(name))

			restartButton := widget.NewButton(i18n.T("Restart"), func(serviceName string) func() {
				return func() {
					runServiceAction("restart", serviceName)
				}
			}(name))

			row := container.NewVBox(
				container.NewBorder(
					nil,
					nil,
					infoBox,
					nil,
					nil,
				),
				container.NewHBox(logsButton, unitButton, startButton, stopButton, restartButton),
				widget.NewSeparator(),
			)

			rows = append(rows, row)
		}

		return rows
	}

	buildFavoriteContainerRows := func() []fyne.CanvasObject {
		if len(appstate.GetConfig().FavoriteContainers) == 0 {
			return []fyne.CanvasObject{widget.NewLabel(i18n.T("No favorite containers"))}
		}

		containers, err := system.ListDockerContainers()
		if err != nil {
			return []fyne.CanvasObject{widget.NewLabel(i18n.T("Could not load favorite containers"))}
		}

		containerMap := make(map[string]system.DockerContainerInfo, len(containers))
		for _, c := range containers {
			containerMap[c.Names] = c
		}

		runContainerAction := func(action, containerName string) {
			dialog.ShowConfirm(
				i18n.T("Confirmation"),
				i18n.Tf("Run %s for container %s?", strings.ToUpper(action), containerName),
				func(ok bool) {
					if !ok {
						return
					}

					go func() {
						err := system.ControlDockerContainer(action, containerName)

						fyne.Do(func() {
							if err != nil {
								if system.IsPermissionError(err) {
									ShowErrorMsg(parent, system.BuildPermissionHint("docker", action, containerName))
								} else {
									ShowError(parent, err)
								}
								return
							}

							refreshButtonTapped()
						})
					}()
				},
				parent,
			)
		}

		rows := make([]fyne.CanvasObject, 0, len(appstate.GetConfig().FavoriteContainers))

		for _, name := range appstate.GetConfig().FavoriteContainers {
			c, ok := containerMap[name]

			statusText := i18n.T("unavailable")
			if ok {
				statusText = c.State
			}

			nameLabel := widget.NewLabel(name)
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}

			statusTextObj := makeStatusText(statusText)

			infoBox := container.NewVBox(
				nameLabel,
				statusTextObj,
			)

			logsButton := widget.NewButton(i18n.T("Logs"), func(containerName string) func() {
				return func() {
					showLogsWindow(
						i18n.Tf("Container logs: %s", containerName),
						i18n.Tf("Logs for container %s", containerName),
						func() (string, error) {
							return system.GetDockerContainerLogs(containerName, appstate.GetConfig().DefaultLogLines)
						},
						appstate.GetConfig(),
					)
				}
			}(name))

			inspectButton := widget.NewButton(i18n.T("Inspect"), func(containerName string) func() {
				return func() {
					go func() {
						result, err := system.GetDockerContainerInspect(containerName)

						fyne.Do(func() {
							if err != nil {
								if system.IsPermissionError(err) {
									ShowErrorMsg(parent, system.BuildPermissionHint("docker", "inspect", containerName))
								} else {
									ShowError(parent, err)
								}
								return
							}

							output := widget.NewMultiLineEntry()
							output.SetText(result)
							output.Wrapping = fyne.TextWrapWord
							output.Disable()

							dialog.ShowCustom(
								i18n.Tf("Container inspect: %s", containerName),
								i18n.T("Close"),
								container.NewPadded(output),
								parent,
							)
						})
					}()
				}
			}(name))

			startButton := widget.NewButton(i18n.T("Start"), func(containerName string) func() {
				return func() {
					runContainerAction("start", containerName)
				}
			}(name))

			stopButton := widget.NewButton(i18n.T("Stop"), func(containerName string) func() {
				return func() {
					runContainerAction("stop", containerName)
				}
			}(name))

			restartButton := widget.NewButton(i18n.T("Restart"), func(containerName string) func() {
				return func() {
					runContainerAction("restart", containerName)
				}
			}(name))

			row := container.NewVBox(
				container.NewBorder(
					nil,
					nil,
					infoBox,
					nil,
					nil,
				),
				container.NewHBox(logsButton, inspectButton, startButton, stopButton, restartButton),
				widget.NewSeparator(),
			)

			rows = append(rows, row)
		}

		return rows
	}

	// Per-core widgets are created once and updated in place on each tick;
	// rebuilding the whole widget tree every refresh causes needless layout
	// and GC churn.
	type coreRow struct {
		label *widget.Label
		bar   *widget.ProgressBar
	}
	var coreRows []coreRow

	updateUI := func(stats system.Stats, heavy bool) {
		// Recorded on every tick, including the light one: the numbers are
		// already in hand, so history costs no extra system calls.
		system.RecordSample(stats)
		samples := system.MetricHistory()
		updateCharts(samples, cpuChart, ramChart, netRecvChart, netSentChart)

		cpuValueLabel.SetText(fmt.Sprintf("CPU: %.1f%%", stats.CPUPercent))
		cpuBar.SetValue(stats.CPUPercent / 100)

		if len(coreRows) != len(stats.PerCPU) {
			perCPUBars.Objects = nil
			coreRows = coreRows[:0]
			for range stats.PerCPU {
				label := widget.NewLabel("")
				bar := widget.NewProgressBar()
				coreRows = append(coreRows, coreRow{label: label, bar: bar})
				perCPUBars.Add(container.NewVBox(label, bar))
			}
			if len(stats.PerCPU) == 0 {
				perCPUBars.Add(widget.NewLabel(i18n.T("No per-core data")))
			}
		}
		for i, core := range stats.PerCPU {
			coreRows[i].label.SetText(i18n.Tf("Core %d: %.1f%%", core.Core, core.Percent))
			coreRows[i].bar.SetValue(core.Percent / 100)
		}
		perCPUBars.Refresh()

		// The totals alone would not explain the chart below them, which plots a
		// rate; showing both makes the pairing legible.
		recvLabel.SetText(fmt.Sprintf("RX: %s  (%s)",
			system.FormatBytes(stats.NetRecv), formatRate(latestRecvRate(samples))))
		sentLabel.SetText(fmt.Sprintf("TX: %s  (%s)",
			system.FormatBytes(stats.NetSent), formatRate(latestSentRate(samples))))

		// Sensor readings spawn `sensors`/`nvidia-smi`; refresh them only on
		// the heavy tick (they are also carried on stats only when heavy).
		if heavy {
			if len(stats.Temperatures) == 0 {
				tempLabel.SetText(i18n.T("No temperature data"))
			} else {
				parts := make([]string, 0, len(stats.Temperatures))
				for _, s := range stats.Temperatures {
					parts = append(parts, fmt.Sprintf("• %s: %.0f%s", s.Name, s.Temp, s.Unit))
				}
				tempLabel.SetText(strings.Join(parts, "\n"))
			}

			fans := system.GetFanSpeeds()
			if len(fans) > 0 {
				parts := make([]string, 0, len(fans))
				for _, f := range fans {
					parts = append(parts, fmt.Sprintf("• %s: %.0f %s", f.Name, f.Speed, f.Unit))
				}
				fanLabel.SetText(strings.Join(parts, "\n"))
			} else {
				fanLabel.SetText("")
			}

			volts := system.GetVoltages()
			if len(volts) > 0 {
				parts := make([]string, 0, len(volts))
				for _, v := range volts {
					parts = append(parts, fmt.Sprintf("• %s: %.3f %s", v.Name, v.Value, v.Unit))
				}
				voltageLabel.SetText(strings.Join(parts, "\n"))
			} else {
				voltageLabel.SetText("")
			}
		}

		ramValueLabel.SetText(fmt.Sprintf("RAM: %.1f%%", stats.RAMPercent))
		ramDetailsLabel.SetText(fmt.Sprintf(
			"%s / %s",
			system.FormatBytes(stats.RAMUsed),
			system.FormatBytes(stats.RAMTotal),
		))
		ramBar.SetValue(stats.RAMPercent / 100)

		diskValueLabel.SetText(i18n.Tf("Disk: %.1f%%", stats.DiskPercent))
		diskDetailsLabel.SetText(fmt.Sprintf(
			"%s / %s",
			system.FormatBytes(stats.DiskUsed),
			system.FormatBytes(stats.DiskTotal),
		))
		diskBar.SetValue(stats.DiskPercent / 100)

		if stats.UptimeKnown {
			uptimeLabel.SetText(i18n.T("Uptime") + ": " + system.FormatUptime(stats.UptimeSeconds))
		} else {
			uptimeLabel.SetText(i18n.T("Uptime") + ": " + i18n.T("unavailable"))
		}

		if !heavy {
			statusLabel.SetText(i18n.T("Updated") + ": " + time.Now().Format("15:04:05"))
			return
		}

		if stats.ServiceManagerAvailable {
			systemdLabel.SetText(system.ServiceManagerName() + ": " + i18n.T("available"))
			if stats.ServiceCountKnown {
				serviceCountLabel.SetText(i18n.Tf("Services: %d", stats.ServiceCount))
			} else {
				serviceCountLabel.SetText(i18n.T("Services") + ": " + i18n.T("unavailable"))
			}
		} else {
			systemdLabel.SetText(system.ServiceManagerName() + ": " + i18n.T("unavailable"))
			serviceCountLabel.SetText(i18n.T("Services") + ": " + i18n.T("unavailable"))
		}

		if stats.DockerAvailable {
			dockerLabel.SetText("Docker: " + i18n.T("available"))
			if stats.DockerCountKnown {
				dockerCountLabel.SetText(i18n.Tf("Containers: %d", stats.DockerContainerCount))
				dockerRunningLabel.SetText(i18n.Tf("Running: %d", stats.DockerRunningCount))
			} else {
				dockerCountLabel.SetText(i18n.T("Containers") + ": " + i18n.T("unavailable"))
				dockerRunningLabel.SetText(i18n.T("Running") + ": " + i18n.T("unavailable"))
			}
		} else {
			dockerLabel.SetText("Docker: " + i18n.T("unavailable"))
			dockerCountLabel.SetText(i18n.T("Containers") + ": " + i18n.T("unavailable"))
			dockerRunningLabel.SetText(i18n.T("Running") + ": " + i18n.T("unavailable"))
		}

		favoriteServicesBox.Objects = buildFavoriteServiceRows()
		favoriteServicesBox.Refresh()

		favoriteContainersBox.Objects = buildFavoriteContainerRows()
		favoriteContainersBox.Refresh()

		problems := system.BuildProblems(
			stats,
			appstate.GetConfig().FavoriteServices,
			appstate.GetConfig().FavoriteContainers,
		)
		if len(problems) == 0 {
			problemsLabel.SetText(i18n.T("No problems detected"))
		} else {
			problemsLabel.SetText(joinLines(problems))
		}

		notifyProblems(problems)
		topProcesses, err := system.ListTopProcesses(5)
		if err != nil {
			topProcessesLabel.SetText(i18n.T("Could not load the process list"))
		} else {
			topProcessesLabel.SetText(formatTopProcesses(topProcesses))
		}
		statusLabel.SetText(i18n.T("Updated") + ": " + time.Now().Format("15:04:05"))
	}

	showError := func(err error) {
		if err == nil {
			return
		}
		statusLabel.SetText(i18n.Tf("Error: %s", err.Error()))
	}

	refreshStats = func(heavy bool) {
		var stats system.Stats
		var err error
		if heavy {
			stats, err = system.GetStats()
		} else {
			stats, err = system.GetStatsLight()
		}
		if err != nil {
			fyne.Do(func() {
				showError(err)
			})
			return
		}

		fyne.Do(func() {
			updateUI(stats, heavy)
		})
	}

	refreshButton := widget.NewButtonWithIcon(i18n.T("Refresh"), theme.ViewRefreshIcon(), func() {
		refreshButtonTapped()
	})

	cpuCard := NewStatCard(
		"CPU",
		i18n.T("Current processor load"),
		container.NewVBox(
			cpuValueLabel,
			cpuBar,
			cpuChart.object(),
		),
	)

	ramCard := NewStatCard(
		"RAM",
		i18n.T("Memory usage"),
		container.NewVBox(
			ramValueLabel,
			ramBar,
			ramDetailsLabel,
			ramChart.object(),
		),
	)

	diskCard := NewStatCard(
		i18n.T("Disk"),
		i18n.T("Disk usage"),
		container.NewVBox(
			diskValueLabel,
			diskBar,
			diskDetailsLabel,
		),
	)

	uptimeCard := NewStatCard(
		i18n.T("Uptime"),
		i18n.T("How long the system has been running"),
		container.NewVBox(
			uptimeLabel,
		),
	)

	servicesCard := NewStatCard(
		i18n.T("Services"),
		i18n.T("Service status and count"),
		container.NewVBox(
			systemdLabel,
			serviceCountLabel,
		),
	)

	dockerCard := NewStatCard(
		"Docker",
		i18n.T("Container status and count"),
		container.NewVBox(
			dockerLabel,
			dockerCountLabel,
			dockerRunningLabel,
		),
	)

	favoriteServicesCard := NewStatCard(
		i18n.T("Favorite services"),
		i18n.T("Quick actions for important services"),
		favoriteServicesBox,
	)

	favoriteContainersCard := NewStatCard(
		i18n.T("Favorite containers"),
		i18n.T("Quick actions for important containers"),
		favoriteContainersBox,
	)

	problemsCard := NewStatCard(
		i18n.T("Problems"),
		i18n.T("Problems that need attention"),
		container.NewVBox(
			problemsLabel,
		),
	)

	topProcessesCard := NewStatCard(
		i18n.T("Top processes"),
		i18n.T("The heaviest processes by CPU"),
		container.NewVBox(
			topProcessesLabel,
		),
	)

	networkCard := NewStatCard(
		i18n.T("Network"),
		i18n.T("Data sent / received"),
		container.NewVBox(
			recvLabel,
			netRecvChart.object(),
			sentLabel,
			netSentChart.object(),
		),
	)

	perCPUCard := NewStatCard(
		i18n.T("CPU cores"),
		i18n.T("Load on each processor core"),
		container.NewVBox(
			perCPUBars,
		),
	)

	tempCard := NewStatCard(
		i18n.T("Temperature"),
		i18n.T("CPU/GPU temperature"),
		container.NewVBox(tempLabel, fanLabel, voltageLabel),
	)

	// GridWithColumns sizes each row to its tallest card, so content cannot
	// overlap its neighbours. The whole tab is inside a VScroll, which handles
	// narrow windows.
	metricsGrid := container.NewGridWithColumns(4, cpuCard, ramCard, diskCard, uptimeCard)
	statusGrid := container.NewGridWithColumns(3, servicesCard, dockerCard, networkCard)
	detailGrid := container.NewGridWithColumns(2,
		problemsCard, tempCard,
		favoriteServicesCard, favoriteContainersCard,
		topProcessesCard, perCPUCard,
	)

	content := container.NewVBox(
		subtitle,
		widget.NewSeparator(),
		metricsGrid,
		statusGrid,
		detailGrid,
		widget.NewSeparator(),
		container.NewHBox(refreshButton, statusLabel),
	)

	go refreshStats(true)

	// Heavy metrics (systemctl/docker listing, full process scan, sensors) are
	// refreshed at most every ~5s regardless of the light poll rate, so the
	// monitor does not itself become a load source.
	const heavyEvery = 5 * time.Second
	var heavyMu sync.Mutex
	lastHeavy := time.Now()

	autoRefresh := newAutoRefresher(func() {
		heavyMu.Lock()
		heavy := time.Since(lastHeavy) >= heavyEvery
		if heavy {
			lastHeavy = time.Now()
		}
		heavyMu.Unlock()

		refreshStats(heavy)
	})
	RegisterCloser(autoRefresh.Stop)
	autoRefresh.SetEnabled(appstate.GetConfig().DashboardAutoRefresh)

	RegisterRefresh("Dashboard", func() { refreshStats(true) })

	// Wrap in a vertical scroll so the dashboard grid does not clip on small
	// windows.
	return container.NewVScroll(container.NewPadded(content))
}

// latestRecvRate and latestSentRate read the most recent network rate, or 0
// before there are two samples to derive one from.
func latestRecvRate(samples []system.Sample) float64 {
	if len(samples) == 0 {
		return 0
	}
	return samples[len(samples)-1].NetRecvPerSec
}

func latestSentRate(samples []system.Sample) float64 {
	if len(samples) == 0 {
		return 0
	}
	return samples[len(samples)-1].NetSentPerSec
}

// formatRate renders a bytes-per-second value using the same units as the
// totals beside it.
func formatRate(bytesPerSec float64) string {
	if bytesPerSec < 0 {
		bytesPerSec = 0
	}
	return system.FormatBytes(uint64(bytesPerSec)) + "/s"
}

// updateCharts feeds one history snapshot into the dashboard's sparklines.
func updateCharts(samples []system.Sample, cpu, ram, netRecv, netSent *sparkline) {
	cpuValues := make([]float64, len(samples))
	ramValues := make([]float64, len(samples))
	recvValues := make([]float64, len(samples))
	sentValues := make([]float64, len(samples))

	for i, s := range samples {
		cpuValues[i] = s.CPUPercent
		ramValues[i] = s.RAMPercent
		recvValues[i] = s.NetRecvPerSec
		sentValues[i] = s.NetSentPerSec
	}

	cpu.setValues(cpuValues)
	ram.setValues(ramValues)
	netRecv.setValues(recvValues)
	netSent.setValues(sentValues)
}

func joinLines(items []string) string {
	if len(items) == 0 {
		return ""
	}

	lines := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		lines = append(lines, "• "+item)
	}

	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n")
}

func formatTopProcesses(items []system.ProcessUsageInfo) string {
	if len(items) == 0 {
		return i18n.T("No process data")
	}

	lines := make([]string, 0, len(items))

	for _, item := range items {
		lines = append(lines,
			fmt.Sprintf(
				"• %s (PID %d) | CPU %.1f%% | RAM %s",
				item.ProcessName,
				item.PID,
				item.CPUPercent,
				system.FormatBytes(item.MemoryBytes),
			),
		)
	}

	return strings.Join(lines, "\n")
}
