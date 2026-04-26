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
	fig := core.NewFigure(1240, 620)

	heatAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.14},
		Max: geom.Pt{X: 0.31, Y: 0.88},
	})
	heatAx.SetTitle("Annotated Heatmap")
	heatAx.SetXLabel("column")
	heatAx.SetYLabel("row")
	heatMap := "viridis"
	heatAx.AnnotatedHeatmap(annotatedData(), core.AnnotatedHeatmapOptions{
		MatShowOptions: core.MatShowOptions{
			Colormap:     &heatMap,
			Aspect:       "equal",
			IntegerTicks: boolPtr(true),
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
	meshData := waveGrid(8, 10, 0.35)
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
	spyAx.Spy(sparsePattern(18, 18), core.SpyOptions{
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

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      1240,
		Height:     620,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("error creating renderer: %v\n", err)
		return
	}
	if err := core.SavePNG(fig, r, "arrays_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved arrays_basic.png")
}

func boolPtr(v bool) *bool {
	return &v
}

func annotatedData() [][]float64 {
	return [][]float64{
		{0.12, 0.28, 0.46, 0.64, 0.82},
		{0.18, 0.34, 0.58, 0.74, 0.88},
		{0.24, 0.42, 0.63, 0.79, 0.91},
		{0.16, 0.38, 0.61, 0.83, 0.97},
	}
}

func waveGrid(rows, cols int, phase float64) [][]float64 {
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

func sparsePattern(rows, cols int) [][]float64 {
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
