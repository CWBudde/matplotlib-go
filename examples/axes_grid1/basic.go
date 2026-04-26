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
	fig := core.NewFigure(1100, 720)

	grid := fig.NewImageGrid(
		2,
		2,
		geom.Rect{
			Min: geom.Pt{X: 0.06, Y: 0.12},
			Max: geom.Pt{X: 0.58, Y: 0.88},
		},
		core.WithAxesDividerHorizontalSpace(0.03),
		core.WithAxesDividerVerticalSpace(0.04),
		core.WithAxesDividerWidthScales(1.2, 1),
		core.WithAxesDividerHeightScales(1, 1.1),
	)
	if grid == nil {
		fmt.Println("image grid creation failed")
		return
	}

	for row := range 2 {
		for col := range 2 {
			ax := grid.At(row, col)
			ax.SetTitle(fmt.Sprintf("Tile %d,%d", row+1, col+1))
			ax.MatShow(surface(24, 24, float64(row*2+col)))
			ax.AddAnchoredText("image grid", core.AnchoredTextOptions{
				Location: core.LegendLowerRight,
			})
		}
	}

	rgb := fig.NewRGBAxes(
		geom.Rect{
			Min: geom.Pt{X: 0.66, Y: 0.18},
			Max: geom.Pt{X: 0.96, Y: 0.84},
		},
		core.WithAxesDividerHorizontalSpace(0.025),
	)
	if rgb == nil {
		fmt.Println("rgb axes creation failed")
		return
	}

	channels := []struct {
		ax    *core.Axes
		title string
		phase float64
	}{
		{ax: rgb.Red, title: "Red", phase: 0},
		{ax: rgb.Green, title: "Green", phase: 1.2},
		{ax: rgb.Blue, title: "Blue", phase: 2.4},
	}
	for _, channel := range channels {
		channel.ax.SetTitle(channel.title)
		channel.ax.MatShow(channelSurface(28, 28, channel.phase))
	}

	fig.AddAnchoredText("axes_grid1-style layout\nImageGrid + RGBAxes", core.AnchoredTextOptions{
		Location: core.LegendUpperRight,
	})

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      1100,
		Height:     720,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("error creating renderer: %v\n", err)
		return
	}
	if err := core.SavePNG(fig, r, "axes_grid1_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved axes_grid1_basic.png")
}

func surface(rows, cols int, phase float64) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			values[y][x] = 0.5 + 0.25*math.Sin((xx+phase)*2*math.Pi) + 0.25*math.Cos((yy+phase*0.3)*3*math.Pi)
		}
	}
	return values
}

func channelSurface(rows, cols int, phase float64) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			values[y][x] = 0.5 + 0.5*math.Sin((xx*2+yy*1.5+phase)*math.Pi)
		}
	}
	return values
}
