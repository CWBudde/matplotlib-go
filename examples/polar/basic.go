package main

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func main() {
	fig := core.NewFigure(720, 720)
	ax := fig.AddPolarAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.10},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})

	ax.SetTitle("Polar Axes")
	ax.SetXLabel("theta")
	ax.SetYLabel("radius")
	ax.YScale = transform.NewLinear(0, 1.1)
	lineColor := render.Color{R: 0.16, G: 0.33, B: 0.73, A: 1}
	fillColor := render.Color{R: 0.36, G: 0.56, B: 0.92, A: 0.2}
	lineWidth := 2.2

	thetaGrid := ax.AddGrid(core.AxisBottom)
	thetaGrid.Color = render.Color{R: 0.8, G: 0.82, B: 0.86, A: 1}
	thetaGrid.LineWidth = 0.9

	radiusGrid := ax.AddGrid(core.AxisLeft)
	radiusGrid.Color = render.Color{R: 0.82, G: 0.84, B: 0.88, A: 0.9}
	radiusGrid.LineWidth = 0.8

	// Five-lobed radius curve shows how ordinary x/y plot data is interpreted
	// as theta/radius once the axes use the polar projection.
	n := 720
	theta := make([]float64, n)
	radius := make([]float64, n)
	for i := range n {
		theta[i] = 2 * math.Pi * float64(i) / float64(n-1)
		radius[i] = 0.55 + 0.35*math.Cos(5*theta[i])
	}

	ax.Plot(theta, radius, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
		Label:     "r = 0.55 + 0.35 cos(5theta)",
	})
	ax.FillToBaseline(theta, radius, core.FillOptions{
		Color: &fillColor,
	})

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      720,
		Height:     720,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("Error creating renderer: %v\n", err)
		return
	}

	if err := core.SavePNG(fig, r, "polar_basic.png"); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Println("Created polar_basic.png")
}
