package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
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


	problemsLabel.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel("Статус: ожидание")

	var refreshStats func()
	var refreshButtonTapped func()

	refreshButtonTapped = func() {
		go refreshStats()
	}

	buildFavoriteServiceRows := func() []fyne.CanvasObject {
	if len(appstate.Config.FavoriteServices) == 0 {
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

	rows := make([]fyne.CanvasObject, 0, len(appstate.Config.FavoriteServices))

	for _, name := range appstate.Config.FavoriteServices {
		svc, ok := serviceMap[name]

		statusText := "unavailable"
		if ok {
			statusText = svc.ActiveState
		}

		nameLabel := widget.NewLabel(name)
		statusLabel := widget.NewLabel(statusText)

		logsButton := widget.NewButton("Logs", func(serviceName string) func() {
			return func() {
				showLogsWindow(
					"Service Logs: "+serviceName,
					"Logs for service "+serviceName,
					func() (string, error) {
						return system.GetServiceLogs(serviceName, appstate.Config.DefaultLogLines)
					},
					appstate.Config,
				)
			}
		}(name))

		restartButton := widget.NewButton("Restart", func(serviceName string) func() {
			return func() {
				dialog.ShowConfirm(
					"Confirm restart",
					"Restart service "+serviceName+"?",
					func(ok bool) {
						if !ok {
							return
						}

						go func() {
							err := system.ControlService("restart", serviceName)
							fyne.Do(func() {
								if err != nil {
									if system.IsPermissionError(err) {
										ShowErrorMsg(parent, system.BuildPermissionHint("service", "restart", serviceName))
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
		}(name))

		row := container.NewBorder(
			nil,
			nil,
			container.NewVBox(nameLabel, statusLabel),
			container.NewHBox(logsButton, restartButton),
			nil,
		)

		rows = append(rows, row)
	}

	return rows
}

	buildFavoriteContainerRows := func() []fyne.CanvasObject {
		if len(appstate.Config.FavoriteContainers) == 0 {
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

		rows := make([]fyne.CanvasObject, 0, len(appstate.Config.FavoriteContainers))

		for _, name := range appstate.Config.FavoriteContainers {
			c, ok := containerMap[name]

			statusText := "unavailable"
			if ok {
				statusText = c.State
			}

			nameLabel := widget.NewLabel(name)
			statusLabel := widget.NewLabel(statusText)

			logsButton := widget.NewButton("Logs", func(containerName string) func() {
				return func() {
					showLogsWindow(
						"Docker Logs: "+containerName,
						"Logs for container "+containerName,
						func() (string, error) {
							return system.GetDockerContainerLogs(containerName, appstate.Config.DefaultLogLines)
						},
						appstate.Config,
					)
				}
			}(name))

			restartButton := widget.NewButton("Restart", func(containerName string) func() {
				return func() {
					dialog.ShowConfirm(
						"Confirm restart",
						"Restart container "+containerName+"?",
						func(ok bool) {
							if !ok {
								return
							}

							go func() {
								err := system.ControlDockerContainer("restart", containerName)
								fyne.Do(func() {
									if err != nil {
										if system.IsPermissionError(err) {
											ShowErrorMsg(parent, system.BuildPermissionHint("docker", "restart", containerName))
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
			}(name))

			row := container.NewBorder(
				nil,
				nil,
				container.NewVBox(nameLabel, statusLabel),
				container.NewHBox(logsButton, restartButton),
				nil,
			)

			rows = append(rows, row)
		}

		return rows
	}

	updateUI := func(stats system.Stats) {
		cpuValueLabel.SetText(fmt.Sprintf("CPU: %.1f%%", stats.CPUPercent))
		cpuBar.SetValue(stats.CPUPercent / 100)

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
			appstate.Config.FavoriteServices,
			appstate.Config.FavoriteContainers,
		)
		problemsLabel.SetText(joinLines(problems))

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

	topRow := container.NewGridWithColumns(3, cpuCard, ramCard, diskCard)
	middleRow := container.NewGridWithColumns(3, uptimeCard, servicesCard, dockerCard)
	bottomRow := container.NewGridWithColumns(3, favoriteServicesCard, favoriteContainersCard, problemsCard)

	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		topRow,
		middleRow,
		bottomRow,
		widget.NewSeparator(),
		refreshButton,
		statusLabel,
	)

	go refreshStats()

	if appstate.Config.DashboardAutoRefresh {
		go func() {
			ticker := time.NewTicker(time.Duration(appstate.Config.RefreshIntervalSeconds) * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				refreshStats()
			}
		}()
	}

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