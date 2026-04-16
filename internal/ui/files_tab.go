package ui

import (
	"fmt"
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
		fileContent.SetText("")
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

		fileContent.SetText(content)
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

	fileList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
	}

	fileList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
	}

	fileList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
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
			container.NewHBox(openButton, upButton, viewButton, infoButton),
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

	loadPath(pathEntry.Text)

	return container.NewPadded(contentContainer)
}