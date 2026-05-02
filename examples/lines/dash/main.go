// Package demonstrates dash patterns with matplotlib-go.
package main

import (
	"log"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax.SetTitle("Dash Patterns")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 5)

	// Dash lengths are specified in Go renderer pixels. The Python reference
	// converts these same values to Matplotlib points.
	specs := []struct {
		y       float64
		pattern []float64
		color   render.Color
	}{
		{4, nil, render.Color{R: 0, G: 0, B: 0, A: 1}},
		{3, []float64{10, 4}, render.Color{R: 0.8, G: 0, B: 0, A: 1}},
		{2, []float64{6, 2, 2, 2}, render.Color{R: 0, G: 0.6, B: 0, A: 1}},
		{1, []float64{2, 2}, render.Color{R: 0, G: 0, B: 0.8, A: 1}},
	}

	for _, spec := range specs {
		line := &core.Line2D{
			XY: []geom.Pt{
				{X: 1, Y: spec.y},
				{X: 9, Y: spec.y},
			},
			W:      3.0,
			Col:    spec.color,
			Dashes: spec.pattern,
		}
		ax.Add(line)
	}

	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      640,
		Height:     360,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if createErr != nil {
		log.Fatalf("Failed to create renderer: %v", createErr)
	}

	err := core.SavePNG(fig, r, "dashes.png")
	if err != nil {
		log.Fatalf("Failed to save PNG: %v", err)
	}

	log.Println("saved dashes.png")
}
