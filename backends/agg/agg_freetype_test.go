//go:build freetype

package agg

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestUsesDejaVuSansWithoutFallback(t *testing.T) {
	r := mustNew(t, 200, 100)
	if r.fontPath == "" {
		t.Fatal("expected DejaVu Sans font path to be configured")
	}
	if r.ctx.textForceAuto {
		t.Fatal("expected raster FreeType text to use Matplotlib's default hinting mode without forced autohint")
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

func TestRasterTextKerningMatchesSharedGlyphLayout(t *testing.T) {
	r := mustNew(t, 260, 120)

	for _, dpi := range []uint{72, 96, 144} {
		r.SetResolution(dpi)
		for _, size := range []float64{12, 24, 72} {
			for _, text := range []string{"Tr", "Te", "To", "Ta", "AV", "WA", "Yo"} {
				t.Run(text, func(t *testing.T) {
					metrics := r.MeasureText(text, size, "")
					layout, ok := render.LayoutTextGlyphs(text, geom.Pt{}, r.fontPixelSize(size), r.fontPath)
					if !ok {
						t.Fatalf("LayoutTextGlyphs(%q) failed", text)
					}
					if math.Abs(metrics.W-quantize(layout.Advance)) > 1e-6 {
						t.Fatalf("MeasureText(%q).W=%v, layout advance=%v glyphs=%+v", text, metrics.W, layout.Advance, layout.Glyphs)
					}
					first := r.MeasureText(text[:1], size, "").W
					second := r.MeasureText(text[1:], size, "").W
					if layout.Glyphs[1].Kern < -0.5 && metrics.W >= first+second-0.5 {
						t.Fatalf("kerned pair %q should be narrower than separate glyphs: pair=%v singles=%v kern=%v", text, metrics.W, first+second, layout.Glyphs[1].Kern)
					}
				})
			}
		}
	}
}

func TestRasterTextBoundsMatchSharedGlyphLayout(t *testing.T) {
	r := mustNew(t, 260, 120)
	r.SetResolution(96)

	for _, text := range []string{"Tr", "Te", "AV"} {
		t.Run(text, func(t *testing.T) {
			const size = 48.0
			bounds, ok := r.MeasureTextBounds(text, size, "")
			if !ok {
				t.Fatalf("MeasureTextBounds(%q) failed", text)
			}
			layout, ok := render.LayoutTextGlyphs(text, geom.Pt{}, r.fontPixelSize(size), r.fontPath)
			if !ok {
				t.Fatalf("LayoutTextGlyphs(%q) failed", text)
			}
			if math.Abs(bounds.X-layout.Bounds.X) > 1e-6 ||
				math.Abs(bounds.Y-layout.Bounds.Y) > 1e-6 ||
				math.Abs(bounds.W-layout.Bounds.W) > 1e-6 ||
				math.Abs(bounds.H-layout.Bounds.H) > 1e-6 {
				t.Fatalf("bounds mismatch for %q: renderer=%+v layout=%+v", text, bounds, layout.Bounds)
			}
		})
	}
}

func TestKerningMetricsMatchMatplotlibRendererAgg(t *testing.T) {
	r := mustNew(t, 260, 120)
	if r.fontPath == "" {
		t.Fatal("expected DejaVu Sans font path to be configured")
	}

	cases := []mplTextMetricCase{}
	for _, dpi := range []uint{72, 96, 144} {
		for _, size := range []float64{12, 24, 72} {
			for _, text := range []string{"Tr", "Te", "To", "Ta", "AV", "WA", "Yo"} {
				cases = append(cases, mplTextMetricCase{Text: text, Size: size, DPI: dpi})
			}
		}
	}

	mplMetrics := runMatplotlibTextMetrics(t, r.fontPath, cases)
	if len(mplMetrics) != len(cases) {
		t.Fatalf("matplotlib returned %d metrics, want %d", len(mplMetrics), len(cases))
	}

	for i, tc := range cases {
		tc := tc
		mpl := mplMetrics[i]
		t.Run(tc.name(), func(t *testing.T) {
			r.SetResolution(tc.DPI)
			goMetrics := r.MeasureText(tc.Text, tc.Size, "")
			goBounds, ok := r.MeasureTextBounds(tc.Text, tc.Size, "")
			if !ok {
				t.Fatalf("MeasureTextBounds(%q) failed", tc.Text)
			}

			goDescent := math.Max(0, goBounds.Y+goBounds.H)
			assertClose(t, "advance", goMetrics.W, mpl.Width, 1.0)
			assertClose(t, "ink height", goBounds.H, mpl.Height, 1.5)
			assertClose(t, "ink descent", goDescent, mpl.Descent, 1.5)
		})
	}
}

func TestMeasureTextUsesStableFontLineMetrics(t *testing.T) {
	r := mustNew(t, 200, 100)

	caps := r.MeasureText("H", 24, "")
	descender := r.MeasureText("g", 24, "")
	fontHeights, ok := r.MeasureFontHeights(24, "")
	if !ok {
		t.Fatal("expected renderer to expose font height metrics")
	}

	if caps.Ascent <= 0 || descender.Ascent <= 0 {
		t.Fatalf("expected positive ascent, got caps=%+v descender=%+v", caps, descender)
	}
	if caps.Ascent != descender.Ascent || caps.Descent != descender.Descent {
		t.Fatalf("expected font line metrics to stay stable across strings, got caps=%+v descender=%+v", caps, descender)
	}
	if math.Abs(caps.Ascent-fontHeights.Ascent) > 1e-6 || math.Abs(caps.Descent-fontHeights.Descent) > 1e-6 {
		t.Fatalf("MeasureText should use font height metrics, got caps=%+v font=%+v", caps, fontHeights)
	}
	if fontHeights.LineGap < 0 {
		t.Fatalf("unexpected negative line gap: %+v", fontHeights)
	}
}

type mplTextMetricCase struct {
	Text string  `json:"text"`
	Size float64 `json:"size"`
	DPI  uint    `json:"dpi"`
}

func (c mplTextMetricCase) name() string {
	return c.Text + "_" + strings.TrimRight(strings.TrimRight(jsonFloat(c.Size), "0"), ".") + "pt_" + jsonUint(c.DPI) + "dpi"
}

type mplTextMetric struct {
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Descent float64 `json:"descent"`
}

func runMatplotlibTextMetrics(t *testing.T, fontPath string, cases []mplTextMetricCase) []mplTextMetric {
	t.Helper()

	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 unavailable; skipping Matplotlib RendererAgg text metric parity test")
	}

	payload, err := json.Marshal(cases)
	if err != nil {
		t.Fatalf("marshal metric cases: %v", err)
	}

	script := `
import json
import sys

try:
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib as mpl
    from matplotlib.backends.backend_agg import FigureCanvasAgg
    from matplotlib.figure import Figure
    from matplotlib.font_manager import FontProperties
except Exception as exc:
    print(f"matplotlib import failed: {exc}", file=sys.stderr)
    sys.exit(78)

font_path = sys.argv[1]
cases = json.loads(sys.argv[2])
out = []
with mpl.rc_context({"text.hinting": "no_hinting"}):
    for case in cases:
        fig = Figure(figsize=(2, 2), dpi=case["dpi"])
        canvas = FigureCanvasAgg(fig)
        renderer = canvas.get_renderer()
        prop = FontProperties(fname=font_path, size=case["size"])
        width, height, descent = renderer.get_text_width_height_descent(case["text"], prop, False)
        out.append({"width": width, "height": height, "descent": descent})
print(json.dumps(out))
`

	cmd := exec.Command(python, "-c", script, fontPath, string(payload))
	cmd.Env = append(os.Environ(), "MPLCONFIGDIR="+t.TempDir())
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 78 {
			t.Skip(strings.TrimSpace(string(exitErr.Stderr)))
		}
		t.Fatalf("run Matplotlib RendererAgg text metric helper: %v", err)
	}

	var metrics []mplTextMetric
	if err := json.Unmarshal(out, &metrics); err != nil {
		t.Fatalf("decode Matplotlib text metrics: %v\n%s", err, out)
	}
	return metrics
}

func assertClose(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s mismatch: got=%v want=%v tolerance=%v", name, got, want, tol)
	}
}

func jsonFloat(v float64) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func jsonUint(v uint) string {
	data, _ := json.Marshal(v)
	return string(data)
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
