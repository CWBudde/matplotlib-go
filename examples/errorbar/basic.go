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
	fig := core.NewFigure(800, 500)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.14},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})

	x := []float64{1, 2, 3, 4, 5, 6, 7}
	y := []float64{1.8, 2.5, 3.2, 2.8, 3.4, 2.9, 3.7}
	xErr := []float64{0.15, 0.20, 0.18, 0.22, 0.25, 0.19, 0.17}
	yErr := []float64{0.28, 0.20, 0.24, 0.30, 0.22, 0.26, 0.18}

	lineColor := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 0.95}
	pointColor := render.Color{R: 0.17, G: 0.63, B: 0.17, A: 0.95}
	errColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	lineWidth := 2.0
	errorWidth := 1.2
	errorCap := 6.0
	pointSize := 4.0

	ax.SetTitle("Error Bars with Scatter + Line")
	ax.SetXLabel("Sample")
	ax.SetYLabel("Response")
	ax.XScale = transform.NewLinear(0, 8)
	ax.YScale = transform.NewLinear(0, 5)

	ax.Plot(x, y, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
		Label:     "Trend",
	})
	ax.Scatter(x, y, core.ScatterOptions{
		Color: &pointColor,
		Size:  &pointSize,
		Label: "Samples",
	})
	ax.ErrorBar(x, y, xErr, yErr, core.ErrorBarOptions{
		Color:     &errColor,
		LineWidth: &errorWidth,
		CapSize:   &errorCap,
		Label:     "1σ",
	})
	ax.AddYGrid()

	renderer, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      800,
		Height:     500,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72.0,
	}, backends.TextCapabilities)
	if createErr != nil {
		fmt.Printf("Error creating renderer: %v\n", createErr)
		return
	}

	ax.AutoScale(0.06)
	if err := core.SavePNG(fig, renderer, "errorbar_basic.png"); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}
	fmt.Println("Created errorbar_basic.png")
}
