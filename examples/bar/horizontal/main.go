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
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax.SetTitle("Horizontal Bars")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 6)

	y := []float64{1, 2, 3, 4, 5}
	widths := []float64{3, 7, 2, 8, 5}
	height := 0.6
	color := render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1}
	orientation := core.BarHorizontal
	ax.Bar(y, widths, core.BarOptions{
		Width:       &height,
		Color:       &color,
		Orientation: &orientation,
	})

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

	if err := core.SavePNG(fig, r, "bar_horizontal.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved bar_horizontal.png")
}
