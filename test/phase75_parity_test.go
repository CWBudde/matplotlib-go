package test

import (
	"image"
	"math"
	"testing"

	"matplotlib-go/backends/agg"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestUnstructuredShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "unstructured_showcase", renderUnstructuredShowcase)
}

func TestArraysShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "arrays_showcase", renderArraysShowcase)
}

func renderUnstructuredShowcase() image.Image {
	fig := core.NewFigure(1320, 520)

	tri, values := parityTriangulation()

	meshAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.16},
		Max: geom.Pt{X: 0.31, Y: 0.88},
	})
	configureParityUnstructuredAxes(meshAx, "Triangulation")
	meshColor := render.Color{R: 0.18, G: 0.24, B: 0.34, A: 1}
	meshWidth := 1.35
	meshAx.TriPlot(tri, core.TriPlotOptions{
		Color:     &meshColor,
		LineWidth: &meshWidth,
		Label:     "triplot",
	})
	meshAx.AddAnchoredText("explicit triangular mesh", core.AnchoredTextOptions{
		Location: core.LegendLowerRight,
	})

	colorAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.37, Y: 0.16},
		Max: geom.Pt{X: 0.63, Y: 0.88},
	})
	configureParityUnstructuredAxes(colorAx, "Tripcolor + Tricontour")
	cmap := "viridis"
	edgeColor := render.Color{R: 1, G: 1, B: 1, A: 0.55}
	edgeWidth := 0.6
	colorAx.TriColor(tri, values, core.TriColorOptions{
		Colormap:  &cmap,
		EdgeColor: &edgeColor,
		EdgeWidth: &edgeWidth,
		Label:     "tripcolor",
	})
	contourColor := render.Color{R: 0.08, G: 0.12, B: 0.18, A: 0.95}
	contourWidth := 1.15
	colorAx.TriContour(tri, values, core.ContourOptions{
		Color:      &contourColor,
		LineWidth:  &contourWidth,
		LevelCount: 6,
		LabelLines: true,
		LabelColor: &contourColor,
	})

	fillAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.69, Y: 0.16},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})
	configureParityUnstructuredAxes(fillAx, "Filled Tricontour")
	fillMap := "plasma"
	fillAx.TriContourf(tri, values, core.ContourOptions{
		Colormap:   &fillMap,
		LevelCount: 7,
		Label:      "tricontourf",
	})
	highlight := render.Color{R: 1, G: 1, B: 1, A: 0.88}
	highlightWidth := 0.95
	fillAx.TriContour(tri, values, core.ContourOptions{
		Color:      &highlight,
		LineWidth:  &highlightWidth,
		LevelCount: 7,
	})

	fig.AddAnchoredText("unstructured gallery family\ntriangulation, tripcolor, tricontour", core.AnchoredTextOptions{
		Location: core.LegendUpperRight,
	})

	r, err := agg.New(1320, 520, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderArraysShowcase() image.Image {
	fig := core.NewFigure(1240, 620)

	heatAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.14},
		Max: geom.Pt{X: 0.31, Y: 0.88},
	})
	heatAx.SetTitle("Annotated Heatmap")
	heatAx.SetXLabel("column")
	heatAx.SetYLabel("row")
	heatMap := "viridis"
	heatAx.AnnotatedHeatmap(parityAnnotatedData(), core.AnnotatedHeatmapOptions{
		MatShowOptions: core.MatShowOptions{
			Colormap:     &heatMap,
			Aspect:       "equal",
			IntegerTicks: parityBoolPtr(true),
		},
		Format:        "%.2f",
		FontSize:      10,
		TextColor:     render.Color{R: 0.12, G: 0.12, B: 0.14, A: 1},
		TextColorHigh: render.Color{R: 1, G: 1, B: 1, A: 1},
	})

	meshAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.37, Y: 0.14},
		Max: geom.Pt{X: 0.63, Y: 0.88},
	})
	meshAx.SetTitle("PColorMesh + Contour")
	meshAx.SetXLabel("x bin")
	meshAx.SetYLabel("y bin")
	meshMap := "plasma"
	meshEdges := render.Color{R: 1, G: 1, B: 1, A: 0.48}
	meshEdgeWidth := 0.65
	meshData := parityWaveGrid(8, 10, 0.35)
	meshAx.PColorMesh(meshData, core.MeshOptions{
		Colormap:  &meshMap,
		EdgeColor: &meshEdges,
		EdgeWidth: &meshEdgeWidth,
		Label:     "pcolormesh",
	})
	contourColor := render.Color{R: 0.14, G: 0.10, B: 0.16, A: 0.95}
	contourWidth := 1.1
	meshAx.Contour(meshData, core.ContourOptions{
		Color:      &contourColor,
		LineWidth:  &contourWidth,
		LevelCount: 6,
		LabelLines: true,
		LabelColor: &contourColor,
	})

	spyAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.69, Y: 0.14},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})
	spyAx.SetTitle("Spy")
	spyAx.SetXLabel("column")
	spyAx.SetYLabel("row")
	spyColor := render.Color{R: 0.16, G: 0.38, B: 0.72, A: 1}
	spyAx.Spy(paritySparsePattern(18, 18), core.SpyOptions{
		Precision:  0.1,
		MarkerSize: 10,
		Color:      &spyColor,
		Label:      "spy",
	})
	spyAx.AddAnchoredText("sparse structure view", core.AnchoredTextOptions{
		Location: core.LegendLowerRight,
	})

	fig.AddAnchoredText("arrays gallery family\nheatmap, quad mesh, sparse matrix", core.AnchoredTextOptions{
		Location: core.LegendUpperRight,
	})

	r, err := agg.New(1240, 620, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func configureParityUnstructuredAxes(ax *core.Axes, title string) {
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(-0.1, 3.1)
	ax.SetYLim(-0.15, 2.65)
	_ = ax.SetAspect("equal")
}

func parityTriangulation() (core.Triangulation, []float64) {
	tri := core.Triangulation{
		X: []float64{0.0, 0.85, 1.75, 2.85, 0.2, 1.1, 2.1, 0.55, 1.55, 2.55},
		Y: []float64{0.0, 0.2, 0.05, 0.3, 1.0, 1.15, 1.25, 2.15, 2.3, 2.05},
		Triangles: [][3]int{
			{0, 1, 4},
			{1, 5, 4},
			{1, 2, 5},
			{2, 6, 5},
			{2, 3, 6},
			{4, 5, 7},
			{5, 8, 7},
			{5, 6, 8},
			{6, 9, 8},
		},
	}

	values := make([]float64, len(tri.X))
	for i := range values {
		values[i] = math.Sin(tri.X[i]*1.4) + 0.7*math.Cos((tri.Y[i]+0.15)*2.1)
	}
	return tri, values
}

func parityAnnotatedData() [][]float64 {
	return [][]float64{
		{0.12, 0.28, 0.46, 0.64, 0.82},
		{0.18, 0.34, 0.58, 0.74, 0.88},
		{0.24, 0.42, 0.63, 0.79, 0.91},
		{0.16, 0.38, 0.61, 0.83, 0.97},
	}
}

func parityWaveGrid(rows, cols int, phase float64) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			values[y][x] = 0.55 + 0.25*math.Sin((xx*2.3+phase)*math.Pi) + 0.20*math.Cos((yy*2.8-phase*0.4)*math.Pi)
		}
	}
	return values
}

func paritySparsePattern(rows, cols int) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		for x := range cols {
			if x == y || x+y == cols-1 || (x+2*y)%7 == 0 || (2*x+y)%11 == 0 {
				values[y][x] = 1
			}
		}
	}
	return values
}

func parityBoolPtr(v bool) *bool {
	return &v
}
