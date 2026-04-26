package main

import (
	"fmt"
	"math"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(1320, 520)

	// Shared triangulation mirrors the Python reference exactly so mesh,
	// pseudocolor, and contour behavior differ only by renderer capability.
	tri, values := sampleTriangulation()

	meshAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.16},
		Max: geom.Pt{X: 0.31, Y: 0.88},
	})
	configureUnstructuredAxes(meshAx, "Triangulation")
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
	configureUnstructuredAxes(colorAx, "Tripcolor + Tricontour")
	cmap := "viridis"
	edgeColor := render.Color{R: 1, G: 1, B: 1, A: 0.55}
	edgeWidth := 0.6
	// Flat tripcolor plus labeled tricontours matches the middle reference
	// panel and exercises scalar values defined at triangulation vertices.
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
	configureUnstructuredAxes(fillAx, "Filled Tricontour")
	fillMap := "plasma"
	// The final panel overlays line contours on filled contours to make level
	// placement visible without a colorbar.
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

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      1320,
		Height:     520,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("error creating renderer: %v\n", err)
		return
	}
	if err := core.SavePNG(fig, r, "unstructured_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved unstructured_basic.png")
}

func configureUnstructuredAxes(ax *core.Axes, title string) {
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(-0.1, 3.1)
	ax.SetYLim(-0.15, 2.65)
	_ = ax.SetAspect("equal")
}

func sampleTriangulation() (core.Triangulation, []float64) {
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
