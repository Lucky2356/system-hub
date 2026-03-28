package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Lucky2356/system-hub/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildSettingsTab(parent fyne.Window, cfg config.Config) fyne.CanvasObject {
	refreshIntervalEntry := widget.NewEntry()
	refreshIntervalEntry.SetText(strconv.Itoa(cfg.RefreshIntervalSeconds))

	defaultLogLinesEntry := widget.NewEntry()
	defaultLogLinesEntry.SetText(strconv.Itoa(cfg.DefaultLogLines))

	dashboardAutoRefreshCheck := widget.NewCheck("Dashboard auto refresh", nil)
	dashboardAutoRefreshCheck.SetChecked(cfg.DashboardAutoRefresh)

	servicesAutoRefreshCheck := widget.NewCheck("Services auto refresh", nil)
	servicesAutoRefreshCheck.SetChecked(cfg.ServicesAutoRefresh)

	dockerAutoRefreshCheck := widget.NewCheck("Docker auto refresh", nil)
	dockerAutoRefreshCheck.SetChecked(cfg.DockerAutoRefresh)

	logsAutoRefreshCheck := widget.NewCheck("Logs tab auto refresh", nil)
	logsAutoRefreshCheck.SetChecked(cfg.LogsAutoRefresh)

	logViewerAutoRefreshCheck := widget.NewCheck("Log viewer auto refresh", nil)
	logViewerAutoRefreshCheck.SetChecked(cfg.LogViewerAutoRefresh)

	configPath, err := config.ConfigFilePath()
	if err != nil {
		configPath = "unavailable"
	}

	statusLabel := widget.NewLabel("Измените настройки и нажмите Save")

	saveButton := widget.NewButton("Save settings", func() {
		refreshInterval, err := strconv.Atoi(strings.TrimSpace(refreshIntervalEntry.Text))
		if err != nil || refreshInterval <= 0 {
			dialog.ShowError(fmt.Errorf("refresh interval must be a positive number"), parent)
			return
		}

		defaultLogLines, err := strconv.Atoi(strings.TrimSpace(defaultLogLinesEntry.Text))
		if err != nil || defaultLogLines <= 0 {
			dialog.ShowError(fmt.Errorf("default log lines must be a positive number"), parent)
			return
		}

		newCfg := config.Config{
			RefreshIntervalSeconds: refreshInterval,
			DefaultLogLines:        defaultLogLines,

			DashboardAutoRefresh: dashboardAutoRefreshCheck.Checked,
			ServicesAutoRefresh:  servicesAutoRefreshCheck.Checked,
			DockerAutoRefresh:    dockerAutoRefreshCheck.Checked,
			LogsAutoRefresh:      logsAutoRefreshCheck.Checked,
			LogViewerAutoRefresh: logViewerAutoRefreshCheck.Checked,
		}

		if err := config.Save(newCfg); err != nil {
			dialog.ShowError(err, parent)
			statusLabel.SetText("Ошибка сохранения: " + err.Error())
			return
		}

		statusLabel.SetText("Настройки сохранены")
		dialog.ShowInformation(
			"Settings saved",
			"Настройки сохранены.\n\nНекоторые изменения полностью применятся после перезапуска приложения.",
			parent,
		)
	})

	resetButton := widget.NewButton("Reset to defaults", func() {
		defaultCfg := config.DefaultConfig()

		refreshIntervalEntry.SetText(strconv.Itoa(defaultCfg.RefreshIntervalSeconds))
		defaultLogLinesEntry.SetText(strconv.Itoa(defaultCfg.DefaultLogLines))

		dashboardAutoRefreshCheck.SetChecked(defaultCfg.DashboardAutoRefresh)
		servicesAutoRefreshCheck.SetChecked(defaultCfg.ServicesAutoRefresh)
		dockerAutoRefreshCheck.SetChecked(defaultCfg.DockerAutoRefresh)
		logsAutoRefreshCheck.SetChecked(defaultCfg.LogsAutoRefresh)
		logViewerAutoRefreshCheck.SetChecked(defaultCfg.LogViewerAutoRefresh)

		statusLabel.SetText("Значения сброшены к default")
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Application config"),
		widget.NewSeparator(),

		widget.NewLabel("Refresh interval (seconds)"),
		refreshIntervalEntry,

		widget.NewLabel("Default log lines"),
		defaultLogLinesEntry,

		widget.NewSeparator(),

		dashboardAutoRefreshCheck,
		servicesAutoRefreshCheck,
		dockerAutoRefreshCheck,
		logsAutoRefreshCheck,
		logViewerAutoRefreshCheck,

		widget.NewSeparator(),

		widget.NewLabel("Config file"),
		widget.NewLabel(configPath),

		widget.NewSeparator(),

		container.NewHBox(
			saveButton,
			resetButton,
		),

		statusLabel,
	)

	return container.NewPadded(form)
}