package ui

import (
	"github.com/Lucky2356/system-hub/internal/i18n"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildCommandsTab(parent fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel(i18n.T("Commands"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel(i18n.T("Safely run read-only diagnostic commands"))

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
	argEntry.SetPlaceHolder(i18n.T("Argument"))

	statusLabel := newDataLabel(i18n.T("Status: waiting"))

	outputEntry := widget.NewMultiLineEntry()
	outputEntry.Wrapping = fyne.TextWrapWord
	outputEntry.Disable()

	updateSelectedCommandUI := func() {
		selectedTitle := commandSelect.Selected
		cmd, ok := commandMap[selectedTitle]
		if !ok {
			descriptionLabel.SetText(i18n.T("No command selected"))
			argEntry.Hide()
			return
		}

		descriptionLabel.SetText(cmd.Description)

		if cmd.NeedsArg {
			argEntry.Show()
			if cmd.ArgHint != "" {
				argEntry.SetPlaceHolder(cmd.ArgHint)
			} else {
				argEntry.SetPlaceHolder(i18n.T("Enter an argument"))
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
			statusLabel.SetText(i18n.T("Status: no command selected"))
			return
		}

		statusLabel.SetText(i18n.T("Status: running the command..."))
		outputEntry.SetText("")

		go func() {
			result, err := system.RunSafeCommand(cmd.Key, argEntry.Text)

			fyne.Do(func() {
				if err != nil {
					outputEntry.SetText(result)
					ShowError(parent, err)
					statusLabel.SetText(i18n.T("Status: the command failed"))
					return
				}

				outputEntry.SetText(result)
				statusLabel.SetText(i18n.T("Status: command completed"))
			})
		}()
	}

	commandSelect.OnChanged = func(string) {
		updateSelectedCommandUI()
	}

	runButton := widget.NewButton(i18n.T("Run"), runCommand)
	clearButton := widget.NewButton(i18n.T("Clear"), func() {
		outputEntry.SetText("")
		statusLabel.SetText(i18n.T("Status: cleared"))
	})

	content := container.NewBorder(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
			widget.NewLabel(i18n.T("Command")),
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
