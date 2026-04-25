package core

import (
	"testing"

	"matplotlib-go/internal/geom"
)

func TestRelativeAnchoredBoxLocatorCentersBox(t *testing.T) {
	locator := RelativeAnchoredBoxLocator{
		X:      0.5,
		Y:      0.5,
		HAlign: BoxAlignCenter,
		VAlign: BoxAlignMiddle,
	}

	rect := locator.Rect(
		geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 110, Y: 100}},
		20,
		10,
	)

	if rect != (geom.Rect{
		Min: geom.Pt{X: 50, Y: 55},
		Max: geom.Pt{X: 70, Y: 65},
	}) {
		t.Fatalf("relative locator rect = %+v", rect)
	}
}

func TestAnchoredOffsetLocatorAppliesCornerAndPixelOffset(t *testing.T) {
	locator := NewAnchoredOffsetLocator(LegendUpperLeft, 8, 5, 3)
	rect := locator.Rect(
		geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 110, Y: 100}},
		20,
		10,
	)

	if rect != (geom.Rect{
		Min: geom.Pt{X: 23, Y: 31},
		Max: geom.Pt{X: 43, Y: 41},
	}) {
		t.Fatalf("offset locator rect = %+v", rect)
	}
}

func TestAnchoredTextBoxBoxRectUsesLocator(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	box := ax.AddAnchoredText("Centered", AnchoredTextOptions{
		Location: LegendUpperLeft,
		Locator: RelativeAnchoredBoxLocator{
			X:      0.5,
			Y:      0.5,
			HAlign: BoxAlignCenter,
			VAlign: BoxAlignMiddle,
		},
	})

	var r figureLayoutRecordingRenderer
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	rect, ok := box.boxRect(&r, ctx)
	if !ok {
		t.Fatal("boxRect() returned !ok")
	}

	if !floatApprox(rect.Min.X+rect.W()/2, ctx.Clip.Min.X+ctx.Clip.W()/2, 1e-9) {
		t.Fatalf("box center x = %v, want %v", rect.Min.X+rect.W()/2, ctx.Clip.Min.X+ctx.Clip.W()/2)
	}
	if !floatApprox(rect.Min.Y+rect.H()/2, ctx.Clip.Min.Y+ctx.Clip.H()/2, 1e-9) {
		t.Fatalf("box center y = %v, want %v", rect.Min.Y+rect.H()/2, ctx.Clip.Min.Y+ctx.Clip.H()/2)
	}
}

func TestLegendBoxRectUsesLocator(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})

	legend := ax.AddLegend()
	legend.SetLocator(RelativeAnchoredBoxLocator{
		X:      0.5,
		Y:      0.5,
		HAlign: BoxAlignCenter,
		VAlign: BoxAlignMiddle,
	})

	var r legendRecordingRenderer
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	rect, ok := legend.boxRect(&r, ctx)
	if !ok {
		t.Fatal("boxRect() returned !ok")
	}

	if !floatApprox(rect.Min.X+rect.W()/2, ctx.Clip.Min.X+ctx.Clip.W()/2, 1e-9) {
		t.Fatalf("legend center x = %v, want %v", rect.Min.X+rect.W()/2, ctx.Clip.Min.X+ctx.Clip.W()/2)
	}
	if !floatApprox(rect.Min.Y+rect.H()/2, ctx.Clip.Min.Y+ctx.Clip.H()/2, 1e-9) {
		t.Fatalf("legend center y = %v, want %v", rect.Min.Y+rect.H()/2, ctx.Clip.Min.Y+ctx.Clip.H()/2)
	}
}
