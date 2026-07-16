package ui

import (
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildCommandsTab(parent fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Команды")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Безопасный запуск read-only команд для диагностики")

	commands := system.GetSafeCommands()
	commandMap := make(map[string]system.SafeCommand, len(commands))
	commandTitles := make([]string, 0, len(commands))

	for _, cmd := range commands {
		commandMap[cmd.Title] = cmd
		commandTitles = append(commandTitles, cmd.Title)
	}

	commandSelect := widget.NewSelect(commandTitles, nil)
	if len(commandTitles) > 0 {
		commandSelect.SetSelected(commandTitles[0])
	}

	descriptionLabel := widget.NewLabel("")
	descriptionLabel.Wrapping = fyne.TextWrapWord

	argEntry := widget.NewEntry()
	argEntry.SetPlaceHolder("Аргумент")

	statusLabel := widget.NewLabel("Статус: ожидание")

	outputEntry := widget.NewMultiLineEntry()
	outputEntry.Wrapping = fyne.TextWrapWord
	outputEntry.Disable()

	updateSelectedCommandUI := func() {
		selectedTitle := commandSelect.Selected
		cmd, ok := commandMap[selectedTitle]
		if !ok {
			descriptionLabel.SetText("Команда не выбрана")
			argEntry.Hide()
			return
		}

		descriptionLabel.SetText(cmd.Description)

		if cmd.NeedsArg {
			argEntry.Show()
			if cmd.ArgHint != "" {
				argEntry.SetPlaceHolder(cmd.ArgHint)
			} else {
				argEntry.SetPlaceHolder("Введите аргумент")
			}
		} else {
			argEntry.SetText("")
			argEntry.Hide()
		}
	}

	runCommand := func() {
		selectedTitle := commandSelect.Selected
		cmd, ok := commandMap[selectedTitle]
		if !ok {
			statusLabel.SetText("Статус: команда не выбрана")
			return
		}

		statusLabel.SetText("Статус: выполнение команды...")
		outputEntry.SetText("")

		go func() {
			result, err := system.RunSafeCommand(cmd.Key, argEntry.Text)

			fyne.Do(func() {
				if err != nil {
					outputEntry.SetText(result)
					ShowError(parent, err)
					statusLabel.SetText("Статус: команда завершилась с ошибкой")
					return
				}

				outputEntry.SetText(result)
				statusLabel.SetText("Статус: команда выполнена")
			})
		}()
	}

	commandSelect.OnChanged = func(string) {
		updateSelectedCommandUI()
	}

	runButton := widget.NewButton("Выполнить", runCommand)
	clearButton := widget.NewButton("Очистить", func() {
		outputEntry.SetText("")
		statusLabel.SetText("Статус: очищено")
	})

	content := container.NewBorder(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
			widget.NewLabel("Команда"),
			commandSelect,
			descriptionLabel,
			argEntry,
			container.NewHBox(runButton, clearButton),
			statusLabel,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewPadded(outputEntry),
	)

	updateSelectedCommandUI()

	return container.NewPadded(content)
}
