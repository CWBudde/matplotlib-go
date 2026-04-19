package core

import (
	"math"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
	"matplotlib-go/transform"
)

func TestRCEffectivePrecedence(t *testing.T) {
	fig := NewFigure(800, 600, style.WithDPI(110))
	axInherit := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	if got := axInherit.effectiveRC(fig).DPI; got != 110 {
		t.Fatalf("expected axes inherit figure DPI, got %v", got)
	}
	axOverride := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.2, Y: 0.2}, Max: geom.Pt{X: 0.8, Y: 0.8}}, style.WithDPI(200))
	if got := axOverride.effectiveRC(fig).DPI; got != 200 {
		t.Fatalf("expected axes override DPI=200, got %v", got)
	}
}

// artist with custom z
type zArtist struct {
	z   float64
	id  int
	hit *[]int
}

func (a zArtist) Draw(_ render.Renderer, _ *DrawContext) { *a.hit = append(*a.hit, a.id) }
func (a zArtist) Z() float64                             { return a.z }
func (a zArtist) Bounds(*DrawContext) geom.Rect          { return geom.Rect{} }

func TestZOrderStableSortAndTraversal(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	var order []int
	// Insertion order: ids 1..5, with equal z for 2 and 3
	ax.Add(zArtist{z: 0, id: 1, hit: &order})
	ax.Add(zArtist{z: 1, id: 2, hit: &order})
	ax.Add(zArtist{z: 1, id: 3, hit: &order})
	ax.Add(zArtist{z: -1, id: 4, hit: &order})
	ax.Add(zArtist{z: 2, id: 5, hit: &order})

	var r render.NullRenderer
	DrawFigure(fig, &r)

	// Expected draw order: z=-1 (id4), z=0 (id1), z=1 (ids 2 then 3), z=2 (id5)
	want := []int{4, 1, 2, 3, 5}
	if len(order) != len(want) {
		t.Fatalf("draw count mismatch: got %d want %d", len(order), len(want))
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order mismatch at %d: got %v want %v (full=%v)", i, order[i], want[i], order)
		}
	}
}

func TestTitleFontSizeUsesTitleOnlyCompensation(t *testing.T) {
	ctx := &DrawContext{RC: style.RC{FontSize: 12}}

	got := titleFontSize(ctx)
	want := 12.0024

	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("titleFontSize() = %v, want %v", got, want)
	}
}

type axesLabelRecordingRenderer struct {
	render.NullRenderer
	bounds         map[string]render.TextBounds
	rotatedText    []string
	rotatedAnchors []geom.Pt
}

func (r *axesLabelRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	switch text {
	case "4":
		return render.TextMetrics{W: 5, H: 10, Ascent: 8, Descent: 2}
	case "Value":
		return render.TextMetrics{W: size * 2, H: 10, Ascent: 8, Descent: 2}
	default:
		return render.TextMetrics{W: float64(len(text)) * size * 0.5, H: size, Ascent: size * 0.8, Descent: size * 0.2}
	}
}

func (r *axesLabelRecordingRenderer) MeasureTextBounds(text string, _ float64, _ string) (render.TextBounds, bool) {
	b, ok := r.bounds[text]
	return b, ok
}

func (r *axesLabelRecordingRenderer) DrawText(_ string, _ geom.Pt, _ float64, _ render.Color) {}

func (r *axesLabelRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, _ float64, _ float64, _ render.Color) {
	r.rotatedText = append(r.rotatedText, text)
	r.rotatedAnchors = append(r.rotatedAnchors, anchor)
}

func TestDrawAxesLabels_YLabelUsesTickBoundsAndLabelPad(t *testing.T) {
	ax := &Axes{
		YAxis:  NewYAxis(),
		YLabel: "Value",
	}
	ax.YAxis.Locator = staticLocator{4}
	ax.YAxis.Formatter = ScalarFormatter{Prec: 0}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72
	px := geom.Rect{
		Min: geom.Pt{X: 50, Y: 350},
		Max: geom.Pt{X: 150, Y: 450},
	}

	r := &axesLabelRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"4": {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer r.End()

	drawAxesLabels(ax, r, ctx, px)

	if len(r.rotatedText) != 1 || r.rotatedText[0] != "Value" {
		t.Fatalf("unexpected rotated text draws: %v", r.rotatedText)
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: getSpinePosition(ax.YAxis.Side, ctx), Y: 4})
	tickLabelMinX := tickPos.X - tickLabelPadPx(ax.YAxis, ctx) - (1 + 5.0) + 1
	want := geom.Pt{
		X: math.Min(spinePixelX(AxisLeft, px), tickLabelMinX) - axisLabelPadPx(ctx),
		Y: px.Min.Y + px.H()/2,
	}
	if r.rotatedAnchors[0] != want {
		t.Fatalf("ylabel anchor = %+v, want %+v", r.rotatedAnchors[0], want)
	}
}

func TestDrawContextTransformsExposeCoordinateSpaces(t *testing.T) {
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(-5, 5),
			DataToAxes:  transform.NewScaleTransform(transform.NewLinear(0, 10), transform.NewLinear(-5, 5)),
			AxesToPixel: transform.NewDisplayRectTransform(geom.Rect{Min: geom.Pt{X: 50, Y: 100}, Max: geom.Pt{X: 250, Y: 300}}),
		},
		Clip:       geom.Rect{Min: geom.Pt{X: 50, Y: 100}, Max: geom.Pt{X: 250, Y: 300}},
		FigureRect: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 400, Y: 500}},
	}

	if got := ctx.TransData().Apply(geom.Pt{X: 2.5, Y: 0}); got != (geom.Pt{X: 100, Y: 200}) {
		t.Fatalf("transData point = %+v, want {100 200}", got)
	}
	if got := ctx.TransAxes().Apply(geom.Pt{X: 0.25, Y: 0.75}); got != (geom.Pt{X: 100, Y: 150}) {
		t.Fatalf("transAxes point = %+v, want {100 150}", got)
	}
	if got := ctx.TransFigure().Apply(geom.Pt{X: 0.25, Y: 0.75}); got != (geom.Pt{X: 100, Y: 125}) {
		t.Fatalf("transFigure point = %+v, want {100 125}", got)
	}
	if got := ctx.TransformFor(BlendCoords(CoordFigure, CoordAxes)).Apply(geom.Pt{X: 0.5, Y: 0.25}); got != (geom.Pt{X: 200, Y: 250}) {
		t.Fatalf("blended transform point = %+v, want {200 250}", got)
	}
}
