package agg

import (
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
