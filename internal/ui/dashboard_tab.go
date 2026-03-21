package ui

import (
	"fmt"
	"time"

	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildDashboardTab() fyne.CanvasObject {
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

		uptimeLabel.SetText("Uptime: " + system.FormatUptime(stats.UptimeSeconds))

		if system.IsSystemdAvailable() {
			systemdLabel.SetText("systemd: available")
		} else {
			systemdLabel.SetText("systemd: unavailable")
		}

		if system.IsDockerAvailable() {
			dockerLabel.SetText("Docker: available")
		} else {
			dockerLabel.SetText("Docker: unavailable")
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

	cpuCard := widget.NewCard("CPU", "Текущая загрузка процессора", container.NewVBox(
		cpuValueLabel,
		cpuBar,
	))

	ramCard := widget.NewCard("RAM", "Использование оперативной памяти", container.NewVBox(
		ramValueLabel,
		ramBar,
		ramDetailsLabel,
	))

	diskCard := widget.NewCard("Disk", "Использование диска", container.NewVBox(
		diskValueLabel,
		diskBar,
		diskDetailsLabel,
	))

	uptimeCard := widget.NewCard("Uptime", "Время непрерывной работы системы", container.NewVBox(
		uptimeLabel,
	))

	systemdCard := widget.NewCard("Services", "Доступность systemd", container.NewVBox(
		systemdLabel,
	))

	dockerCard := widget.NewCard("Docker", "Доступность Docker CLI", container.NewVBox(
		dockerLabel,
	))

	topRow := container.NewGridWithColumns(3, cpuCard, ramCard, diskCard)
	bottomRow := container.NewGridWithColumns(3, uptimeCard, systemdCard, dockerCard)

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

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			refreshStats()
		}
	}()

	return container.NewPadded(content)
}