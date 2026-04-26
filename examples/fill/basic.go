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

	// Fill a single series down to the zero baseline, matching ax.fill_between(x, 0, y).
	x := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	y := []float64{0.5, 1.8, 2.3, 1.2, 2.8, 1.9, 2.1, 1.5, 0.8}
	fillColor := render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.7}
	edgeColor := render.Color{R: 0.1, G: 0.3, B: 0.5, A: 1.0}
	edgeWidth := 2.0
	baseline := 0.0

	ax.SetTitle("Fill to Baseline")
	ax.SetXLim(0, 10)
	ax.SetYLim(-1, 3)
	ax.FillToBaseline(x, y, core.FillOptions{
		Color:     &fillColor,
		EdgeColor: &edgeColor,
		EdgeWidth: &edgeWidth,
		Baseline:  &baseline,
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

	if err := core.SavePNG(fig, r, "fill_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved fill_basic.png")
}
