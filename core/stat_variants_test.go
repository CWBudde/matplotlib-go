package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAxesStackPlot_CumulativeLayers(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	fills := ax.StackPlot(
		[]float64{0, 1, 2},
		[][]float64{
			{1, 2, 3},
			{4, 5, 6},
		},
		StackPlotOptions{Labels: []string{"a", "b"}},
	)

	if len(fills) != 2 {
		t.Fatalf("got %d fills, want 2", len(fills))
	}
	if fills[0].Label != "a" || fills[1].Label != "b" {
		t.Fatalf("labels = %q, %q", fills[0].Label, fills[1].Label)
	}
	assertFloatSlices(t, "first lower", fills[0].Y2, []float64{0, 0, 0})
	assertFloatSlices(t, "first upper", fills[0].Y1, []float64{1, 2, 3})
	assertFloatSlices(t, "second lower", fills[1].Y2, []float64{1, 2, 3})
	assertFloatSlices(t, "second upper", fills[1].Y1, []float64{5, 7, 9})
}

func TestAxesECDF_ComputesSortedStepData(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	line := ax.ECDF([]float64{3, 1, 2, 2})
	if line == nil {
		t.Fatal("expected ECDF line")
	}
	if line.DrawStyle != LineDrawStyleStepsPost {
		t.Fatalf("draw style = %v, want steps-post", line.DrawStyle)
	}
	want := []geom.Pt{
		{X: 1, Y: 0},
		{X: 1, Y: 0.25},
		{X: 2, Y: 0.5},
		{X: 2, Y: 0.75},
		{X: 3, Y: 1},
	}
	if len(line.XY) != len(want) {
		t.Fatalf("got %d points, want %d", len(line.XY), len(want))
	}
	for i := range want {
		if line.XY[i] != want[i] {
			t.Fatalf("point %d = %+v, want %+v", i, line.XY[i], want[i])
		}
	}
}

func TestAxesECDF_CompressKeepsDuplicateProbabilities(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	line := ax.ECDF([]float64{3, 1, 2, 2}, ECDFOptions{Compress: true})
	if line == nil {
		t.Fatal("expected ECDF line")
	}
	want := []geom.Pt{
		{X: 1, Y: 0},
		{X: 1, Y: 0.25},
		{X: 2, Y: 0.75},
		{X: 3, Y: 1},
	}
	if len(line.XY) != len(want) {
		t.Fatalf("got %d points, want %d", len(line.XY), len(want))
	}
	for i := range want {
		if line.XY[i] != want[i] {
			t.Fatalf("point %d = %+v, want %+v", i, line.XY[i], want[i])
		}
	}
}

func TestHist2D_CumulativeProbability(t *testing.T) {
	hist := &Hist2D{
		Data:       []float64{0.2, 0.4, 1.2, 2.2},
		BinEdges:   []float64{0, 1, 2, 3},
		Norm:       HistNormProbability,
		Cumulative: true,
	}

	_, counts := hist.BinCounts()
	assertFloatSlices(t, "counts", counts, []float64{0.5, 0.75, 1})
}

func TestHist2D_StepFilledDrawsClosedPath(t *testing.T) {
	hist := &Hist2D{
		Data:      []float64{0.2, 0.4, 1.2},
		BinEdges:  []float64{0, 1, 2},
		HistType:  HistTypeStepFilled,
		Color:     render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1},
		EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		EdgeWidth: 1,
	}

	r := &recordingRenderer{}
	hist.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one path call, got %d", len(r.pathCalls))
	}
	call := r.pathCalls[0]
	if call.paint.Fill.A == 0 || call.paint.Stroke.A == 0 {
		t.Fatalf("unexpected paint = %+v", call.paint)
	}
	if len(call.path.C) == 0 || call.path.C[len(call.path.C)-1] != geom.ClosePath {
		t.Fatalf("expected closed path, got %+v", call.path.C)
	}
}

func TestAxesHistMulti_UsesSharedEdgesAndStackedBaselines(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	hists := ax.HistMulti(
		[][]float64{
			{0.2, 0.4, 1.2},
			{0.7, 1.4, 1.6},
		},
		MultiHistOptions{
			BinEdges: []float64{0, 1, 2},
			Stacked:  true,
		},
	)

	if len(hists) != 2 {
		t.Fatalf("got %d hists, want 2", len(hists))
	}
	edges0, counts0 := hists[0].BinCounts()
	edges1, _ := hists[1].BinCounts()
	assertFloatSlices(t, "edges0", edges0, []float64{0, 1, 2})
	assertFloatSlices(t, "edges1", edges1, []float64{0, 1, 2})
	assertFloatSlices(t, "baselines", hists[1].Baselines, counts0)
}

func assertFloatSlices(t *testing.T, name string, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d (%v)", name, len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %v, want %v (all %v)", name, i, got[i], want[i], got)
		}
	}
}
