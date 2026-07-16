package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/activity"
	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/i18n"
	"github.com/Lucky2356/system-hub/internal/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// shortImageID trims an image ID for display. Slicing [:12] directly panics on
// a shorter ID, which docker can return for some images.
func shortImageID(id string) string {
	const width = 12
	if len(id) <= width {
		return id
	}
	return id[:width]
}

func buildDockerTab(parent fyne.Window, cfg config.Config) fyne.CanvasObject {
	title := widget.NewLabel("Docker")
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel(i18n.T("View and manage Docker containers"))

	modeContainers, modeImages := i18n.T("Containers"), i18n.T("Images")
	modeSelect := widget.NewSelect([]string{modeContainers, modeImages}, nil)
	modeSelect.SetSelected(modeContainers)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(i18n.T("Search containers (for example: nginx, postgres...)"))

	// The visible labels are translated; the values they map to are Docker's own
	// state words and must not be.
	statusFilterLabels := []string{
		i18n.T("All"), i18n.T("Running"), i18n.T("Stopped"), i18n.T("Paused"),
	}
	statusFilterValues := map[string]string{
		statusFilterLabels[0]: "",
		statusFilterLabels[1]: "running",
		statusFilterLabels[2]: "exited",
		statusFilterLabels[3]: "paused",
	}
	statusFilter := widget.NewSelect(statusFilterLabels, nil)
	statusFilter.SetSelected(statusFilterLabels[0])

	sortByName, sortByStatus := i18n.T("Name"), i18n.T("Status")
	sortSelect := widget.NewSelect([]string{sortByName, sortByStatus}, nil)
	sortSelect.SetSelected(sortByName)

	statusLabel := widget.NewLabel(i18n.T("Status: waiting"))

	var allContainers []system.DockerContainerInfo
	var filteredContainers []system.DockerContainerInfo
	var allImages []system.DockerImageInfo
	var filteredImages []system.DockerImageInfo
	selectedIndex := -1
	lastSelectedContainerName := ""

	autoRefreshCheck := widget.NewCheck(
		i18n.Tf("Auto-refresh (%d s)", appstate.GetConfig().RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.DockerAutoRefresh)

	detailsButton := widget.NewButtonWithIcon(i18n.T("Details"), theme.InfoIcon(), nil)
	inspectButton := widget.NewButtonWithIcon(i18n.T("Inspect"), theme.SearchIcon(), nil)
	logsButton := widget.NewButtonWithIcon(i18n.T("Logs"), theme.DocumentIcon(), nil)
	startButton := widget.NewButtonWithIcon(i18n.T("Start"), theme.MediaPlayIcon(), nil)
	stopButton := widget.NewButtonWithIcon(i18n.T("Stop"), theme.MediaStopIcon(), nil)
	restartButton := widget.NewButtonWithIcon(i18n.T("Restart"), theme.ViewRefreshIcon(), nil)
	favoriteButton := widget.NewButtonWithIcon("", theme.RadioButtonIcon(), nil)

	startButton.Importance = widget.HighImportance
	stopButton.Importance = widget.DangerImportance

	pullButton := widget.NewButtonWithIcon(i18n.T("Pull image"), theme.DownloadIcon(), nil)
	removeImageButton := widget.NewButtonWithIcon(i18n.T("Remove image"), theme.DeleteIcon(), nil)
	removeImageButton.Importance = widget.DangerImportance
	removeImageButton.Disable()

	hideContainerButtons := func() {
		detailsButton.Hide()
		inspectButton.Hide()
		logsButton.Hide()
		startButton.Hide()
		stopButton.Hide()
		restartButton.Hide()
		favoriteButton.Hide()
		pullButton.Show()
		removeImageButton.Show()
	}

	showContainerButtons := func() {
		detailsButton.Show()
		inspectButton.Show()
		logsButton.Show()
		startButton.Show()
		stopButton.Show()
		restartButton.Show()
		favoriteButton.Show()
		pullButton.Hide()
		removeImageButton.Hide()
	}

	isImagesMode := func() bool {
		return modeSelect.Selected == modeImages
	}

	detailsButton.Disable()
	inspectButton.Disable()
	logsButton.Disable()
	startButton.Disable()
	stopButton.Disable()
	restartButton.Disable()
	favoriteButton.Disable()

	updateActionButtons := func() {
		if isImagesMode() {
			hasSelection := selectedIndex >= 0 && selectedIndex < len(filteredImages)
			removeImageButton.Disable()
			if hasSelection {
				removeImageButton.Enable()
			}
			return
		}

		hasSelection := selectedIndex >= 0 && selectedIndex < len(filteredContainers)
		if hasSelection {
			detailsButton.Enable()
			inspectButton.Enable()
			logsButton.Enable()
			startButton.Enable()
			stopButton.Enable()
			restartButton.Enable()
			favoriteButton.Enable()

			c := filteredContainers[selectedIndex]
			if isFavoriteContainer(c.Names) {
				favoriteButton.SetIcon(theme.RadioButtonCheckedIcon())
				favoriteButton.SetText(i18n.T("In favorites"))
			} else {
				favoriteButton.SetIcon(theme.RadioButtonIcon())
				favoriteButton.SetText(i18n.T("Add to favorites"))
			}
			return
		}

		detailsButton.Disable()
		inspectButton.Disable()
		logsButton.Disable()
		startButton.Disable()
		stopButton.Disable()
		restartButton.Disable()
		favoriteButton.Disable()
		favoriteButton.SetIcon(theme.RadioButtonIcon())
		favoriteButton.SetText(i18n.T("Add to favorites"))
	}

	getSelectedContainer := func() (*system.DockerContainerInfo, bool) {
		if selectedIndex < 0 || selectedIndex >= len(filteredContainers) {
			return nil, false
		}
		c := filteredContainers[selectedIndex]
		return &c, true
	}

	sortContainers := func(items []system.DockerContainerInfo) {
		switch sortSelect.Selected {
		case sortByStatus:
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].State == items[j].State {
					return items[i].Names < items[j].Names
				}
				return items[i].State < items[j].State
			})
		default:
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].Names < items[j].Names
			})
		}
	}

	filterContainers := func(query string) []system.DockerContainerInfo {
		query = strings.ToLower(strings.TrimSpace(query))
		status := statusFilterValues[statusFilter.Selected]

		var result []system.DockerContainerInfo
		for _, c := range allContainers {
			if status != "" && !strings.EqualFold(c.State, status) {
				continue
			}

			if query == "" ||
				strings.Contains(strings.ToLower(c.Names), query) ||
				strings.Contains(strings.ToLower(c.Image), query) ||
				strings.Contains(strings.ToLower(c.State), query) ||
				strings.Contains(strings.ToLower(c.Status), query) {
				result = append(result, c)
			}
		}

		sortContainers(result)
		return result
	}

	filterImages := func(query string) []system.DockerImageInfo {
		query = strings.ToLower(strings.TrimSpace(query))

		var result []system.DockerImageInfo
		for _, img := range allImages {
			if query == "" ||
				strings.Contains(strings.ToLower(img.Repository), query) ||
				strings.Contains(strings.ToLower(img.Tag), query) ||
				strings.Contains(strings.ToLower(img.ID), query) {
				result = append(result, img)
			}
		}
		return result
	}

	containerList := widget.NewList(
		func() int {
			if isImagesMode() {
				return len(filteredImages)
			}
			return len(filteredContainers)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("")

			stateLabel := widget.NewLabel("")
			stateLabel.Alignment = fyne.TextAlignTrailing
			stateLabel.TextStyle = fyne.TextStyle{Bold: true}

			return container.NewBorder(nil, nil, nil, stateLabel, nameLabel)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 {
				return
			}

			row := obj.(*fyne.Container)
			nameLabel := row.Objects[0].(*widget.Label)
			stateLabel := row.Objects[1].(*widget.Label)

			if isImagesMode() {
				if id >= len(filteredImages) {
					return
				}
				img := filteredImages[id]
				nameLabel.SetText(fmt.Sprintf("%s:%s", img.Repository, img.Tag))
				stateLabel.SetText(fmt.Sprintf("%s  %s", shortImageID(img.ID), img.Size))
				stateLabel.Importance = widget.MediumImportance
			} else {
				if id >= len(filteredContainers) {
					return
				}
				c := filteredContainers[id]

				prefix := ""
				if isFavoriteContainer(c.Names) {
					prefix = "★ "
				}

				nameLabel.SetText(fmt.Sprintf("%s%s (%s)", prefix, c.Names, c.Image))
				stateLabel.SetText(c.State)
				stateLabel.Importance = StatusImportance(c.State)
			}

			stateLabel.Refresh()
		},
	)

	refreshList := func() {
		if isImagesMode() {
			filteredImages = filterImages(searchEntry.Text)
		} else {
			filteredContainers = filterContainers(searchEntry.Text)
		}

		restoredIndex := -1
		if !isImagesMode() && lastSelectedContainerName != "" {
			for i, c := range filteredContainers {
				if c.Names == lastSelectedContainerName {
					restoredIndex = i
					break
				}
			}
		}

		selectedIndex = restoredIndex
		containerList.Refresh()

		if restoredIndex >= 0 {
			containerList.Select(restoredIndex)
		} else {
			containerList.UnselectAll()
		}

		updateActionButtons()

		count := 0
		if isImagesMode() {
			count = len(filteredImages)
		} else {
			count = len(filteredContainers)
		}

		if count == 0 {
			statusLabel.SetText(i18n.T("Status: 0 results"))
			return
		}

		if isImagesMode() {
			statusLabel.SetText(i18n.Tf("Status: %d images | updated %s", count, time.Now().Format("15:04:05")))
		} else {
			statusLabel.SetText(i18n.Tf("Status: %d containers | updated %s", count, time.Now().Format("15:04:05")))
		}
	}

	showErrorState := func(err error) {
		if isImagesMode() {
			allImages = nil
			filteredImages = nil
		} else {
			allContainers = nil
			filteredContainers = nil
		}
		selectedIndex = -1
		lastSelectedContainerName = ""
		containerList.UnselectAll()
		containerList.Refresh()
		updateActionButtons()
		statusLabel.SetText(i18n.Tf("Status: load error — %s", err.Error()))
	}

	refreshContainers := func() {
		containers, err := system.ListDockerContainers()
		if err != nil {
			fyne.Do(func() {
				showErrorState(err)
			})
			return
		}

		fyne.Do(func() {
			allContainers = containers
			refreshList()
		})
	}

	refreshImages := func() {
		images, err := system.ListDockerImages()
		if err != nil {
			fyne.Do(func() {
				showErrorState(err)
			})
			return
		}

		fyne.Do(func() {
			allImages = images
			refreshList()
		})
	}

	refreshData := func() {
		if isImagesMode() {
			go refreshImages()
		} else {
			go refreshContainers()
		}
	}

	showContainerDetails := func(c system.DockerContainerInfo) {
		content := container.NewVBox(
			widget.NewLabel(i18n.T("Name")+": "+c.Names),
			widget.NewLabel(i18n.T("Image")+": "+c.Image),
			widget.NewLabel(i18n.T("State")+": "+c.State),
			widget.NewLabel(i18n.T("Status")+": "+c.Status),
			widget.NewLabel("ID: "+c.ID),
		)

		dialog.ShowCustom(
			i18n.T("Container details"),
			i18n.T("Close"),
			container.NewPadded(content),
			parent,
		)
	}

	runContainerAction := func(action string) {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText(i18n.T("Select a container from the list"))
			updateActionButtons()
			return
		}

		message := i18n.Tf("Run %s for container %s?", strings.ToUpper(action), c.Names)

		dialog.ShowConfirm(i18n.T("Confirmation"), message, func(confirmed bool) {
			if !confirmed {
				return
			}

			statusLabel.SetText(i18n.Tf("Status: running %s for %s...", action, c.Names))
			detailsButton.Disable()
			logsButton.Disable()
			startButton.Disable()
			stopButton.Disable()
			restartButton.Disable()
			favoriteButton.Disable()

			go func(containerName string) {
				err := system.ControlDockerContainer(action, containerName)
				if err != nil {
					activity.Add("docker", action, containerName, "failed", err.Error())

					fyne.Do(func() {
						if system.IsPermissionError(err) {
							ShowErrorMsg(parent, system.BuildPermissionHint("docker", action, containerName))
						} else {
							ShowError(parent, err)
						}

						statusLabel.SetText(i18n.T("Status: action failed"))
						updateActionButtons()
					})
					return
				}

				activity.Add("docker", action, containerName, "success", "")

				fyne.Do(func() {
					dialog.ShowInformation(
						i18n.T("Done"),
						i18n.Tf("Command %s for %s completed.", strings.ToUpper(action), containerName),
						parent,
					)
				})

				refreshContainers()
			}(c.Names)
		}, parent)
	}

	showContainerLogs := func(containerName string) {
		showLogsWindow(
			i18n.Tf("Container logs: %s", containerName),
			i18n.Tf("Logs for container %s", containerName),
			func() (string, error) {
				return system.GetDockerContainerLogs(containerName, cfg.DefaultLogLines)
			},
			cfg,
		)
	}

	containerList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		if !isImagesMode() && id >= 0 && id < len(filteredContainers) {
			lastSelectedContainerName = filteredContainers[id].Names
		}
		updateActionButtons()
	}

	containerList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
		lastSelectedContainerName = ""
		updateActionButtons()
	}

	searchEntry.OnChanged = func(string) {
		refreshList()
	}

	statusFilter.OnChanged = func(string) {
		refreshList()
	}

	sortSelect.OnChanged = func(string) {
		refreshList()
	}

	modeSelect.OnChanged = func(mode string) {
		switch mode {
		case modeImages:
			subtitle.SetText(i18n.T("View and manage Docker images"))
			searchEntry.SetPlaceHolder(i18n.T("Search images (for example: nginx, ubuntu...)"))
			statusFilter.Hide()
			sortSelect.Hide()
			hideContainerButtons()
		default:
			subtitle.SetText(i18n.T("View and manage Docker containers"))
			searchEntry.SetPlaceHolder(i18n.T("Search containers (for example: nginx, postgres...)"))
			statusFilter.Show()
			sortSelect.Show()
			showContainerButtons()
		}
		selectedIndex = -1
		containerList.UnselectAll()
		refreshData()
	}

	autoRefresh := newAutoRefresher(refreshData)
	RegisterCloser(autoRefresh.Stop)
	autoRefreshCheck.OnChanged = autoRefresh.SetEnabled
	autoRefresh.SetEnabled(cfg.DockerAutoRefresh)

	pullButton.OnTapped = func() {
		imageEntry := widget.NewEntry()
		imageEntry.SetPlaceHolder("nginx:latest, ubuntu:22.04...")

		dialog.ShowCustomConfirm(
			i18n.T("Pull a Docker image"),
			i18n.T("Pull"),
			i18n.T("Cancel"),
			container.NewPadded(container.NewVBox(
				widget.NewLabel(i18n.T("Enter the name of the image to pull:")),
				imageEntry,
			)),
			func(confirmed bool) {
				if !confirmed || strings.TrimSpace(imageEntry.Text) == "" {
					return
				}

				imageName := strings.TrimSpace(imageEntry.Text)
				statusLabel.SetText(i18n.Tf("Status: pulling %s...", imageName))
				pullButton.Disable()

				go func() {
					err := system.PullDockerImage(imageName)

					fyne.Do(func() {
						pullButton.Enable()
						if err != nil {
							ShowError(parent, err)
							statusLabel.SetText(i18n.T("Status: pull failed"))
							return
						}

						activity.Add("docker", "pull", imageName, "success", "")
						dialog.ShowInformation(i18n.T("Done"), i18n.Tf("Image %s pulled.", imageName), parent)
						refreshImages()
					})
				}()
			},
			parent,
		)
	}

	removeImageButton.OnTapped = func() {
		if selectedIndex < 0 || selectedIndex >= len(filteredImages) {
			return
		}
		img := filteredImages[selectedIndex]
		imageName := img.Repository + ":" + img.Tag

		dialog.ShowConfirm(
			i18n.T("Remove image"),
			i18n.Tf("Remove image %s?", imageName),
			func(confirmed bool) {
				if !confirmed {
					return
				}

				statusLabel.SetText(i18n.Tf("Status: removing %s...", imageName))
				removeImageButton.Disable()

				go func() {
					err := system.RemoveDockerImage(imageName)

					fyne.Do(func() {
						removeImageButton.Enable()
						if err != nil {
							ShowError(parent, err)
							statusLabel.SetText(i18n.T("Status: remove failed"))
							return
						}

						activity.Add("docker", "rmi", imageName, "success", "")
						statusLabel.SetText(i18n.T("Status: image removed"))
						refreshImages()
					})
				}()
			},
			parent,
		)
	}

	favoriteButton.OnTapped = func() {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText(i18n.T("Select a container from the list"))
			updateActionButtons()
			return
		}

		if err := toggleFavoriteContainer(c.Names); err != nil {
			activity.Add("docker", "favorite", c.Names, "failed", err.Error())
			ShowError(parent, err)
			statusLabel.SetText(i18n.T("Status: could not save favorites"))
			return
		}

		if isFavoriteContainer(c.Names) {
			activity.Add("docker", "favorite", c.Names, "success", "added to favorites")
		} else {
			activity.Add("docker", "favorite", c.Names, "success", "removed from favorites")
		}

		refreshList()
	}

	detailsButton.OnTapped = func() {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText(i18n.T("Select a container from the list"))
			updateActionButtons()
			return
		}

		showContainerDetails(*c)
	}

	inspectButton.OnTapped = func() {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText(i18n.T("Select a container from the list"))
			updateActionButtons()
			return
		}

		statusLabel.SetText(i18n.T("Status: loading docker inspect..."))

		go func(containerName string) {
			result, err := system.GetDockerContainerInspect(containerName)

			fyne.Do(func() {
				if err != nil {
					if system.IsPermissionError(err) {
						ShowErrorMsg(parent, system.BuildPermissionHint("docker", "inspect", containerName))
					} else {
						ShowError(parent, err)
					}
					statusLabel.SetText(i18n.T("Status: docker inspect failed"))
					return
				}

				output := widget.NewMultiLineEntry()
				output.SetText(result)
				output.Wrapping = fyne.TextWrapWord
				output.Disable()

				dialog.ShowCustom(
					i18n.Tf("Container inspect: %s", containerName),
					i18n.T("Close"),
					container.NewPadded(output),
					parent,
				)

				statusLabel.SetText(i18n.T("Status: docker inspect loaded"))
			})
		}(c.Names)
	}

	logsButton.OnTapped = func() {
		c, ok := getSelectedContainer()
		if !ok {
			statusLabel.SetText(i18n.T("Select a container from the list"))
			updateActionButtons()
			return
		}

		showContainerLogs(c.Names)
	}

	startButton.OnTapped = func() {
		runContainerAction("start")
	}

	stopButton.OnTapped = func() {
		runContainerAction("stop")
	}

	restartButton.OnTapped = func() {
		runContainerAction("restart")
	}

	refreshButton := widget.NewButton(i18n.T("Refresh"), func() {
		go refreshData()
	})

	actionsRow := container.NewHBox(
		refreshButton,
		autoRefreshCheck,
		favoriteButton,
		detailsButton,
		inspectButton,
		logsButton,
		startButton,
		stopButton,
		restartButton,
		pullButton,
		removeImageButton,
	)

	// The tab opens in Containers mode, so show that mode's buttons. This used
	// to call hideContainerButtons(), which left the container actions hidden
	// and the image-only actions showing until the user toggled the mode.
	showContainerButtons()

	content := container.NewBorder(
		container.NewPadded(
			container.NewVBox(
				title,
				subtitle,
				widget.NewSeparator(),
				modeSelect,
				container.NewGridWithColumns(2, searchEntry, statusFilter),
				sortSelect,
				actionsRow,
				statusLabel,
				widget.NewSeparator(),
			),
		),
		nil,
		nil,
		nil,
		container.NewPadded(containerList),
	)

	go refreshContainers()
	RegisterRefresh("Docker", refreshData)

	return content
}
