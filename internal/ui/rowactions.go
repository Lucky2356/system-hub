package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// tappableRow wraps a list row so a right-click and a double-click reach it.
//
// widget.List does not forward secondary or double taps to the objects it
// renders, so without this the only way to act on a row was to select it and
// then walk the mouse up to a button bar. The row itself is now the handle:
// double-click for the primary action, right-click for a menu of the rest —
// the actions live where the object is.
//
// The list reuses row objects as the user scrolls, so the callbacks read the
// row's *current* index through the closures the update function sets, not a
// captured one.
type tappableRow struct {
	widget.BaseWidget
	content     fyne.CanvasObject
	onSecondary func(*fyne.PointEvent)
	onDouble    func()
}

func newTappableRow(content fyne.CanvasObject) *tappableRow {
	r := &tappableRow{content: content}
	r.ExtendBaseWidget(r)
	return r
}

func (r *tappableRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}

// TappedSecondary opens the row's context menu.
func (r *tappableRow) TappedSecondary(e *fyne.PointEvent) {
	if r.onSecondary != nil {
		r.onSecondary(e)
	}
}

// DoubleTapped runs the row's primary action.
func (r *tappableRow) DoubleTapped(*fyne.PointEvent) {
	if r.onDouble != nil {
		r.onDouble()
	}
}

// rowAction is one entry in a row's context menu.
type rowAction struct {
	label string
	do    func()
}

// showRowMenu pops up a context menu for a row at the click position.
func showRowMenu(w fyne.Window, e *fyne.PointEvent, actions []rowAction) {
	items := make([]*fyne.MenuItem, 0, len(actions))
	for _, a := range actions {
		items = append(items, fyne.NewMenuItem(a.label, a.do))
	}
	widget.ShowPopUpMenuAtPosition(fyne.NewMenu("", items...), w.Canvas(), e.AbsolutePosition)
}
