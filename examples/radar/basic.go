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
	fig := core.NewFigure(720, 720)
	labels := []string{"Speed", "Power", "Range", "Handling", "Comfort"}
	ax, err := fig.AddRadarAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.10},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	}, labels)
	if err != nil {
		fmt.Printf("Error creating radar axes: %v\n", err)
		return
	}

	ax.SetTitle("Radar Projection")
	ax.YScale = transform.NewLinear(0, 1)
	ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0.25, 0.5, 0.75, 1.0}}
	ax.YAxis.MinorLocator = nil
	ax.YAxis.Formatter = core.PercentFormatter{Decimals: 0}

	// Split theta and radius grids so their styling matches the Python polar
	// axes calls, while the radial ticks use percent labels.
	thetaGrid := ax.AddGrid(core.AxisBottom)
	thetaGrid.Color = render.Color{R: 0.78, G: 0.80, B: 0.84, A: 1}
	thetaGrid.LineWidth = 0.8

	radiusGrid := ax.AddGrid(core.AxisLeft)
	radiusGrid.Color = render.Color{R: 0.80, G: 0.83, B: 0.88, A: 1}
	radiusGrid.LineWidth = 0.8

	// Radar data is closed explicitly by repeating the first angle/value pair.
	angles := core.RadarAngles(len(labels))
	values := []float64{0.72, 0.88, 0.64, 0.79, 0.58}
	closedAngles := append(append([]float64(nil), angles...), angles[0])
	closedValues := append(append([]float64(nil), values...), values[0])

	lineColor := render.Color{R: 0.15, G: 0.35, B: 0.70, A: 1}
	fillColor := render.Color{R: 0.18, G: 0.50, B: 0.82, A: 0.22}
	lineWidth := 2.2
	ax.FillToBaseline(closedAngles, closedValues, core.FillOptions{
		Color: &fillColor,
	})
	ax.Plot(closedAngles, closedValues, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
		Label:     "model A",
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

	if err := core.SavePNG(fig, r, "radar_basic.png"); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Println("Created radar_basic.png")
}
