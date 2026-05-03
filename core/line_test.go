package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

type recordingRenderer struct {
	render.NullRenderer
	pathCalls []recordedPathCall
}

type recordedPathCall struct {
	path  geom.Path
	paint render.Paint
}

func (r *recordingRenderer) Path(p geom.Path, paint *render.Paint) {
	if paint == nil {
		r.pathCalls = append(r.pathCalls, recordedPathCall{path: p})
		return
	}
	r.pathCalls = append(r.pathCalls, recordedPathCall{
		path:  p,
		paint: *paint,
	})
}

func TestLine2D_EmptyData(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{}, // empty data
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	// Should not panic with empty data
	var r render.NullRenderer
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	// This should not panic
	line.Draw(&r, ctx)
}

func TestLine2D_DefaultsToButtCaps(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 10),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	line.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one Path call, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.LineCap != render.CapButt {
		t.Fatalf("expected default line cap %v, got %v", render.CapButt, r.pathCalls[0].paint.LineCap)
	}
	if r.pathCalls[0].paint.LineJoin != render.JoinRound {
		t.Fatalf("expected default line join %v, got %v", render.JoinRound, r.pathCalls[0].paint.LineJoin)
	}
	if r.pathCalls[0].paint.Snap != render.SnapAuto {
		t.Fatalf("expected default line snap %v, got %v", render.SnapAuto, r.pathCalls[0].paint.Snap)
	}
}

func TestLine2D_SingletonData(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{{X: 5, Y: 0.5}}, // single point
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	// Should not panic with singleton data
	var r render.NullRenderer
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	// This should not panic
	line.Draw(&r, ctx)
}

func TestLine2D_BasicFunctionality(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0.2},
			{X: 3, Y: 0.9},
			{X: 6, Y: 0.4},
			{X: 10, Y: 0.8},
		},
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 0, A: 1},
		z:   1.0,
	}

	// Test Z() method
	if line.Z() != 1.0 {
		t.Errorf("Expected Z() = 1.0, got %f", line.Z())
	}

	// Test Bounds() returns the data bounding box of the line
	bounds := line.Bounds(nil)
	if bounds.Min.X != 0 || bounds.Min.Y != 0 || bounds.Max.X != 10 || bounds.Max.Y != 0.9 {
		t.Errorf("Unexpected bounds, got %+v", bounds)
	}

	// Test Draw() method doesn't panic
	var r render.NullRenderer
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	// This should not panic
	line.Draw(&r, ctx)
}

func TestLine2D_AsArtist(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		W:   1.0,
		Col: render.Color{R: 1, G: 1, B: 1, A: 1},
	}

	// Test that Line2D implements Artist interface
	var _ Artist = line

	// Test integration with Axes
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.15}, Max: geom.Pt{X: 0.95, Y: 0.9}})
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 1)
	ax.Add(line)

	// Test that the figure can be drawn without panic
	var r render.NullRenderer
	DrawFigure(fig, &r)
}

func TestLine2D_Bounds(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{{X: 2, Y: -1}, {X: 5, Y: 3}, {X: 8, Y: 0}},
	}
	b := line.Bounds(nil)
	if b.Min.X != 2 || b.Min.Y != -1 || b.Max.X != 8 || b.Max.Y != 3 {
		t.Errorf("unexpected bounds: %v", b)
	}
}

func TestLine2D_BoundsEmpty(t *testing.T) {
	line := &Line2D{}
	b := line.Bounds(nil)
	if b.W() != 0 || b.H() != 0 {
		t.Errorf("empty line should have zero bounds: %v", b)
	}
}

func TestAutoScale(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Add(&Line2D{XY: []geom.Pt{{X: 2, Y: -1}, {X: 8, Y: 5}}})
	ax.Add(&Line2D{XY: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 3}}})

	ax.AutoScale(0)
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()

	if xMin != 0 || xMax != 10 {
		t.Errorf("x limits: got [%v, %v], want [0, 10]", xMin, xMax)
	}
	if yMin != -1 || yMax != 5 {
		t.Errorf("y limits: got [%v, %v], want [-1, 5]", yMin, yMax)
	}
}

func TestAutoScaleWithMargin(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Add(&Line2D{XY: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 10}}})

	ax.AutoScale(0.1) // 10% margin
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()

	// 10% of span=10 is 1, so limits should be [-1, 11]
	if xMin != -1 || xMax != 11 {
		t.Errorf("x limits: got [%v, %v], want [-1, 11]", xMin, xMax)
	}
	if yMin != -1 || yMax != 11 {
		t.Errorf("y limits: got [%v, %v], want [-1, 11]", yMin, yMax)
	}
}

func TestAutoScaleNoData(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.XScale = transform.NewLinear(0, 1)
	ax.YScale = transform.NewLinear(0, 1)

	// AutoScale with no artists should not change limits
	ax.AutoScale(0.05)
	xMin, xMax := ax.XScale.Domain()
	if xMin != 0 || xMax != 1 {
		t.Errorf("limits should be unchanged with no data: got [%v, %v]", xMin, xMax)
	}
}

func TestPlotAutoScalesWithDefaultMargin(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Plot([]float64{0, 10}, []float64{-1, 3})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if !floatApprox(xMin, -0.5, 1e-12) || !floatApprox(xMax, 10.5, 1e-12) {
		t.Fatalf("x limits = [%v, %v], want [-0.5, 10.5]", xMin, xMax)
	}
	if !floatApprox(yMin, -1.2, 1e-12) || !floatApprox(yMax, 3.2, 1e-12) {
		t.Fatalf("y limits = [%v, %v], want [-1.2, 3.2]", yMin, yMax)
	}
}

func TestPlotAutoScaleExpandsAcrossLines(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Plot([]float64{0, 1}, []float64{0, 1})
	ax.Plot([]float64{-2, 4}, []float64{-3, 5})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if !floatApprox(xMin, -2.3, 1e-12) || !floatApprox(xMax, 4.3, 1e-12) {
		t.Fatalf("x limits = [%v, %v], want [-2.3, 4.3]", xMin, xMax)
	}
	if !floatApprox(yMin, -3.4, 1e-12) || !floatApprox(yMax, 5.4, 1e-12) {
		t.Fatalf("y limits = [%v, %v], want [-3.4, 5.4]", yMin, yMax)
	}
}

func TestPlotAutoScalePreservesExplicitLimits(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.SetXLim(0, 10)
	ax.SetYLim(-2, 2)
	ax.Plot([]float64{-100, 100}, []float64{-50, 50})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if xMin != 0 || xMax != 10 {
		t.Fatalf("x limits = [%v, %v], want [0, 10]", xMin, xMax)
	}
	if yMin != -2 || yMax != 2 {
		t.Fatalf("y limits = [%v, %v], want [-2, 2]", yMin, yMax)
	}
}

func TestPlotAutoScalePreservesOnlyExplicitAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.SetXLim(0, 10)
	ax.Plot([]float64{-100, 100}, []float64{-50, 50})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if xMin != 0 || xMax != 10 {
		t.Fatalf("x limits = [%v, %v], want [0, 10]", xMin, xMax)
	}
	if yMin != -55 || yMax != 55 {
		t.Fatalf("y limits = [%v, %v], want [-55, 55]", yMin, yMax)
	}
}
