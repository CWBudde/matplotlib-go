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

	// Basic bar chart matching the Python reference module.
	categories := []float64{1, 2, 3, 4, 5}
	values := []float64{3, 7, 2, 8, 5}
	width := 0.6
	color := render.Color{R: 0.2, G: 0.6, B: 0.8, A: 1}

	ax.SetTitle("Basic Bars")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 10)
	ax.Bar(categories, values, core.BarOptions{
		Width: &width,
		Color: &color,
	})

	// Go examples construct the renderer explicitly; Python hides this in savefig.
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

	if err := core.SavePNG(fig, r, "bar_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved bar_basic.png")
}
