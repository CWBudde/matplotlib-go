package core

import (
	"math"
	"testing"

	"matplotlib-go/internal/geom"
)

func TestAddSkewXAxesConfiguresProjection(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}

	if got := ax.ProjectionName(); got != "skewx" {
		t.Fatalf("projection name = %q, want skewx", got)
	}
	if ax.XAxisTop == nil {
		t.Fatal("skewx axes should configure a top x axis")
	}

	xMin, xMax := ax.XScale.Domain()
	if !approx(xMin, -40, 1e-9) || !approx(xMax, 50, 1e-9) {
		t.Fatalf("temperature domain = (%v, %v), want (-40, 50)", xMin, xMax)
	}
	yMin, yMax := ax.YScale.Domain()
	if !approx(yMin, 1050, 1e-9) || !approx(yMax, 100, 1e-9) {
		t.Fatalf("pressure domain = (%v, %v), want (1050, 100)", yMin, yMax)
	}
	if !ax.YInverted() {
		t.Fatal("skewx pressure axis should decrease upward")
	}
}

func TestSkewXTransformRoundTrip(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}
	if err := ax.SetSkewXAngle(35); err != nil {
		t.Fatalf("SetSkewXAngle: %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	points := []geom.Pt{
		{X: 0, Y: 1000},
		{X: 12, Y: 700},
		{X: -18, Y: 300},
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

func TestSkewXAngleValidationAndEffect(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}

	if err := ax.SetSkewXAngle(math.Inf(1)); err == nil {
		t.Fatal("expected infinite skewx angle to be rejected")
	}
	if err := ax.SetSkewXAngle(90); err == nil {
		t.Fatal("expected vertical skewx angle to be rejected")
	}

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	bottom := ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 1000})
	top := ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 200})
	if !(top.X > bottom.X) {
		t.Fatalf("default skew should shift upper pressure levels right: bottom=%+v top=%+v", bottom, top)
	}

	if err := ax.SetSkewXAngle(-25); err != nil {
		t.Fatalf("SetSkewXAngle(-25): %v", err)
	}
	ctx = newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	bottom = ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 1000})
	top = ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 200})
	if !(top.X < bottom.X) {
		t.Fatalf("negative skew should shift upper pressure levels left: bottom=%+v top=%+v", bottom, top)
	}
}

func TestSkewXProjectionRejectsAngleOnOtherAxes(t *testing.T) {
	ax := NewFigure(480, 360).AddAxes(unitRect())
	if err := ax.SetSkewXAngle(30); err == nil {
		t.Fatal("expected SetSkewXAngle on rectilinear axes to fail")
	}
}

func TestSkewXGridUsesSampledProjectionPathsAndAxisLocators(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	grid := NewGrid(AxisLeft)
	grid.Minor = false

	r := &recordingRenderer{}
	grid.Draw(r, ctx)

	if got := len(r.pathCalls); got != 7 {
		t.Fatalf("expected 7 major pressure grid paths from axis locator fallback, got %d", got)
	}
	for i, call := range r.pathCalls {
		if got := len(call.path.V); got != geoGridSegments+1 {
			t.Fatalf("grid path %d vertex count = %d, want %d", i, got, geoGridSegments+1)
		}
	}
}
