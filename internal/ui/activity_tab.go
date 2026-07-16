package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/activity"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildActivityTab() fyne.CanvasObject {
	title := widget.NewLabel("История действий")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("История действий, выполненных из приложения")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Фильтр: nginx, docker, restart...")

	statusLabel := widget.NewLabel("Статус: ожидание")

	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.Disable()

	renderEntries := func(items []activity.Entry, filter string) string {
		filter = strings.ToLower(strings.TrimSpace(filter))

		lines := make([]string, 0, len(items))

		for i := len(items) - 1; i >= 0; i-- {
			e := items[i]

			line := fmt.Sprintf(
				"%s | %s | %s | %s | %s",
				e.Time.Format("15:04:05"),
				e.Target,
				e.Action,
				e.Name,
				e.Status,
			)

			if strings.TrimSpace(e.Details) != "" {
				line += " | " + e.Details
			}

			if filter != "" && !strings.Contains(strings.ToLower(line), filter) {
				continue
			}

			lines = append(lines, line)
		}

		if len(lines) == 0 {
			return "Пока нет действий"
		}

		return strings.Join(lines, "\n")
	}

	refreshView := func() {
		entries := activity.List()
		logOutput.SetText(renderEntries(entries, searchEntry.Text))
		statusLabel.SetText(fmt.Sprintf("Статус: %d записей | обновлено %s", len(entries), time.Now().Format("15:04:05")))
	}

	searchEntry.OnChanged = func(string) {
		refreshView()
	}

	refreshButton := widget.NewButton("Обновить", func() {
		refreshView()
	})

	clearButton := widget.NewButton("Очистить", func() {
		activity.Clear()
		refreshView()
	})

	content := container.NewBorder(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
			searchEntry,
			container.NewHBox(refreshButton, clearButton),
			statusLabel,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewPadded(logOutput),
	)

	refreshView()

	return container.NewPadded(content)
}
