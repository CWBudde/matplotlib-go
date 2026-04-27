package main

import (
	"fmt"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/color"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax.SetTitle("Multi-Series Plot")
	ax.SetXLim(0, 8)
	ax.SetYLim(0, 6)

	tab10 := color.Tab10

	// The reference intentionally mixes line, scatter, and bar artists on one
	// axes so their draw order and color cycling can be compared together.
	ax.Plot(
		[]float64{1, 2, 3, 4, 5, 6},
		[]float64{1.5, 2.8, 2.2, 3.5, 3.8, 4.2},
		core.PlotOptions{
			Color:     &tab10[0],
			LineWidth: floatPtr(2),
		},
	)

	size := core.ScatterAreaFromRadius(8.0, 100.0)
	edgeWidth := 0.0
	ax.Scatter(
		[]float64{1.5, 2.5, 3.5, 4.5, 5.5},
		[]float64{2.2, 3.1, 2.9, 4.1, 4.5},
		core.ScatterOptions{
			Color:     &tab10[1],
			Size:      &size,
			EdgeWidth: &edgeWidth,
		},
	)

	width := 0.4
	ax.Bar(
		[]float64{2, 3, 4, 5},
		[]float64{3.8, 2.5, 4.8, 3.2},
		core.BarOptions{
			Color: &tab10[2],
			Width: &width,
		},
	)

	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      640,
		Height:     360,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if createErr != nil {
		fmt.Printf("error creating renderer: %v\n", createErr)
		return
	}

	if err := core.SavePNG(fig, r, "multi_series_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved multi_series_basic.png")
}

func floatPtr(v float64) *float64 {
	return &v
}
