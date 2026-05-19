package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/system"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildDashboardTab(parent fyne.Window, cfg config.Config) fyne.CanvasObject {
	title := widget.NewLabel("System Hub")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("MVP dashboard: CPU / RAM / Disk / Overview")

	cpuValueLabel := widget.NewLabel("CPU: ...")
	cpuBar := widget.NewProgressBar()

	ramValueLabel := widget.NewLabel("RAM: ...")
	ramDetailsLabel := widget.NewLabel("...")
	ramBar := widget.NewProgressBar()

	diskValueLabel := widget.NewLabel("Disk: ...")
	diskDetailsLabel := widget.NewLabel("...")
	diskBar := widget.NewProgressBar()

	uptimeLabel := widget.NewLabel("Uptime: ...")
	systemdLabel := widget.NewLabel("systemd: ...")
	dockerLabel := widget.NewLabel("Docker: ...")

	serviceCountLabel := widget.NewLabel("Services: ...")
	dockerCountLabel := widget.NewLabel("Containers: ...")
	dockerRunningLabel := widget.NewLabel("Running: ...")

	favoriteServicesBox := container.NewVBox(widget.NewLabel("No favorite services"))
	favoriteContainersBox := container.NewVBox(widget.NewLabel("No favorite containers"))
	problemsLabel := widget.NewLabel("Checking problems...")
	topProcessesLabel := widget.NewLabel("Loading top processes...")
	topProcessesLabel.Wrapping = fyne.TextWrapWord

	recvLabel := widget.NewLabel("RX: ...")
	sentLabel := widget.NewLabel("TX: ...")

	perCPUBars := container.NewVBox(widget.NewLabel("Loading per-CPU info..."))

	tempLabel := widget.NewLabel("Loading temperatures...")
	tempLabel.Wrapping = fyne.TextWrapWord
	fanLabel := widget.NewLabel("")
	fanLabel.Wrapping = fyne.TextWrapWord
	voltageLabel := widget.NewLabel("")
	voltageLabel.Wrapping = fyne.TextWrapWord


	problemsLabel.Wrapping = fyne.TextWrapWord
	topProcessesLabel.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel("Статус: ожидание")

	var lastProblems []string
	
	makeStatusText := func(status string) *canvas.Text {
		text := canvas.NewText(status, color.NRGBA{R: 160, G: 160, B: 160, A: 255})
		text.TextSize = 12

		switch strings.ToLower(strings.TrimSpace(status)) {
		case "active", "running":
			text.Color = color.NRGBA{R: 60, G: 180, B: 90, A: 255}
		case "failed", "exited":
			text.Color = color.NRGBA{R: 220, G: 70, B: 70, A: 255}
		case "inactive", "dead", "unavailable":
			text.Color = color.NRGBA{R: 180, G: 140, B: 50, A: 255}
		default:
			text.Color = color.NRGBA{R: 160, G: 160, B: 160, A: 255}
		}

		return text
	}

	var refreshStats func()
	var refreshButtonTapped func()

	refreshButtonTapped = func() {
		go refreshStats()
	}

	buildFavoriteServiceRows := func() []fyne.CanvasObject {
		if len(appstate.GetConfig().FavoriteServices) == 0 {
			return []fyne.CanvasObject{widget.NewLabel("No favorite services")}
		}

		services, err := system.ListServices()
		if err != nil {
			return []fyne.CanvasObject{widget.NewLabel("Unable to load favorite services")}
		}

		serviceMap := make(map[string]system.ServiceInfo, len(services))
		for _, svc := range services {
			serviceMap[svc.Name] = svc
		}

		runServiceAction := func(action, serviceName string) {
			dialog.ShowConfirm(
				"Confirm "+action,
				fmt.Sprintf("%s service %s?", cases.Title(language.English).String(action), serviceName),
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

			statusText := "unavailable"
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

			logsButton := widget.NewButton("Logs", func(serviceName string) func() {
				return func() {
					showLogsWindow(
						"Service Logs: "+serviceName,
						"Logs for service "+serviceName,
						func() (string, error) {
							return system.GetServiceLogs(serviceName, appstate.GetConfig().DefaultLogLines)
						},
						appstate.GetConfig(),
					)
				}
			}(name))

			unitButton := widget.NewButton("Unit", func(serviceName string) func() {
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
						"Unit file opened",
						"Файл открыт во вкладке Files:\n"+unitPath,
						parent,
					)
				}
			}(name))

			startButton := widget.NewButton("Start", func(serviceName string) func() {
				return func() {
					runServiceAction("start", serviceName)
				}
			}(name))

			stopButton := widget.NewButton("Stop", func(serviceName string) func() {
				return func() {
					runServiceAction("stop", serviceName)
				}
			}(name))

			restartButton := widget.NewButton("Restart", func(serviceName string) func() {
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
			return []fyne.CanvasObject{widget.NewLabel("No favorite containers")}
		}

		containers, err := system.ListDockerContainers()
		if err != nil {
			return []fyne.CanvasObject{widget.NewLabel("Unable to load favorite containers")}
		}

		containerMap := make(map[string]system.DockerContainerInfo, len(containers))
		for _, c := range containers {
			containerMap[c.Names] = c
		}

		runContainerAction := func(action, containerName string) {
			dialog.ShowConfirm(
				"Confirm "+action,
				fmt.Sprintf("%s container %s?", cases.Title(language.English).String(action), containerName),
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

			statusText := "unavailable"
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

			logsButton := widget.NewButton("Logs", func(containerName string) func() {
				return func() {
					showLogsWindow(
						"Docker Logs: "+containerName,
						"Logs for container "+containerName,
						func() (string, error) {
							return system.GetDockerContainerLogs(containerName, appstate.GetConfig().DefaultLogLines)
						},
						appstate.GetConfig(),
					)
				}
			}(name))

			inspectButton := widget.NewButton("Inspect", func(containerName string) func() {
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
								"Docker Inspect: "+containerName,
								"Закрыть",
								container.NewPadded(output),
								parent,
							)
						})
					}()
				}
			}(name))

			startButton := widget.NewButton("Start", func(containerName string) func() {
				return func() {
					runContainerAction("start", containerName)
				}
			}(name))

			stopButton := widget.NewButton("Stop", func(containerName string) func() {
				return func() {
					runContainerAction("stop", containerName)
				}
			}(name))

			restartButton := widget.NewButton("Restart", func(containerName string) func() {
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

	updateUI := func(stats system.Stats) {
		cpuValueLabel.SetText(fmt.Sprintf("CPU: %.1f%%", stats.CPUPercent))
		cpuBar.SetValue(stats.CPUPercent / 100)

		perCPUBars.Objects = nil
		for _, core := range stats.PerCPU {
			bar := widget.NewProgressBar()
			bar.SetValue(core.Percent / 100)
			label := widget.NewLabel(fmt.Sprintf("Core %d: %.1f%%", core.Core, core.Percent))
			perCPUBars.Add(container.NewVBox(label, bar))
		}
		if len(stats.PerCPU) == 0 {
			perCPUBars.Add(widget.NewLabel("No per-CPU data available"))
		}
		perCPUBars.Refresh()

		recvLabel.SetText(fmt.Sprintf("RX: %s", system.FormatBytes(stats.NetRecv)))
		sentLabel.SetText(fmt.Sprintf("TX: %s", system.FormatBytes(stats.NetSent)))

		if len(stats.Temperatures) == 0 {
			tempLabel.SetText("No temperature data")
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

		ramValueLabel.SetText(fmt.Sprintf("RAM: %.1f%%", stats.RAMPercent))
		ramDetailsLabel.SetText(fmt.Sprintf(
			"%s / %s",
			system.FormatBytes(stats.RAMUsed),
			system.FormatBytes(stats.RAMTotal),
		))
		ramBar.SetValue(stats.RAMPercent / 100)

		diskValueLabel.SetText(fmt.Sprintf("Disk: %.1f%%", stats.DiskPercent))
		diskDetailsLabel.SetText(fmt.Sprintf(
			"%s / %s",
			system.FormatBytes(stats.DiskUsed),
			system.FormatBytes(stats.DiskTotal),
		))
		diskBar.SetValue(stats.DiskPercent / 100)

		if stats.UptimeKnown {
			uptimeLabel.SetText("Uptime: " + system.FormatUptime(stats.UptimeSeconds))
		} else {
			uptimeLabel.SetText("Uptime: unavailable")
		}

		if stats.SystemdAvailable {
			systemdLabel.SetText("systemd: available")
			if stats.ServiceCountKnown {
				serviceCountLabel.SetText(fmt.Sprintf("Services: %d", stats.ServiceCount))
			} else {
				serviceCountLabel.SetText("Services: unavailable")
			}
		} else {
			systemdLabel.SetText("systemd: unavailable")
			serviceCountLabel.SetText("Services: unavailable")
		}

		if stats.DockerAvailable {
			dockerLabel.SetText("Docker: available")
			if stats.DockerCountKnown {
				dockerCountLabel.SetText(fmt.Sprintf("Containers: %d", stats.DockerContainerCount))
				dockerRunningLabel.SetText(fmt.Sprintf("Running: %d", stats.DockerRunningCount))
			} else {
				dockerCountLabel.SetText("Containers: unavailable")
				dockerRunningLabel.SetText("Running: unavailable")
			}
		} else {
			dockerLabel.SetText("Docker: unavailable")
			dockerCountLabel.SetText("Containers: unavailable")
			dockerRunningLabel.SetText("Running: unavailable")
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
		problemsLabel.SetText(joinLines(problems))

		for _, p := range problems {
			found := false
			for _, old := range lastProblems {
				if old == p {
					found = true
					break
				}
			}
			if !found {
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "System Hub Alert",
					Content: p,
				})
			}
		}
		lastProblems = problems
		topProcesses, err := system.ListTopProcesses(5)
		if err != nil {
			topProcessesLabel.SetText("Unable to load top processes")
		} else {
			topProcessesLabel.SetText(formatTopProcesses(topProcesses))
		}
		statusLabel.SetText("Обновлено: " + time.Now().Format("15:04:05"))
	}

	showError := func(err error) {
		if err == nil {
			return
		}
		statusLabel.SetText("Ошибка: " + err.Error())
	}

	refreshStats = func() {
		stats, err := system.GetStats()
		if err != nil {
			fyne.Do(func() {
				showError(err)
			})
			return
		}

		fyne.Do(func() {
			updateUI(stats)
		})
	}

	refreshButton := widget.NewButton("Обновить", func() {
		refreshButtonTapped()
	})

	cpuCard := NewStatCard(
		"CPU",
		"Текущая загрузка процессора",
		container.NewVBox(
			cpuValueLabel,
			cpuBar,
		),
	)

	ramCard := NewStatCard(
		"RAM",
		"Использование оперативной памяти",
		container.NewVBox(
			ramValueLabel,
			ramBar,
			ramDetailsLabel,
		),
	)

	diskCard := NewStatCard(
		"Disk",
		"Использование диска",
		container.NewVBox(
			diskValueLabel,
			diskBar,
			diskDetailsLabel,
		),
	)

	uptimeCard := NewStatCard(
		"Uptime",
		"Время непрерывной работы системы",
		container.NewVBox(
			uptimeLabel,
		),
	)

	servicesCard := NewStatCard(
		"Services",
		"Статус и количество сервисов",
		container.NewVBox(
			systemdLabel,
			serviceCountLabel,
		),
	)

	dockerCard := NewStatCard(
		"Docker",
		"Статус и количество контейнеров",
		container.NewVBox(
			dockerLabel,
			dockerCountLabel,
			dockerRunningLabel,
		),
	)

	favoriteServicesCard := NewStatCard(
		"Favorite Services",
		"Быстрые действия для важных сервисов",
		favoriteServicesBox,
	)

	favoriteContainersCard := NewStatCard(
		"Favorite Containers",
		"Быстрые действия для важных контейнеров",
		favoriteContainersBox,
	)

	problemsCard := NewStatCard(
		"Problems",
		"Проблемы, требующие внимания",
		container.NewVBox(
			problemsLabel,
		),
	)

	topProcessesCard := NewStatCard(
		"Top Processes",
		"Самые тяжёлые процессы по CPU",
		container.NewVBox(
			topProcessesLabel,
		),
	)

	networkCard := NewStatCard(
		"Network",
		"Передано / получено данных",
		container.NewVBox(
			recvLabel,
			sentLabel,
		),
	)

	perCPUCard := NewStatCard(
		"Per-Core CPU",
		"Загрузка каждого ядра процессора",
		container.NewVBox(
			perCPUBars,
		),
	)

	tempCard := NewStatCard(
		"Temperature",
		"Температура CPU/GPU",
		container.NewVBox(tempLabel, fanLabel, voltageLabel),
	)

	topRow := container.NewGridWithColumns(3, cpuCard, ramCard, diskCard)
	middleRow := container.NewGridWithColumns(4, uptimeCard, servicesCard, dockerCard, networkCard)
	bottomRow := container.NewGridWithColumns(4,
		favoriteServicesCard,
		favoriteContainersCard,
		problemsCard,
		tempCard,
	)

	extraRow := container.NewGridWithColumns(2, topProcessesCard, perCPUCard)

	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		topRow,
		middleRow,
		bottomRow,
		extraRow,
		widget.NewSeparator(),
		refreshButton,
		statusLabel,
	)

	go refreshStats()

	stopDashboardRefresh := make(chan struct{})
	if appstate.GetConfig().DashboardAutoRefresh {
		go func() {
			ticker := time.NewTicker(time.Duration(appstate.GetConfig().RefreshIntervalSeconds) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					refreshStats()
				case <-stopDashboardRefresh:
					return
				}
			}
		}()
	}

	RegisterRefresh("Dashboard", refreshStats)

	return container.NewPadded(content)
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
		return "No process data"
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