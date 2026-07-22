package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// emptyState is a centred message shown over a list that has nothing to show,
// so an unavailable backend reads as an explanation instead of a blank pane
// with a raw error in the status bar.
type emptyState struct {
	overlay *fyne.Container
	label   *widget.Label
}

// newListWithEmptyState stacks a message over a list. The message is hidden
// while the list has rows and shown, with the given text, when it does not.
func newListWithEmptyState(list fyne.CanvasObject) (fyne.CanvasObject, *emptyState) {
	label := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{})
	label.Wrapping = fyne.TextWrapWord

	es := &emptyState{
		overlay: container.NewCenter(label),
		label:   label,
	}
	es.overlay.Hide()

	return container.NewStack(list, es.overlay), es
}

// show displays the message; hide returns to the list.
func (e *emptyState) show(message string) {
	e.label.SetText(message)
	e.overlay.Show()
}

func (e *emptyState) hide() {
	e.overlay.Hide()
}

// setTextIfChanged updates a multi-line entry only when its content differs.
//
// widget.Entry.SetText resets the scroll position, so calling it every time the
// log is re-read — which follow mode does once a second — yanked the view back
// to the top while the user was reading further down. Comparing first keeps the
// scroll steady when nothing changed.
func setTextIfChanged(entry *widget.Entry, text string) {
	if entry.Text == text {
		return
	}
	entry.SetText(text)
}

// loadInitialData reports whether a tab should fetch its data as soon as it is
// built. It is always true in the app; the layout tests turn it off.
//
// The seam exists because those tests measure geometry, and a tab that starts
// polling the host while it is being measured races Fyne's font cache through
// fyne.Do — a race the real app does not have, since there fyne.Do marshals the
// work to the main goroutine rather than running it inline. This mirrors the
// injectable runner in internal/system/exec.go, which the command tests use for
// the same reason.
var loadInitialData = true

// startInitialLoad runs a tab's first data fetch unless the layout tests have
// disabled it.
func startInitialLoad(load func()) {
	if loadInitialData {
		go load()
	}
}

// NewStatCard builds the standard card: a title, a caption and a padded content
// area.
func NewStatCard(title string, subtitle string, content fyne.CanvasObject) *widget.Card {
	return widget.NewCard(
		title,
		subtitle,
		container.NewPadded(content),
	)
}

// Dashboard cell sizes. They are fixed because the dashboard lays out in a
// wrapping grid, which is what keeps its minimum width at one cell instead of
// the whole row. Anything placed in a cell has to fit it — labels truncate,
// lists scroll — or it will draw over its neighbour.
//
// The tile height is measured, not guessed: the tallest composition is the
// network tile at 219px (two values and two charts), and
// TestDashboardTilesFitTheirCells fails if any tile outgrows the cell again.
// At 150 the uptime tile drew across the CPU chart, and at 190 the network
// tile's second chart spilled past its own bottom edge.
var (
	dashboardTileSize  = fyne.NewSize(240, 225)
	dashboardPanelSize = fyne.NewSize(320, 225)
)

// NewTile builds a dashboard tile: a title and content, no caption.
//
// The captions ("CPU" / "Current processor load") were noise — the title
// already said it — and they were not free: a Card sizes itself to its title
// and subtitle, so a sentence-long caption is what made these cards ~490px wide
// to hold a single line of content.
func NewTile(title string, content fyne.CanvasObject) *widget.Card {
	return widget.NewCard(title, "", container.NewPadded(content))
}

// newDataLabel creates a label for text that comes from the host — statuses,
// paths, names, error messages — rather than from our own captions.
//
// The distinction matters for layout, not style. A widget.Label reports the full
// width of its text as its minimum size, so a label holding host output lets
// that output dictate how wide the window must be. That is not theoretical: the
// docker-unavailable error is 1424px wide as a plain label, and because
// container.AppTabs takes its minimum as the maximum over every tab, that one
// string forced the whole window to 2015px — wider than a 1080p screen, and
// wider than the app could ever be resized back down from.
//
// Truncating makes the minimum a constant (~13px) regardless of the text, so the
// layout is decided by the design and not by whatever the host returns.
func newDataLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Truncation = fyne.TextTruncateEllipsis
	return label
}

// newWrappingDataLabel is for host text that should stay fully readable in a
// block (problem lists, details panes) rather than being cut to one line.
// Wrapping also keeps the minimum width bounded, because the text reflows
// instead of extending.
func newWrappingDataLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

// toolbarCell is sized for the longest button caption in either language
// ("Переименовать", "Диагностический пакет" wraps to its own row anyway).
var toolbarCell = fyne.NewSize(150, 36)

// newToolbarRow lays actions out in a grid that wraps instead of one row that
// does not.
//
// An HBox reports the sum of its children as its minimum width, so a row of
// eleven buttons — which is what the Files tab had — made the tab, and therefore
// the whole window, that wide forever. A wrapping grid reports a single cell
// instead, so the same buttons reflow onto more rows in a narrow window.
func newToolbarRow(actions ...fyne.CanvasObject) fyne.CanvasObject {
	return container.NewGridWrap(toolbarCell, actions...)
}
