package ui

import (
	"testing"

	"fyne.io/fyne/v2"
)

// size100 is a 100x100 box, so a position maps directly onto a percentage and
// the expectations below read as coordinates.
var size100 = fyne.NewSize(100, 100)

func TestSparklinePointsInvertsY(t *testing.T) {
	// A canvas grows downward, so the highest value must produce the *smallest*
	// Y. Getting this backwards draws a chart that is upside down but otherwise
	// plausible, which is exactly what a screenshot fails to catch.
	points := sparklinePoints([]float64{0, 50, 100}, 100, size100)

	if len(points) != 3 {
		t.Fatalf("got %d points, want 3", len(points))
	}

	if points[0].Y != 100 {
		t.Errorf("0%% should sit at the bottom (Y=100), got Y=%v", points[0].Y)
	}
	if points[1].Y != 50 {
		t.Errorf("50%% should sit at the middle (Y=50), got Y=%v", points[1].Y)
	}
	if points[2].Y != 0 {
		t.Errorf("100%% should sit at the top (Y=0), got Y=%v", points[2].Y)
	}
}

func TestSparklinePointsSpreadAcrossTheWidth(t *testing.T) {
	points := sparklinePoints([]float64{1, 2, 3, 4, 5}, 5, size100)

	if points[0].X != 0 {
		t.Errorf("the first point should start at X=0, got %v", points[0].X)
	}
	if last := points[len(points)-1].X; last != 100 {
		t.Errorf("the last point should reach the full width (X=100), got %v", last)
	}
}

func TestSparklinePointsUseTheFixedScaleWhenGiven(t *testing.T) {
	// A percentage chart is pinned to 0-100: an idle CPU must look idle, not be
	// stretched to fill the card.
	points := sparklinePoints([]float64{5, 5, 5}, 100, size100)

	for i, p := range points {
		if p.Y != 95 {
			t.Errorf("point %d: a flat 5%% on a 0-100 scale should sit at Y=95, got %v", i, p.Y)
		}
	}
}

func TestSparklinePointsAutoScaleToThePeak(t *testing.T) {
	// Network traffic has no natural ceiling, so scale 0 means "fit the data":
	// the peak touches the top and the rest is relative to it.
	points := sparklinePoints([]float64{0, 200, 400}, 0, size100)

	if points[2].Y != 0 {
		t.Errorf("the peak should touch the top (Y=0), got %v", points[2].Y)
	}
	if points[1].Y != 50 {
		t.Errorf("half the peak should sit at the middle (Y=50), got %v", points[1].Y)
	}
}

func TestSparklinePointsClampAboveTheScale(t *testing.T) {
	// CPU percentages can momentarily exceed 100 on some hosts; the line must
	// stay inside the card instead of being drawn over the card above it.
	points := sparklinePoints([]float64{0, 150}, 100, size100)

	if points[1].Y < 0 {
		t.Errorf("a value above the scale must clamp to the top, got Y=%v", points[1].Y)
	}
}

func TestSparklinePointsHandleAnAllZeroSeries(t *testing.T) {
	// An idle network auto-scales, so the scale would be 0: the series must draw
	// flat along the bottom rather than divide by zero.
	points := sparklinePoints([]float64{0, 0, 0}, 0, size100)

	for i, p := range points {
		if p.Y != 100 {
			t.Errorf("point %d: an idle series should lie on the bottom (Y=100), got %v", i, p.Y)
		}
	}
}

func TestSparklinePointsNeedTwoValues(t *testing.T) {
	// One point is not a line; the caller relies on nil to hide the segments.
	if got := sparklinePoints([]float64{42}, 100, size100); got != nil {
		t.Errorf("a single value cannot form a segment, got %v", got)
	}
	if got := sparklinePoints(nil, 100, size100); got != nil {
		t.Errorf("an empty series should produce no points, got %v", got)
	}
}
