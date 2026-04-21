package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestRectangleDrawAndBounds(t *testing.T) {
	rect := &Rectangle{
		Patch: Patch{
			FaceColor: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
			EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
			EdgeWidth: 2,
		},
		XY:     geom.Pt{X: 1, Y: 2},
		Width:  3,
		Height: 4,
	}

	r := &recordingRenderer{}
	rect.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one path call, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.Fill.A <= 0 {
		t.Fatalf("expected rectangle fill paint, got %+v", r.pathCalls[0].paint)
	}
	if r.pathCalls[0].paint.Stroke.A <= 0 {
		t.Fatalf("expected rectangle stroke paint, got %+v", r.pathCalls[0].paint)
	}

	bounds := rect.Bounds(nil)
	want := geom.Rect{
		Min: geom.Pt{X: 1, Y: 2},
		Max: geom.Pt{X: 4, Y: 6},
	}
	if bounds != want {
		t.Fatalf("bounds = %+v, want %+v", bounds, want)
	}
}

func TestFancyBboxPatchRoundUsesCurvesAndHatch(t *testing.T) {
	box := &FancyBboxPatch{
		Patch: Patch{
			FaceColor:    render.Color{R: 0.3, G: 0.6, B: 0.9, A: 0.7},
			EdgeColor:    render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1},
			EdgeWidth:    1.5,
			Hatch:        "/",
			HatchColor:   render.Color{R: 0.2, G: 0.2, B: 0.2, A: 1},
			HatchSpacing: 6,
		},
		XY:           geom.Pt{X: 1, Y: 1},
		Width:        2.5,
		Height:       1.5,
		Pad:          0.2,
		BoxStyle:     BoxStyleRound,
		RoundingSize: 0.25,
	}

	r := &recordingRenderer{}
	box.Draw(r, createTestDrawContext())

	if len(r.pathCalls) < 2 {
		t.Fatalf("expected fill path plus hatch strokes, got %d calls", len(r.pathCalls))
	}
	if len(r.pathCalls[0].path.C) == 0 {
		t.Fatal("expected rounded path commands")
	}
	hasCurve := false
	for _, cmd := range r.pathCalls[0].path.C {
		if cmd == geom.CubicTo {
			hasCurve = true
			break
		}
	}
	if !hasCurve {
		t.Fatalf("expected rounded fancy box path to use cubic curves, got %v", r.pathCalls[0].path.C)
	}
}

func TestPatchAutoScaleIgnoresNonDataCoords(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.AddPatch(&Rectangle{
		Patch:  Patch{FaceColor: render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1}},
		XY:     geom.Pt{X: 2, Y: 3},
		Width:  2,
		Height: 3,
	})
	ax.AddPatch(&Circle{
		Patch:  Patch{FaceColor: render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1}},
		Center: geom.Pt{X: 0.5, Y: 0.5},
		Radius: 0.35,
		Coords: Coords(CoordAxes),
	})

	ax.AutoScale(0)
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()

	if xMin != 2 || xMax != 4 {
		t.Fatalf("x domain = [%v, %v], want [2, 4]", xMin, xMax)
	}
	if yMin != 3 || yMax != 6 {
		t.Fatalf("y domain = [%v, %v], want [3, 6]", yMin, yMax)
	}
}

func TestPathPatchLegendEntryIncludesHatch(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.AddPatch(&PathPatch{
		Patch: Patch{
			FaceColor:  render.Color{R: 0.7, G: 0.7, B: 0.2, A: 1},
			EdgeColor:  render.Color{R: 0.2, G: 0.2, B: 0.1, A: 1},
			EdgeWidth:  1,
			Hatch:      "x",
			HatchColor: render.Color{R: 0, G: 0, B: 0, A: 1},
			Label:      "region",
		},
		Path: patchRectPath(geom.Rect{
			Min: geom.Pt{X: 1, Y: 1},
			Max: geom.Pt{X: 3, Y: 2},
		}),
	})

	entries := ax.AddLegend().collectEntries()
	if len(entries) != 1 {
		t.Fatalf("expected one legend entry, got %d", len(entries))
	}
	if entries[0].Label != "region" || entries[0].kind != legendEntryPatch {
		t.Fatalf("unexpected legend entry: %+v", entries[0])
	}
	if entries[0].patchHatch != "x" {
		t.Fatalf("expected hatch metadata to be preserved, got %+v", entries[0])
	}
}

func TestFancyArrowBoundsAndClosedPath(t *testing.T) {
	arrow := &FancyArrow{
		Patch: Patch{
			FaceColor: render.Color{R: 0.9, G: 0.3, B: 0.2, A: 1},
			EdgeColor: render.Color{R: 0.2, G: 0.1, B: 0.1, A: 1},
			EdgeWidth: 1,
		},
		XY:         geom.Pt{X: 1, Y: 1},
		DX:         4,
		DY:         0,
		Width:      0.4,
		HeadWidth:  1.2,
		HeadLength: 1.1,
	}

	bounds := arrow.Bounds(nil)
	if bounds.Min.X != 1 || bounds.Max.X != 5 {
		t.Fatalf("x bounds = [%v, %v], want [1, 5]", bounds.Min.X, bounds.Max.X)
	}
	if bounds.Min.Y >= 1 || bounds.Max.Y <= 1 {
		t.Fatalf("expected arrow bounds to span around y=1, got %+v", bounds)
	}

	r := &recordingRenderer{}
	arrow.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one arrow path call, got %d", len(r.pathCalls))
	}
	path := r.pathCalls[0].path
	if len(path.C) == 0 || path.C[len(path.C)-1] != geom.ClosePath {
		t.Fatalf("expected closed arrow polygon, got %v", path.C)
	}
}
