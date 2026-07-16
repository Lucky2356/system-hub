package ui

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/i18n"

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

	dashboardAutoRefreshCheck := widget.NewCheck(i18n.T("Auto-refresh the dashboard"), nil)
	dashboardAutoRefreshCheck.SetChecked(cfg.DashboardAutoRefresh)

	servicesAutoRefreshCheck := widget.NewCheck(i18n.T("Auto-refresh services"), nil)
	servicesAutoRefreshCheck.SetChecked(cfg.ServicesAutoRefresh)

	dockerAutoRefreshCheck := widget.NewCheck(i18n.T("Auto-refresh Docker"), nil)
	dockerAutoRefreshCheck.SetChecked(cfg.DockerAutoRefresh)

	logsAutoRefreshCheck := widget.NewCheck(i18n.T("Auto-refresh the logs tab"), nil)
	logsAutoRefreshCheck.SetChecked(cfg.LogsAutoRefresh)

	logViewerAutoRefreshCheck := widget.NewCheck(i18n.T("Auto-refresh the log window"), nil)
	logViewerAutoRefreshCheck.SetChecked(cfg.LogViewerAutoRefresh)

	themeSelect := widget.NewSelect([]string{"dark", "light"}, nil)
	themeSelect.SetSelected(cfg.Theme)
	if cfg.Theme == "" {
		themeSelect.SetSelected("dark")
	}

	// The language Select shows human names but stores the config codes, so the
	// selection survives a language switch.
	languageLabels := map[string]string{
		i18n.Auto:    i18n.T("System language"),
		i18n.Russian: "Русский",
		i18n.English: "English",
	}
	languageCodes := map[string]string{}
	languageOptions := make([]string, 0, len(languageLabels))
	for _, code := range []string{i18n.Auto, i18n.Russian, i18n.English} {
		languageOptions = append(languageOptions, languageLabels[code])
		languageCodes[languageLabels[code]] = code
	}

	languageSelect := widget.NewSelect(languageOptions, nil)
	languageSelect.SetSelected(languageLabels[cfg.Language])

	configPath, err := config.ConfigFilePath()
	if err != nil {
		configPath = i18n.T("unavailable")
	}

	statusLabel := widget.NewLabel(i18n.T("Change the settings and press «Save settings»"))

	saveButton := widget.NewButton(i18n.T("Save settings"), func() {
		refreshInterval, err := strconv.Atoi(strings.TrimSpace(refreshIntervalEntry.Text))
		if err != nil || refreshInterval <= 0 {
			dialog.ShowError(errors.New(i18n.T("the refresh interval must be a positive number")), parent)
			return
		}

		defaultLogLines, err := strconv.Atoi(strings.TrimSpace(defaultLogLinesEntry.Text))
		if err != nil || defaultLogLines <= 0 {
			dialog.ShowError(errors.New(i18n.T("the log line count must be a positive number")), parent)
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
			Language:           languageCodes[languageSelect.Selected],
		}

		if err := config.Save(newCfg); err != nil {
			dialog.ShowError(err, parent)
			statusLabel.SetText(i18n.Tf("Save error: %s", err.Error()))
			return
		}
		appstate.SetConfig(newCfg)

		ApplyTheme(fyne.CurrentApp(), newCfg.Theme)

		statusLabel.SetText(i18n.T("Settings saved"))
		dialog.ShowInformation(
			i18n.T("Settings saved"),
			i18n.T("Settings saved.\n\nSome changes take full effect after restarting the application."),
			parent,
		)

		// Applied last: on a language change this rebuilds the window, which
		// replaces the widgets the lines above are still touching.
		i18n.SetLanguage(newCfg.Language)
	})

	resetButton := widget.NewButton(i18n.T("Reset settings"), func() {
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
		languageSelect.SetSelected(languageLabels[defaultCfg.Language])
		statusLabel.SetText(i18n.T("Values reset to defaults"))
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle(i18n.T("Settings"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(i18n.T("Application configuration")),
		widget.NewSeparator(),

		widget.NewLabel(i18n.T("Refresh interval (s)")),
		refreshIntervalEntry,

		widget.NewLabel(i18n.T("Default log lines")),
		defaultLogLinesEntry,

		widget.NewSeparator(),

		dashboardAutoRefreshCheck,
		servicesAutoRefreshCheck,
		dockerAutoRefreshCheck,
		logsAutoRefreshCheck,
		logViewerAutoRefreshCheck,

		widget.NewSeparator(),

		widget.NewLabel(i18n.T("Theme")),
		themeSelect,

		widget.NewLabel(i18n.T("Language")),
		languageSelect,

		widget.NewSeparator(),

		widget.NewLabel(i18n.T("Configuration file")),
		widget.NewLabel(configPath),

		widget.NewSeparator(),

		container.NewHBox(
			saveButton,
			resetButton,
		),

		container.NewHBox(
			widget.NewButton(i18n.T("Export configuration"), func() {
				dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil || writer == nil {
						return
					}
					cfgPath, err := config.ConfigFilePath()
					if err != nil {
						_ = writer.Close()
						ShowError(parent, err)
						return
					}

					data, err := os.ReadFile(cfgPath)
					if err != nil {
						_ = writer.Close()
						ShowError(parent, err)
						return
					}

					if _, err := writer.Write(data); err != nil {
						_ = writer.Close()
						ShowError(parent, err)
						return
					}

					// Close reports flush errors, so an exported config cannot
					// be silently truncated.
					if err := writer.Close(); err != nil {
						ShowError(parent, err)
						return
					}
					statusLabel.SetText(i18n.T("Settings exported"))
				}, parent)
			}),
			widget.NewButton(i18n.T("Import configuration"), func() {
				dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil || reader == nil {
						return
					}
					defer func() { _ = reader.Close() }()

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
					statusLabel.SetText(i18n.T("Settings imported. Restart the application."))
				}, parent)
			}),
		),

		statusLabel,
	)

	// The form is taller than the default 900x550 window: without a scroll it
	// cut off at the auto-refresh checkboxes, leaving the theme, the language
	// and both buttons unreachable with no hint that anything was below.
	return container.NewVScroll(container.NewPadded(form))
}
