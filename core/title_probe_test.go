package core

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	"matplotlib-go/backends/agg"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
	"matplotlib-go/test/imagecmp"
)

func TestTitleTuneProbe(t *testing.T) {
	if os.Getenv("MPL_TITLE_PROBE") != "1" {
		t.Skip("probe only")
	}

	wantPath := filepath.Join("..", "testdata", "matplotlib_ref", "title_strict.png")
	want, err := imagecmp.LoadPNG(wantPath)
	if err != nil {
		t.Fatalf("load matplotlib reference %s: %v", wantPath, err)
	}

	bestMean := math.MaxFloat64
	bestPSNR := -1.0
	bestScale, bestX, bestY := 0.0, 0.0, 0.0

	for scale := 0.995; scale <= 1.0100001; scale += 0.00025 {
		for x := -0.75; x <= -0.10; x += 0.0078125 {
			for y := -0.35; y <= 0.10; y += 0.0078125 {
				titleScaleFactor = scale
				titleXAdjustPxValue = x
				titleBaselineAdjustPxV = y

				got := renderProbeTitleStrict(t)
				diff, err := imagecmp.ComparePNG(got, want, 1)
				if err != nil {
					t.Fatalf("compare image: %v", err)
				}
				if diff.MeanAbs < bestMean-1e-9 || (math.Abs(diff.MeanAbs-bestMean) < 1e-9 && diff.PSNR > bestPSNR) {
					bestMean = diff.MeanAbs
					bestPSNR = diff.PSNR
					bestScale, bestX, bestY = scale, x, y
					t.Logf("best mean=%.4f psnr=%.4f scale=%.4f x=%.4f y=%.4f", bestMean, bestPSNR, bestScale, bestX, bestY)
				}
			}
		}
	}

	fmt.Printf("best mean=%.4f psnr=%.4f scale=%.4f x=%.4f y=%.4f\n", bestMean, bestPSNR, bestScale, bestX, bestY)
}

func renderProbeTitleStrict(t *testing.T) image.Image {
	t.Helper()

	fig := NewFigure(320, 80)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.40},
		Max: geom.Pt{X: 0.95, Y: 0.85},
	})
	ax.SetTitle("Histogram Strategies")
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	r, err := agg.New(320, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("new renderer: %v", err)
	}
	if setter, ok := any(r).(interface{ SetResolution(uint) }); ok {
		setter.SetResolution(uint(style.Default.DPI))
	}
	DrawFigure(fig, r)
	return r.GetImage()
}
