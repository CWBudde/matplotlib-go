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

	// Left side mirrors the Python ImageGrid reference: a 2x2 image tile layout
	// with small gutters and per-tile anchored labels.
	grid := fig.NewImageGrid(
		2,
		2,
		geom.Rect{
			Min: geom.Pt{X: 0.06, Y: 0.12},
			Max: geom.Pt{X: 0.60, Y: 0.88},
		},
		core.WithAxesDividerHorizontalSpace(0.18/11.0),
		core.WithAxesDividerVerticalSpace(0.20/7.2),
	)
	if grid == nil {
		fmt.Println("image grid creation failed")
		return
	}

	for row := range 2 {
		for col := range 2 {
			ax := grid.At(row, col)
			ax.SetTitle(fmt.Sprintf("Tile %d,%d", row+1, col+1))
			// Use the same deterministic phase formula as the Python reference
			// so each tile differs without needing external image assets.
			ax.ImShow(surface(24, 24, float64(row*2+col)))
			ax.AddAnchoredText("image grid", core.AnchoredTextOptions{
				Location: core.LegendLowerRight,
			})
		}
	}

	// Right side mirrors the RGB channel axes: three shared small images in a
	// single row. The Go helper exposes this as RGBAxes.
	rgb := fig.NewRGBAxes(
		geom.Rect{
			Min: geom.Pt{X: 0.66, Y: 0.34},
			Max: geom.Pt{X: 0.98, Y: 0.56},
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
		phase int
		cmap  string
	}{
		{ax: rgb.Red, title: "Red", phase: 0, cmap: "red channel"},
		{ax: rgb.Green, title: "Green", phase: 1, cmap: "green channel"},
		{ax: rgb.Blue, title: "Blue", phase: 2, cmap: "blue channel"},
	}
	for _, channel := range channels {
		channel.ax.SetTitle(channel.title)
		channel.ax.ImShow(channelSurface(28, 28, channel.phase), core.ImShowOptions{
			Colormap: &channel.cmap,
		})
		channel.ax.XAxis.Locator = core.FixedLocator{TicksList: []float64{0, 10, 20}}
		channel.ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 10, 20}}
	}

	fig.AddAnchoredText("axes_grid1-style layout\nImageGrid + RGB channel views", core.AnchoredTextOptions{
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

func channelSurface(rows, cols int, phase int) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			switch phase {
			case 1:
				values[y][x] = 0.5 + 0.32*math.Sin(yy*4*math.Pi) + 0.18*math.Cos(xx*2*math.Pi)
			case 2:
				dx := xx - 0.5
				dy := yy - 0.5
				values[y][x] = 0.58 + 0.36*math.Sin((xx+yy)*3*math.Pi) - 0.18*math.Hypot(dx, dy)
			default:
				dx := xx - 0.35
				dy := yy - 0.42
				values[y][x] = 0.35 + 0.65*math.Exp(-7*(dx*dx+dy*dy))
			}
		}
	}
	return values
}
