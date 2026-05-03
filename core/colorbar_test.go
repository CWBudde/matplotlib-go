package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestFigureAddColorbarConfiguresAxes(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{
		{0, 1},
		{2, 3},
	})

	cbAx := fig.AddColorbar(ax, img, ColorbarOptions{Label: "Intensity"})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}

	base := geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	}
	wantWidth := resolvedColorbarWidth(fig, base, 0, defaultColorbarAspect)
	wantPadding := resolvedColorbarPadding(base, 0)
	if wantPadding != 0.05 {
		t.Fatalf("default colorbar padding = %v, want matplotlib default 0.05", wantPadding)
	}
	if got, want := ax.RectFraction.Max.X, base.Max.X-wantWidth-wantPadding; !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected parent to reserve colorbar space: got right=%v want %v", got, want)
	}
	if got, want := cbAx.RectFraction.W(), wantWidth; !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected colorbar width to follow default aspect: got %v want %v", got, want)
	}
	if got, want := cbAx.RectFraction.Max.X, base.Max.X; !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected colorbar to occupy reserved right edge: got right=%v want %v", got, want)
	}
	if cbAx.RectFraction.Min.X <= ax.RectFraction.Max.X {
		t.Fatalf("expected colorbar to be placed to the right, got %+v", cbAx.RectFraction)
	}
	if cbAx.RectFraction.Min.Y != ax.RectFraction.Min.Y || cbAx.RectFraction.Max.Y != ax.RectFraction.Max.Y {
		t.Fatalf("expected colorbar to share vertical extent, got %+v", cbAx.RectFraction)
	}
	if cbAx.XAxis.ShowSpine || cbAx.XAxis.ShowTicks || cbAx.XAxis.ShowLabels {
		t.Fatalf("expected hidden colorbar x-axis, got %+v", cbAx.XAxis)
	}
	if cbAx.YAxis.ShowSpine || cbAx.YAxis.ShowTicks || cbAx.YAxis.ShowLabels {
		t.Fatalf("expected hidden left-side y-axis, got %+v", cbAx.YAxis)
	}
	if cbAx.YAxisRight == nil {
		t.Fatal("expected explicit right-side y-axis")
	}
	if cbAx.YAxisRight.Side != AxisRight {
		t.Fatalf("expected right-side y-axis, got %v", cbAx.YAxisRight.Side)
	}
	if !cbAx.YAxisRight.ShowLabels || !cbAx.YAxisRight.ShowTicks {
		t.Fatalf("expected visible right-side colorbar ticks and labels, got %+v", cbAx.YAxisRight)
	}
	if cbAx.YAxis.ShowLabels || cbAx.YAxis.ShowTicks {
		t.Fatalf("expected hidden left-side colorbar ticks and labels, got %+v", cbAx.YAxis)
	}
	if cbAx.effectiveYLabelSide() != AxisRight {
		t.Fatalf("expected colorbar label on right side")
	}
	if cbAx.YLabel != "Intensity" {
		t.Fatalf("unexpected colorbar label %q", cbAx.YLabel)
	}

	yMin, yMax := cbAx.YScale.Domain()
	if yMin != 0 || yMax != 3 {
		t.Fatalf("unexpected colorbar limits %v..%v", yMin, yMax)
	}
}

func TestColorbarDrawRendersGradientAndTickLabels(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{
		{0, 1},
		{2, 3},
	})

	cbAx := fig.AddColorbar(ax, img, ColorbarOptions{Label: "Value"})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}

	var r colorbarRecordingRenderer
	DrawFigure(fig, &r)

	if r.imageCount < 1 {
		t.Fatalf("expected heatmap image, got %d", r.imageCount)
	}
	if r.pathCount == 0 {
		t.Fatal("expected colorbar border/axes paths to be rendered")
	}
	if len(r.texts) == 0 {
		t.Fatal("expected tick labels or colorbar label to be rendered")
	}
}

func TestColorbarDrawSnapsRangeLegendToPixels(t *testing.T) {
	var r colorbarRecordingRenderer
	clip := geom.Rect{
		Min: geom.Pt{X: 10.4, Y: 20.6},
		Max: geom.Pt{X: 42.6, Y: 80.2},
	}

	(&Colorbar{Colormap: "inferno", Alpha: 1, BorderColor: render.Color{A: 1}, BorderWidth: 1}).Draw(&r, &DrawContext{Clip: clip})

	if len(r.imageRects) != 0 {
		t.Fatalf("image rect count = %d, want 0", len(r.imageRects))
	}
	if len(r.paths) != 257 {
		t.Fatalf("path count = %d, want 256 color cells plus outline", len(r.paths))
	}
	wantFirstCell := []geom.Pt{
		{X: 10, Y: 80},
		{X: 43, Y: 80},
		{X: 43, Y: 81},
		{X: 10, Y: 81},
	}
	for i, want := range wantFirstCell {
		if !floatApprox(r.paths[0].V[i].X, want.X, 1e-12) || !floatApprox(r.paths[0].V[i].Y, want.Y, 1e-12) {
			t.Fatalf("first cell vertex %d = %+v, want %+v", i, r.paths[0].V[i], want)
		}
	}
	wantOutline := []geom.Pt{
		{X: 10.5, Y: 21.5},
		{X: 43.5, Y: 21.5},
		{X: 43.5, Y: 80.5},
		{X: 10.5, Y: 80.5},
	}
	outline := r.paths[len(r.paths)-1]
	if len(outline.V) < len(wantOutline) {
		t.Fatalf("outline vertices = %v, want at least %d", outline.V, len(wantOutline))
	}
	for i, want := range wantOutline {
		if !floatApprox(outline.V[i].X, want.X, 1e-12) || !floatApprox(outline.V[i].Y, want.Y, 1e-12) {
			t.Fatalf("outline vertex %d = %+v, want %+v", i, outline.V[i], want)
		}
	}
}

type colorbarRecordingRenderer struct {
	render.NullRenderer
	imageCount int
	pathCount  int
	texts      []string
	imageRects []geom.Rect
	paths      []geom.Path
}

func (r *colorbarRecordingRenderer) Image(_ render.Image, dst geom.Rect) {
	r.imageCount++
	r.imageRects = append(r.imageRects, dst)
}

func (r *colorbarRecordingRenderer) Path(path geom.Path, _ *render.Paint) {
	r.pathCount++
	r.paths = append(r.paths, path)
}

func (r *colorbarRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *colorbarRecordingRenderer) DrawText(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
}
