package core

import (
	"math"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestAxesPColorMeshAndColorbar(t *testing.T) {
	fig := NewFigure(800, 500)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.8, Y: 0.9},
	})

	edgeWidth := 1.0
	mesh := ax.PColorMesh([][]float64{
		{0, 1},
		{2, 3},
	}, MeshOptions{
		XEdges:    []float64{2, 4, 8},
		YEdges:    []float64{-1, 1, 5},
		EdgeWidth: &edgeWidth,
		Label:     "mesh",
	})
	if mesh == nil {
		t.Fatal("expected quad mesh")
	}
	if len(mesh.FaceColors) != 4 {
		t.Fatalf("expected 4 face colors, got %d", len(mesh.FaceColors))
	}
	mapping := mesh.ScalarMap()
	if mapping.Colormap != "viridis" || mapping.VMin != 0 || mapping.VMax != 3 {
		t.Fatalf("unexpected scalar map %+v", mapping)
	}
	bounds := mesh.Bounds(nil)
	if bounds.Min != (geom.Pt{X: 2, Y: -1}) || bounds.Max != (geom.Pt{X: 8, Y: 5}) {
		t.Fatalf("unexpected bounds %+v", bounds)
	}

	cb := fig.AddColorbar(ax, mesh, ColorbarOptions{Label: "density"})
	if cb == nil {
		t.Fatal("expected colorbar axes for mesh")
	}
	yMin, yMax := cb.YScale.Domain()
	if yMin != 0 || yMax != 3 {
		t.Fatalf("unexpected colorbar limits %v..%v", yMin, yMax)
	}
}

func TestAxesHist2DCounts(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	result := ax.Hist2D(
		[]float64{0.1, 0.2, 0.8, 0.9},
		[]float64{0.1, 0.2, 0.8, 0.9},
		Hist2DOptions{
			XBinEdges: []float64{0, 0.5, 1},
			YBinEdges: []float64{0, 0.5, 1},
			Label:     "hist2d",
		},
	)
	if result == nil || result.Mesh == nil {
		t.Fatal("expected hist2d result mesh")
	}
	if got := result.Counts[0][0]; got != 2 {
		t.Fatalf("lower-left count = %v, want 2", got)
	}
	if got := result.Counts[1][1]; got != 2 {
		t.Fatalf("upper-right count = %v, want 2", got)
	}
	if got := result.Counts[0][1] + result.Counts[1][0]; got != 0 {
		t.Fatalf("unexpected off-diagonal counts %v", got)
	}
}

func TestAxesContourAndContourf(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	grid := [][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}

	contours := ax.Contour(grid, ContourOptions{
		Levels:     []float64{1.5, 2.5},
		LabelLines: true,
		Label:      "contour",
	})
	if contours == nil || contours.Lines == nil {
		t.Fatal("expected contour lines")
	}
	if len(contours.labels) != 2 {
		t.Fatalf("expected one label per contour level, got %d", len(contours.labels))
	}

	filled := ax.Contourf(grid, ContourOptions{
		Levels: []float64{0, 1, 2, 3, 4},
		Label:  "filled",
	})
	if filled == nil || filled.Fills == nil {
		t.Fatal("expected filled contours")
	}
	mapping := filled.ScalarMap()
	if mapping.VMin != 0 || mapping.VMax != 4 {
		t.Fatalf("unexpected filled contour scalar map %+v", mapping)
	}
}

func TestContourLevelsUseNiceLocatorForImplicitCounts(t *testing.T) {
	levels := contourLevels([]float64{0.287, 1.0}, nil, 6, false)
	want := []float64{0.3, 0.45, 0.6, 0.75, 0.9}
	if len(levels) != len(want) {
		t.Fatalf("levels = %v, want %v", levels, want)
	}
	for i := range want {
		if !approx(levels[i], want[i], 1e-12) {
			t.Fatalf("levels = %v, want %v", levels, want)
		}
	}

	filled := contourLevels([]float64{0.287, 1.0}, nil, 6, true)
	if len(filled) < 2 || filled[0] > 0.287 || filled[len(filled)-1] < 1.0 {
		t.Fatalf("filled levels should cover data range, got %v", filled)
	}
}

func TestTriangulationArtists(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 3}, {0, 3, 2}},
	}

	lineWidth := 2.0
	plot := ax.TriPlot(tri, TriPlotOptions{LineWidth: &lineWidth, Label: "mesh"})
	if plot == nil {
		t.Fatal("expected triplot collection")
	}
	if len(plot.Segments) != 5 {
		t.Fatalf("expected 5 unique edges, got %d", len(plot.Segments))
	}

	colorMesh := ax.TriColor(tri, []float64{0, 1, 2, 3}, TriColorOptions{Label: "tripcolor"})
	if colorMesh == nil {
		t.Fatal("expected tripcolor collection")
	}
	if len(colorMesh.Polygons) != 2 || len(colorMesh.FaceColors) != 2 {
		t.Fatalf("unexpected tripcolor polygon/color counts: %d / %d", len(colorMesh.Polygons), len(colorMesh.FaceColors))
	}

	contours := ax.TriContour(tri, []float64{0, 1, 2, 3}, ContourOptions{Levels: []float64{1.5}})
	if contours == nil || contours.Lines == nil {
		t.Fatal("expected tricontour lines")
	}

	filled := ax.TriContourf(tri, []float64{0, 1, 2, 3}, ContourOptions{Levels: []float64{0, 1, 2, 3}})
	if filled == nil || filled.Fills == nil {
		t.Fatal("expected tricontourf fills")
	}
}

func TestContourLabelsDrawOverlay(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	contours := ax.Contour([][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}, ContourOptions{
		Levels:     []float64{1.5},
		LabelLines: true,
	})
	if contours == nil {
		t.Fatal("expected contour set")
	}

	var renderer contourTextRenderer
	DrawFigure(fig, &renderer)
	if len(renderer.texts) == 0 {
		t.Fatal("expected contour labels to be rendered")
	}
}

func TestContourInlineLabelAngleUsesMatplotlibDisplayConvention(t *testing.T) {
	screen := []geom.Pt{
		{X: 0, Y: 10},
		{X: 5, Y: 5},
		{X: 10, Y: 0},
	}
	data := []geom.Pt{
		{X: 0, Y: 0},
		{X: 5, Y: 5},
		{X: 10, Y: 10},
	}

	angle, parts := splitContourPolylineForLabel(data, screen, 1, 4, 0)
	if len(parts) == 0 {
		t.Fatal("expected split contour parts")
	}
	if !approx(angle, math.Pi/4, 1e-12) {
		t.Fatalf("angle = %v, want %v", angle, math.Pi/4)
	}
}

func TestContourRotatedTextAnchorKeepsCenterFixed(t *testing.T) {
	center := geom.Pt{X: 100, Y: 200}
	angle := math.Pi / 6
	layout := singleLineTextLayout{
		TextLineLayout: render.TextLineLayout{
			Height: 12,
		},
	}

	anchor := contourRotatedTextAnchor(center, layout, angle)
	got := geom.Pt{
		X: anchor.X - 6*math.Sin(angle),
		Y: anchor.Y - 6*math.Cos(angle),
	}
	if !approx(got.X, center.X, 1e-12) || !approx(got.Y, center.Y, 1e-12) {
		t.Fatalf("rotated center = %+v, want %+v", got, center)
	}
}

type contourTextRenderer struct {
	render.NullRenderer
	texts []string
}

func (r *contourTextRenderer) DrawText(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
}

func (r *contourTextRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}
