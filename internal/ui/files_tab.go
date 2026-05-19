package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildFilesTab(parent fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Files")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Read-only просмотр важных директорий и конфигов")

	presets := system.GetPresetPaths()
	presetMap := make(map[string]string, len(presets))
	presetTitles := make([]string, 0, len(presets))

	for _, p := range presets {
		presetTitles = append(presetTitles, p.Title)
		presetMap[p.Title] = p.Path
	}

	pathEntry := widget.NewEntry()

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Фильтр по имени файла...")

	statusLabel := widget.NewLabel("Статус: ожидание")

	fileContent := widget.NewMultiLineEntry()
	fileContent.Wrapping = fyne.TextWrapWord
	fileContent.Disable()

	var allEntries []system.FileEntry
	var filteredEntries []system.FileEntry
	selectedIndex := -1
	currentPath := ""
	currentFilePath := ""
	updateEditMode := func(enabled bool) {
		if enabled {
			fileContent.Enable()
			statusLabel.SetText("Статус: режим редактирования")
			return
		}

		fileContent.Disable()
	}

	presetSelect := widget.NewSelect(presetTitles, func(selected string) {
		path := presetMap[selected]
		pathEntry.SetText(path)
	})

	if len(presetTitles) > 0 {
		presetSelect.SetSelected(presetTitles[0])
		currentPath = presetMap[presetTitles[0]]
		pathEntry.SetText(currentPath)
	}

	filterEntries := func(query string) []system.FileEntry {
		query = strings.ToLower(strings.TrimSpace(query))
		if query == "" {
			return allEntries
		}

		result := make([]system.FileEntry, 0)
		for _, item := range allEntries {
			if strings.Contains(strings.ToLower(item.Name), query) {
				result = append(result, item)
			}
		}
		return result
	}

	fileList := widget.NewList(
		func() int {
			return len(filteredEntries)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("file")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(filteredEntries) {
				return
			}

			item := filteredEntries[id]
			label := obj.(*widget.Label)

			prefix := "[F]"
			if item.IsDir {
				prefix = "[D]"
			}

			suffix := ""
			if !item.IsDir {
				suffix = " (" + system.FormatFileSize(item.Size) + ")"
			}

			label.SetText(prefix + " " + item.Name + suffix)
		},
	)

	refreshList := func() {
		filteredEntries = filterEntries(searchEntry.Text)
		fileList.Refresh()
		statusLabel.SetText(fmt.Sprintf("Статус: %d записей", len(filteredEntries)))
	}

	loadPath := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			statusLabel.SetText("Статус: путь пустой")
			return
		}

		entries, err := system.ListFiles(path)
		if err != nil {
			ShowError(parent, err)
			statusLabel.SetText("Статус: ошибка загрузки")
			return
		}

		currentPath = path
		allEntries = entries
		selectedIndex = -1
		currentFilePath = ""
		fileContent.SetText("")
		updateEditMode(false)
		refreshList()
		statusLabel.SetText("Статус: открыта директория " + currentPath)
	}

	openSelected := func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredEntries) {
			statusLabel.SetText("Статус: выбери файл или папку")
			return
		}

		item := filteredEntries[selectedIndex]

		if item.IsDir {
			pathEntry.SetText(item.FullPath)
			loadPath(item.FullPath)
			return
		}

		content, err := system.ReadTextFile(item.FullPath, 1024*1024)
		if err != nil {
			ShowError(parent, err)
			statusLabel.SetText("Статус: ошибка чтения файла")
			return
		}

		currentFilePath = item.FullPath
		fileContent.SetText(content)
		updateEditMode(false)
		statusLabel.SetText("Статус: открыт файл " + item.FullPath)
	}

	upButton := widget.NewButton("Up", func() {
		if strings.TrimSpace(currentPath) == "" {
			return
		}

		parentPath := strings.TrimSpace(currentPath)
		parentPath = strings.TrimSuffix(parentPath, "/")

		if parentPath == "" {
			return
		}

		next := parentPath
		lastSlash := strings.LastIndex(next, "/")
		if lastSlash <= 0 {
			next = "/"
		} else {
			next = next[:lastSlash]
		}

		pathEntry.SetText(next)
		loadPath(next)
	})

	openButton := widget.NewButton("Open", func() {
		loadPath(pathEntry.Text)
	})

	viewButton := widget.NewButton("View", func() {
		openSelected()
	})

	editButton := widget.NewButton("Edit", func() {
		if strings.TrimSpace(currentFilePath) == "" {
			statusLabel.SetText("Статус: сначала открой файл")
			return
		}

		updateEditMode(true)
	})

	reloadButton := widget.NewButton("Reload", func() {
		if strings.TrimSpace(currentFilePath) == "" {
			statusLabel.SetText("Статус: сначала открой файл")
			return
		}

		content, err := system.ReadTextFile(currentFilePath, 1024*1024)
		if err != nil {
			ShowError(parent, err)
			statusLabel.SetText("Статус: ошибка перезагрузки файла")
			return
		}

		fileContent.SetText(content)
		updateEditMode(false)
		statusLabel.SetText("Статус: файл перезагружен")
	})

	saveButton := widget.NewButton("Save", func() {
		if strings.TrimSpace(currentFilePath) == "" {
			statusLabel.SetText("Статус: сначала открой файл")
			return
		}

		err := system.WriteTextFile(currentFilePath, fileContent.Text)
		if err != nil {
			if system.IsPermissionError(err) {
				ShowErrorMsg(parent, "Недостаточно прав для сохранения файла")
			} else {
				ShowError(parent, err)
			}
			statusLabel.SetText("Статус: ошибка сохранения файла")
			return
		}

		updateEditMode(false)
		statusLabel.SetText("Статус: файл сохранён")
	})

	infoButton := widget.NewButton("Info", func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredEntries) {
			statusLabel.SetText("Статус: выбери файл или папку")
			return
		}

		item := filteredEntries[selectedIndex]

		text := fmt.Sprintf(
			"Name: %s\nPath: %s\nType: %s\nSize: %s",
			item.Name,
			item.FullPath,
			map[bool]string{true: "Directory", false: "File"}[item.IsDir],
			system.FormatFileSize(item.Size),
		)

		dialog.ShowInformation("File Info", text, parent)
	})

	newFileButton := widget.NewButton("New File", func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("filename.txt")

		dialog.ShowCustomConfirm("New File", "Create", "Cancel",
			container.NewPadded(container.NewVBox(
				widget.NewLabel("Enter filename in current directory:"),
				nameEntry,
			)),
			func(confirmed bool) {
				if !confirmed || strings.TrimSpace(nameEntry.Text) == "" {
					return
				}
				newPath := filepath.Join(currentPath, strings.TrimSpace(nameEntry.Text))
				if err := system.CreateFile(newPath); err != nil {
					ShowError(parent, err)
					return
				}
				statusLabel.SetText("Статус: файл создан")
				loadPath(currentPath)
			},
			parent,
		)
	})

	newDirButton := widget.NewButton("New Dir", func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("newdir")

		dialog.ShowCustomConfirm("New Directory", "Create", "Cancel",
			container.NewPadded(container.NewVBox(
				widget.NewLabel("Enter directory name:"),
				nameEntry,
			)),
			func(confirmed bool) {
				if !confirmed || strings.TrimSpace(nameEntry.Text) == "" {
					return
				}
				newPath := filepath.Join(currentPath, strings.TrimSpace(nameEntry.Text))
				if err := system.CreateDirectory(newPath); err != nil {
					ShowError(parent, err)
					return
				}
				statusLabel.SetText("Статус: директория создана")
				loadPath(currentPath)
			},
			parent,
		)
	})

	renameButton := widget.NewButton("Rename", func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredEntries) {
			statusLabel.SetText("Статус: выбери файл или папку")
			return
		}

		item := filteredEntries[selectedIndex]
		nameEntry := widget.NewEntry()
		nameEntry.SetText(item.Name)

		dialog.ShowCustomConfirm("Rename", "Rename", "Cancel",
			container.NewPadded(container.NewVBox(
				widget.NewLabel("New name for " + item.Name + ":"),
				nameEntry,
			)),
			func(confirmed bool) {
				if !confirmed || strings.TrimSpace(nameEntry.Text) == "" {
					return
				}
				newPath := filepath.Join(currentPath, strings.TrimSpace(nameEntry.Text))
				if err := system.RenameFile(item.FullPath, newPath); err != nil {
					if system.IsPermissionError(err) {
						ShowErrorMsg(parent, "Недостаточно прав для переименования")
					} else {
						ShowError(parent, err)
					}
					return
				}
				statusLabel.SetText("Статус: переименовано")
				loadPath(currentPath)
			},
			parent,
		)
	})

	deleteButton := widget.NewButton("Delete", func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredEntries) {
			statusLabel.SetText("Статус: выбери файл или папку")
			return
		}

		item := filteredEntries[selectedIndex]
		label := "файл"
		if item.IsDir {
			label = "директорию (рекурсивно)"
		}

		dialog.ShowConfirm("Delete", fmt.Sprintf("Удалить %s %s?", label, item.Name), func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := system.DeleteFile(item.FullPath); err != nil {
				if system.IsPermissionError(err) {
					ShowErrorMsg(parent, "Недостаточно прав для удаления")
				} else {
					ShowError(parent, err)
				}
				return
			}
			statusLabel.SetText("Статус: удалено")
			currentFilePath = ""
			fileContent.SetText("")
			loadPath(currentPath)
		}, parent)
	})

	fileList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
	}

	fileList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
	}

	searchEntry.OnChanged = func(string) {
		refreshList()
	}

	fileListContainer := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Directory"),
			presetSelect,
			pathEntry,
			searchEntry,
			container.NewHBox(openButton, upButton, viewButton, editButton, reloadButton, saveButton, infoButton, newFileButton, newDirButton, renameButton, deleteButton),
			statusLabel,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewPadded(fileList),
	)

	contentContainer := container.NewBorder(
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
		),
		nil,
		fileListContainer,
		nil,
		container.NewPadded(fileContent),
	)

	RegisterFilesOpener(func(path string) error {
		path = strings.TrimSpace(path)
		if path == "" {
			return fmt.Errorf("path is empty")
		}

		dir := filepath.Dir(path)
		pathEntry.SetText(dir)
		loadPath(dir)

		content, err := system.ReadTextFile(path, 1024*1024)
		if err != nil {
			return err
		}

		currentFilePath = path
		fileContent.SetText(content)
		updateEditMode(false)
		statusLabel.SetText("Статус: открыт файл " + path)

		return nil
	})

	loadPath(pathEntry.Text)

	return container.NewPadded(contentContainer)
}