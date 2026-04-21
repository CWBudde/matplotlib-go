package svg

import (
	"image"
	"image/color"
	"os"
	"regexp"
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

func renderSVGDocument(t *testing.T, draw func(*Renderer)) string {
	t.Helper()

	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	draw(r)

	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	return r.renderSVG()
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

func TestGlyphRunUsesFontKeyAndOffsetsWithoutAccumulatingOffset(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.GlyphRun(render.GlyphRun{
			Glyphs: []render.Glyph{
				{ID: 'A', Advance: 8, Offset: geom.Pt{X: 2, Y: 3}},
				{ID: 'B', Advance: 7, Offset: geom.Pt{X: 1, Y: -1}},
			},
			Origin:  geom.Pt{X: 10, Y: 20},
			Size:    12,
			FontKey: "sans-serif",
		}, render.Color{A: 1})
	})

	if !strings.Contains(content, `font-family="DejaVu Sans, Arial, sans-serif"`) {
		t.Fatalf("glyph run should honor sans-serif font selection, got %q", content)
	}
	if !strings.Contains(content, `<text x="12.000000" y="23.000000"`) {
		t.Fatalf("first glyph should render at origin plus its offset, got %q", content)
	}
	if !strings.Contains(content, `<text x="19.000000" y="19.000000"`) {
		t.Fatalf("second glyph should advance from origin without accumulating prior offsets, got %q", content)
	}
}

func TestBeginResetsLastFontKey(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}

	if err := r.Begin(viewport); err != nil {
		t.Fatalf("first Begin failed: %v", err)
	}
	r.MeasureText("sample", 12, "monospace")
	r.DrawText("first", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("first End failed: %v", err)
	}

	if err := r.Begin(viewport); err != nil {
		t.Fatalf("second Begin failed: %v", err)
	}
	r.DrawText("second", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("second End failed: %v", err)
	}

	content := r.renderSVG()
	if strings.Contains(content, `font-family="DejaVu Sans Mono, monospace"`) {
		t.Fatalf("plain text should not inherit font family from a previous drawing session, got %q", content)
	}
	if !strings.Contains(content, `font-family="DejaVu Sans, Arial, sans-serif"`) {
		t.Fatalf("plain text should fall back to default sans font family, got %q", content)
	}
}

func TestRenderSVGPreservesClipStackAcrossSaveRestore(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipRect(geom.Rect{
			Min: geom.Pt{X: 5, Y: 5},
			Max: geom.Pt{X: 50, Y: 60},
		})
		r.Save()
		r.ClipRect(geom.Rect{
			Min: geom.Pt{X: 10, Y: 20},
			Max: geom.Pt{X: 30, Y: 40},
		})
		r.DrawText("inner", geom.Pt{X: 15, Y: 25}, 12, render.Color{A: 1})
		r.Restore()
		r.DrawText("outer", geom.Pt{X: 15, Y: 25}, 12, render.Color{A: 1})
	})

	if strings.Count(content, "<clipPath") != 2 {
		t.Fatalf("expected two clip path definitions after nested clipping, got %q", content)
	}
	if !strings.Contains(content, `<rect x="5.000000" y="5.000000" width="45.000000" height="55.000000" />`) {
		t.Fatalf("missing outer clip rect in SVG defs: %q", content)
	}
	if !strings.Contains(content, `<rect x="10.000000" y="20.000000" width="20.000000" height="20.000000" />`) {
		t.Fatalf("missing intersected inner clip rect in SVG defs: %q", content)
	}

	re := regexp.MustCompile(`<g clip-path="url\(#(clip\d+)\)"><text[^>]*>inner</text></g>\s*<g clip-path="url\(#(clip\d+)\)"><text[^>]*>outer</text></g>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) != 3 {
		t.Fatalf("expected clipped inner and restored outer groups, got %q", content)
	}
	if matches[1] == matches[2] {
		t.Fatalf("expected restore to switch back to the outer clip, got %q", content)
	}
}

func TestPathSerializesStrokeFillOpacityAndDashes(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 10, Y: 15})
		path.LineTo(geom.Pt{X: 60, Y: 15})
		path.LineTo(geom.Pt{X: 60, Y: 45})
		path.Close()

		r.Path(path, &render.Paint{
			LineWidth:  2.5,
			LineJoin:   render.JoinRound,
			LineCap:    render.CapSquare,
			MiterLimit: 7,
			Stroke:     render.Color{R: 1, G: 0, B: 0, A: 0.25},
			Fill:       render.Color{G: 1, A: 0.5},
			Dashes:     []float64{4, 2, 1, 3},
		})
	})

	for _, want := range []string{
		`fill="rgb(0,255,0)"`,
		`fill-opacity="0.500000"`,
		`stroke="rgb(255,0,0)"`,
		`stroke-opacity="0.250000"`,
		`stroke-width="2.500000"`,
		`stroke-linejoin="round"`,
		`stroke-linecap="square"`,
		`stroke-miterlimit="7.000000"`,
		`stroke-dasharray="4.000000,2.000000,1.000000,3.000000"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected serialized path attribute %q in %q", want, content)
		}
	}
}

func TestImageSerializesEmbeddedPNGAndNormalizesDestinationRect(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.SetRGBA(0, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		r.Image(render.NewImageData(img), geom.Rect{
			Min: geom.Pt{X: 30, Y: 40},
			Max: geom.Pt{X: 10, Y: 20},
		})
	})

	for _, want := range []string{
		`<image x="10.000000" y="20.000000" width="20.000000" height="20.000000"`,
		`preserveAspectRatio="none"`,
		`href="data:image/png;base64,`,
		`xlink:href="data:image/png;base64,`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected serialized image attribute %q in %q", want, content)
		}
	}
}

func TestTextEscapingAndRotationSerialization(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.DrawTextRotated(`A<&"B`, geom.Pt{X: 90, Y: 70}, 12, 0.5, render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.75})
	})

	for _, want := range []string{
		`transform="rotate(-28.647890 90.000000 70.000000)"`,
		`fill="rgb(51,102,153)"`,
		`fill-opacity="0.750000"`,
		`A&lt;&amp;&#34;B`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected rotated text attribute %q in %q", want, content)
		}
	}
}

func TestDrawTextVerticalEmitsOneNodePerRune(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.DrawTextVertical("AB", geom.Pt{X: 90, Y: 60}, 12, render.Color{A: 1})
	})

	if strings.Count(content, "<text") != 2 {
		t.Fatalf("expected one text node per rune for vertical text, got %q", content)
	}
	if !strings.Contains(content, ">A</text>") || !strings.Contains(content, ">B</text>") {
		t.Fatalf("expected both vertical glyph nodes in %q", content)
	}
}
