package agg

import (
	"bytes"
	"testing"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestRasterTextUsesEmbeddedFontFace(t *testing.T) {
	r, err := New(220, 120, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	face := render.FontFace{Family: "DejaVu Sans", Data: dejavusans.TTF}
	metrics, ok := r.measureRasterText("Ag", face, 24)
	if !ok {
		t.Fatal("measureRasterText with embedded face failed")
	}
	if metrics.W <= 0 || metrics.H <= 0 {
		t.Fatalf("embedded raster metrics = %+v, want positive dimensions", metrics)
	}
	if _, ok := rasterFontHeightMetrics(face, 24, 72); !ok {
		t.Fatal("rasterFontHeightMetrics with embedded face failed")
	}

	if !r.drawRasterText("Ag", face, geom.Pt{X: 20, Y: 60}, 24, render.Color{R: 0, G: 0, B: 0, A: 1}) {
		t.Fatal("drawRasterText with embedded face failed")
	}
}

func TestDrawRasterTextUsesSharedCombiningMarkShape(t *testing.T) {
	face := render.FontFace{Family: "DejaVu Sans", Data: dejavusans.TTF}
	decomposed := renderRasterTextPixels(t, "e\u0301", face)
	precomposed := renderRasterTextPixels(t, "\u00e9", face)

	if !bytes.Equal(decomposed, precomposed) {
		t.Fatal("decomposed e-acute raster output differs from precomposed e-acute")
	}
}

func TestMeasureRasterTextUsesSharedShapedAdvance(t *testing.T) {
	r, err := New(220, 120, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	face := render.FontFace{Family: "DejaVu Sans", Data: dejavusans.TTF}
	const size = 72.0
	shaped, ok := render.ShapeText("fi", geom.Pt{}, r.fontPixelSize(size), render.TextShapingOptions{FontKey: fontReference(face)})
	if !ok || len(shaped.Glyphs) != 1 {
		t.Fatalf("ShapeText(fi) = %+v, %v; want one ligature glyph", shaped, ok)
	}

	metrics, ok := r.measureRasterText("fi", face, size)
	if !ok {
		t.Fatal("measureRasterText(fi) failed")
	}
	if metrics.W != quantize(shaped.Advance.X) {
		t.Fatalf("measureRasterText(fi).W = %v, want shaped advance %v", metrics.W, quantize(shaped.Advance.X))
	}
}

func renderRasterTextPixels(t *testing.T, text string, face render.FontFace) []byte {
	t.Helper()

	r, err := New(180, 120, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !r.drawRasterText(text, face, geom.Pt{X: 20, Y: 80}, 48, render.Color{R: 0, G: 0, B: 0, A: 1}) {
		t.Fatalf("drawRasterText(%q) failed", text)
	}
	return append([]byte(nil), r.GetImage().Pix...)
}
