package core

import (
	"math"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

type polarTextRenderer struct {
	recordingRenderer
	texts   []string
	origins []geom.Pt
}

func (r *polarTextRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *polarTextRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, origin)
}

func TestAddPolarAxesConfiguresProjection(t *testing.T) {
	fig := NewFigure(400, 400)
	ax := fig.AddPolarAxes(unitRect())
	if ax == nil {
		t.Fatal("expected polar axes")
	}

	if got := ax.ProjectionName(); got != "polar" {
		t.Fatalf("projection name = %q, want polar", got)
	}
	if ax.ShowFrame {
		t.Fatal("polar axes should disable rectangular frame fallback")
	}

	xMin, xMax := ax.XScale.Domain()
	if !approx(xMin, 0, 1e-9) || !approx(xMax, 2*math.Pi, 1e-9) {
		t.Fatalf("theta domain = (%v, %v), want (0, 2pi)", xMin, xMax)
	}
	yMin, yMax := ax.YScale.Domain()
	if !approx(yMin, 0, 1e-9) || !approx(yMax, 1, 1e-9) {
		t.Fatalf("radius domain = (%v, %v), want (0, 1)", yMin, yMax)
	}

	center := geom.Pt{X: 200, Y: 200}
	if !ax.ContainsDisplayPoint(center) {
		t.Fatal("polar axes should contain its center")
	}
	if ax.ContainsDisplayPoint(geom.Pt{X: 40, Y: 40}) {
		t.Fatal("polar axes should reject points in the layout corner outside the circular frame")
	}
}

func TestPolarAxesPixelToDataRoundTrip(t *testing.T) {
	fig := NewFigure(400, 400)
	ax := fig.AddPolarAxes(unitRect())
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	got := ctx.DataToPixel.Apply(geom.Pt{X: math.Pi / 2, Y: 1})
	center, radius := polarCenterAndRadius(ax.adjustedLayout(fig))
	want := geom.Pt{X: center.X, Y: center.Y - radius}
	if !approx(got.X, want.X, 1e-6) || !approx(got.Y, want.Y, 1e-6) {
		t.Fatalf("polar top point = %+v, want %+v", got, want)
	}

	data, ok := ax.PixelToData(got)
	if !ok {
		t.Fatal("PixelToData should invert polar coordinates")
	}
	if !approx(data.X, math.Pi/2, 1e-6) || !approx(data.Y, 1, 1e-6) {
		t.Fatalf("PixelToData = %+v, want theta=pi/2 radius=1", data)
	}
}

func TestPolarGridAndTicksUseCurvedGeometry(t *testing.T) {
	fig := NewFigure(400, 400)
	ax := fig.AddPolarAxes(unitRect())
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	radialGrid := NewGrid(AxisLeft)
	radialGrid.Locator = FixedLocator{TicksList: []float64{0.5}}
	radialGrid.Minor = false

	thetaGrid := NewGrid(AxisBottom)
	thetaGrid.Locator = FixedLocator{TicksList: []float64{math.Pi / 2}}
	thetaGrid.Minor = false

	r := &recordingRenderer{}
	radialGrid.Draw(r, ctx)
	thetaGrid.Draw(r, ctx)
	ax.XAxis.Draw(r, ctx)

	center, radius := polarCenterAndRadius(ax.adjustedLayout(fig))

	var foundCircle bool
	var foundSpoke bool
	for _, call := range r.pathCalls {
		if len(call.path.C) > 0 && call.path.C[len(call.path.C)-1] == geom.ClosePath && len(call.path.V) >= polarCircleSegments {
			foundCircle = true
		}
		if len(call.path.C) == 2 && len(call.path.V) == 2 {
			p1, p2 := call.path.V[0], call.path.V[1]
			if approx(p1.X, center.X, 1e-6) && approx(p1.Y, center.Y, 1e-6) &&
				approx(p2.X, center.X, 1e-6) && approx(p2.Y, center.Y-radius, 1e-6) {
				foundSpoke = true
			}
		}
	}

	if !foundCircle {
		t.Fatal("expected a closed circular path for polar spine/grid")
	}
	if !foundSpoke {
		t.Fatal("expected a radial spoke path for angular grid lines")
	}
}

func TestPolarTickLabelsUseAngularFormatting(t *testing.T) {
	fig := NewFigure(400, 400)
	ax := fig.AddPolarAxes(unitRect())
	ax.XAxis.Locator = FixedLocator{TicksList: []float64{0, math.Pi / 2}}
	ax.XAxis.MinorLocator = nil
	ax.YAxis.Locator = FixedLocator{TicksList: []float64{0.5}}
	ax.YAxis.MinorLocator = nil

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &polarTextRenderer{}

	ax.XAxis.DrawTickLabels(r, ctx)
	ax.YAxis.DrawTickLabels(r, ctx)

	if len(r.texts) != 3 {
		t.Fatalf("expected 3 polar tick labels, got %d (%v)", len(r.texts), r.texts)
	}
	if r.texts[0] != "0 deg" || r.texts[1] != "90 deg" {
		t.Fatalf("theta labels = %v, want [0 deg 90 deg ...]", r.texts)
	}
	if r.texts[2] == "" {
		t.Fatal("expected a radial tick label")
	}
}
