package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// parseLogLines reads the user-entered log line count, falling back to the
// default when the field is empty or not a positive number.
func parseLogLines(text string) int {
	const fallback = 100

	n, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func buildReportTab(parent fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Диагностический отчёт")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Экспорт снимка состояния системы, избранного и логов")

	linesEntry := widget.NewEntry()
	linesEntry.SetText("100")

	statusLabel := widget.NewLabel("Статус: ожидание")

	reportOutput := widget.NewMultiLineEntry()
	reportOutput.Wrapping = fyne.TextWrapWord

	buildReport := func() {
		statusLabel.SetText("Статус: сбор отчёта...")

		logLines := parseLogLines(linesEntry.Text)

		go func() {
			report, err := system.BuildDiagnosticsReport(system.DiagnosticsReportParams{
				FavoriteServices:   appstate.GetConfig().FavoriteServices,
				FavoriteContainers: appstate.GetConfig().FavoriteContainers,
				LogLines:           logLines,
			})

			if err != nil {
				fyne.Do(func() {
					ShowError(parent, err)
					statusLabel.SetText("Статус: ошибка сборки отчёта")
				})
				return
			}

			fyne.Do(func() {
				reportOutput.SetText(report)
				statusLabel.SetText("Статус: отчёт собран")
			})
		}()
	}

	buildBundle := func() {
		statusLabel.SetText("Статус: сбор диагностического пакета...")

		logLines := parseLogLines(linesEntry.Text)

		go func() {
			report, err := system.BuildDiagnosticsBundle(system.DiagnosticsBundleParams{
				FavoriteServices:   appstate.GetConfig().FavoriteServices,
				FavoriteContainers: appstate.GetConfig().FavoriteContainers,
				LogLines:           logLines,
				TopProcessLimit:    5,
			})

			if err != nil {
				fyne.Do(func() {
					ShowError(parent, err)
					statusLabel.SetText("Статус: ошибка сборки диагностического пакета")
				})
				return
			}

			fyne.Do(func() {
				reportOutput.SetText(report)
				statusLabel.SetText("Статус: диагностический пакет собран")
			})
		}()
	}

	saveReport := func() {
		reportText := reportOutput.Text
		if reportText == "" {
			statusLabel.SetText("Статус: сначала собери отчёт")
			return
		}

		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				ShowError(parent, err)
				statusLabel.SetText("Статус: ошибка сохранения")
				return
			}
			if writer == nil {
				statusLabel.SetText("Статус: сохранение отменено")
				return
			}
			if _, writeErr := writer.Write([]byte(reportText)); writeErr != nil {
				_ = writer.Close()
				ShowError(parent, writeErr)
				statusLabel.SetText("Статус: ошибка записи файла")
				return
			}

			// Close surfaces flush errors; ignoring it can truncate the report.
			if err := writer.Close(); err != nil {
				ShowError(parent, err)
				statusLabel.SetText("Статус: ошибка записи файла")
				return
			}

			statusLabel.SetText("Статус: отчёт сохранён")
		}, parent)
	}

	saveToDefaultLocation := func() {
		reportText := reportOutput.Text
		if reportText == "" {
			statusLabel.SetText("Статус: сначала собери отчёт")
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			ShowError(parent, err)
			statusLabel.SetText("Статус: ошибка определения домашней папки")
			return
		}

		filename := fmt.Sprintf("system-hub-report-%s.txt", time.Now().Format("20060102-150405"))
		fullPath := filepath.Join(homeDir, filename)

		if err := os.WriteFile(fullPath, []byte(reportText), 0o644); err != nil {
			ShowError(parent, err)
			statusLabel.SetText("Статус: ошибка сохранения файла")
			return
		}

		statusLabel.SetText("Статус: отчёт сохранён в " + fullPath)
	}

	buildButton := widget.NewButton("Собрать отчёт", buildReport)
	bundleButton := widget.NewButton("Диагностический пакет", buildBundle)
	saveButton := widget.NewButton("Сохранить как...", saveReport)
	quickSaveButton := widget.NewButton("Быстро сохранить в домашнюю папку", saveToDefaultLocation)

	content := container.NewBorder(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
			widget.NewLabel("Строк логов для каждого избранного сервиса/контейнера"),
			linesEntry,
			container.NewHBox(buildButton, bundleButton, saveButton, quickSaveButton),
			statusLabel,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewPadded(reportOutput),
	)

	return container.NewPadded(content)
}
