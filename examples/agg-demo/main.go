package main

import (
	"fmt"
	"math"

	"matplotlib-go/backends/agg"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/transform"
)

func main() {
	// Create a figure with dimensions 800x500
	fig := core.NewFigure(800, 500)

	// Add axes
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-1.2, 1.2)

	// Generate sine wave data
	var sinePts []geom.Pt
	for i := 0; i <= 100; i++ {
		x := float64(i) / 10.0
		y := math.Sin(x)
		sinePts = append(sinePts, geom.Pt{X: x, Y: y})
	}

	// Generate cosine wave data
	var cosPts []geom.Pt
	for i := 0; i <= 100; i++ {
		x := float64(i) / 10.0
		y := math.Cos(x)
		cosPts = append(cosPts, geom.Pt{X: x, Y: y})
	}

	// Add sine line (blue)
	ax.Add(&core.Line2D{
		XY:  sinePts,
		W:   2.5,
		Col: render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}, // Tab blue
	})

	// Add cosine line (orange)
	ax.Add(&core.Line2D{
		XY:  cosPts,
		W:   2.5,
		Col: render.Color{R: 1.0, G: 0.50, B: 0.05, A: 1}, // Tab orange
	})

	// Create AGG renderer with white background
	r := agg.New(800, 500, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Save as PNG using the AGG backend
	err := core.SavePNG(fig, r, "agg_demo.png")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Created agg_demo.png using the AGG anti-aliased renderer!")
}
