package main

import (
	"fmt"
	"math"
	"math/rand/v2"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func main() {
	// Generate normally-distributed data using Box-Muller transform.
	rng := rand.New(rand.NewPCG(42, 0))
	n := 500
	data := make([]float64, n)
	for i := range data {
		u1 := rng.Float64()
		u2 := rng.Float64()
		data[i] = math.Sqrt(-2*math.Log(u1))*math.Cos(2*math.Pi*u2)*1.5 + 5.0
	}

	// --- Plot 1: Basic histogram (auto bins) ---
	fig := core.NewFigure(800, 500)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})
	ax.SetTitle("Histogram — Auto Bins (Sturges)")
	ax.SetXLabel("Value")
	ax.SetYLabel("Count")

	blue := render.Color{R: 0.26, G: 0.53, B: 0.80, A: 0.8}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 0.8
	ax.Hist(data, core.HistOptions{
		Color:     &blue,
		EdgeColor: &black,
		EdgeWidth: &ew,
		Label:     "normal(5, 1.5)",
	})
	ax.AutoScale(0.05)
	ax.AddXGrid()
	ax.AddYGrid()

	r1, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      800,
		Height:     500,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72.0,
	}, backends.TextCapabilities)
	if createErr != nil {
		fmt.Printf("Error: %v\n", createErr)
		return
	}
	if err := core.SavePNG(fig, r1, "hist_basic.png"); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// --- Plot 2: Density normalization with custom bin count ---
	fig2 := core.NewFigure(800, 500)
	ax2 := fig2.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})
	ax2.SetTitle("Histogram — Density Normalization, 20 Bins")
	ax2.SetXLabel("Value")
	ax2.SetYLabel("Density")

	green := render.Color{R: 0.20, G: 0.65, B: 0.30, A: 0.8}
	bins := 20
	ax2.Hist(data, core.HistOptions{
		Bins:      bins,
		Norm:      core.HistNormDensity,
		Color:     &green,
		EdgeColor: &black,
		EdgeWidth: &ew,
		Label:     "density",
	})
	ax2.AutoScale(0.05)
	ax2.AddXGrid()
	ax2.AddYGrid()

	r2, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      800,
		Height:     500,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72.0,
	}, backends.TextCapabilities)
	if createErr != nil {
		fmt.Printf("Error: %v\n", createErr)
		return
	}
	if err := core.SavePNG(fig2, r2, "hist_density.png"); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Created hist_basic.png and hist_density.png")
}
