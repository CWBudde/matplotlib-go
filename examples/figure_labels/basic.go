package main

import (
	"fmt"
	"math"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(1100, 720)
	fig.SetSuptitle("Shared-Axis Figure Labels")
	fig.SetSupXLabel("time [s]")
	fig.SetSupYLabel("amplitude")

	grid := fig.Subplots(2, 2, core.WithSubplotPadding(0.10, 0.92, 0.14, 0.86), core.WithSubplotSpacing(0.10, 0.12))

	for row := range grid {
		for col, ax := range grid[row] {
			x := make([]float64, 180)
			y := make([]float64, 180)
			for i := range x {
				xv := 2 * math.Pi * float64(i) / float64(len(x)-1)
				x[i] = xv
				y[i] = math.Sin(xv+float64(row)*0.5) * (1 + float64(col)*0.2)
			}

			ax.Plot(x, y, core.PlotOptions{
				Label: fmt.Sprintf("series %d", row*2+col+1),
			})
			ax.SetTitle(fmt.Sprintf("Panel %d", row*2+col+1))
			ax.SetXLabel("local x")
			ax.SetYLabel("local y")
			ax.SetXLim(0, 2*math.Pi)
			ax.SetYLim(-1.6, 1.6)
			ax.AddXGrid()
			ax.AddYGrid()
		}
	}

	grid[0][0].AddAnchoredText("upper-left\nnote")
	grid[1][1].AddAnchoredText("lower-right", core.AnchoredTextOptions{
		Location: core.LegendLowerRight,
	})
	fig.AddAnchoredText("Figure note", core.AnchoredTextOptions{
		Location: core.LegendUpperRight,
	})
	fig.AddLegend()

	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      1100,
		Height:     720,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        96,
	}, backends.TextCapabilities)
	if createErr != nil {
		fmt.Printf("error creating renderer: %v\n", createErr)
		return
	}

	if err := core.SavePNG(fig, r, "figure_labels_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved figure_labels_basic.png")
}
