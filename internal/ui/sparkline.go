package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// sparkline draws a series as a small line chart.
//
// It is built from canvas.Line segments rather than a raster: the series is at
// most HistoryCapacity points, and lines are drawn by the GPU and stay crisp at
// any scale, whereas a raster would have to be re-rasterised on every resize
// and on every tick.
type sparkline struct {
	widgetBase *fyne.Container
	lines      []*canvas.Line
	colour     color.Color

	values []float64
	// maxValue fixes the top of the chart. Zero means "scale to the data",
	// which suits network traffic; percentages pass 100 so that a flat 5% load
	// looks flat rather than filling the card.
	maxValue float64
}

// sparklineHeight keeps the chart readable without crowding the card's numbers.
const sparklineHeight = 32

// newSparkline creates an empty chart. maxValue fixes the vertical scale; pass
// 0 to scale to whatever the data holds.
func newSparkline(colour color.Color, maxValue float64) *sparkline {
	s := &sparkline{
		widgetBase: container.NewWithoutLayout(),
		colour:     colour,
		maxValue:   maxValue,
	}
	s.widgetBase.Resize(fyne.NewSize(0, sparklineHeight))
	return s
}

// object returns the canvas object to place in a card.
func (s *sparkline) object() fyne.CanvasObject {
	// The layout holds the sparkline so a resize redraws it: the segment
	// positions are absolute, so without this the chart would keep the width it
	// happened to have when the data last changed.
	return container.New(&sparklineLayout{chart: s}, s.widgetBase)
}

// setValues replaces the series and redraws.
func (s *sparkline) setValues(values []float64) {
	s.values = values
	s.refresh()
}

func (s *sparkline) refresh() {
	size := s.widgetBase.Size()
	if size.Width <= 0 || size.Height <= 0 {
		return
	}

	points := sparklinePoints(s.values, s.maxValue, size)

	// Fewer than two points means there is no segment to draw yet.
	if len(points) < 2 {
		for _, line := range s.lines {
			line.Hide()
		}
		return
	}

	segments := len(points) - 1
	s.ensureLines(segments)

	for i := 0; i < segments; i++ {
		line := s.lines[i]
		line.Position1 = points[i]
		line.Position2 = points[i+1]
		line.StrokeColor = s.colour
		line.StrokeWidth = 1.5
		line.Show()
		line.Refresh()
	}

	// Hide any leftovers from a longer previous series instead of reallocating.
	for i := segments; i < len(s.lines); i++ {
		s.lines[i].Hide()
	}
}

// sparklinePoints maps a series onto canvas positions.
//
// maxValue fixes the top of the chart; pass 0 to scale to the series' own peak.
// Values above the scale are clamped rather than drawn outside the card.
func sparklinePoints(values []float64, maxValue float64, size fyne.Size) []fyne.Position {
	if len(values) < 2 {
		return nil
	}

	scale := maxValue
	if scale <= 0 {
		for _, v := range values {
			if v > scale {
				scale = v
			}
		}
	}
	if scale <= 0 {
		// An all-zero series (an idle network) draws flat along the bottom
		// rather than dividing by zero.
		scale = 1
	}

	stepX := size.Width / float32(len(values)-1)
	points := make([]fyne.Position, len(values))

	for i, v := range values {
		ratio := v / scale
		if ratio > 1 {
			ratio = 1
		}
		if ratio < 0 {
			ratio = 0
		}
		// Y grows downward on a canvas, so a high value sits near the top.
		points[i] = fyne.NewPos(float32(i)*stepX, size.Height-float32(ratio)*size.Height)
	}

	return points
}

// ensureLines grows the segment pool to n. Lines are reused across ticks: the
// dashboard redraws every couple of seconds, and rebuilding the objects each
// time would churn the canvas and the GC.
func (s *sparkline) ensureLines(n int) {
	for len(s.lines) < n {
		line := canvas.NewLine(s.colour)
		line.StrokeWidth = 1.5
		s.lines = append(s.lines, line)
		s.widgetBase.Add(line)
	}
}

// sparklineLayout gives the chart the full width of its card and a fixed
// height, and redraws it when the width changes.
type sparklineLayout struct {
	chart *sparkline
}

func (l *sparklineLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(size.Width, sparklineHeight))
		o.Move(fyne.NewPos(0, 0))
	}
	l.chart.refresh()
}

func (l *sparklineLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, sparklineHeight)
}
