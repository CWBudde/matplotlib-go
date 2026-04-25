package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestInsetAxesUsesParentAxesFractionBounds(t *testing.T) {
	fig := NewFigure(500, 300)
	parent := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.20},
		Max: geom.Pt{X: 0.90, Y: 0.80},
	})
	inset := parent.InsetAxes(geom.Rect{
		Min: geom.Pt{X: 0.60, Y: 0.55},
		Max: geom.Pt{X: 0.95, Y: 0.95},
	})
	if inset == nil {
		t.Fatal("InsetAxes returned nil")
	}
	if inset.AxesLocator() == nil {
		t.Fatal("inset axes should have a locator")
	}

	want := geom.Rect{
		Min: geom.Pt{X: 0.58, Y: 0.53},
		Max: geom.Pt{X: 0.86, Y: 0.77},
	}
	assertRectApproxTol(t, inset.RectFraction, want, 1e-12)
}

func TestInsetAxesFollowsParentAdjustedLayout(t *testing.T) {
	fig := NewFigure(400, 400)
	parent := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.10},
		Max: geom.Pt{X: 0.90, Y: 0.90},
	})
	if err := parent.SetBoxAspect(0.5); err != nil {
		t.Fatalf("SetBoxAspect: %v", err)
	}
	inset := parent.InsetAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	if inset == nil {
		t.Fatal("InsetAxes returned nil")
	}

	syncAxesLocators(fig, &render.NullRenderer{})
	want := pixelRectToFigureFraction(parent.adjustedLayout(fig), fig.DisplayRect())
	assertRectApproxTol(t, inset.RectFraction, want, 1e-12)
}

func TestInsetAxesProjectionAndSharing(t *testing.T) {
	fig := NewFigure(400, 300)
	parent := fig.AddAxes(unitRect())
	inset := parent.InsetAxes(
		geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.5, Y: 0.5}},
		WithInsetProjection("mollweide"),
		WithInsetSharedX(parent),
	)
	if inset == nil {
		t.Fatal("InsetAxes returned nil")
	}
	if got := inset.ProjectionName(); got != "mollweide" {
		t.Fatalf("projection name = %q, want mollweide", got)
	}
	if inset.xScaleRoot() != parent.xScaleRoot() {
		t.Fatal("inset should share x scale root with parent")
	}
}

func TestZoomedInsetConfiguresLimitsAndIndicator(t *testing.T) {
	fig := NewFigure(500, 300)
	parent := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.20},
		Max: geom.Pt{X: 0.90, Y: 0.80},
	})
	parent.SetXLim(0, 10)
	parent.SetYLim(-1, 1)

	inset, indicator := parent.ZoomedInset(
		geom.Rect{Min: geom.Pt{X: 0.55, Y: 0.50}, Max: geom.Pt{X: 0.95, Y: 0.92}},
		[2]float64{2, 4},
		[2]float64{-0.25, 0.75},
	)
	if inset == nil || indicator == nil {
		t.Fatal("ZoomedInset returned nil")
	}
	xMin, xMax := inset.XScale.Domain()
	yMin, yMax := inset.YScale.Domain()
	if !approx(xMin, 2, 1e-12) || !approx(xMax, 4, 1e-12) ||
		!approx(yMin, -0.25, 1e-12) || !approx(yMax, 0.75, 1e-12) {
		t.Fatalf("inset limits = x(%v,%v) y(%v,%v)", xMin, xMax, yMin, yMax)
	}

	ctx := newAxesDrawContext(parent, fig, fig.DisplayRect(), parent.adjustedLayout(fig))
	r := &recordingRenderer{}
	indicator.DrawOverlay(r, ctx)
	if len(r.pathCalls) != 3 {
		t.Fatalf("expected zoom rectangle plus two connectors, got %d path calls", len(r.pathCalls))
	}
	if got := len(r.pathCalls[0].path.V); got != 4 {
		t.Fatalf("zoom rectangle vertices = %d, want 4", got)
	}
}

func assertRectApproxTol(t *testing.T, got, want geom.Rect, tol float64) {
	t.Helper()
	if !approx(got.Min.X, want.Min.X, tol) ||
		!approx(got.Min.Y, want.Min.Y, tol) ||
		!approx(got.Max.X, want.Max.X, tol) ||
		!approx(got.Max.Y, want.Max.Y, tol) {
		t.Fatalf("rect = %+v, want %+v", got, want)
	}
}
