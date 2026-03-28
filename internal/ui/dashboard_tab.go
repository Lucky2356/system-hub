package ui

import (
	"fmt"
	"time"

	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/system"
	"github.com/Lucky2356/system-hub/internal/appstate"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildDashboardTab(cfg config.Config) fyne.CanvasObject {
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

	statusLabel := widget.NewLabel("Статус: ожидание")

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

		statusLabel.SetText("Обновлено: " + time.Now().Format("15:04:05"))
	}

	showError := func(err error) {
		statusLabel.SetText("Ошибка: " + err.Error())
	}

	refreshStats := func() {
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
		go refreshStats()
	})

	cpuCard := NewStatCard("CPU", "Текущая загрузка процессора",
		container.NewVBox(
			cpuValueLabel,
			cpuBar,
		),
	)

	ramCard := NewStatCard("RAM", "Использование оперативной памяти",
		container.NewVBox(
			ramValueLabel,
			ramBar,
			ramDetailsLabel,
		),
	)

	diskCard := NewStatCard("Disk", "Использование диска",
		container.NewVBox(
			diskValueLabel,
			diskBar,
			diskDetailsLabel,
		),
	)

	uptimeCard := NewStatCard("Uptime", "Время непрерывной работы системы",
		container.NewVBox(
			uptimeLabel,
		),
	)

	servicesCard := NewStatCard("Services", "Статус и количество сервисов",
		container.NewVBox(
			systemdLabel,
			serviceCountLabel,
		),
	)

	dockerCard := NewStatCard("Docker", "Статус и количество контейнеров",
		container.NewVBox(
			dockerLabel,
			dockerCountLabel,
			dockerRunningLabel,
		),
	)

	topRow := container.NewGridWithColumns(3, cpuCard, ramCard, diskCard)
	bottomRow := container.NewGridWithColumns(3, uptimeCard, servicesCard, dockerCard)

	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		topRow,
		bottomRow,
		widget.NewSeparator(),
		refreshButton,
		statusLabel,
	)

	go refreshStats()

	if appstate.Config.DashboardAutoRefresh {
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.RefreshIntervalSeconds) * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				refreshStats()
			}
		}()
	}

	return container.NewPadded(content)
}