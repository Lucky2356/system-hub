package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/appstate"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func buildSettingsTab(parent fyne.Window, cfg config.Config) fyne.CanvasObject {
	refreshIntervalEntry := widget.NewEntry()
	refreshIntervalEntry.SetText(strconv.Itoa(cfg.RefreshIntervalSeconds))

	defaultLogLinesEntry := widget.NewEntry()
	defaultLogLinesEntry.SetText(strconv.Itoa(cfg.DefaultLogLines))

	dashboardAutoRefreshCheck := widget.NewCheck("Автообновление дашборда", nil)
	dashboardAutoRefreshCheck.SetChecked(cfg.DashboardAutoRefresh)

	servicesAutoRefreshCheck := widget.NewCheck("Автообновление сервисов", nil)
	servicesAutoRefreshCheck.SetChecked(cfg.ServicesAutoRefresh)

	dockerAutoRefreshCheck := widget.NewCheck("Автообновление Docker", nil)
	dockerAutoRefreshCheck.SetChecked(cfg.DockerAutoRefresh)

	logsAutoRefreshCheck := widget.NewCheck("Автообновление вкладки логов", nil)
	logsAutoRefreshCheck.SetChecked(cfg.LogsAutoRefresh)

	logViewerAutoRefreshCheck := widget.NewCheck("Автообновление окна логов", nil)
	logViewerAutoRefreshCheck.SetChecked(cfg.LogViewerAutoRefresh)

	themeSelect := widget.NewSelect([]string{"dark", "light"}, nil)
	themeSelect.SetSelected(cfg.Theme)
	if cfg.Theme == "" {
		themeSelect.SetSelected("dark")
	}

	configPath, err := config.ConfigFilePath()
	if err != nil {
		configPath = "недоступно"
	}

	statusLabel := widget.NewLabel("Измените настройки и нажмите «Сохранить настройки»")

	saveButton := widget.NewButton("Сохранить настройки", func() {
		refreshInterval, err := strconv.Atoi(strings.TrimSpace(refreshIntervalEntry.Text))
		if err != nil || refreshInterval <= 0 {
			dialog.ShowError(fmt.Errorf("интервал обновления должен быть положительным числом"), parent)
			return
		}

		defaultLogLines, err := strconv.Atoi(strings.TrimSpace(defaultLogLinesEntry.Text))
		if err != nil || defaultLogLines <= 0 {
			dialog.ShowError(fmt.Errorf("количество строк логов должно быть положительным числом"), parent)
			return
		}

		appTheme := themeSelect.Selected
		if appTheme == "" {
			appTheme = "dark"
		}

		newCfg := config.Config{
			RefreshIntervalSeconds: refreshInterval,
			DefaultLogLines:        defaultLogLines,
			LogFile:                cfg.LogFile,

			DashboardAutoRefresh: dashboardAutoRefreshCheck.Checked,
			ServicesAutoRefresh:  servicesAutoRefreshCheck.Checked,
			DockerAutoRefresh:    dockerAutoRefreshCheck.Checked,
			LogsAutoRefresh:      logsAutoRefreshCheck.Checked,
			LogViewerAutoRefresh: logViewerAutoRefreshCheck.Checked,

			FavoriteServices:   cfg.FavoriteServices,
			FavoriteContainers: cfg.FavoriteContainers,
			Theme:              appTheme,
		}

		if err := config.Save(newCfg); err != nil {
			dialog.ShowError(err, parent)
			statusLabel.SetText("Ошибка сохранения: " + err.Error())
			return
		}
		appstate.SetConfig(newCfg)

		if newCfg.Theme == "light" {
			fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
		} else {
			fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
		}

		statusLabel.SetText("Настройки сохранены")
		dialog.ShowInformation(
			"Настройки сохранены",
			"Настройки сохранены.\n\nНекоторые изменения полностью применятся после перезапуска приложения.",
			parent,
		)
	})

	resetButton := widget.NewButton("Сбросить настройки", func() {
		defaultCfg := config.DefaultConfig()

		refreshIntervalEntry.SetText(strconv.Itoa(defaultCfg.RefreshIntervalSeconds))
		defaultLogLinesEntry.SetText(strconv.Itoa(defaultCfg.DefaultLogLines))

		dashboardAutoRefreshCheck.SetChecked(defaultCfg.DashboardAutoRefresh)
		servicesAutoRefreshCheck.SetChecked(defaultCfg.ServicesAutoRefresh)
		dockerAutoRefreshCheck.SetChecked(defaultCfg.DockerAutoRefresh)
		logsAutoRefreshCheck.SetChecked(defaultCfg.LogsAutoRefresh)
		logViewerAutoRefreshCheck.SetChecked(defaultCfg.LogViewerAutoRefresh)

		themeSelect.SetSelected(defaultCfg.Theme)
		if defaultCfg.Theme == "" {
			themeSelect.SetSelected("dark")
		}
		statusLabel.SetText("Значения сброшены по умолчанию")
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Настройки", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Конфигурация приложения"),
		widget.NewSeparator(),

		widget.NewLabel("Интервал обновления (сек)"),
		refreshIntervalEntry,

		widget.NewLabel("Строк логов по умолчанию"),
		defaultLogLinesEntry,

		widget.NewSeparator(),

		dashboardAutoRefreshCheck,
		servicesAutoRefreshCheck,
		dockerAutoRefreshCheck,
		logsAutoRefreshCheck,
		logViewerAutoRefreshCheck,

		widget.NewSeparator(),

		widget.NewLabel("Тема"),
		themeSelect,

		widget.NewSeparator(),

		widget.NewLabel("Файл конфигурации"),
		widget.NewLabel(configPath),

		widget.NewSeparator(),

		container.NewHBox(
			saveButton,
			resetButton,
		),

		container.NewHBox(
			widget.NewButton("Экспорт конфигурации", func() {
				dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil || writer == nil {
						return
					}
					defer writer.Close()

					cfgPath, err := config.ConfigFilePath()
					if err != nil {
						ShowError(parent, err)
						return
					}

					data, err := os.ReadFile(cfgPath)
					if err != nil {
						ShowError(parent, err)
						return
					}

					if _, err := writer.Write(data); err != nil {
						ShowError(parent, err)
					}
				}, parent)
			}),
			widget.NewButton("Импорт конфигурации", func() {
				dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil || reader == nil {
						return
					}
					defer reader.Close()

					data, err := io.ReadAll(reader)
					if err != nil {
						ShowError(parent, err)
						return
					}

					cfgPath, err := config.ConfigFilePath()
					if err != nil {
						ShowError(parent, err)
						return
					}

					if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
						ShowError(parent, err)
						return
					}

					loaded, err := config.Load()
					if err != nil {
						ShowError(parent, err)
						return
					}

					appstate.SetConfig(loaded)
					statusLabel.SetText("Настройки импортированы. Перезапустите приложение.")
				}, parent)
			}),
		),

		statusLabel,
	)

	return container.NewPadded(form)
}