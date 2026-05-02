package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestAddRadarAxesConfiguresProjection(t *testing.T) {
	fig := NewFigure(400, 400)
	ax, err := fig.AddRadarAxes(unitRect(), []string{"Speed", "Power", "Range", "Cost"})
	if err != nil {
		t.Fatalf("AddRadarAxes: %v", err)
	}

	if got := ax.ProjectionName(); got != "radar" {
		t.Fatalf("projection name = %q, want radar", got)
	}
	if ax.ShowFrame {
		t.Fatal("radar axes should disable rectangular frame fallback")
	}

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	center, radius := polarCenterAndRadius(ax.adjustedLayout(fig))
	top := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: 1})
	wantTop := geom.Pt{X: center.X, Y: center.Y - radius}
	if !approx(top.X, wantTop.X, 1e-6) || !approx(top.Y, wantTop.Y, 1e-6) {
		t.Fatalf("theta=0 point = %+v, want %+v", top, wantTop)
	}
	right := ctx.DataToPixel.Apply(geom.Pt{X: math.Pi / 2, Y: 1})
	wantRight := geom.Pt{X: center.X + radius, Y: center.Y}
	if !approx(right.X, wantRight.X, 1e-6) || !approx(right.Y, wantRight.Y, 1e-6) {
		t.Fatalf("theta=pi/2 point = %+v, want %+v", right, wantRight)
	}

	r := &polarTextRenderer{}
	ax.XAxis.DrawTickLabels(r, ctx)
	if len(r.texts) != 4 {
		t.Fatalf("expected 4 radar spoke labels, got %d (%v)", len(r.texts), r.texts)
	}
	if r.texts[0] != "Speed" || r.texts[1] != "Power" || r.texts[2] != "Range" || r.texts[3] != "Cost" {
		t.Fatalf("radar spoke labels = %v", r.texts)
	}
}

func TestRadarFrameAndGridUsePolygonGeometry(t *testing.T) {
	fig := NewFigure(400, 400)
	ax, err := fig.AddRadarAxes(unitRect(), []string{"A", "B", "C", "D", "E"})
	if err != nil {
		t.Fatalf("AddRadarAxes: %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	grid := NewGrid(AxisLeft)
	grid.Locator = FixedLocator{TicksList: []float64{0.5}}
	grid.Minor = false

	r := &recordingRenderer{}
	grid.Draw(r, ctx)
	ax.XAxis.Draw(r, ctx)

	center, radius := polarCenterAndRadius(ax.adjustedLayout(fig))
	wantTop := geom.Pt{X: center.X, Y: center.Y - radius}
	var foundOuterPentagon bool
	for _, call := range r.pathCalls {
		if len(call.path.C) == 6 && call.path.C[len(call.path.C)-1] == geom.ClosePath && len(call.path.V) == 5 {
			if approx(call.path.V[0].X, wantTop.X, 1e-6) && approx(call.path.V[0].Y, wantTop.Y, 1e-6) {
				foundOuterPentagon = true
			}
		}
	}
	if !foundOuterPentagon {
		t.Fatal("expected radar outer spine to be a five-sided polygon with its first vertex at theta=0")
	}
}

func TestRadarVariableCountValidation(t *testing.T) {
	fig := NewFigure(400, 400)
	if _, err := fig.AddRadarAxes(unitRect(), []string{"A", "B"}); err == nil {
		t.Fatal("expected AddRadarAxes to reject fewer than 3 labels")
	}

	ax, err := fig.AddRadarAxes(unitRect(), nil)
	if err != nil {
		t.Fatalf("AddRadarAxes with defaults: %v", err)
	}
	if err := ax.SetRadarVariableCount(6); err != nil {
		t.Fatalf("SetRadarVariableCount: %v", err)
	}
	ticks := ax.XAxis.Locator.Ticks(0, 2*math.Pi, 10)
	if len(ticks) != 6 {
		t.Fatalf("radar tick count = %d, want 6", len(ticks))
	}
}
