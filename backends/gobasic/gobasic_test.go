package gobasic

import (
	"math"
	"image"
	"image/color"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestNew(t *testing.T) {
	r := New(100, 50, render.Color{R: 1, G: 1, B: 1, A: 1})

	if r == nil {
		t.Fatal("New returned nil")
	}

	img := r.GetImage()
	if img == nil {
		t.Fatal("GetImage returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 50 {
		t.Errorf("Expected dimensions 100x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Check that background color is set (sample a few pixels)
	expectedColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for _, pt := range []image.Point{{0, 0}, {50, 25}, {99, 49}} {
		if c := img.RGBAAt(pt.X, pt.Y); c != expectedColor {
			t.Errorf("Expected background color %v at %v, got %v", expectedColor, pt, c)
		}
	}
}

func TestBeginEnd(t *testing.T) {
	r := New(100, 50, render.Color{R: 0, G: 0, B: 0, A: 1})

	// Test Begin
	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err != nil {
		t.Errorf("Begin failed: %v", err)
	}

	// Test double Begin should fail
	err = r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err == nil {
		t.Error("Expected Begin to fail when called twice")
	}

	// Test End
	err = r.End()
	if err != nil {
		t.Errorf("End failed: %v", err)
	}

	// Test End without Begin should fail
	err = r.End()
	if err == nil {
		t.Error("Expected End to fail when called without Begin")
	}
}

func TestSaveRestore(t *testing.T) {
	r := New(100, 50, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Set a clip rect
	clipRect := geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 90, Y: 40}}
	r.ClipRect(clipRect)

	// Save state
	r.Save()

	// Modify clip rect
	newClipRect := geom.Rect{Min: geom.Pt{X: 20, Y: 20}, Max: geom.Pt{X: 80, Y: 30}}
	r.ClipRect(newClipRect)

	// Restore should bring back original clip rect
	r.Restore()

	// Can't easily test the clip rect is restored without internal access,
	// but at least ensure Save/Restore don't crash
}

func TestPathFill(t *testing.T) {
	r := New(100, 50, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Create a simple triangle path
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{{X: 50, Y: 10}, {X: 30, Y: 40}, {X: 70, Y: 40}},
	}

	paint := render.Paint{
		Fill: render.Color{R: 1, G: 0, B: 0, A: 1}, // Red fill
	}

	// Should not crash
	r.Path(path, &paint)

	// Check that some pixels changed from white background
	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false

	// Check a few pixels around the triangle center
	for y := 20; y <= 30; y++ {
		for x := 45; x <= 55; x++ {
			if c := img.RGBAAt(x, y); c != whiteColor {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}

	if !changed {
		t.Error("Expected some pixels to change from background color after drawing triangle")
	}
}

func TestPathStroke(t *testing.T) {
	r := New(100, 50, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Create a simple line path
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 10, Y: 25}, {X: 90, Y: 25}},
	}

	paint := render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 1, A: 1}, // Blue stroke
		LineWidth: 2.0,
	}

	// Should not crash
	r.Path(path, &paint)

	// Check that some pixels changed from white background
	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false

	// Check pixels along the line
	for x := 20; x <= 80; x += 10 {
		if c := img.RGBAAt(x, 25); c != whiteColor {
			changed = true
			break
		}
	}

	if !changed {
		t.Error("Expected some pixels to change from background color after drawing line")
	}
}

func TestClipRectAllowsPathDrawingAcrossIndependentRegions(t *testing.T) {
	r := New(240, 240, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 240, Y: 240}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	quadrants := []geom.Rect{
		{Min: geom.Pt{X: 20, Y: 20}, Max: geom.Pt{X: 100, Y: 100}},
		{Min: geom.Pt{X: 140, Y: 20}, Max: geom.Pt{X: 220, Y: 100}},
		{Min: geom.Pt{X: 20, Y: 140}, Max: geom.Pt{X: 100, Y: 220}},
		{Min: geom.Pt{X: 140, Y: 140}, Max: geom.Pt{X: 220, Y: 220}},
	}
	colors := []render.Color{
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 1, G: 0.5, B: 0, A: 1},
	}

	for i, clip := range quadrants {
		r.Save()
		r.ClipRect(clip)
		path := geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.LineTo, geom.ClosePath},
			V: []geom.Pt{
				{X: clip.Min.X + 8, Y: clip.Min.Y + 8},
				{X: clip.Max.X - 8, Y: clip.Min.Y + 8},
				{X: clip.Max.X - 8, Y: clip.Max.Y - 8},
				{X: clip.Min.X + 8, Y: clip.Max.Y - 8},
			},
		}
		r.Path(path, &render.Paint{Fill: colors[i]})
		r.Restore()
	}

	img := r.GetImage()
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for i, clip := range quadrants {
		center := image.Point{
			X: int((clip.Min.X + clip.Max.X) / 2),
			Y: int((clip.Min.Y + clip.Max.Y) / 2),
		}
		if got := img.RGBAAt(center.X, center.Y); got == white {
			t.Fatalf("quadrant %d center remained background at %v", i, center)
		}
	}
}

func TestMeasureText(t *testing.T) {
	r := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Test empty string
	metrics := r.MeasureText("", 12, "default")
	if metrics.W != 0 || metrics.H != 0 {
		t.Errorf("Expected zero metrics for empty string, got W=%v, H=%v", metrics.W, metrics.H)
	}

	// Test basic text
	metrics = r.MeasureText("Hello", 13, "default")
	if metrics.W <= 0 || metrics.H <= 0 {
		t.Errorf("Expected positive metrics for text, got W=%v, H=%v", metrics.W, metrics.H)
	}

	// Test scaling - larger size should give larger metrics
	metricsSmall := r.MeasureText("Test", 10, "default")
	metricsLarge := r.MeasureText("Test", 20, "default")
	if metricsLarge.W <= metricsSmall.W || metricsLarge.H <= metricsSmall.H {
		t.Errorf("Expected larger metrics for larger size, got small: W=%v,H=%v, large: W=%v,H=%v",
			metricsSmall.W, metricsSmall.H, metricsLarge.W, metricsLarge.H)
	}
}

func TestDrawTextRenders(t *testing.T) {
	r := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 100}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Text should render visible glyphs in GoBasic.
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1} // black
	origin := geom.Pt{X: 10, Y: 50}
	r.DrawText("Hello, World!", origin, 13, textColor)

	// Verify that the image has changed from the white background.
	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if c := img.RGBAAt(x, y); c != whiteColor {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}

	if !changed {
		t.Fatal("Expected at least one pixel to change after DrawText")
	}
}

func TestDrawTextRotatedRenders(t *testing.T) {
	r := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 100}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	r.DrawTextRotated("Hello", geom.Pt{X: 100, Y: 50}, 13, math.Pi/4, textColor)

	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if c := img.RGBAAt(x, y); c != whiteColor {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}

	if !changed {
		t.Fatal("Expected at least one pixel to change after DrawTextRotated")
	}
}

func TestGlyphRun(t *testing.T) {
	r := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 100}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Test GlyphRun - should not panic even with limited implementation
	glyphRun := render.GlyphRun{
		Glyphs:  []render.Glyph{{ID: 'H', Advance: 7, Offset: geom.Pt{}}},
		Origin:  geom.Pt{X: 10, Y: 50},
		Size:    13,
		FontKey: "default",
	}
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}

	// Should render a glyph when the ID maps to a visible rune.
	r.GlyphRun(glyphRun, textColor)

	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if c := img.RGBAAt(x, y); c != whiteColor {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}
	if !changed {
		t.Fatal("Expected DrawText to render at least one pixel")
	}
}
