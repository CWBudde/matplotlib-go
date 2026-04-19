package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/transform"
)

func TestAxis_Draw(t *testing.T) {
	// Test drawing a basic X axis
	axis := NewXAxis()

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxis_YAxis(t *testing.T) {
	// Test drawing a basic Y axis
	axis := NewYAxis()

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxis_CustomSettings(t *testing.T) {
	// Test axis with custom settings
	axis := &Axis{
		Side:       AxisTop,
		Locator:    LinearLocator{},
		Formatter:  ScalarFormatter{Prec: 2},
		Color:      render.Color{R: 1, G: 0, B: 0, A: 1}, // red
		LineWidth:  2.0,
		TickSize:   10.0,
		ShowSpine:  true,
		ShowTicks:  true,
		ShowLabels: false,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxis_DisabledComponents(t *testing.T) {
	// Test axis with components disabled
	axis := NewXAxis()
	axis.ShowSpine = false
	axis.ShowTicks = false
	axis.ShowLabels = false

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic even with everything disabled
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxes_SetLimits(t *testing.T) {
	// Test the convenience methods for setting limits
	axes := &Axes{
		XScale: nil,
		YScale: nil,
		XAxis:  NewXAxis(),
		YAxis:  NewYAxis(),
	}

	// Test SetXLim
	axes.SetXLim(-5, 10)
	xMin, xMax := axes.XScale.Domain()
	if xMin != -5 || xMax != 10 {
		t.Errorf("SetXLim failed: expected (-5, 10), got (%v, %v)", xMin, xMax)
	}

	// Test SetYLim
	axes.SetYLim(0, 100)
	yMin, yMax := axes.YScale.Domain()
	if yMin != 0 || yMax != 100 {
		t.Errorf("SetYLim failed: expected (0, 100), got (%v, %v)", yMin, yMax)
	}

	// Test SetXLimLog
	axes.SetXLimLog(1, 1000, 10)
	xMin, xMax = axes.XScale.Domain()
	if xMin != 1 || xMax != 1000 {
		t.Errorf("SetXLimLog failed: expected (1, 1000), got (%v, %v)", xMin, xMax)
	}

	// Check that locator was updated to logarithmic
	if logLoc, ok := axes.XAxis.Locator.(LogLocator); !ok || logLoc.Base != 10 {
		t.Errorf("SetXLimLog should update locator to LogLocator with base 10")
	}
}

func TestGrid_Draw(t *testing.T) {
	// Test grid drawing
	grid := NewGrid(AxisBottom)

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	grid.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestGrid_Disabled(t *testing.T) {
	// Test grid with major disabled
	grid := NewGrid(AxisLeft)
	grid.Major = false

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not draw anything
	grid.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestGrid_MinorDraw(t *testing.T) {
	grid := NewGrid(AxisBottom)
	grid.Minor = true
	grid.MinorDashes = []float64{2, 3}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	_ = renderer.Begin(geom.Rect{})
	// Should not panic with minor grid enabled
	grid.Draw(renderer, ctx)
	_ = renderer.End()
}

func TestGrid_MinorOnlyDraw(t *testing.T) {
	grid := NewGrid(AxisLeft)
	grid.Major = false
	grid.Minor = true

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	_ = renderer.Begin(geom.Rect{})
	grid.Draw(renderer, ctx)
	_ = renderer.End()
}

func TestGrid_CustomLocator(t *testing.T) {
	grid := NewGrid(AxisBottom)
	grid.Locator = LogLocator{Base: 10}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	_ = renderer.Begin(geom.Rect{})
	grid.Draw(renderer, ctx)
	_ = renderer.End()
}

func TestAxis_DrawTickLabels_UsesStepPrecisionForScalarFormatter(t *testing.T) {
	axis := NewYAxis()
	ctx := createTestDrawContext()
	ctx.DataToPixel.YScale = transform.NewLinear(0, 0.196)

	var r axisLabelRecordingRenderer
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	want := []string{"0.000", "0.025", "0.050", "0.075", "0.100", "0.125", "0.150", "0.175", "0.200"}
	if len(r.texts) != len(want) {
		t.Fatalf("unexpected tick label count: got %v want %v", r.texts, want)
	}
	for i := range want {
		if r.texts[i] != want[i] {
			t.Fatalf("tick label %d mismatch: got %q want %q", i, r.texts[i], want[i])
		}
	}
}

type axisLabelRecordingRenderer struct {
	render.NullRenderer
	texts []string
}

func (r *axisLabelRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *axisLabelRecordingRenderer) DrawText(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
}

func TestAxes_AddGrid(t *testing.T) {
	// Test adding grids to axes
	axes := &Axes{
		Artists: []Artist{},
		XAxis:   NewXAxis(),
		YAxis:   NewYAxis(),
	}

	initialCount := len(axes.Artists)

	// Add X grid
	xGrid := axes.AddXGrid()
	if len(axes.Artists) != initialCount+1 {
		t.Errorf("AddXGrid should add one artist, got %d artists", len(axes.Artists))
	}
	if xGrid.Axis != AxisBottom {
		t.Errorf("AddXGrid should create grid for AxisBottom, got %v", xGrid.Axis)
	}

	// Add Y grid
	yGrid := axes.AddYGrid()
	if len(axes.Artists) != initialCount+2 {
		t.Errorf("AddYGrid should add second artist, got %d artists", len(axes.Artists))
	}
	if yGrid.Axis != AxisLeft {
		t.Errorf("AddYGrid should create grid for AxisLeft, got %v", yGrid.Axis)
	}
}
