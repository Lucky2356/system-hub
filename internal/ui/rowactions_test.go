package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestTappableRowForwardsSecondaryAndDouble(t *testing.T) {
	test.NewApp()

	row := newTappableRow(widget.NewLabel("row"))

	secondary := 0
	double := 0
	row.onSecondary = func(*fyne.PointEvent) { secondary++ }
	row.onDouble = func() { double++ }

	row.TappedSecondary(&fyne.PointEvent{})
	row.DoubleTapped(&fyne.PointEvent{})

	if secondary != 1 {
		t.Errorf("secondary tap fired %d times, want 1", secondary)
	}
	if double != 1 {
		t.Errorf("double tap fired %d times, want 1", double)
	}
}

func TestTappableRowIsSafeWithoutCallbacks(t *testing.T) {
	test.NewApp()

	// The list sets the callbacks in its update function; a row that has been
	// created but not yet updated must not panic on a stray click.
	row := newTappableRow(widget.NewLabel("row"))
	row.TappedSecondary(&fyne.PointEvent{})
	row.DoubleTapped(&fyne.PointEvent{})
}
