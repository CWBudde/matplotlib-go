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
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})

	const (
		rows = 80
		cols = 120
	)
	data := make([][]float64, rows)
	for row := 0; row < rows; row++ {
		data[row] = make([]float64, cols)
		for col := 0; col < cols; col++ {
			x := (float64(col)/float64(cols-1))*4 - 2
			y := (float64(row)/float64(rows-1))*4 - 2
			r := math.Hypot(x, y)
			data[row][col] = math.Sin(3*r) * math.Exp(-0.6*r)
		}
	}

	cmap := "inferno"
	img := ax.Image(data, core.ImageOptions{Colormap: &cmap})
	ax.SetTitle("Heatmap with Colorbar")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(0, cols)
	ax.SetYLim(0, rows)
	ax.AddXGrid()
	ax.AddYGrid()

	fig.AddColorbar(ax, img, core.ColorbarOptions{Label: "Intensity"})

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

	if err := core.SavePNG(fig, r, "colorbar_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved colorbar_basic.png")
}
