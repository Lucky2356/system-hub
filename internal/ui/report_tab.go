package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/i18n"
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
	title := widget.NewLabel(i18n.T("Diagnostic report"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel(i18n.T("Export a snapshot of the system state, favorites and logs"))

	linesEntry := widget.NewEntry()
	linesEntry.SetText("100")

	statusLabel := newDataLabel(i18n.T("Status: waiting"))

	reportOutput := widget.NewMultiLineEntry()
	reportOutput.Wrapping = fyne.TextWrapWord

	buildReport := func() {
		statusLabel.SetText(i18n.T("Status: building the report..."))

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
					statusLabel.SetText(i18n.T("Status: could not build the report"))
				})
				return
			}

			fyne.Do(func() {
				reportOutput.SetText(report)
				statusLabel.SetText(i18n.T("Status: report ready"))
			})
		}()
	}

	buildBundle := func() {
		statusLabel.SetText(i18n.T("Status: building the diagnostic bundle..."))

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
					statusLabel.SetText(i18n.T("Status: could not build the diagnostic bundle"))
				})
				return
			}

			fyne.Do(func() {
				reportOutput.SetText(report)
				statusLabel.SetText(i18n.T("Status: diagnostic bundle ready"))
			})
		}()
	}

	saveReport := func() {
		reportText := reportOutput.Text
		if reportText == "" {
			statusLabel.SetText(i18n.T("Status: build the report first"))
			return
		}

		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				ShowError(parent, err)
				statusLabel.SetText(i18n.T("Status: save failed"))
				return
			}
			if writer == nil {
				statusLabel.SetText(i18n.T("Status: save cancelled"))
				return
			}
			if _, writeErr := writer.Write([]byte(reportText)); writeErr != nil {
				_ = writer.Close()
				ShowError(parent, writeErr)
				statusLabel.SetText(i18n.T("Status: file write error"))
				return
			}

			// Close surfaces flush errors; ignoring it can truncate the report.
			if err := writer.Close(); err != nil {
				ShowError(parent, err)
				statusLabel.SetText(i18n.T("Status: file write error"))
				return
			}

			statusLabel.SetText(i18n.T("Status: report saved"))
		}, parent)
	}

	saveToDefaultLocation := func() {
		reportText := reportOutput.Text
		if reportText == "" {
			statusLabel.SetText(i18n.T("Status: build the report first"))
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			ShowError(parent, err)
			statusLabel.SetText(i18n.T("Status: could not resolve the home folder"))
			return
		}

		filename := fmt.Sprintf("system-hub-report-%s.txt", time.Now().Format("20060102-150405"))
		fullPath := filepath.Join(homeDir, filename)

		if err := os.WriteFile(fullPath, []byte(reportText), 0o644); err != nil {
			ShowError(parent, err)
			statusLabel.SetText(i18n.T("Status: could not save the file"))
			return
		}

		statusLabel.SetText(i18n.Tf("Status: report saved to %s", fullPath))
	}

	buildButton := widget.NewButton(i18n.T("Build report"), buildReport)
	bundleButton := widget.NewButton(i18n.T("Diagnostic bundle"), buildBundle)
	saveButton := widget.NewButton(i18n.T("Save as..."), saveReport)
	quickSaveButton := widget.NewButton(i18n.T("Quick save to the home folder"), saveToDefaultLocation)

	content := container.NewBorder(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
			widget.NewLabel(i18n.T("Log lines per favorite service/container")),
			linesEntry,
			newToolbarRow(buildButton, bundleButton, saveButton, quickSaveButton),
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
