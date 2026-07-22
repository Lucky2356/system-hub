package ui

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/Lucky2356/system-hub/internal/i18n"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildFilesTab(parent fyne.Window) fyne.CanvasObject {
	// No bold title: the sidebar already names the section. Only the one-line
	// caption remains, which also trims height — this tab was the tallest.
	subtitle := widget.NewLabel(i18n.T("Browse important directories and config files"))

	presets := system.GetPresetPaths()
	presetMap := make(map[string]string, len(presets))
	presetTitles := make([]string, 0, len(presets))

	for _, p := range presets {
		presetTitles = append(presetTitles, p.Title)
		presetMap[p.Title] = p.Path
	}

	pathEntry := widget.NewEntry()

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(i18n.T("Filter by file name..."))

	statusLabel := newDataLabel(i18n.T("Status: waiting"))

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
			statusLabel.SetText(i18n.T("Status: edit mode"))
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
		statusLabel.SetText(i18n.Tf("Status: %d entries", len(filteredEntries)))
	}

	loadPath := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			statusLabel.SetText(i18n.T("Status: the path is empty"))
			return
		}

		entries, err := system.ListFiles(path)
		if err != nil {
			ShowError(parent, err)
			statusLabel.SetText(i18n.T("Status: load error"))
			return
		}

		currentPath = path
		allEntries = entries
		selectedIndex = -1
		currentFilePath = ""
		fileContent.SetText("")
		updateEditMode(false)
		refreshList()
		statusLabel.SetText(i18n.Tf("Status: opened directory %s", currentPath))
	}

	openSelected := func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredEntries) {
			statusLabel.SetText(i18n.T("Status: select a file or folder"))
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
			statusLabel.SetText(i18n.T("Status: could not read the file"))
			return
		}

		currentFilePath = item.FullPath
		fileContent.SetText(content)
		updateEditMode(false)
		statusLabel.SetText(i18n.Tf("Status: opened file %s", item.FullPath))
	}

	upButton := widget.NewButton(i18n.T("Up"), func() {
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

	openButton := widget.NewButton(i18n.T("Open"), func() {
		loadPath(pathEntry.Text)
	})

	viewButton := widget.NewButton(i18n.T("View"), func() {
		openSelected()
	})

	editButton := widget.NewButton(i18n.T("Edit"), func() {
		if strings.TrimSpace(currentFilePath) == "" {
			statusLabel.SetText(i18n.T("Status: open a file first"))
			return
		}

		updateEditMode(true)
	})

	reloadButton := widget.NewButton(i18n.T("Reload"), func() {
		if strings.TrimSpace(currentFilePath) == "" {
			statusLabel.SetText(i18n.T("Status: open a file first"))
			return
		}

		content, err := system.ReadTextFile(currentFilePath, 1024*1024)
		if err != nil {
			ShowError(parent, err)
			statusLabel.SetText(i18n.T("Status: could not reload the file"))
			return
		}

		fileContent.SetText(content)
		updateEditMode(false)
		statusLabel.SetText(i18n.T("Status: file reloaded"))
	})

	saveButton := widget.NewButton(i18n.T("Save"), func() {
		if strings.TrimSpace(currentFilePath) == "" {
			statusLabel.SetText(i18n.T("Status: open a file first"))
			return
		}

		err := system.WriteTextFile(currentFilePath, fileContent.Text)
		if err != nil {
			if system.IsPermissionError(err) {
				ShowErrorMsg(parent, i18n.T("Not enough permissions to save the file"))
			} else {
				ShowError(parent, err)
			}
			statusLabel.SetText(i18n.T("Status: could not save the file"))
			return
		}

		updateEditMode(false)
		statusLabel.SetText(i18n.T("Status: file saved"))
	})

	infoButton := widget.NewButton(i18n.T("Info"), func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredEntries) {
			statusLabel.SetText(i18n.T("Status: select a file or folder"))
			return
		}

		item := filteredEntries[selectedIndex]

		kind := i18n.T("File")
		if item.IsDir {
			kind = i18n.T("Directory")
		}

		text := i18n.Tf(
			"Name: %s\nPath: %s\nType: %s\nSize: %s",
			item.Name,
			item.FullPath,
			kind,
			system.FormatFileSize(item.Size),
		)

		dialog.ShowInformation(i18n.T("File details"), text, parent)
	})

	newFileButton := widget.NewButton(i18n.T("New file"), func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("filename.txt")

		dialog.ShowCustomConfirm(i18n.T("New file"), i18n.T("Create"), i18n.T("Cancel"),
			container.NewPadded(container.NewVBox(
				widget.NewLabel(i18n.T("Enter a file name in the current directory:")),
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
				statusLabel.SetText(i18n.T("Status: file created"))
				loadPath(currentPath)
			},
			parent,
		)
	})

	newDirButton := widget.NewButton(i18n.T("New folder"), func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("newdir")

		dialog.ShowCustomConfirm(i18n.T("New directory"), i18n.T("Create"), i18n.T("Cancel"),
			container.NewPadded(container.NewVBox(
				widget.NewLabel(i18n.T("Enter a directory name:")),
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
				statusLabel.SetText(i18n.T("Status: directory created"))
				loadPath(currentPath)
			},
			parent,
		)
	})

	renameButton := widget.NewButton(i18n.T("Rename"), func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredEntries) {
			statusLabel.SetText(i18n.T("Status: select a file or folder"))
			return
		}

		item := filteredEntries[selectedIndex]
		nameEntry := widget.NewEntry()
		nameEntry.SetText(item.Name)

		dialog.ShowCustomConfirm(i18n.T("Rename"), i18n.T("Rename"), i18n.T("Cancel"),
			container.NewPadded(container.NewVBox(
				widget.NewLabel(i18n.Tf("New name for %s:", item.Name)),
				nameEntry,
			)),
			func(confirmed bool) {
				if !confirmed || strings.TrimSpace(nameEntry.Text) == "" {
					return
				}
				newPath := filepath.Join(currentPath, strings.TrimSpace(nameEntry.Text))
				if err := system.RenameFile(item.FullPath, newPath); err != nil {
					if system.IsPermissionError(err) {
						ShowErrorMsg(parent, i18n.T("Not enough permissions to rename"))
					} else {
						ShowError(parent, err)
					}
					return
				}
				statusLabel.SetText(i18n.T("Status: renamed"))
				loadPath(currentPath)
			},
			parent,
		)
	})

	deleteButton := widget.NewButton(i18n.T("Delete"), func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredEntries) {
			statusLabel.SetText(i18n.T("Status: select a file or folder"))
			return
		}

		item := filteredEntries[selectedIndex]

		message := i18n.Tf("Delete file %s?", item.Name)
		if item.IsDir {
			message = i18n.Tf("Delete directory %s? Only empty directories can be removed.", item.Name)
		}

		dialog.ShowConfirm(i18n.T("Delete"), message, func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := system.DeleteFile(item.FullPath); err != nil {
				if system.IsPermissionError(err) {
					ShowErrorMsg(parent, i18n.T("Not enough permissions to delete"))
				} else {
					ShowError(parent, err)
				}
				return
			}
			statusLabel.SetText(i18n.T("Status: deleted"))
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
			widget.NewLabel(i18n.T("Directory")),
			presetSelect,
			pathEntry,
			searchEntry,
			statusLabel,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewPadded(fileList),
	)

	// The action row spans the full window width up top, not the narrow left
	// column: eleven buttons wrapped into six rows inside the column and made
	// this the tallest tab. Across the whole width they fit one or two rows.
	toolbar := newToolbarRow(openButton, upButton, viewButton, editButton, reloadButton, saveButton, infoButton, newFileButton, newDirButton, renameButton, deleteButton)

	contentContainer := container.NewBorder(
		container.NewVBox(
			subtitle,
			toolbar,
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
			return errors.New(i18n.T("the path is empty"))
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
		statusLabel.SetText(i18n.Tf("Status: opened file %s", path))

		return nil
	})

	loadPath(pathEntry.Text)

	return container.NewPadded(contentContainer)
}
