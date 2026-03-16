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
	fig := core.NewFigure(800, 500)

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	ax.XScale = transform.NewLinear(0, 2*math.Pi)
	ax.YScale = transform.NewLinear(-1.5, 1.5)

	// Title, xlabel, ylabel (ylabel is rotated 90° in AGG backend)
	ax.SetTitle("Text Labels: Title, X-Label, Rotated Y-Label")
	ax.SetXLabel("Angle (radians)")
	ax.SetYLabel("Amplitude")

	// Minor ticks for finer detail
	ax.XAxis.MinorLocator = core.MinorLinearLocator{N: 5}
	ax.YAxis.MinorLocator = core.MinorLinearLocator{N: 4}

	ax.AddXGrid()
	ax.AddYGrid()

	// Generate data
	n := 200
	x := make([]float64, n)
	y := make([]float64, n)
	for i := range n {
		x[i] = 2 * math.Pi * float64(i) / float64(n-1)
		y[i] = math.Sin(x[i])
	}

	ax.Plot(x, y, core.PlotOptions{Label: "sin(x)"})

	r, err := agg.New(800, 500, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := core.SavePNG(fig, r, "labels.png"); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Created labels.png — title, x-label, rotated y-label, and tick labels")
}
