package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
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
	subtitle := widget.NewLabel("Обзор системы: CPU / RAM / диск / сеть / сервисы / Docker")

	cpuValueLabel := widget.NewLabel("CPU: ...")
	cpuBar := widget.NewProgressBar()

	ramValueLabel := widget.NewLabel("RAM: ...")
	ramDetailsLabel := widget.NewLabel("...")
	ramBar := widget.NewProgressBar()

	diskValueLabel := widget.NewLabel("Диск: ...")
	diskDetailsLabel := widget.NewLabel("...")
	diskBar := widget.NewProgressBar()

	uptimeLabel := widget.NewLabel("Аптайм: ...")
	systemdLabel := widget.NewLabel("systemd: ...")
	dockerLabel := widget.NewLabel("Docker: ...")

	serviceCountLabel := widget.NewLabel("Сервисов: ...")
	dockerCountLabel := widget.NewLabel("Контейнеров: ...")
	dockerRunningLabel := widget.NewLabel("Запущено: ...")

	favoriteServicesBox := container.NewVBox(widget.NewLabel("Нет избранных сервисов"))
	favoriteContainersBox := container.NewVBox(widget.NewLabel("Нет избранных контейнеров"))
	problemsLabel := widget.NewLabel("Проверка проблем...")
	topProcessesLabel := widget.NewLabel("Загрузка списка процессов...")
	topProcessesLabel.Wrapping = fyne.TextWrapWord

	recvLabel := widget.NewLabel("RX: ...")
	sentLabel := widget.NewLabel("TX: ...")

	perCPUBars := container.NewVBox(widget.NewLabel("Загрузка данных по ядрам..."))

	tempLabel := widget.NewLabel("Загрузка температур...")
	tempLabel.Wrapping = fyne.TextWrapWord
	fanLabel := widget.NewLabel("")
	fanLabel.Wrapping = fyne.TextWrapWord
	voltageLabel := widget.NewLabel("")
	voltageLabel.Wrapping = fyne.TextWrapWord

	problemsLabel.Wrapping = fyne.TextWrapWord
	topProcessesLabel.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel("Статус: ожидание")

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
			return []fyne.CanvasObject{widget.NewLabel("Нет избранных сервисов")}
		}

		services, err := system.ListServices()
		if err != nil {
			return []fyne.CanvasObject{widget.NewLabel("Не удалось загрузить избранные сервисы")}
		}

		serviceMap := make(map[string]system.ServiceInfo, len(services))
		for _, svc := range services {
			serviceMap[svc.Name] = svc
		}

		runServiceAction := func(action, serviceName string) {
			dialog.ShowConfirm(
				"Подтверждение",
				fmt.Sprintf("Выполнить %s для сервиса %s?", strings.ToUpper(action), serviceName),
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

			statusText := "недоступен"
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

			logsButton := widget.NewButton("Логи", func(serviceName string) func() {
				return func() {
					showLogsWindow(
						"Логи сервиса: "+serviceName,
						"Логи сервиса "+serviceName,
						func() (string, error) {
							return system.GetServiceLogs(serviceName, appstate.GetConfig().DefaultLogLines)
						},
						appstate.GetConfig(),
					)
				}
			}(name))

			unitButton := widget.NewButton("Юнит", func(serviceName string) func() {
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
						"Юнит-файл открыт",
						"Файл открыт во вкладке «Файлы»:\n"+unitPath,
						parent,
					)
				}
			}(name))

			startButton := widget.NewButton("Старт", func(serviceName string) func() {
				return func() {
					runServiceAction("start", serviceName)
				}
			}(name))

			stopButton := widget.NewButton("Стоп", func(serviceName string) func() {
				return func() {
					runServiceAction("stop", serviceName)
				}
			}(name))

			restartButton := widget.NewButton("Рестарт", func(serviceName string) func() {
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
			return []fyne.CanvasObject{widget.NewLabel("Нет избранных контейнеров")}
		}

		containers, err := system.ListDockerContainers()
		if err != nil {
			return []fyne.CanvasObject{widget.NewLabel("Не удалось загрузить избранные контейнеры")}
		}

		containerMap := make(map[string]system.DockerContainerInfo, len(containers))
		for _, c := range containers {
			containerMap[c.Names] = c
		}

		runContainerAction := func(action, containerName string) {
			dialog.ShowConfirm(
				"Подтверждение",
				fmt.Sprintf("Выполнить %s для контейнера %s?", strings.ToUpper(action), containerName),
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

			statusText := "недоступен"
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

			logsButton := widget.NewButton("Логи", func(containerName string) func() {
				return func() {
					showLogsWindow(
						"Логи контейнера: "+containerName,
						"Логи контейнера "+containerName,
						func() (string, error) {
							return system.GetDockerContainerLogs(containerName, appstate.GetConfig().DefaultLogLines)
						},
						appstate.GetConfig(),
					)
				}
			}(name))

			inspectButton := widget.NewButton("Инспекция", func(containerName string) func() {
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
								"Инспекция контейнера: "+containerName,
								"Закрыть",
								container.NewPadded(output),
								parent,
							)
						})
					}()
				}
			}(name))

			startButton := widget.NewButton("Старт", func(containerName string) func() {
				return func() {
					runContainerAction("start", containerName)
				}
			}(name))

			stopButton := widget.NewButton("Стоп", func(containerName string) func() {
				return func() {
					runContainerAction("stop", containerName)
				}
			}(name))

			restartButton := widget.NewButton("Рестарт", func(containerName string) func() {
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
				perCPUBars.Add(widget.NewLabel("Нет данных по ядрам"))
			}
		}
		for i, core := range stats.PerCPU {
			coreRows[i].label.SetText(fmt.Sprintf("Ядро %d: %.1f%%", core.Core, core.Percent))
			coreRows[i].bar.SetValue(core.Percent / 100)
		}
		perCPUBars.Refresh()

		recvLabel.SetText(fmt.Sprintf("RX: %s", system.FormatBytes(stats.NetRecv)))
		sentLabel.SetText(fmt.Sprintf("TX: %s", system.FormatBytes(stats.NetSent)))

		// Sensor readings spawn `sensors`/`nvidia-smi`; refresh them only on
		// the heavy tick (they are also carried on stats only when heavy).
		if heavy {
			if len(stats.Temperatures) == 0 {
				tempLabel.SetText("Нет данных о температуре")
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

		diskValueLabel.SetText(fmt.Sprintf("Диск: %.1f%%", stats.DiskPercent))
		diskDetailsLabel.SetText(fmt.Sprintf(
			"%s / %s",
			system.FormatBytes(stats.DiskUsed),
			system.FormatBytes(stats.DiskTotal),
		))
		diskBar.SetValue(stats.DiskPercent / 100)

		if stats.UptimeKnown {
			uptimeLabel.SetText("Аптайм: " + system.FormatUptime(stats.UptimeSeconds))
		} else {
			uptimeLabel.SetText("Аптайм: недоступно")
		}

		if !heavy {
			statusLabel.SetText("Обновлено: " + time.Now().Format("15:04:05"))
			return
		}

		if stats.SystemdAvailable {
			systemdLabel.SetText("systemd: доступен")
			if stats.ServiceCountKnown {
				serviceCountLabel.SetText(fmt.Sprintf("Сервисов: %d", stats.ServiceCount))
			} else {
				serviceCountLabel.SetText("Сервисов: недоступно")
			}
		} else {
			systemdLabel.SetText("systemd: недоступен")
			serviceCountLabel.SetText("Сервисов: недоступно")
		}

		if stats.DockerAvailable {
			dockerLabel.SetText("Docker: доступен")
			if stats.DockerCountKnown {
				dockerCountLabel.SetText(fmt.Sprintf("Контейнеров: %d", stats.DockerContainerCount))
				dockerRunningLabel.SetText(fmt.Sprintf("Запущено: %d", stats.DockerRunningCount))
			} else {
				dockerCountLabel.SetText("Контейнеров: недоступно")
				dockerRunningLabel.SetText("Запущено: недоступно")
			}
		} else {
			dockerLabel.SetText("Docker: недоступен")
			dockerCountLabel.SetText("Контейнеров: недоступно")
			dockerRunningLabel.SetText("Запущено: недоступно")
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
			problemsLabel.SetText("Проблем не обнаружено")
		} else {
			problemsLabel.SetText(joinLines(problems))
		}

		notifyProblems(problems)
		topProcesses, err := system.ListTopProcesses(5)
		if err != nil {
			topProcessesLabel.SetText("Не удалось загрузить список процессов")
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

	refreshButton := widget.NewButtonWithIcon("Обновить", theme.ViewRefreshIcon(), func() {
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
		"Диск",
		"Использование диска",
		container.NewVBox(
			diskValueLabel,
			diskBar,
			diskDetailsLabel,
		),
	)

	uptimeCard := NewStatCard(
		"Аптайм",
		"Время непрерывной работы системы",
		container.NewVBox(
			uptimeLabel,
		),
	)

	servicesCard := NewStatCard(
		"Сервисы",
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
		"Избранные сервисы",
		"Быстрые действия для важных сервисов",
		favoriteServicesBox,
	)

	favoriteContainersCard := NewStatCard(
		"Избранные контейнеры",
		"Быстрые действия для важных контейнеров",
		favoriteContainersBox,
	)

	problemsCard := NewStatCard(
		"Проблемы",
		"Проблемы, требующие внимания",
		container.NewVBox(
			problemsLabel,
		),
	)

	topProcessesCard := NewStatCard(
		"Топ процессов",
		"Самые тяжёлые процессы по CPU",
		container.NewVBox(
			topProcessesLabel,
		),
	)

	networkCard := NewStatCard(
		"Сеть",
		"Передано / получено данных",
		container.NewVBox(
			recvLabel,
			sentLabel,
		),
	)

	perCPUCard := NewStatCard(
		"Ядра CPU",
		"Загрузка каждого ядра процессора",
		container.NewVBox(
			perCPUBars,
		),
	)

	tempCard := NewStatCard(
		"Температура",
		"Температура CPU/GPU",
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
		return "Нет данных о процессах"
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
