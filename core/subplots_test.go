package core

import (
	"math"
	"testing"
)

func TestFigureSubplotsGridLayout(t *testing.T) {
	fig := NewFigure(1000, 800)
	grid := fig.Subplots(
		2,
		2,
		WithSubplotPadding(0.1, 0.9, 0.1, 0.9),
		WithSubplotSpacing(0.1, 0.1),
	)
	if len(grid) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(grid))
	}
	if len(grid[0]) != 2 || len(grid[1]) != 2 {
		t.Fatalf("expected 2 columns, got %d and %d", len(grid[0]), len(grid[1]))
	}

	got := grid[0][0].RectFraction
	if !floatApprox(got.Min.X, 0.1, 1e-12) ||
		!floatApprox(got.Max.X, 0.45, 1e-12) ||
		!floatApprox(got.Min.Y, 0.55, 1e-12) ||
		!floatApprox(got.Max.Y, 0.9, 1e-12) {
		t.Fatalf("unexpected top-left rect: %+v", got)
	}

	got = grid[0][1].RectFraction
	if !floatApprox(got.Min.X, 0.55, 1e-12) ||
		!floatApprox(got.Max.X, 0.9, 1e-12) ||
		!floatApprox(got.Min.Y, 0.55, 1e-12) ||
		!floatApprox(got.Max.Y, 0.9, 1e-12) {
		t.Fatalf("unexpected top-right rect: %+v", got)
	}

	got = grid[1][0].RectFraction
	if !floatApprox(got.Min.X, 0.1, 1e-12) ||
		!floatApprox(got.Max.X, 0.45, 1e-12) ||
		!floatApprox(got.Min.Y, 0.1, 1e-12) ||
		!floatApprox(got.Max.Y, 0.45, 1e-12) {
		t.Fatalf("unexpected bottom-left rect: %+v", got)
	}

	got = grid[1][1].RectFraction
	if !floatApprox(got.Min.X, 0.55, 1e-12) ||
		!floatApprox(got.Max.X, 0.9, 1e-12) ||
		!floatApprox(got.Min.Y, 0.1, 1e-12) ||
		!floatApprox(got.Max.Y, 0.45, 1e-12) {
		t.Fatalf("unexpected bottom-right rect: %+v", got)
	}
}

func TestFigureSubplotsAxisSharing(t *testing.T) {
	fig := NewFigure(1000, 800)
	grid := fig.Subplots(2, 2, WithSubplotShareX(), WithSubplotShareY())
	if len(grid) != 2 || len(grid[0]) != 2 {
		t.Fatalf("expected 2x2 grid, got %dx%d", len(grid), len(grid[0]))
	}

	grid[1][0].SetXLim(-2, 2)
	grid[1][1].SetYLim(-5, 5)

	xMin, xMax := grid[0][0].effectiveXScale().Domain()
	axXMin, axXMax := grid[1][0].effectiveXScale().Domain()
	if xMin != axXMin || xMax != axXMax {
		t.Fatalf("x scale not shared: got base=%v..%v and shared=%v..%v", xMin, xMax, axXMin, axXMax)
	}

	yMin, yMax := grid[0][0].effectiveYScale().Domain()
	axYMin, axYMax := grid[0][1].effectiveYScale().Domain()
	if yMin != axYMin || yMax != axYMax {
		t.Fatalf("y scale not shared: got base=%v..%v and shared=%v..%v", yMin, yMax, axYMin, axYMax)
	}

	if grid[1][1].XAxis != grid[0][0].XAxis {
		t.Fatalf("x-axis object not shared")
	}
	if grid[1][1].YAxis != grid[0][0].YAxis {
		t.Fatalf("y-axis object not shared")
	}
}

func floatApprox(a, b, eps float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	return math.Abs(a-b) <= eps
}
