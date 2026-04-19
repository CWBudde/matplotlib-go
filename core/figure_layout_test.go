package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestDrawFigure_RendersFigureLevelLabels(t *testing.T) {
	fig := NewFigure(800, 600)
	fig.SetSuptitle("Overview")
	fig.SetSupXLabel("time")
	fig.SetSupYLabel("value")

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.92, Y: 0.82},
	})
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowLabels = false

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "Overview") {
		t.Fatalf("missing suptitle draw: %v", r.texts)
	}
	if !containsString(r.texts, "time") {
		t.Fatalf("missing supxlabel draw: %v", r.texts)
	}
	if !containsString(r.rotatedText, "value") {
		t.Fatalf("missing supylabel draw: %v", r.rotatedText)
	}
}

func TestFigureLegendCollectsAcrossAxes(t *testing.T) {
	fig := NewFigure(800, 600)
	left := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.15},
		Max: geom.Pt{X: 0.45, Y: 0.85},
	})
	right := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.55, Y: 0.15},
		Max: geom.Pt{X: 0.90, Y: 0.85},
	})

	left.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	right.Scatter([]float64{0.5}, []float64{0.5}, ScatterOptions{Label: "samples"})
	fig.AddLegend()

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "signal") || !containsString(r.texts, "samples") {
		t.Fatalf("unexpected figure legend labels: %v", r.texts)
	}
	if r.pathCount < 4 {
		t.Fatalf("expected figure legend box and samples, got %d paths", r.pathCount)
	}
}

func TestAnchoredTextBoxDrawsFigureAndAxesBoxes(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	})
	ax.AddAnchoredText("Peak\nstable")
	fig.AddAnchoredText("Global", AnchoredTextOptions{Location: LegendLowerRight})

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "Peak") || !containsString(r.texts, "stable") || !containsString(r.texts, "Global") {
		t.Fatalf("unexpected anchored text draws: %v", r.texts)
	}
	if r.pathCount < 2 {
		t.Fatalf("expected anchored text boxes to draw borders, got %d paths", r.pathCount)
	}
}

func TestDrawFigure_AlignsYLabelsAcrossColumn(t *testing.T) {
	fig := NewFigure(800, 600)
	top := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.55},
		Max: geom.Pt{X: 0.90, Y: 0.90},
	})
	bottom := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.10},
		Max: geom.Pt{X: 0.90, Y: 0.45},
	})

	top.SetYLabel("Top Y")
	bottom.SetYLabel("Bottom Y")
	top.YAxis.Locator = staticLocator{1000}
	top.YAxis.Formatter = ScalarFormatter{Prec: 0}
	bottom.YAxis.Locator = staticLocator{1}
	bottom.YAxis.Formatter = ScalarFormatter{Prec: 0}

	r := figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"1000": {X: 1, Y: -8, W: 24, H: 10},
			"1":    {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	DrawFigure(fig, &r)

	topAnchor, okTop := r.rotatedAnchor("Top Y")
	bottomAnchor, okBottom := r.rotatedAnchor("Bottom Y")
	if !okTop || !okBottom {
		t.Fatalf("missing ylabel draws: %v", r.rotatedText)
	}
	if topAnchor.X != bottomAnchor.X {
		t.Fatalf("expected aligned ylabel anchors, got top=%+v bottom=%+v", topAnchor, bottomAnchor)
	}
}

func TestDrawFigure_AlignsXLabelsAcrossRow(t *testing.T) {
	fig := NewFigure(800, 600)
	left := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.18},
		Max: geom.Pt{X: 0.45, Y: 0.82},
	})
	right := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.55, Y: 0.18},
		Max: geom.Pt{X: 0.90, Y: 0.82},
	})

	left.SetXLabel("Left X")
	right.SetXLabel("Right X")
	left.XAxis.Locator = staticLocator{1000}
	left.XAxis.Formatter = ScalarFormatter{Prec: 0}
	right.XAxis.Locator = staticLocator{1}
	right.XAxis.Formatter = ScalarFormatter{Prec: 0}

	r := figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"1000": {X: 1, Y: -14, W: 24, H: 20},
			"1":    {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	DrawFigure(fig, &r)

	leftOrigin, okLeft := r.textOrigin("Left X")
	rightOrigin, okRight := r.textOrigin("Right X")
	if !okLeft || !okRight {
		t.Fatalf("missing xlabel draws: %v", r.texts)
	}
	if leftOrigin.Y != rightOrigin.Y {
		t.Fatalf("expected aligned xlabel origins, got left=%+v right=%+v", leftOrigin, rightOrigin)
	}
}

func TestDrawFigure_AlignsTitlesAcrossRow(t *testing.T) {
	fig := NewFigure(800, 600)
	left := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.18},
		Max: geom.Pt{X: 0.45, Y: 0.82},
	})
	right := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.55, Y: 0.18},
		Max: geom.Pt{X: 0.90, Y: 0.82},
	})

	left.SetTitle("Left Title")
	right.SetTitle("Right Title")
	if err := left.SetXTickLabelPosition("top"); err != nil {
		t.Fatalf("left SetXTickLabelPosition(top): %v", err)
	}
	if err := right.SetXTickLabelPosition("top"); err != nil {
		t.Fatalf("right SetXTickLabelPosition(top): %v", err)
	}
	left.TopAxis().Locator = staticLocator{1000}
	left.TopAxis().Formatter = ScalarFormatter{Prec: 0}
	right.TopAxis().Locator = staticLocator{1}
	right.TopAxis().Formatter = ScalarFormatter{Prec: 0}

	r := figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"1000": {X: 1, Y: -14, W: 24, H: 20},
			"1":    {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	DrawFigure(fig, &r)

	leftOrigin, okLeft := r.textOrigin("Left Title")
	rightOrigin, okRight := r.textOrigin("Right Title")
	if !okLeft || !okRight {
		t.Fatalf("missing title draws: %v", r.texts)
	}
	if leftOrigin.Y != rightOrigin.Y {
		t.Fatalf("expected aligned title origins, got left=%+v right=%+v", leftOrigin, rightOrigin)
	}
}

type figureLayoutRecordingRenderer struct {
	render.NullRenderer
	bounds         map[string]render.TextBounds
	texts          []string
	origins        []geom.Pt
	rotatedText    []string
	rotatedAnchors []geom.Pt
	pathCount      int
}

func (r *figureLayoutRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *figureLayoutRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	metrics := render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
	if bounds, ok := r.bounds[text]; ok {
		if bounds.W > metrics.W {
			metrics.W = bounds.W
		}
		ascent := -bounds.Y
		if ascent > metrics.Ascent {
			metrics.Ascent = ascent
		}
		descent := bounds.Y + bounds.H
		if descent > metrics.Descent {
			metrics.Descent = descent
		}
		metrics.H = metrics.Ascent + metrics.Descent
	}
	return metrics
}

func (r *figureLayoutRecordingRenderer) MeasureTextBounds(text string, _ float64, _ string) (render.TextBounds, bool) {
	b, ok := r.bounds[text]
	return b, ok
}

func (r *figureLayoutRecordingRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, origin)
}

func (r *figureLayoutRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, _ float64, _ float64, _ render.Color) {
	r.rotatedText = append(r.rotatedText, text)
	r.rotatedAnchors = append(r.rotatedAnchors, anchor)
}

func (r *figureLayoutRecordingRenderer) textOrigin(text string) (geom.Pt, bool) {
	for i, candidate := range r.texts {
		if candidate == text {
			return r.origins[i], true
		}
	}
	return geom.Pt{}, false
}

func (r *figureLayoutRecordingRenderer) rotatedAnchor(text string) (geom.Pt, bool) {
	for i, candidate := range r.rotatedText {
		if candidate == text {
			return r.rotatedAnchors[i], true
		}
	}
	return geom.Pt{}, false
}
