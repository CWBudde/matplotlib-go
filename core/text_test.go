package core

import (
	"math"
	"strings"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestNormalizeDisplayText_ReplacesBasicMathTokens(t *testing.T) {
	got := normalizeDisplayText(`\\mu = 1.2, \\Delta x \\rightarrow \\pi`)
	want := "μ = 1.2, Δ x → π"
	if got != want {
		t.Fatalf("unexpected normalized text: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_ParsesInlineMath(t *testing.T) {
	got := normalizeDisplayText(`signal $\\alpha^2 + \\beta_i$ peak`)
	want := "signal α² + βᵢ peak"
	if got != want {
		t.Fatalf("unexpected inline math normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_FormatsFractionsAndRoots(t *testing.T) {
	got := normalizeDisplayText(`$\\frac{1}{\\sqrt{2}}$`)
	want := "1⁄√2"
	if got != want {
		t.Fatalf("unexpected fraction/root normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_HandlesGroupedScripts(t *testing.T) {
	got := normalizeDisplayText(`$x_{\\mathrm{max}}$`)
	want := "xₘₐₓ"
	if got != want {
		t.Fatalf("unexpected grouped subscript normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_HandlesAccentsAndOperators(t *testing.T) {
	got := normalizeDisplayText(`$\\hat{x} + \\sin(\\theta) \\leq \\overline{AB}$`)
	want := "x̂ + sin(θ) ≤ A̅B̅"
	if got != want {
		t.Fatalf("unexpected accent/operator normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_PreservesUnmatchedDollar(t *testing.T) {
	got := normalizeDisplayText(`cost is $5`)
	want := "cost is $5"
	if got != want {
		t.Fatalf("unexpected unmatched dollar normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_IgnoresLimitModifiers(t *testing.T) {
	got := normalizeDisplayText(`$\\displaystyle \\sum\\limits_{i=1}^n$`)
	want := "∑ᵢ₌₁ⁿ"
	if got != want {
		t.Fatalf("unexpected limit-modifier normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_HandlesMatrixEnvironment(t *testing.T) {
	got := normalizeDisplayText(`$\\begin{pmatrix} a & b \\\\ c & d \\end{pmatrix}$`)
	want := "(a b; c d)"
	if got != want {
		t.Fatalf("unexpected matrix normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_HandlesMiddleDelimiter(t *testing.T) {
	got := normalizeDisplayText(`$\\left\\langle{a}\\middle|b\\right\\rangle$`)
	want := "⟨a|b⟩"
	if got != want {
		t.Fatalf("unexpected middle-delimiter normalization: got %q want %q", got, want)
	}
}

func TestLayoutMathTextScripts(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `x_i^2`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if layout.Width <= 0 || layout.Ascent <= 0 || layout.Descent <= 0 || layout.Height != layout.Ascent+layout.Descent {
		t.Fatalf("invalid layout metrics: %+v", layout)
	}
	if len(layout.Rules) != 0 {
		t.Fatalf("unexpected rules in script-only layout: %+v", layout.Rules)
	}
	if !containsMathRun(layout.Runs, "x", 20) || !containsMathRun(layout.Runs, "i", 14) || !containsMathRun(layout.Runs, "2", 14) {
		t.Fatalf("missing expected script runs: %+v", layout.Runs)
	}

	var subY, superY float64
	for _, run := range layout.Runs {
		switch run.Text {
		case "i":
			subY = run.Offset.Y
		case "2":
			superY = run.Offset.Y
		}
	}
	if subY <= 0 || superY >= 0 {
		t.Fatalf("script baselines not shifted as expected: sub=%v super=%v runs=%+v", subY, superY, layout.Runs)
	}
}

func TestLayoutMathTextFractionAddsRule(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\\frac{1}{2}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(layout.Rules) != 1 {
		t.Fatalf("expected one fraction rule, got %+v", layout.Rules)
	}
	if layout.Rules[0].Rect.Max.X <= layout.Rules[0].Rect.Min.X {
		t.Fatalf("unexpected fraction rule rect: %+v", layout.Rules[0].Rect)
	}
}

func TestLayoutMathTextSqrtHasVinculum(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\\sqrt[3]{x + 1}`, 18, "")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(layout.Rules) != 1 {
		t.Fatalf("expected sqrt rule, got %+v", layout.Rules)
	}
	if !containsMathRun(layout.Runs, "√", 18) || !containsMathRun(layout.Runs, "3", 9.9) {
		t.Fatalf("missing sqrt/index runs: %+v", layout.Runs)
	}
	if layout.Rules[0].Rect.Min.X <= 0 || layout.Rules[0].Rect.Max.X <= layout.Rules[0].Rect.Min.X {
		t.Fatalf("unexpected sqrt rule rect: %+v", layout.Rules[0].Rect)
	}
}

func TestLayoutMathTextStacksLargeOperatorLimits(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\\sum\\limits_{i=1}^n`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if !containsMathRun(layout.Runs, "∑", 24) || !containsMathRun(layout.Runs, "i=1", 14) || !containsMathRun(layout.Runs, "n", 14) {
		t.Fatalf("missing expected limit runs: %+v", layout.Runs)
	}

	sumW := r.MeasureText("∑", 24, "DejaVu Sans").W
	subW := r.MeasureText("i=1", 14, "DejaVu Sans").W
	superW := r.MeasureText("n", 14, "DejaVu Sans").W

	var sumX, subX, superX, subY, superY float64
	for _, run := range layout.Runs {
		switch run.Text {
		case "∑":
			sumX = run.Offset.X
		case "i=1":
			subX = run.Offset.X
			subY = run.Offset.Y
		case "n":
			superX = run.Offset.X
			superY = run.Offset.Y
		}
	}

	if subY <= 0 || superY >= 0 {
		t.Fatalf("large-operator limits not stacked vertically: sub=%v super=%v runs=%+v", subY, superY, layout.Runs)
	}
	sumCenter := sumX + sumW/2
	subCenter := subX + subW/2
	superCenter := superX + superW/2
	if math.Abs(subCenter-sumCenter) > 0.01 || math.Abs(superCenter-sumCenter) > 0.01 {
		t.Fatalf("large-operator limits not centered over operator: sumCenter=%v subCenter=%v superCenter=%v runs=%+v", sumCenter, subCenter, superCenter, layout.Runs)
	}
}

func TestLayoutMathTextSupportsFencedDelimiters(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\left(\frac{1}{2}\right)`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	var leftSize, rightSize float64
	for _, run := range layout.Runs {
		switch run.Text {
		case "(":
			leftSize = run.FontSize
		case ")":
			rightSize = run.FontSize
		}
	}
	if leftSize <= 20 || rightSize <= 20 {
		t.Fatalf("expected stretched delimiters larger than base size: left=%v right=%v runs=%+v", leftSize, rightSize, layout.Runs)
	}
}

func TestLayoutMathTextSupportsMiddleDelimiters(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\left\langle \frac{1}{2} \middle| x \right\rangle`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	var leftSize, middleSize, rightSize float64
	var leftX, middleX, rightX float64
	for _, run := range layout.Runs {
		switch run.Text {
		case "⟨":
			leftSize = run.FontSize
			leftX = run.Offset.X
		case "|":
			middleSize = run.FontSize
			middleX = run.Offset.X
		case "⟩":
			rightSize = run.FontSize
			rightX = run.Offset.X
		}
	}

	if leftSize <= 20 || middleSize <= 20 || rightSize <= 20 {
		t.Fatalf("expected stretched fence delimiters larger than base size: left=%v middle=%v right=%v runs=%+v", leftSize, middleSize, rightSize, layout.Runs)
	}
	if leftX >= middleX || middleX >= rightX {
		t.Fatalf("expected middle delimiter to be between outer delimiters: left=%v middle=%v right=%v runs=%+v", leftX, middleX, rightX, layout.Runs)
	}
}

func TestLayoutMathTextSupportsOmittedFenceDelimiters(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\left. x \right|`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	for _, run := range layout.Runs {
		if run.Text == "." {
			t.Fatalf("null delimiter should not render as a visible glyph: %+v", layout.Runs)
		}
	}

	var sawX, sawBar bool
	for _, run := range layout.Runs {
		switch strings.TrimSpace(run.Text) {
		case "x":
			sawX = true
		case "|":
			sawBar = true
		}
	}
	if !sawX || !sawBar {
		t.Fatalf("missing expected fence runs: %+v", layout.Runs)
	}
}

func TestLayoutMathTextSupportsStyleSwitches(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\mathrm{r} + \mathsf{s} + \mathtt{t}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	var romanKey, sansKey, monoKey string
	for _, run := range layout.Runs {
		switch run.Text {
		case "r":
			romanKey = strings.ToLower(run.FontKey)
		case "s":
			sansKey = strings.ToLower(run.FontKey)
		case "t":
			monoKey = strings.ToLower(run.FontKey)
		}
	}

	if romanKey == "" || sansKey == "" || monoKey == "" {
		t.Fatalf("missing styled run font keys: %+v", layout.Runs)
	}
	if !strings.Contains(romanKey, "serif") {
		t.Fatalf("roman style did not resolve serif font key: %q", romanKey)
	}
	if !strings.Contains(sansKey, "sans") {
		t.Fatalf("sans style did not resolve sans font key: %q", sansKey)
	}
	if !strings.Contains(monoKey, "mono") {
		t.Fatalf("monospace style did not resolve mono font key: %q", monoKey)
	}
}

func TestLayoutMathTextSupportsSpacingCommands(t *testing.T) {
	var r textRecordingRenderer
	compact, ok := LayoutMathText(&r, `ab`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("compact LayoutMathText returned !ok")
	}
	wide, ok := LayoutMathText(&r, `a\quad b`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("wide LayoutMathText returned !ok")
	}
	if wide.Width <= compact.Width+10 {
		t.Fatalf("spacing command did not widen layout enough: compact=%v wide=%v", compact.Width, wide.Width)
	}
}

func TestLayoutMathTextSupportsMatrixEnvironments(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\begin{pmatrix} a & b \\\\ c & d \end{pmatrix}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	var leftX, rightX, firstColTopY, firstColBottomY, secondColTopX float64
	var sawLeft, sawRight, sawA, sawB, sawC, sawD bool
	for _, run := range layout.Runs {
		text := strings.TrimSpace(run.Text)
		switch text {
		case "(":
			leftX = run.Offset.X
			sawLeft = true
		case ")":
			rightX = run.Offset.X
			sawRight = true
		case "a":
			firstColTopY = run.Offset.Y
			sawA = true
		case "b":
			if !sawA {
				t.Fatalf("expected a to be laid out before b: %+v", layout.Runs)
			}
			secondColTopX = run.Offset.X
			sawB = true
		case "c":
			firstColBottomY = run.Offset.Y
			sawC = true
		case "d":
			sawD = true
		}
	}

	if !sawLeft || !sawRight || !sawA || !sawB || !sawC || !sawD {
		t.Fatalf("missing expected matrix runs: %+v", layout.Runs)
	}
	if rightX <= leftX {
		t.Fatalf("expected right delimiter after left delimiter: left=%v right=%v runs=%+v", leftX, rightX, layout.Runs)
	}
	if secondColTopX <= leftX {
		t.Fatalf("expected second matrix column to be offset to the right: left=%v secondCol=%v runs=%+v", leftX, secondColTopX, layout.Runs)
	}
	if firstColBottomY <= firstColTopY {
		t.Fatalf("expected second matrix row below first row: top=%v bottom=%v runs=%+v", firstColTopY, firstColBottomY, layout.Runs)
	}
}

func TestLayoutMathTextSupportsArrayEnvironments(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\begin{array}{cc} a & b \\\\ c & d \end{array}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	for _, run := range layout.Runs {
		if run.Text == "(" || run.Text == ")" || run.Text == "[" || run.Text == "]" {
			t.Fatalf("array environment should not add fences: %+v", layout.Runs)
		}
	}

	var sawA, sawD bool
	for _, run := range layout.Runs {
		switch strings.TrimSpace(run.Text) {
		case "a":
			sawA = true
		case "d":
			sawD = true
		}
	}
	if !sawA || !sawD {
		t.Fatalf("missing expected array runs: %+v", layout.Runs)
	}
}

func TestLayoutDisplayTextMixedInlineMath(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := layoutDisplayText(&r, `phase $\\frac{1}{2}$ peak`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("layoutDisplayText returned !ok")
	}
	if layout.Width <= 0 || layout.Ascent <= 0 || layout.Descent <= 0 {
		t.Fatalf("invalid layout metrics: %+v", layout)
	}
	if len(layout.Rules) != 1 {
		t.Fatalf("expected one fraction rule, got %+v", layout.Rules)
	}
	if !containsMathRun(layout.Runs, "phase ", 20) || !containsMathRun(layout.Runs, "1", 15) || !containsMathRun(layout.Runs, "2", 15) || !containsMathRun(layout.Runs, " peak", 20) {
		t.Fatalf("missing expected mixed inline runs: %+v", layout.Runs)
	}
}

func TestAlignedTextOrigin(t *testing.T) {
	anchor := geom.Pt{X: 100, Y: 50}
	metrics := render.TextMetrics{W: 40, Ascent: 8, Descent: 2}

	got := alignedTextOrigin(anchor, metrics, TextAlignCenter, TextVAlignTop)
	if got.X != 80 || got.Y != 58 {
		t.Fatalf("unexpected text origin: %+v", got)
	}
}

func TestAxesTextDrawsNormalizedContent(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.5, 0.5, `\\alpha + \\beta`, TextOptions{
		HAlign:   TextAlignCenter,
		VAlign:   TextVAlignMiddle,
		FontSize: 12,
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 1 {
		t.Fatalf("expected exactly one text draw, got %d (%v)", len(r.texts), r.texts)
	}
	if r.texts[0] != "α + β" {
		t.Fatalf("unexpected rendered text %q", r.texts[0])
	}
}

func TestAxesTextDrawsFullMathLayoutRuns(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.5, 0.5, `$\\frac{1}{2}$`, TextOptions{
		HAlign:   TextAlignCenter,
		VAlign:   TextVAlignMiddle,
		FontSize: 12,
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if containsTextString(r.texts, "1⁄2") {
		t.Fatalf("full math expression fell back to normalized text draw: %v", r.texts)
	}
	if !containsTextString(r.texts, "1") || !containsTextString(r.texts, "2") {
		t.Fatalf("expected structured math runs for fraction, got %v", r.texts)
	}
	if r.pathCount == 0 {
		t.Fatalf("expected fraction rule path, got %d paths", r.pathCount)
	}
}

func TestAxesTextDrawsMixedInlineMathLayoutRuns(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.5, 0.5, `phase $\\frac{1}{2}$ peak`, TextOptions{
		HAlign:   TextAlignCenter,
		VAlign:   TextVAlignMiddle,
		FontSize: 12,
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if containsTextString(r.texts, "phase 1⁄2 peak") {
		t.Fatalf("mixed inline math fell back to normalized text draw: %v", r.texts)
	}
	if !containsTextString(r.texts, "phase ") || !containsTextString(r.texts, "1") || !containsTextString(r.texts, "2") || !containsTextString(r.texts, " peak") {
		t.Fatalf("expected stitched mixed inline runs, got %v", r.texts)
	}
	if r.pathCount == 0 {
		t.Fatalf("expected fraction rule path, got %d paths", r.pathCount)
	}
}

func TestDrawDisplayTextVerticalFullMathUsesPaths(t *testing.T) {
	var r verticalMathTextRecordingRenderer
	drawDisplayTextVertical(&r, `$\\frac{1}{2}$`, geom.Pt{X: 100, Y: 60}, 12, render.Color{A: 1}, "DejaVu Sans")

	if len(r.verticalTexts) != 0 {
		t.Fatalf("full math expression unexpectedly used DrawTextVertical fallback: %v", r.verticalTexts)
	}
	if !containsTextString(r.textPathCalls, "1") || !containsTextString(r.textPathCalls, "2") {
		t.Fatalf("expected fraction runs to resolve through TextPath, got %v", r.textPathCalls)
	}
	if r.pathCount < 3 {
		t.Fatalf("expected fraction rule plus glyph paths, got %d paths", r.pathCount)
	}
}

func TestAnnotationDrawOverlayRendersArrowAndText(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Annotate("peak", 0.5, 0.5)

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 1 || r.texts[0] != "peak" {
		t.Fatalf("unexpected annotation texts: %v", r.texts)
	}
	if r.pathCount < 2 {
		t.Fatalf("expected annotation arrow line and head, got %d paths", r.pathCount)
	}
}

type textRecordingRenderer struct {
	render.NullRenderer
	pathCount int
	texts     []string
	origins   []geom.Pt
}

func (r *textRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *textRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *textRecordingRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, origin)
}

type verticalMathTextRecordingRenderer struct {
	textRecordingRenderer
	verticalTexts []string
	textPathCalls []string
}

func (r *verticalMathTextRecordingRenderer) DrawTextVertical(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.verticalTexts = append(r.verticalTexts, text)
}

func (r *verticalMathTextRecordingRenderer) TextPath(text string, origin geom.Pt, _ float64, _ string) (geom.Path, bool) {
	r.textPathCalls = append(r.textPathCalls, text)
	return patchRectPath(geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - 4},
		Max: geom.Pt{X: origin.X + 4, Y: origin.Y},
	}), true
}

func TestAxesTextSupportsAxesAndBlendedCoordinates(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.25, 0.75, "axes", TextOptions{
		Coords:  Coords(CoordAxes),
		OffsetX: 5,
		OffsetY: -7,
	})
	ax.Text(0.25, 0.75, "blend", TextOptions{
		Coords: BlendCoords(CoordFigure, CoordAxes),
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 2 {
		t.Fatalf("expected 2 text draws, got %d", len(r.texts))
	}

	wantAxes := geom.Pt{X: 245, Y: 173}
	if r.origins[0] != wantAxes {
		t.Fatalf("axes coords origin = %+v, want %+v", r.origins[0], wantAxes)
	}

	wantBlend := geom.Pt{X: 200, Y: 180}
	if r.origins[1] != wantBlend {
		t.Fatalf("blended coords origin = %+v, want %+v", r.origins[1], wantBlend)
	}
}

func TestAxesLabelsDrawMathTextAccordingToExpressionScope(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle(`$\\alpha^2$`)
	ax.SetXLabel(`phase $\\theta$`)
	ax.SetYLabel(`amp $\\frac{1}{2}$`)

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if containsTextString(r.texts, "α²") {
		t.Fatalf("full math title unexpectedly collapsed to normalized text: %v", r.texts)
	}
	if !containsTextString(r.texts, "α") || !containsTextString(r.texts, "2") {
		t.Fatalf("missing structured title math runs: %v", r.texts)
	}
	if containsTextString(r.texts, "phase θ") {
		t.Fatalf("mixed inline xlabel unexpectedly collapsed to normalized text: %v", r.texts)
	}
	if !containsTextString(r.texts, "phase ") || !containsTextString(r.texts, "θ") {
		t.Fatalf("missing structured xlabel runs: %v", r.texts)
	}
	if containsTextString(r.texts, "amp 1⁄2") {
		t.Fatalf("mixed inline ylabel unexpectedly collapsed to normalized text: %v", r.texts)
	}
	if !containsTextString(r.texts, "amp ") || !containsTextString(r.texts, "1") || !containsTextString(r.texts, "2") {
		t.Fatalf("missing structured ylabel runs: %v", r.texts)
	}
}

func containsTextString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsMathRun(runs []MathTextLayoutRun, text string, size float64) bool {
	for _, run := range runs {
		if run.Text == text && almostEqualFloat(run.FontSize, size) {
			return true
		}
	}
	return false
}

func almostEqualFloat(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 1e-9
}
