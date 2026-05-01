package core

import (
	"path/filepath"
	"strings"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// testPNGSVGRenderer is a minimal renderer that satisfies render.Renderer (via
// the embedded NullRenderer), render.PNGExporter, and render.SVGExporter. It
// records which export path was exercised so tests can assert dispatch.
type testPNGSVGRenderer struct {
	render.NullRenderer
	savedPNG bool
	savedSVG bool
	pngPath  string
	svgPath  string
}

func newTestPNGSVGRenderer() *testPNGSVGRenderer { return &testPNGSVGRenderer{} }

func (r *testPNGSVGRenderer) SavePNG(path string) error {
	r.savedPNG = true
	r.pngPath = path
	return nil
}

func (r *testPNGSVGRenderer) SaveSVG(path string) error {
	r.savedSVG = true
	r.svgPath = path
	return nil
}

// defaultAxesRect is the unit-square rect used by SaveFig tests when adding
// axes; the dispatch logic does not depend on its precise value.
var defaultAxesRect = geom.Rect{
	Min: geom.Pt{X: 0.1, Y: 0.1},
	Max: geom.Pt{X: 0.9, Y: 0.9},
}

func TestSaveFig_DispatchesByExtension_PNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.png")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedPNG {
		t.Fatal("expected PNG path to be exercised")
	}
	if r.savedSVG {
		t.Fatal("did not expect SVG path")
	}
	if r.pngPath != path {
		t.Fatalf("SavePNG received path %q, want %q", r.pngPath, path)
	}
}

func TestSaveFig_DispatchesByExtension_SVG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.svg")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path); err != nil {
		t.Fatalf("SaveFig: %v", err)
	}
	if !r.savedSVG {
		t.Fatal("expected SVG path to be exercised")
	}
	if r.savedPNG {
		t.Fatal("did not expect PNG path")
	}
	if r.svgPath != path {
		t.Fatalf("SaveSVG received path %q, want %q", r.svgPath, path)
	}
}

func TestSaveFig_RejectsUnknownExtension(t *testing.T) {
	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)
	r := newTestPNGSVGRenderer()

	err := SaveFig(fig, r, "/tmp/out.tiff")
	if err == nil {
		t.Fatal("expected error for unknown extension")
	}
	if !strings.Contains(err.Error(), ".tiff") {
		t.Fatalf("error should mention extension .tiff, got: %v", err)
	}
	if !strings.Contains(err.Error(), ".png") || !strings.Contains(err.Error(), ".svg") {
		t.Fatalf("error should list supported extensions, got: %v", err)
	}
	if r.savedPNG || r.savedSVG {
		t.Fatal("no exporter should have been invoked for unknown extension")
	}
}

func TestSaveFig_NoExtensionRejected(t *testing.T) {
	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)
	r := newTestPNGSVGRenderer()

	err := SaveFig(fig, r, "/tmp/out")
	if err == nil {
		t.Fatal("expected error for missing extension")
	}
	if !strings.Contains(err.Error(), ".png") || !strings.Contains(err.Error(), ".svg") {
		t.Fatalf("error should list supported extensions, got: %v", err)
	}
	if r.savedPNG || r.savedSVG {
		t.Fatal("no exporter should have been invoked for missing extension")
	}
}

func TestSaveFig_UppercaseExtensionWorks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.PNG")

	fig := NewFigure(100, 80)
	fig.AddAxes(defaultAxesRect)

	r := newTestPNGSVGRenderer()
	if err := SaveFig(fig, r, path); err != nil {
		t.Fatalf("SaveFig with .PNG: %v", err)
	}
	if !r.savedPNG {
		t.Fatal("uppercase extension should still hit PNG path")
	}
}
