package ui

import (
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

func buildServicesTab(parent fyne.Window, cfg config.Config) fyne.CanvasObject {
	// The sidebar names the section; the caption keeps the useful part — which
	// service manager is backing this (systemd / Windows Services).
	subtitle := widget.NewLabel(i18n.Tf("View and manage services (%s)", system.ServiceManagerName()))

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(i18n.T("Search (for example: ssh, docker...)"))

	// The visible labels are translated; the values they map to are systemd's
	// own ActiveState words and must not be.
	statusFilterLabels := []string{
		i18n.T("All"), i18n.T("Active"), i18n.T("Inactive"), i18n.T("Failed"), i18n.T("Starting"),
	}
	statusFilterValues := map[string]string{
		statusFilterLabels[0]: "",
		statusFilterLabels[1]: "active",
		statusFilterLabels[2]: "inactive",
		statusFilterLabels[3]: "failed",
		statusFilterLabels[4]: "activating",
	}
	statusFilter := widget.NewSelect(statusFilterLabels, nil)
	statusFilter.SetSelected(statusFilterLabels[0])

	sortByName, sortByStatus := i18n.T("Name"), i18n.T("Status")
	sortSelect := widget.NewSelect([]string{sortByName, sortByStatus}, nil)
	sortSelect.SetSelected(sortByName)

	statusLabel := newDataLabel(i18n.T("Status: waiting"))

	var allServices []system.ServiceInfo
	var filteredServices []system.ServiceInfo
	selectedIndex := -1
	lastSelectedServiceName := ""

	autoRefreshCheck := widget.NewCheck(
		i18n.Tf("Auto-refresh (%d s)", appstate.GetConfig().RefreshIntervalSeconds),
		nil,
	)
	autoRefreshCheck.SetChecked(cfg.ServicesAutoRefresh)

	detailsButton := widget.NewButtonWithIcon(i18n.T("Details"), theme.InfoIcon(), nil)
	openUnitButton := widget.NewButtonWithIcon(i18n.T("Unit file"), theme.FileTextIcon(), nil)
	logsButton := widget.NewButtonWithIcon(i18n.T("Logs"), theme.DocumentIcon(), nil)
	startButton := widget.NewButtonWithIcon(i18n.T("Start"), theme.MediaPlayIcon(), nil)
	stopButton := widget.NewButtonWithIcon(i18n.T("Stop"), theme.MediaStopIcon(), nil)
	restartButton := widget.NewButtonWithIcon(i18n.T("Restart"), theme.ViewRefreshIcon(), nil)
	enableButton := widget.NewButtonWithIcon(i18n.T("Enable"), theme.ConfirmIcon(), nil)
	disableButton := widget.NewButtonWithIcon(i18n.T("Disable"), theme.CancelIcon(), nil)
	favoriteButton := widget.NewButtonWithIcon("", theme.RadioButtonIcon(), nil)

	startButton.Importance = widget.HighImportance
	stopButton.Importance = widget.DangerImportance

	detailsButton.Disable()
	openUnitButton.Disable()
	logsButton.Disable()
	startButton.Disable()
	stopButton.Disable()
	restartButton.Disable()
	enableButton.Disable()
	disableButton.Disable()
	favoriteButton.Disable()

	updateActionButtons := func() {
		hasSelection := selectedIndex >= 0 && selectedIndex < len(filteredServices)
		if hasSelection {
			detailsButton.Enable()
			openUnitButton.Enable()
			logsButton.Enable()
			startButton.Enable()
			stopButton.Enable()
			restartButton.Enable()
			enableButton.Enable()
			disableButton.Enable()
			favoriteButton.Enable()

			svc := filteredServices[selectedIndex]
			if isFavoriteService(svc.Name) {
				favoriteButton.SetIcon(theme.RadioButtonCheckedIcon())
				favoriteButton.SetText(i18n.T("In favorites"))
			} else {
				favoriteButton.SetIcon(theme.RadioButtonIcon())
				favoriteButton.SetText(i18n.T("Add to favorites"))
			}
			return
		}

		detailsButton.Disable()
		openUnitButton.Disable()
		logsButton.Disable()
		startButton.Disable()
		stopButton.Disable()
		restartButton.Disable()
		enableButton.Disable()
		disableButton.Disable()
		favoriteButton.Disable()
		favoriteButton.SetIcon(theme.RadioButtonIcon())
		favoriteButton.SetText(i18n.T("Add to favorites"))
	}

	getSelectedService := func() (*system.ServiceInfo, bool) {
		if selectedIndex < 0 || selectedIndex >= len(filteredServices) {
			return nil, false
		}

		svc := filteredServices[selectedIndex]
		return &svc, true
	}

	sortServices := func(items []system.ServiceInfo) {
		switch sortSelect.Selected {
		case sortByStatus:
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].ActiveState == items[j].ActiveState {
					return items[i].Name < items[j].Name
				}
				return items[i].ActiveState < items[j].ActiveState
			})
		default:
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].Name < items[j].Name
			})
		}
	}

	filterServices := func(query string) []system.ServiceInfo {
		query = strings.ToLower(strings.TrimSpace(query))
		status := statusFilterValues[statusFilter.Selected]

		var result []system.ServiceInfo
		for _, svc := range allServices {
			if status != "" && !strings.EqualFold(svc.ActiveState, status) {
				continue
			}

			if query == "" ||
				strings.Contains(strings.ToLower(svc.Name), query) ||
				strings.Contains(strings.ToLower(svc.Description), query) {
				result = append(result, svc)
			}
		}

		sortServices(result)
		return result
	}

	// Assigned once the actions below exist; the row callbacks read them at
	// click time, not at construction, so the forward reference is fine.
	var onRowMenu func(id widget.ListItemID, e *fyne.PointEvent)
	var onRowActivate func(id widget.ListItemID)

	serviceList := widget.NewList(
		func() int {
			return len(filteredServices)
		},
		func() fyne.CanvasObject {
			// The state badge leads the row. A trailing state (border layout's
			// right slot, or an HBox spacer) rendered off the visible width and
			// was never seen; a leading, fixed-width badge always is.
			stateLabel := widget.NewLabel(statePlaceholder)
			stateLabel.TextStyle = fyne.TextStyle{Bold: true}

			nameLabel := widget.NewLabel("service")

			return newTappableRow(container.NewHBox(stateLabel, nameLabel))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(filteredServices) {
				return
			}

			svc := filteredServices[id]

			row := obj.(*tappableRow)
			hbox := row.content.(*fyne.Container)
			stateLabel := hbox.Objects[0].(*widget.Label)
			nameLabel := hbox.Objects[1].(*widget.Label)

			prefix := ""
			if isFavoriteService(svc.Name) {
				prefix = "★ "
			}

			nameLabel.SetText(prefix + svc.Name)
			stateLabel.SetText(padState(svc.ActiveState))
			stateLabel.Importance = StatusImportance(svc.ActiveState)
			stateLabel.Refresh()

			rowID := id
			row.onSecondary = func(e *fyne.PointEvent) { onRowMenu(rowID, e) }
			row.onDouble = func() { onRowActivate(rowID) }
		},
	)

	refreshList := func() {
		filteredServices = filterServices(searchEntry.Text)

		restoredIndex := -1
		if lastSelectedServiceName != "" {
			for i, svc := range filteredServices {
				if svc.Name == lastSelectedServiceName {
					restoredIndex = i
					break
				}
			}
		}

		selectedIndex = restoredIndex
		serviceList.Refresh()

		if restoredIndex >= 0 {
			serviceList.Select(restoredIndex)
		} else {
			serviceList.UnselectAll()
		}

		updateActionButtons()

		if len(filteredServices) == 0 {
			statusLabel.SetText(i18n.T("Status: 0 results"))
			return
		}

		statusLabel.SetText(
			i18n.Tf(
				"Status: %d services | updated %s",
				len(filteredServices),
				time.Now().Format("15:04:05"),
			),
		)
	}

	showErrorState := func(err error) {
		allServices = nil
		filteredServices = nil
		selectedIndex = -1
		lastSelectedServiceName = ""
		serviceList.UnselectAll()
		serviceList.Refresh()
		updateActionButtons()
		statusLabel.SetText(i18n.Tf("Status: load error — %s", err.Error()))
	}

	refreshServices := func() {
		services, err := system.ListServices()
		if err != nil {
			fyne.Do(func() {
				showErrorState(err)
			})
			return
		}

		fyne.Do(func() {
			allServices = services
			refreshList()
		})
	}

	showServiceDetails := func(svc system.ServiceInfo) {
		description := widget.NewLabel(svc.Description)
		description.Wrapping = fyne.TextWrapWord

		content := container.NewVBox(
			widget.NewLabel(i18n.T("Name")+": "+svc.Name),
			widget.NewLabel(i18n.T("Load state")+": "+svc.LoadState),
			widget.NewLabel(i18n.T("Active state")+": "+svc.ActiveState),
			widget.NewLabel(i18n.T("Sub state")+": "+svc.SubState),
			widget.NewSeparator(),
			widget.NewLabel(i18n.T("Description")+":"),
			description,
		)

		dialog.ShowCustom(
			i18n.T("Service details"),
			i18n.T("Close"),
			container.NewPadded(content),
			parent,
		)
	}

	runServiceAction := func(action string) {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText(i18n.T("Select a service from the list"))
			updateActionButtons()
			return
		}

		message := i18n.Tf("Run %s for %s?", strings.ToUpper(action), svc.Name)

		dialog.ShowConfirm(i18n.T("Confirmation"), message, func(confirmed bool) {
			if !confirmed {
				return
			}

			statusLabel.SetText(i18n.Tf("Status: running %s for %s...", action, svc.Name))
			detailsButton.Disable()
			logsButton.Disable()
			startButton.Disable()
			stopButton.Disable()
			restartButton.Disable()
			enableButton.Disable()
			disableButton.Disable()
			favoriteButton.Disable()

			go func(serviceName string) {
				err := system.ControlService(action, serviceName)
				if err != nil {
					activity.Add("service", action, serviceName, "failed", err.Error())

					fyne.Do(func() {
						if system.IsPermissionError(err) {
							ShowErrorMsg(parent, system.BuildPermissionHint("service", action, serviceName))
						} else {
							ShowError(parent, err)
						}

						statusLabel.SetText(i18n.T("Status: action failed"))
						updateActionButtons()
					})
					return
				}
				activity.Add("service", action, serviceName, "success", "")

				fyne.Do(func() {
					dialog.ShowInformation(
						i18n.T("Done"),
						i18n.Tf("Command %s for %s completed.", strings.ToUpper(action), serviceName),
						parent,
					)
				})
				refreshServices()
			}(svc.Name)
		}, parent)
	}

	serviceList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		if id >= 0 && id < len(filteredServices) {
			lastSelectedServiceName = filteredServices[id].Name
		}
		updateActionButtons()
	}

	serviceList.OnUnselected = func(id widget.ListItemID) {
		selectedIndex = -1
		lastSelectedServiceName = ""
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

	autoRefresh := newAutoRefresher(refreshServices)
	RegisterCloser(autoRefresh.Stop)
	autoRefreshCheck.OnChanged = autoRefresh.SetEnabled
	autoRefresh.SetEnabled(cfg.ServicesAutoRefresh)

	favoriteButton.OnTapped = func() {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText(i18n.T("Select a service from the list"))
			updateActionButtons()
			return
		}

		if err := toggleFavoriteService(svc.Name); err != nil {
			activity.Add("service", "favorite", svc.Name, "failed", err.Error())
			ShowError(parent, err)
			statusLabel.SetText(i18n.T("Status: could not save favorites"))
			return
		}

		if isFavoriteService(svc.Name) {
			activity.Add("service", "favorite", svc.Name, "success", "added to favorites")
		} else {
			activity.Add("service", "favorite", svc.Name, "success", "removed from favorites")
		}

		refreshList()
	}

	detailsButton.OnTapped = func() {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText(i18n.T("Select a service from the list"))
			updateActionButtons()
			return
		}

		showServiceDetails(*svc)
	}

	openUnitButton.OnTapped = func() {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText(i18n.T("Select a service from the list"))
			updateActionButtons()
			return
		}

		unitPath, err := system.FindServiceUnitFile(svc.Name)
		if err != nil {
			ShowError(parent, err)
			statusLabel.SetText(i18n.T("Status: unit file not found"))
			return
		}

		if err := OpenFileInFiles(unitPath); err != nil {
			ShowError(parent, err)
			statusLabel.SetText(i18n.T("Status: could not open the unit file"))
			return
		}

		dialog.ShowInformation(
			i18n.T("Unit file opened"),
			i18n.Tf("The file is open in the «Files» tab:\n%s", unitPath),
			parent,
		)
		statusLabel.SetText(i18n.T("Status: unit file opened"))
	}

	logsButton.OnTapped = func() {
		svc, ok := getSelectedService()
		if !ok {
			statusLabel.SetText(i18n.T("Select a service from the list"))
			updateActionButtons()
			return
		}

		statusLabel.SetText(i18n.T("Opening the log window..."))

		showLogsWindow(
			i18n.Tf("Service logs: %s", svc.Name),
			i18n.Tf("Logs for service %s", svc.Name),
			func() (string, error) {
				return system.GetServiceLogs(svc.Name, cfg.DefaultLogLines)
			},
			cfg,
		)

		statusLabel.SetText(i18n.T("Log window opened"))
	}

	startButton.OnTapped = func() {
		runServiceAction("start")
	}

	stopButton.OnTapped = func() {
		runServiceAction("stop")
	}

	restartButton.OnTapped = func() {
		runServiceAction("restart")
	}

	enableButton.OnTapped = func() {
		runServiceAction("enable")
	}

	disableButton.OnTapped = func() {
		runServiceAction("disable")
	}

	// A right-click selects the row, then offers the same actions as the button
	// bar. Selecting first means the existing handlers, which act on the
	// selection, need no change.
	onRowMenu = func(id widget.ListItemID, e *fyne.PointEvent) {
		serviceList.Select(id)
		if id < 0 || id >= len(filteredServices) {
			return
		}
		svc := filteredServices[id]

		favourite := i18n.T("Add to favorites")
		if isFavoriteService(svc.Name) {
			favourite = i18n.T("In favorites")
		}

		showRowMenu(parent, e, []rowAction{
			{label: i18n.T("Start"), do: func() { runServiceAction("start") }},
			{label: i18n.T("Stop"), do: func() { runServiceAction("stop") }},
			{label: i18n.T("Restart"), do: func() { runServiceAction("restart") }},
			{label: i18n.T("Logs"), do: func() { logsButton.OnTapped() }},
			{label: i18n.T("Details"), do: func() { detailsButton.OnTapped() }},
			{label: favourite, do: func() { favoriteButton.OnTapped() }},
		})
	}

	// Double-click opens details, the same as the Details button.
	onRowActivate = func(id widget.ListItemID) {
		serviceList.Select(id)
		detailsButton.OnTapped()
	}

	refreshButton := widget.NewButtonWithIcon(i18n.T("Refresh"), theme.ViewRefreshIcon(), func() {
		startInitialLoad(refreshServices)
	})

	// Lifecycle actions the user reaches for most, then the inspection and
	// bookkeeping ones. Both wrap rather than forcing the window wide enough to
	// hold every button on one line.
	actionsRow := newToolbarRow(
		startButton, stopButton, restartButton,
		enableButton, disableButton,
	)
	secondaryRow := container.NewVBox(
		newToolbarRow(
			favoriteButton, detailsButton, logsButton, openUnitButton, refreshButton,
		),
		autoRefreshCheck,
	)

	content := container.NewBorder(
		container.NewPadded(
			container.NewVBox(
				subtitle,
				widget.NewSeparator(),
				container.NewGridWithColumns(3, searchEntry, statusFilter, sortSelect),
				actionsRow,
				secondaryRow,
				statusLabel,
				widget.NewSeparator(),
			),
		),
		nil,
		nil,
		nil,
		container.NewPadded(serviceList),
	)

	startInitialLoad(refreshServices)
	RegisterRefresh("Services", refreshServices)

	return content
}
