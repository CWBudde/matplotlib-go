//go:build freetype

package agg

import (
	"bytes"
	"math"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestUsesDejaVuSansWithoutFallback(t *testing.T) {
	r := mustNew(t, 200, 100)
	if r.fontPath == "" {
		t.Fatal("expected DejaVu Sans font path to be configured")
	}

	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 100}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	samples := []struct {
		text string
		size float64
		pos  geom.Pt
	}{
		{text: "Text Labels", size: 12, pos: geom.Pt{X: 10, Y: 24}},
		{text: "Group", size: 11.64, pos: geom.Pt{X: 10, Y: 44}},
		{text: "0.0", size: 11.64, pos: geom.Pt{X: 10, Y: 64}},
	}
	for _, sample := range samples {
		metrics := r.MeasureText(sample.text, sample.size, "")
		if metrics.W <= 0 || metrics.Ascent <= 0 || metrics.H <= 0 {
			t.Fatalf("invalid metrics for %q: %+v", sample.text, metrics)
		}
		r.DrawText(sample.text, sample.pos, sample.size, white)
	}
	r.DrawTextRotated("Value", geom.Pt{X: 160, Y: 50}, 11.64, math.Pi/2, white)

	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if r.fallback {
		t.Fatal("expected AGG outline FreeType text to be used without falling back to GSV")
	}
}

func TestRasterTextWidthTracksRendererDPI(t *testing.T) {
	r := mustNew(t, 200, 100)

	r.SetResolution(72)
	width72 := r.MeasureText("Basic Bars", 12, "").W

	r.SetResolution(96)
	width96 := r.MeasureText("Basic Bars", 12, "").W

	if width72 <= 0 || width96 <= 0 {
		t.Fatalf("expected positive widths, got 72dpi=%v 96dpi=%v", width72, width96)
	}
	if width96 <= width72 {
		t.Fatalf("expected width to increase with DPI, got 72dpi=%v 96dpi=%v", width72, width96)
	}

	gotRatio := width96 / width72
	wantRatio := 96.0 / 72.0
	if math.Abs(gotRatio-wantRatio) > 0.15 {
		t.Fatalf("unexpected DPI scaling ratio: got=%v want=%v", gotRatio, wantRatio)
	}
}

func TestTrailingSpaceDoesNotRenderDuplicateGlyph(t *testing.T) {
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}

	renderText := func(text string) []byte {
		r := mustNew(t, 160, 80)
		viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 160, Y: 80}}
		if err := r.Begin(viewport); err != nil {
			t.Fatalf("Begin failed: %v", err)
		}
		r.DrawText(text, geom.Pt{X: 20, Y: 42}, 24, textColor)
		if err := r.End(); err != nil {
			t.Fatalf("End failed: %v", err)
		}
		img := r.GetImage()
		return append([]byte(nil), img.Pix...)
	}

	withoutTrailingSpace := renderText("x")
	withTrailingSpace := renderText("x ")
	if !bytes.Equal(withoutTrailingSpace, withTrailingSpace) {
		t.Fatal("expected trailing space to add no ink; raster text appears to replay the previous glyph")
	}
}

func TestInternalSpaceDoesNotReplayPreviousGlyph(t *testing.T) {
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}

	renderText := func(text string) []byte {
		r := mustNew(t, 320, 80)
		viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 320, Y: 80}}
		if err := r.Begin(viewport); err != nil {
			t.Fatalf("Begin failed: %v", err)
		}
		r.DrawText(text, geom.Pt{X: 20, Y: 42}, 24, textColor)
		if err := r.End(); err != nil {
			t.Fatalf("End failed: %v", err)
		}
		return r.GetImage().Pix
	}

	withSingleSpace := append([]byte(nil), renderText("Histogram Strategies")...)
	withoutSpace := append([]byte(nil), renderText("HistogramStrategies")...)
	withDoubleLetter := append([]byte(nil), renderText("HistogrammStrategies")...)

	eqNoSpace := bytes.Equal(withSingleSpace, withoutSpace)
	eqDoubleLetter := bytes.Equal(withSingleSpace, withDoubleLetter)
	if eqNoSpace || eqDoubleLetter {
		t.Fatalf(
			"internal-space rendering collapsed unexpectedly: equals_no_space=%v equals_double_letter=%v",
			eqNoSpace, eqDoubleLetter,
		)
	}
}
