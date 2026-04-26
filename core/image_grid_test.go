package core

import (
	"testing"

	"matplotlib-go/internal/geom"
)

func TestAxesDividerLaysOutGridCells(t *testing.T) {
	fig := NewFigure(400, 400)
	divider := fig.NewAxesDivider(unitRect(), 2, 2, WithAxesDividerHorizontalSpace(0.02), WithAxesDividerVerticalSpace(0.02))

	r00 := divider.AddAxes(0, 0)
	r01 := divider.AddAxes(0, 1)
	r10 := divider.AddAxes(1, 0)
	r11 := divider.AddAxes(1, 1)
	if r00 == nil || r01 == nil || r10 == nil || r11 == nil {
		t.Fatal("AddAxes returned nil for valid cell")
	}

	assertRectApproxTol(t, r00.RectFraction, geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.51}, Max: geom.Pt{X: 0.49, Y: 0.9}}, 1e-12)
	assertRectApproxTol(t, r01.RectFraction, geom.Rect{Min: geom.Pt{X: 0.51, Y: 0.51}, Max: geom.Pt{X: 0.9, Y: 0.9}}, 1e-12)
	assertRectApproxTol(t, r10.RectFraction, geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.10}, Max: geom.Pt{X: 0.49, Y: 0.49}}, 1e-12)
	assertRectApproxTol(t, r11.RectFraction, geom.Rect{Min: geom.Pt{X: 0.51, Y: 0.10}, Max: geom.Pt{X: 0.9, Y: 0.49}}, 1e-12)
}

func TestAxesDividerAddAxesProjectionValidatesInput(t *testing.T) {
	fig := NewFigure(200, 200)
	divider := fig.NewAxesDivider(unitRect(), 1, 2)

	if _, err := divider.AddAxesProjection(-1, 0, "rectilinear"); err == nil {
		t.Fatal("expected out-of-bounds cell error")
	}
	if _, err := divider.AddAxesProjection(2, 0, "rectilinear"); err == nil {
		t.Fatal("expected out-of-bounds cell error")
	}
	if _, err := divider.AddAxesProjection(0, 0, "mollweide"); err != nil {
		t.Fatalf("AddAxesProjection() error = %v", err)
	}
	if _, err := divider.AddAxesProjection(0, 0, "not-a-projection"); err == nil {
		t.Fatal("expected invalid projection error")
	}
}

func TestNewImageGridRespectsWidthAndHeightRatios(t *testing.T) {
	fig := NewFigure(1000, 1000)
	grid := fig.NewImageGrid(
		2,
		2,
		unitRect(),
		WithAxesDividerHorizontalSpace(0),
		WithAxesDividerVerticalSpace(0),
		WithAxesDividerWidthScales(1, 3),
		WithAxesDividerHeightScales(2, 1),
	)
	if grid == nil {
		t.Fatal("NewImageGrid returned nil")
	}

	assertRectApproxTol(t, grid.At(0, 0).RectFraction, geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.36666666666666664}, Max: geom.Pt{X: 0.30000000000000004, Y: 0.9}}, 1e-12)
	assertRectApproxTol(t, grid.At(0, 1).RectFraction, geom.Rect{Min: geom.Pt{X: 0.30000000000000004, Y: 0.36666666666666664}, Max: geom.Pt{X: 0.9, Y: 0.9}}, 1e-12)
}

func TestNewRGBAxesCreatesSharedScales(t *testing.T) {
	fig := NewFigure(1000, 1000)
	grid := fig.NewRGBAxes(
		unitRect(),
		WithAxesDividerHorizontalSpace(0),
	)
	if grid == nil {
		t.Fatal("NewRGBAxes returned nil")
	}

	if grid.Red == nil || grid.Green == nil || grid.Blue == nil {
		t.Fatal("expected three channel axes")
	}
	if grid.Green.xScaleRoot() != grid.Red.xScaleRoot() {
		t.Fatal("green axis should share x scale root with red axis")
	}
	if grid.Blue.yScaleRoot() != grid.Red.yScaleRoot() {
		t.Fatal("blue axis should share y scale root with red axis")
	}
	assertRectApproxTol(t, grid.Red.RectFraction, geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.3666666666666667, Y: 0.9}}, 1e-12)
	assertRectApproxTol(t, grid.Green.RectFraction, geom.Rect{Min: geom.Pt{X: 0.3666666666666667, Y: 0.1}, Max: geom.Pt{X: 0.6333333333333333, Y: 0.9}}, 1e-12)
}
