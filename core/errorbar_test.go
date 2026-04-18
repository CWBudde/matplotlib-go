package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestErrorBar_Draw_Basic(t *testing.T) {
	errBar := &ErrorBar{
		XY: []geom.Pt{
			{X: 1, Y: 2},
			{X: 2, Y: 3},
			{X: 3, Y: 2.5},
		},
		XErr:      []float64{0.2, 0.3, 0.25},
		YErr:      []float64{0.4, 0.2, 0.3},
		LineWidth: 1.2,
		CapSize:   6,
		Color:     render.Color{R: 0, G: 0, B: 0, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	errBar.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestErrorBar_Draw_Empty(t *testing.T) {
	errBar := &ErrorBar{}
	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	errBar.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestErrorBar_Draw_BroadcastError(t *testing.T) {
	errBar := &ErrorBar{
		XY: []geom.Pt{
			{X: 1, Y: 2},
			{X: 2, Y: 3},
			{X: 3, Y: 4},
		},
		XErr:      []float64{0.3},
		YErr:      []float64{0.1},
		LineWidth: 1,
		CapSize:   4,
		Color:     render.Color{R: 0, G: 0, B: 1, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	errBar.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestErrorBar_ZOrder(t *testing.T) {
	errBar := &ErrorBar{z: 1.25}
	if got := errBar.Z(); got != 1.25 {
		t.Errorf("expected Z() = 1.25, got %v", got)
	}
}

func TestErrorBar_Bounds(t *testing.T) {
	errBar := &ErrorBar{
		XY: []geom.Pt{
			{X: 2, Y: 3},
			{X: 5, Y: 5},
		},
		XErr: []float64{0.5},
		YErr: []float64{0.4, 0.6},
	}
	bounds := errBar.Bounds(nil)
	if bounds.Min.X != 1.5 || bounds.Max.X != 5.5 || bounds.Min.Y != 2.4 || bounds.Max.Y != 5.6 {
		t.Errorf("unexpected bounds: %v", bounds)
	}
}

func TestAxes_ErrorBar(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	errBar := ax.ErrorBar(
		[]float64{1, 2, 3},
		[]float64{1.1, 2.2, 3.3},
		[]float64{0.1},
		nil,
	)
	if errBar == nil {
		t.Fatal("ErrorBar should return non-nil for non-empty data")
	}
}

func TestAxes_ErrorBar_Options(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	col := render.Color{R: 1, G: 0, B: 0, A: 1}
	lineWidth := 2.0
	capSize := 6.0
	alpha := 0.8
	errBar := ax.ErrorBar(
		[]float64{1, 2},
		[]float64{3, 4},
		nil,
		[]float64{0.2},
		ErrorBarOptions{
			Color:     &col,
			LineWidth: &lineWidth,
			CapSize:   &capSize,
			Alpha:     &alpha,
			Label:     "test",
		},
	)

	if errBar == nil {
		t.Fatal("expected non-nil error bar")
	}
	if errBar.Label != "test" {
		t.Errorf("expected label 'test', got %q", errBar.Label)
	}
	if errBar.LineWidth != lineWidth {
		t.Errorf("expected line width %v, got %v", lineWidth, errBar.LineWidth)
	}
	if errBar.CapSize != capSize {
		t.Errorf("expected cap size %v, got %v", capSize, errBar.CapSize)
	}
	if errBar.Alpha != alpha {
		t.Errorf("expected alpha %v, got %v", alpha, errBar.Alpha)
	}
}
