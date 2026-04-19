package main

import (
	"fmt"
	"math"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(1000, 700)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.11, Y: 0.14},
		Max: geom.Pt{X: 0.93, Y: 0.88},
	})

	const n = 180
	x := make([]float64, n)
	y := make([]float64, n)
	peakIndex := 0
	peakValue := math.Inf(-1)
	for i := 0; i < n; i++ {
		xv := 6 * math.Pi * float64(i) / float64(n-1)
		value := 0.75*math.Sin(xv) + 0.2*math.Cos(0.5*xv)
		x[i] = xv
		y[i] = value
		if value > peakValue {
			peakValue = value
			peakIndex = i
		}
	}

	ax.Plot(x, y, core.PlotOptions{Label: "signal"})
	ax.AddLegend()
	ax.AddXGrid()
	ax.AddYGrid()
	ax.SetTitle("Text and Arrow Annotations")
	ax.SetXLabel("phase")
	ax.SetYLabel("response")
	ax.SetXLim(0, 6*math.Pi)
	ax.SetYLim(-1.2, 1.2)

	ax.Text(1.6, 0.92, `\\mu = 0.42, \\Delta x = \\pi/4`, core.TextOptions{
		FontSize: 13,
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignBaseline,
	})
	ax.Annotate(`Peak \\rightarrow`, x[peakIndex], y[peakIndex], core.AnnotationOptions{
		OffsetX:  34,
		OffsetY:  -28,
		FontSize: 13,
	})

	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      1000,
		Height:     700,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        96,
	}, backends.TextCapabilities)
	if createErr != nil {
		fmt.Printf("error creating renderer: %v\n", createErr)
		return
	}

	if err := core.SavePNG(fig, r, "annotation_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved annotation_basic.png")
}
