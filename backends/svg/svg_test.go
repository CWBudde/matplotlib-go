package svg

import (
	"os"
	"strings"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func mustNewRenderer(t *testing.T) *Renderer {
	t.Helper()
	r, err := New(180, 120, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	return r
}

func TestNewInvalidDimensions(t *testing.T) {
	r, err := New(0, 10, render.Color{})
	if err == nil || r != nil {
		t.Fatal("expected error for non-positive dimensions")
	}
}

func TestSaveSVG(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	var path geom.Path
	path.MoveTo(geom.Pt{X: 10, Y: 10})
	path.LineTo(geom.Pt{X: 170, Y: 110})
	r.Path(path, &render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 2,
	})

	r.DrawText("line", geom.Pt{X: 20, Y: 30}, 14, render.Color{R: 0, G: 0, B: 0, A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	tmp, err := os.CreateTemp("", "matplotlib-go-svg-*.svg")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmpPath) })

	if err := r.SaveSVG(tmpPath); err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "<svg") || !strings.Contains(content, "</svg>") {
		t.Fatal("SVG output missing root element")
	}
	if !strings.Contains(content, "<path") {
		t.Fatal("SVG output missing path node")
	}
	if !strings.Contains(content, "<text") || !strings.Contains(content, ">line<") {
		t.Fatal("SVG output missing text node")
	}
}

func TestSaveSVGPreservesClip(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	r.ClipRect(geom.Rect{
		Min: geom.Pt{X: 10, Y: 10},
		Max: geom.Pt{X: 50, Y: 50},
	})
	r.DrawText("clipped", geom.Pt{X: 20, Y: 20}, 12, render.Color{R: 1})
	r.End()

	tmp, err := os.CreateTemp("", "matplotlib-go-svg-clip-*.svg")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmpPath) })

	if err := r.SaveSVG(tmpPath); err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<clipPath") {
		t.Fatal("SVG output should contain clipPath definitions")
	}
	if !strings.Contains(content, "clip-path=\"url(#") {
		t.Fatal("SVG output should apply clip-path to content")
	}
}

func TestDrawTextSupportsNegativeCoordinates(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.DrawText("neg", geom.Pt{X: -15, Y: 30}, 12, render.Color{R: 0})
	r.End()

	tmp, err := os.CreateTemp("", "matplotlib-go-svg-negative-*.svg")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmpPath) })

	if err := r.SaveSVG(tmpPath); err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "x=\"-15.000000\"") {
		t.Fatalf("expected preserved negative x coordinate, got %q", content)
	}
}
