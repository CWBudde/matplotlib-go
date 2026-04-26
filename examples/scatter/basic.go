package main

import (
	"fmt"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
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

	ax.SetTitle("Basic Scatter")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)

	// Keep this example intentionally small: one scatter call, matching the
	// Python reference used for visual comparisons.
	size := 8.0
	edgeWidth := 0.0
	ax.Scatter(
		[]float64{2, 4, 6, 8, 3, 7},
		[]float64{3, 6, 4, 7, 8, 2},
		core.ScatterOptions{
			Color:     &render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
			Size:      &size,
			EdgeWidth: &edgeWidth,
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

	if err := core.SavePNG(fig, r, "scatter_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved scatter_basic.png")
}
