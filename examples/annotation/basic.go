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
	fig := core.NewFigure(1040, 720)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	})

	// Match the reference signal: a damped sine wave with a slow cosine offset.
	const n = 240
	x := make([]float64, n)
	y := make([]float64, n)
	for i := 0; i < n; i++ {
		xv := 6 * math.Pi * float64(i) / float64(n-1)
		value := math.Sin(xv)*math.Exp(-0.015*xv) + 0.2*math.Cos(0.5*xv)
		x[i] = xv
		y[i] = value
	}

	ax.SetTitle("Text and Arrow Annotations")
	ax.SetXLabel("phase")
	ax.SetYLabel("response")
	ax.AddXGrid()
	ax.AddYGrid()
	ax.Plot(x, y, core.PlotOptions{Label: "signal"})
	ax.SetXLim(0, 6*math.Pi)
	ax.SetYLim(-1.2, 1.2)
	ax.AddLegend()

	// The reference uses axes-fraction coordinates; these data coordinates place
	// the label at the same relative spot for the fixed x/y limits above.
	ax.Text(0.20*6*math.Pi, 0.96, "m∫T  φ x =  λ/4", core.TextOptions{
		FontSize: 12,
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignBaseline,
	})

	// Annotate the analytical first peak used by the Python reference.
	peakX := math.Pi / 2
	peakY := math.Sin(peakX)*math.Exp(-0.015*peakX) + 0.2*math.Cos(0.5*peakX)
	ax.Annotate("Peak\n= 0.42", peakX, peakY, core.AnnotationOptions{
		OffsetX:  48,
		OffsetY:  -42,
		FontSize: 12,
	})

	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      1040,
		Height:     720,
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
