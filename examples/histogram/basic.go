package main

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	// Generate the same Go-compatible normal sample stream used by the Python
	// reference helper normal_data(42, 0, 500, 5.0, 1.5).
	rng := rand.New(rand.NewPCG(42, 0))
	n := 500
	data := make([]float64, n)
	for i := range data {
		u1 := rng.Float64()
		u2 := rng.Float64()
		data[i] = math.Sqrt(-2*math.Log(u1))*math.Cos(2*math.Pi*u2)*1.5 + 5.0
	}

	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	ax.SetTitle("Basic Histogram")

	blue := render.Color{R: 0.26, G: 0.53, B: 0.80, A: 0.8}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 0.8
	// Bins defaults to auto; for n=500 that selects Sturges, matching bins="sturges".
	ax.Hist(data, core.HistOptions{
		Color:     &blue,
		EdgeColor: &black,
		EdgeWidth: &ew,
	})
	ax.AutoScale(0.05)

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
	if err := core.SavePNG(fig, r, "hist_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved hist_basic.png")
}
