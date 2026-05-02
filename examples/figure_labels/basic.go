package main

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(1100, 720)
	grid := fig.Subplots(
		2,
		2,
		core.WithSubplotPadding(0.083, 0.996, 0.0986, 0.9333),
		core.WithSubplotSpacing(0.067, 0.100),
	)
	fig.SetSuptitle("Shared-Axis Figure Labels")
	fig.SetSupXLabel("time [s]")
	fig.SetSupYLabel("amplitude")

	textBox := &core.TextBBoxOptions{
		FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor: render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
	}

	for row := range grid {
		for col, ax := range grid[row] {
			x := make([]float64, 180)
			y := make([]float64, 180)
			for i := range x {
				xv := 2 * math.Pi * float64(i) / float64(len(x)-1)
				x[i] = xv
				y[i] = math.Sin(xv+float64(row)*0.5) * (1 + float64(col)*0.2)
			}

			lineWidth := 1.5 * fig.RC.DPI / 72.0
			ax.Plot(x, y, core.PlotOptions{
				LineWidth: &lineWidth,
				Label:     fmt.Sprintf("series %d", row*2+col+1),
			})
			ax.SetTitle(fmt.Sprintf("Panel %d", row*2+col+1))
			ax.SetXLabel("local x")
			ax.SetYLabel("local y")
			ax.SetXLim(0, 2*math.Pi)
			ax.SetYLim(-1.6, 1.6)
			xGrid := ax.AddGrid(core.AxisBottom)
			xGrid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
			xGrid.LineWidth = 0.5
			yGrid := ax.AddGrid(core.AxisLeft)
			yGrid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
			yGrid.LineWidth = 0.5
		}
	}

	grid[0][0].Text(0.02, 0.92, "upper-left\nnote", core.TextOptions{
		Coords: core.Coords(core.CoordAxes),
		VAlign: core.TextVAlignTop,
		BBox:   textBox,
	})
	grid[1][1].Text(0.98, 0.08, "lower-right", core.TextOptions{
		Coords: core.Coords(core.CoordAxes),
		HAlign: core.TextAlignRight,
		VAlign: core.TextVAlignBottom,
		BBox:   textBox,
	})
	fig.Text(0.985, 0.94, "Figure note", core.TextOptions{
		HAlign: core.TextAlignRight,
		VAlign: core.TextVAlignTop,
		BBox:   textBox,
	})
	legend := fig.AddLegend()
	legend.SetLocator(core.BBoxToAnchorLocator{
		X:        0.99,
		Y:        0.90,
		Location: core.LegendUpperRight,
	})

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
