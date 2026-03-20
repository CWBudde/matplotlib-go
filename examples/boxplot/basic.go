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
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})

	ax.XScale = transform.NewLinear(0, 4)
	ax.YScale = transform.NewLinear(0, 10)
	ax.SetTitle("Box Plots")
	ax.SetXLabel("Group")
	ax.SetYLabel("Value")

	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	alpha := 0.75
	boxWidth := 0.55
	edgeWidth := 1.2
	whiskerWidth := 1.2
	medianWidth := 1.8
	capWidth := 0.35
	flierSize := 4.0

	series := []struct {
		data     []float64
		position float64
		color    render.Color
	}{
		{
			data:     []float64{0.9, 1.0, 1.1, 1.2, 1.3, 1.45, 1.5, 1.7, 1.8},
			position: 1.0,
			color:    render.Color{R: 0.25, G: 0.55, B: 0.82, A: 1},
		},
		{
			data:     []float64{4.0, 4.2, 4.3, 4.5, 4.8, 5.0, 5.4, 5.8, 9.4},
			position: 2.0,
			color:    render.Color{R: 0.80, G: 0.45, B: 0.20, A: 1},
		},
		{
			data:     []float64{2.0, 2.1, 2.1, 2.2, 2.3, 2.4, 2.4, 2.6, 3.8},
			position: 3.0,
			color:    render.Color{R: 0.35, G: 0.70, B: 0.35, A: 1},
		},
	}

	for _, s := range series {
		pos := s.position
		ax.BoxPlot(s.data, core.BoxPlotOptions{
			Position:     &pos,
			Width:        &boxWidth,
			Color:        &s.color,
			EdgeColor:    &black,
			MedianColor:  &black,
			WhiskerColor: &black,
			CapColor:     &black,
			FlierColor:   &black,
			EdgeWidth:    &edgeWidth,
			WhiskerWidth: &whiskerWidth,
			MedianWidth:  &medianWidth,
			CapWidth:     &capWidth,
			FlierSize:    &flierSize,
			Alpha:        &alpha,
		})
	}

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      800,
		Height:     500,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        72.0,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("Error creating renderer: %v\n", err)
		return
	}

	if err := core.SavePNG(fig, r, "boxplot_basic.png"); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully created boxplot_basic.png")
}
