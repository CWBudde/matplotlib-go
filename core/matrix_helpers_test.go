package core

import (
	"testing"

	"matplotlib-go/internal/geom"
)

func TestAxesMatShowConfiguresMatrixView(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())

	img := ax.MatShow([][]float64{
		{1, 2, 3},
		{4, 5, 6},
	})
	if img == nil {
		t.Fatal("MatShow() returned nil")
	}
	if img.Origin != ImageOriginUpper {
		t.Fatalf("image origin = %v, want %v", img.Origin, ImageOriginUpper)
	}
	if img.XMin != -0.5 || img.XMax != 2.5 || img.YMin != -0.5 || img.YMax != 1.5 {
		t.Fatalf("image extent = [%v,%v]x[%v,%v], want [-0.5,2.5]x[-0.5,1.5]", img.XMin, img.XMax, img.YMin, img.YMax)
	}
	if !ax.YInverted() {
		t.Fatal("MatShow() should invert the y-axis")
	}
	if ax.XAxis == nil || ax.XAxis.ShowTicks || ax.XAxis.ShowLabels {
		t.Fatal("MatShow() should hide bottom x ticks and labels")
	}
	if ax.XAxisTop == nil || !ax.XAxisTop.ShowTicks || !ax.XAxisTop.ShowLabels {
		t.Fatal("MatShow() should show top x ticks and labels")
	}
	if ax.xLabelSide != AxisBottom {
		t.Fatalf("x label side = %v, want AxisBottom", ax.xLabelSide)
	}
	if _, ok := ax.XAxis.Locator.(MaxNLocator); !ok {
		t.Fatalf("x locator = %T, want MaxNLocator", ax.XAxis.Locator)
	}
	if _, ok := ax.XAxisTop.Locator.(MaxNLocator); !ok {
		t.Fatalf("top x locator = %T, want MaxNLocator", ax.XAxisTop.Locator)
	}
	if _, ok := ax.YAxis.Locator.(MaxNLocator); !ok {
		t.Fatalf("y locator = %T, want MaxNLocator", ax.YAxis.Locator)
	}
}

func TestAxesSpySupportsMarkerAndImageModes(t *testing.T) {
	data := [][]float64{
		{0, 1, 0},
		{2, 0, 0},
		{0, 0, 3},
	}

	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	result := ax.Spy(data, SpyOptions{Precision: 0.5})
	if result == nil {
		t.Fatal("Spy() returned nil for default image mode")
	}
	if result.Image == nil {
		t.Fatal("default Spy() should create an Image2D like matplotlib")
	}
	if result.Markers != nil {
		t.Fatal("default Spy() should not create marker collection")
	}
	if got := len(result.Indices); got != 3 {
		t.Fatalf("len(indices) = %d, want 3", got)
	}
	if !ax.YInverted() {
		t.Fatal("Spy() should invert the y-axis")
	}
	if ax.XAxisTop == nil || !ax.XAxisTop.ShowTicks || !ax.XAxisTop.ShowLabels {
		t.Fatal("Spy() should show matrix-style top x ticks and labels")
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	useImage := false
	result = ax.Spy(data, SpyOptions{UseImage: &useImage})
	if result == nil {
		t.Fatal("Spy() returned nil for marker mode")
	}
	if result.Markers == nil {
		t.Fatal("marker mode should create a PathCollection")
	}
	if result.Image != nil {
		t.Fatal("marker mode should not create an Image2D")
	}
}

func TestAxesSpyMarkerSizeUsesMatplotlibPointDiameter(t *testing.T) {
	data := [][]float64{{1}}

	fig := NewFigure(400, 300)
	small := fig.AddAxes(unitRect()).Spy(data, SpyOptions{MarkerSize: 5})
	large := fig.AddAxes(unitRect()).Spy(data, SpyOptions{MarkerSize: 10})

	if small == nil || small.Markers == nil || large == nil || large.Markers == nil {
		t.Fatal("Spy(marker size) should use marker mode")
	}
	if !(large.Markers.Size > small.Markers.Size) {
		t.Fatalf("larger MarkerSize should increase rendered marker scale, got small=%v large=%v", small.Markers.Size, large.Markers.Size)
	}
	if got, want := large.Markers.Size/small.Markers.Size, 2.0; !almostEqualFloat(got, want) {
		t.Fatalf("marker scale ratio = %v, want %v", got, want)
	}
}

func TestAxesAnnotatedHeatmapAddsLabels(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	threshold := 2.5

	result := ax.AnnotatedHeatmap([][]float64{
		{1, 2},
		{3, 4},
	}, AnnotatedHeatmapOptions{
		Format:    "%.1f",
		Threshold: &threshold,
	})
	if result == nil {
		t.Fatal("AnnotatedHeatmap() returned nil")
	}
	if result.Image == nil {
		t.Fatal("AnnotatedHeatmap() should create an image")
	}
	if got := len(result.Labels); got != 4 {
		t.Fatalf("label count = %d, want 4", got)
	}
	if result.Labels[0].Content != "1.0" {
		t.Fatalf("first label text = %q, want %q", result.Labels[0].Content, "1.0")
	}
	if result.Labels[0].Color == result.Labels[3].Color {
		t.Fatal("expected low and high cells to use different text colors")
	}
}

func unitRect() geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	}
}
