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

	// Add axes with margins for labels
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.18},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-1.2, 1.2)

	// Set labels
	ax.SetTitle("Sine and Cosine Waves")
	ax.SetXLabel("x (radians)")
	ax.SetYLabel("y")

	// Add grid lines (behind data)
	ax.AddXGrid()
	ax.AddYGrid()

	// Generate sine wave data
	n := 200
	x := make([]float64, n)
	sinY := make([]float64, n)
	cosY := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = float64(i) / float64(n-1) * 10.0
		sinY[i] = math.Sin(x[i])
		cosY[i] = math.Cos(x[i])
	}

	// Plot sine (auto color cycle: blue)
	ax.Plot(x, sinY, core.PlotOptions{Label: "sin(x)"})

	// Plot cosine (auto color cycle: orange)
	ax.Plot(x, cosY, core.PlotOptions{Label: "cos(x)"})

	// Create AGG renderer with white background
	r, err := agg.New(800, 500, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		fmt.Printf("Error creating renderer: %v\n", err)
		return
	}

	// Save as PNG
	err = core.SavePNG(fig, r, "agg_demo.png")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Created agg_demo.png — sine/cosine plot with AGG anti-aliased rendering,")
	fmt.Println("axis ticks, grid lines, and labels!")
}
