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
	if _, ok := ax.XAxis.Locator.(FixedLocator); !ok {
		t.Fatalf("x locator = %T, want FixedLocator", ax.XAxis.Locator)
	}
	if _, ok := ax.YAxis.Locator.(FixedLocator); !ok {
		t.Fatalf("y locator = %T, want FixedLocator", ax.YAxis.Locator)
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
		t.Fatal("Spy() returned nil for marker mode")
	}
	if result.Markers == nil {
		t.Fatal("marker mode should create a PathCollection")
	}
	if got := len(result.Indices); got != 3 {
		t.Fatalf("len(indices) = %d, want 3", got)
	}
	if !ax.YInverted() {
		t.Fatal("Spy() should invert the y-axis")
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	useImage := true
	result = ax.Spy(data, SpyOptions{UseImage: &useImage})
	if result == nil {
		t.Fatal("Spy() returned nil for image mode")
	}
	if result.Image == nil {
		t.Fatal("image mode should create an Image2D")
	}
	if result.Markers != nil {
		t.Fatal("image mode should not create marker collection")
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
