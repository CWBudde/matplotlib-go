package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestNormalizeDisplayText_ReplacesBasicMathTokens(t *testing.T) {
	got := normalizeDisplayText(`\\mu = 1.2, \\Delta x \\rightarrow \\pi`)
	want := "μ = 1.2, Δ x → π"
	if got != want {
		t.Fatalf("unexpected normalized text: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_ParsesInlineMath(t *testing.T) {
	got := normalizeDisplayText(`signal $\\alpha^2 + \\beta_i$ peak`)
	want := "signal α² + βᵢ peak"
	if got != want {
		t.Fatalf("unexpected inline math normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_FormatsFractionsAndRoots(t *testing.T) {
	got := normalizeDisplayText(`$\\frac{1}{\\sqrt{2}}$`)
	want := "1⁄√2"
	if got != want {
		t.Fatalf("unexpected fraction/root normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_HandlesGroupedScripts(t *testing.T) {
	got := normalizeDisplayText(`$x_{\\mathrm{max}}$`)
	want := "xₘₐₓ"
	if got != want {
		t.Fatalf("unexpected grouped subscript normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_PreservesUnmatchedDollar(t *testing.T) {
	got := normalizeDisplayText(`cost is $5`)
	want := "cost is $5"
	if got != want {
		t.Fatalf("unexpected unmatched dollar normalization: got %q want %q", got, want)
	}
}

func TestAlignedTextOrigin(t *testing.T) {
	anchor := geom.Pt{X: 100, Y: 50}
	metrics := render.TextMetrics{W: 40, Ascent: 8, Descent: 2}

	got := alignedTextOrigin(anchor, metrics, TextAlignCenter, TextVAlignTop)
	if got.X != 80 || got.Y != 58 {
		t.Fatalf("unexpected text origin: %+v", got)
	}
}

func TestAxesTextDrawsNormalizedContent(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.5, 0.5, `\\alpha + \\beta`, TextOptions{
		HAlign:   TextAlignCenter,
		VAlign:   TextVAlignMiddle,
		FontSize: 12,
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 1 {
		t.Fatalf("expected exactly one text draw, got %d (%v)", len(r.texts), r.texts)
	}
	if r.texts[0] != "α + β" {
		t.Fatalf("unexpected rendered text %q", r.texts[0])
	}
}

func TestAnnotationDrawOverlayRendersArrowAndText(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Annotate("peak", 0.5, 0.5)

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 1 || r.texts[0] != "peak" {
		t.Fatalf("unexpected annotation texts: %v", r.texts)
	}
	if r.pathCount < 2 {
		t.Fatalf("expected annotation arrow line and head, got %d paths", r.pathCount)
	}
}

type textRecordingRenderer struct {
	render.NullRenderer
	pathCount int
	texts     []string
	origins   []geom.Pt
}

func (r *textRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *textRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *textRecordingRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, origin)
}

func TestAxesTextSupportsAxesAndBlendedCoordinates(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.25, 0.75, "axes", TextOptions{
		Coords:  Coords(CoordAxes),
		OffsetX: 5,
		OffsetY: -7,
	})
	ax.Text(0.25, 0.75, "blend", TextOptions{
		Coords: BlendCoords(CoordFigure, CoordAxes),
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 2 {
		t.Fatalf("expected 2 text draws, got %d", len(r.texts))
	}

	wantAxes := geom.Pt{X: 245, Y: 173}
	if r.origins[0] != wantAxes {
		t.Fatalf("axes coords origin = %+v, want %+v", r.origins[0], wantAxes)
	}

	wantBlend := geom.Pt{X: 200, Y: 180}
	if r.origins[1] != wantBlend {
		t.Fatalf("blended coords origin = %+v, want %+v", r.origins[1], wantBlend)
	}
}

func TestAxesLabelsDrawNormalizedMathText(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle(`$\\alpha^2$`)
	ax.SetXLabel(`phase $\\theta$`)
	ax.SetYLabel(`amp $\\frac{1}{2}$`)

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if !containsTextString(r.texts, "α²") {
		t.Fatalf("missing normalized title draw: %v", r.texts)
	}
	if !containsTextString(r.texts, "phase θ") {
		t.Fatalf("missing normalized xlabel draw: %v", r.texts)
	}
	if !containsTextString(r.texts, "amp 1⁄2") {
		t.Fatalf("missing normalized ylabel draw: %v", r.texts)
	}
}

func containsTextString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
