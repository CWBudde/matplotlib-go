package main

import (
	"fmt"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/transform"
)

func main() {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})
	ax.XScale = transform.NewLinear(0, 3)
	ax.YScale = transform.NewLinear(0, 3)

	data := [][]float64{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
	}
	ax.Image(data, core.ImageOptions{
		Colormap: strPtr("inferno"),
	})

	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      640,
		Height:     360,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72.0,
	}, backends.TextCapabilities)
	if createErr != nil {
		fmt.Printf("Error creating renderer: %v\n", createErr)
		return
	}

	if err := core.SavePNG(fig, r, "out.png"); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}
	fmt.Println("Created out.png with an image heatmap")
}

func strPtr(s string) *string {
	return &s
}
