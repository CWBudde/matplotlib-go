package render

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestTextPathBuildsValidGlyphOutlines(t *testing.T) {
	withDejaVuFontManager(t)

	textPath, ok := TextPath("Ag", geom.Pt{X: 10, Y: 20}, 14, "DejaVu Sans")
	if !ok {
		t.Fatal("TextPath returned !ok")
	}
	if len(textPath.C) == 0 || len(textPath.V) == 0 {
		t.Fatalf("expected path commands and vertices, got %+v", textPath)
	}
	if !textPath.Validate() {
		t.Fatalf("invalid path: %+v", textPath)
	}
}

func TestTextGlyphLayoutAppliesKerningAtActiveSize(t *testing.T) {
	withDejaVuFontManager(t)

	const size = 72.0
	layout, ok := LayoutTextGlyphs("Tr", geom.Pt{}, size, "DejaVu Sans")
	if !ok || len(layout.Glyphs) != 2 {
		t.Fatalf("LayoutTextGlyphs(Tr) = %+v, %v; want two glyphs", layout, ok)
	}
	separateT, okT := LayoutTextGlyphs("T", geom.Pt{}, size, "DejaVu Sans")
	separateR, okR := LayoutTextGlyphs("r", geom.Pt{}, size, "DejaVu Sans")
	if !okT || !okR {
		t.Fatalf("single-glyph layouts failed: T=%+v/%v r=%+v/%v", separateT, okT, separateR, okR)
	}

	unkerned := separateT.Advance + separateR.Advance
	if layout.Advance >= unkerned-2 {
		t.Fatalf("expected visible negative kerning for Tr at %vpx: kerned=%v unkerned=%v glyphs=%+v", size, layout.Advance, unkerned, layout.Glyphs)
	}
	if math.Abs((layout.Glyphs[1].Origin.X-separateT.Advance)-layout.Glyphs[1].Kern) > 1e-6 {
		t.Fatalf("second glyph origin does not include reported kern: layout=%+v separateT=%+v", layout, separateT)
	}
}

func TestTextGlyphLayoutKerningScalesWithSize(t *testing.T) {
	withDejaVuFontManager(t)

	small, okSmall := LayoutTextGlyphs("Te", geom.Pt{}, 12, "DejaVu Sans")
	large, okLarge := LayoutTextGlyphs("Te", geom.Pt{}, 72, "DejaVu Sans")
	if !okSmall || !okLarge || len(small.Glyphs) != 2 || len(large.Glyphs) != 2 {
		t.Fatalf("unexpected layouts: small=%+v/%v large=%+v/%v", small, okSmall, large, okLarge)
	}
	if small.Glyphs[1].Kern >= 0 || large.Glyphs[1].Kern >= 0 {
		t.Fatalf("expected negative kerning for Te: small=%+v large=%+v", small.Glyphs, large.Glyphs)
	}
	gotRatio := large.Glyphs[1].Kern / small.Glyphs[1].Kern
	if math.Abs(gotRatio-6) > 0.75 {
		t.Fatalf("kerning should scale with font size: small=%v large=%v ratio=%v", small.Glyphs[1].Kern, large.Glyphs[1].Kern, gotRatio)
	}
}

func TestTextPathBoundsUseSharedGlyphLayout(t *testing.T) {
	withDejaVuFontManager(t)

	for _, size := range []float64{12, 24, 72} {
		for _, text := range []string{"Tr", "Te", "AV"} {
			path, ok := TextPath(text, geom.Pt{}, size, "DejaVu Sans")
			if !ok {
				t.Fatalf("TextPath(%q, %v) failed", text, size)
			}
			layout, ok := LayoutTextGlyphs(text, geom.Pt{}, size, "DejaVu Sans")
			if !ok {
				t.Fatalf("LayoutTextGlyphs(%q, %v) failed", text, size)
			}
			bounds, ok := pathBounds(path)
			if !ok {
				t.Fatalf("TextPath(%q, %v) has no bounds", text, size)
			}
			if math.Abs(bounds.X-layout.Bounds.X) > 1e-6 ||
				math.Abs(bounds.Y-layout.Bounds.Y) > 1e-6 ||
				math.Abs(bounds.W-layout.Bounds.W) > 1e-6 ||
				math.Abs(bounds.H-layout.Bounds.H) > 1e-6 {
				t.Fatalf("path/layout bounds mismatch for %q at %vpx: path=%+v layout=%+v", text, size, bounds, layout.Bounds)
			}
		}
	}
}

func TestFontManagerResolveTextRunsUsesRequestedFace(t *testing.T) {
	manager := withDejaVuFontManager(t)

	runs, ok := manager.ResolveTextRuns("ABC", "DejaVu Sans")
	if !ok || len(runs) != 1 {
		t.Fatalf("ResolveTextRuns = %+v, %v; want one run", runs, ok)
	}
	if runs[0].Text != "ABC" || runs[0].Face.Path == "" {
		t.Fatalf("unexpected resolved run: %+v", runs[0])
	}
}

func TestTextPathAndLayoutUseEmbeddedFontFace(t *testing.T) {
	face, ok := embeddedFontFace("DejaVu Sans", FontProperties{Style: FontStyleNormal, Weight: 400})
	if !ok {
		t.Fatal("embeddedFontFace(DejaVu Sans) failed")
	}

	runs := []FontRun{{Text: "Ag", Face: face}}
	layout, ok := LayoutTextGlyphRuns(runs, geom.Pt{}, 14)
	if !ok {
		t.Fatal("LayoutTextGlyphRuns with embedded face failed")
	}
	if len(layout.Glyphs) != 2 {
		t.Fatalf("embedded layout glyph count = %d, want 2", len(layout.Glyphs))
	}
	if layout.Advance <= 0 {
		t.Fatalf("embedded layout advance = %v, want > 0", layout.Advance)
	}

	path, ok := textRunsPath(runs, geom.Pt{}, 14)
	if !ok {
		t.Fatal("textRunsPath with embedded face failed")
	}
	if len(path.C) == 0 || len(path.V) == 0 {
		t.Fatalf("embedded text path is empty: commands=%d vertices=%d", len(path.C), len(path.V))
	}
}

func TestTextPathSkipsMissingFontsAndEmptyInputs(t *testing.T) {
	if path, ok := TextPath("", geom.Pt{}, 12, "missing-font"); ok || len(path.C) != 0 {
		t.Fatalf("empty text should not build a path: %+v, %v", path, ok)
	}
	if path, ok := TextPath("A", geom.Pt{}, 0, "missing-font"); ok || len(path.C) != 0 {
		t.Fatalf("zero size should not build a path: %+v, %v", path, ok)
	}
	if path, ok := TextPath("A", geom.Pt{}, 12, filepath.Join(t.TempDir(), "missing.ttf")); ok || len(path.C) != 0 {
		t.Fatalf("missing font should not build a path: %+v, %v", path, ok)
	}
}

func withDejaVuFontManager(t *testing.T) *FontManager {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "DejaVuSans.ttf")
	if err := os.WriteFile(path, dejavusans.TTF, 0o644); err != nil {
		t.Fatalf("write test font: %v", err)
	}

	manager := NewFontManager()
	manager.AddFontDir(dir)
	previous := defaultFontManager
	defaultFontManager = manager
	t.Cleanup(func() { defaultFontManager = previous })
	return manager
}

func pathBounds(path geom.Path) (TextBounds, bool) {
	if len(path.V) == 0 {
		return TextBounds{}, false
	}
	minX, minY := path.V[0].X, path.V[0].Y
	maxX, maxY := minX, minY
	for _, pt := range path.V[1:] {
		minX = math.Min(minX, pt.X)
		minY = math.Min(minY, pt.Y)
		maxX = math.Max(maxX, pt.X)
		maxY = math.Max(maxY, pt.Y)
	}
	return TextBounds{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}, true
}
