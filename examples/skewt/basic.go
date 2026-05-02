package main

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(720, 640)
	ax, err := fig.AddSkewXAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.14},
		Max: geom.Pt{X: 0.86, Y: 0.88},
	})
	if err != nil {
		fmt.Printf("Error creating skewx axes: %v\n", err)
		return
	}

	ax.SetTitle("Skew-T Style Projection")
	ax.SetXLabel("Temperature (deg C)")
	ax.SetYLabel("Pressure (hPa)")
	ax.SetXLim(-70, 35)
	ax.SetYLim(1050, 180)

	gridColor := render.Color{R: 0.82, G: 0.84, B: 0.88, A: 1}
	xGrid := ax.AddGrid(core.AxisBottom)
	xGrid.Color = gridColor
	xGrid.LineWidth = 0.8
	yGrid := ax.AddGrid(core.AxisLeft)
	yGrid.Color = gridColor
	yGrid.LineWidth = 0.8

	pressure := []float64{1000, 925, 850, 700, 600, 500, 400, 300, 250, 200}
	temperature := []float64{24, 20, 15, 5, -4, -14, -28, -43, -51, -58}
	dewpoint := []float64{18, 14, 8, -4, -14, -25, -38, -50, -57, -64}

	// AddSkewXAxes applies the skew transform, so these profiles are supplied
	// in normal temperature/pressure coordinates just like the Python data.
	tempColor := render.Color{R: 0.78, G: 0.13, B: 0.16, A: 1}
	dewColor := render.Color{R: 0.05, G: 0.48, B: 0.28, A: 1}
	width := 2.4
	ax.Plot(temperature, pressure, core.PlotOptions{
		Color:     &tempColor,
		LineWidth: &width,
		Label:     "temperature",
	})
	ax.Plot(dewpoint, pressure, core.PlotOptions{
		Color:     &dewColor,
		LineWidth: &width,
		Label:     "dewpoint",
	})
	ax.AddLegend()

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      720,
		Height:     640,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("Error creating renderer: %v\n", err)
		return
	}

	if err := core.SavePNG(fig, r, "skewt_basic.png"); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Println("Created skewt_basic.png")
}
