package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildReportTab(parent fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Diagnostics Report")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Экспорт снимка состояния системы, избранного и логов")

	linesEntry := widget.NewEntry()
	linesEntry.SetText("100")

	statusLabel := widget.NewLabel("Статус: ожидание")

	reportOutput := widget.NewMultiLineEntry()
	reportOutput.Wrapping = fyne.TextWrapWord

	buildReport := func() {
		statusLabel.SetText("Статус: сбор отчёта...")

		logLines := 100
		fmt.Sscanf(linesEntry.Text, "%d", &logLines)
		if logLines <= 0 {
			logLines = 100
		}

		go func() {
			report, err := system.BuildDiagnosticsReport(system.DiagnosticsReportParams{
				FavoriteServices:   appstate.Config.FavoriteServices,
				FavoriteContainers: appstate.Config.FavoriteContainers,
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

	saveReport := func() {
		reportText := reportOutput.Text
		if reportText == "" {
			statusLabel.SetText("Статус: сначала собери отчёт")
			return
		}

		filename := fmt.Sprintf("system-hub-report-%s.txt", time.Now().Format("20060102-150405"))

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
			defer writer.Close()

			_, writeErr := writer.Write([]byte(reportText))
			if writeErr != nil {
				ShowError(parent, writeErr)
				statusLabel.SetText("Статус: ошибка записи файла")
				return
			}

			statusLabel.SetText("Статус: отчёт сохранён")
		}, parent)

		_ = filename
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
	saveButton := widget.NewButton("Сохранить как...", saveReport)
	quickSaveButton := widget.NewButton("Быстро сохранить в Home", saveToDefaultLocation)

	content := container.NewBorder(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
			widget.NewLabel("Log lines for each favorite service/container"),
			linesEntry,
			container.NewHBox(buildButton, saveButton, quickSaveButton),
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