package core

import (
	"math"
	"testing"

	"matplotlib-go/internal/geom"
)

func TestAddMollweideAxesConfiguresProjection(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "mollweide")
	if err != nil {
		t.Fatalf("AddAxesProjection(mollweide): %v", err)
	}

	if got := ax.ProjectionName(); got != "mollweide" {
		t.Fatalf("projection name = %q, want mollweide", got)
	}
	if ax.ShowFrame {
		t.Fatal("mollweide axes should disable rectangular frame fallback")
	}

	xMin, xMax := ax.XScale.Domain()
	if !approx(xMin, -math.Pi, 1e-9) || !approx(xMax, math.Pi, 1e-9) {
		t.Fatalf("longitude domain = (%v, %v), want (-pi, pi)", xMin, xMax)
	}
	yMin, yMax := ax.YScale.Domain()
	if !approx(yMin, -math.Pi/2, 1e-9) || !approx(yMax, math.Pi/2, 1e-9) {
		t.Fatalf("latitude domain = (%v, %v), want (-pi/2, pi/2)", yMin, yMax)
	}

	if !ax.ContainsDisplayPoint(geom.Pt{X: 200, Y: 120}) {
		t.Fatal("mollweide axes should contain its center")
	}
	if ax.ContainsDisplayPoint(geom.Pt{X: 10, Y: 10}) {
		t.Fatal("mollweide axes should reject layout corners outside the oval frame")
	}
}

func TestMollweideTransformRoundTrip(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "mollweide")
	if err != nil {
		t.Fatalf("AddAxesProjection(mollweide): %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	points := []geom.Pt{
		{X: 0, Y: 0},
		{X: math.Pi / 3, Y: math.Pi / 6},
		{X: -2 * math.Pi / 3, Y: -math.Pi / 4},
	}
	for _, pt := range points {
		pixel := ctx.DataToPixel.Apply(pt)
		got, ok := ax.PixelToData(pixel)
		if !ok {
			t.Fatalf("PixelToData(%+v) failed", pixel)
		}
		if !approx(got.X, pt.X, 1e-6) || !approx(got.Y, pt.Y, 1e-6) {
			t.Fatalf("round trip = %+v, want %+v", got, pt)
		}
	}
}

func TestMollweideFrameAndGridUseSampledCurves(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "mollweide")
	if err != nil {
		t.Fatalf("AddAxesProjection(mollweide): %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	longitudeGrid := NewGrid(AxisBottom)
	longitudeGrid.Locator = FixedLocator{TicksList: []float64{0}}
	longitudeGrid.Minor = false

	latitudeGrid := NewGrid(AxisLeft)
	latitudeGrid.Locator = FixedLocator{TicksList: []float64{math.Pi / 6}}
	latitudeGrid.Minor = false

	r := &recordingRenderer{}
	longitudeGrid.Draw(r, ctx)
	latitudeGrid.Draw(r, ctx)
	ax.XAxis.Draw(r, ctx)

	if len(r.pathCalls) < 3 {
		t.Fatalf("expected longitude, latitude, and frame paths, got %d", len(r.pathCalls))
	}

	var foundLongitude, foundLatitude, foundFrame bool
	for _, call := range r.pathCalls {
		if len(call.path.V) == geoGridSegments+1 {
			first := call.path.V[0]
			last := call.path.V[len(call.path.V)-1]
			if approx(first.X, last.X, 1e-6) && !approx(first.Y, last.Y, 1e-6) {
				foundLongitude = true
			}
			if !approx(first.X, last.X, 1e-6) && approx(first.Y, last.Y, 1e-6) {
				foundLatitude = true
			}
		}
		if len(call.path.C) > 0 && call.path.C[len(call.path.C)-1] == geom.ClosePath && len(call.path.V) >= geoFrameSegments {
			foundFrame = true
		}
	}

	if !foundLongitude {
		t.Fatal("expected sampled meridian grid path")
	}
	if !foundLatitude {
		t.Fatal("expected sampled parallel grid path")
	}
	if !foundFrame {
		t.Fatal("expected closed oval frame path")
	}
}
