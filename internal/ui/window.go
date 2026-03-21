package ui

import (
	"fmt"
	"time"

	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewMainWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("System Hub")
	w.Resize(fyne.NewSize(900, 550))

	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", buildDashboardTab()),
		container.NewTabItem("Services", buildServicesTab()),
		container.NewTabItem("Docker", buildDockerTab()),
		container.NewTabItem("Logs", buildLogsTab()),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(tabs)
	return w
}

func buildDashboardTab() fyne.CanvasObject {
	title := widget.NewLabel("System Hub")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("MVP dashboard: CPU / RAM / Disk")

	// CPU
	cpuValueLabel := widget.NewLabel("CPU: ...")
	cpuBar := widget.NewProgressBar()

	// RAM
	ramValueLabel := widget.NewLabel("RAM: ...")
	ramDetailsLabel := widget.NewLabel("...")
	ramBar := widget.NewProgressBar()

	// Disk
	diskValueLabel := widget.NewLabel("Disk: ...")
	diskDetailsLabel := widget.NewLabel("...")
	diskBar := widget.NewProgressBar()

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

	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		container.NewGridWithColumns(3, cpuCard, ramCard, diskCard),
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

func buildServicesTab() fyne.CanvasObject {
	title := widget.NewLabel("Services")
	title.TextStyle = fyne.TextStyle{Bold: true}

	info := widget.NewLabel(
		"Здесь позже будет управление systemd-сервисами.\n\n" +
			"План:\n" +
			"- список сервисов\n" +
			"- статус сервиса\n" +
			"- start / stop / restart",
	)

	return container.NewPadded(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			info,
		),
	)
}

func buildDockerTab() fyne.CanvasObject {
	title := widget.NewLabel("Docker")
	title.TextStyle = fyne.TextStyle{Bold: true}

	info := widget.NewLabel(
		"Здесь позже будет управление Docker.\n\n" +
			"План:\n" +
			"- список контейнеров\n" +
			"- статус контейнера\n" +
			"- запуск / остановка",
	)

	return container.NewPadded(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			info,
		),
	)
}

func buildLogsTab() fyne.CanvasObject {
	title := widget.NewLabel("Logs")
	title.TextStyle = fyne.TextStyle{Bold: true}

	info := widget.NewLabel(
		"Здесь позже будет просмотр логов.\n\n" +
			"План:\n" +
			"- выбор источника логов\n" +
			"- текстовый просмотр\n" +
			"- обновление",
	)

	return container.NewPadded(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			info,
		),
	)
}