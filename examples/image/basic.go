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
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})

	// Small 3x3 matrix to make the image extent and orientation easy to inspect.
	data := [][]float64{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
	}
	cmap := "viridis"
	vmin, vmax := 0.0, 8.0
	ax.SetTitle("Image Heatmap")
	ax.SetXLim(0, 3)
	ax.SetYLim(0, 3)
	ax.Image(data, core.ImageOptions{
		Colormap:      &cmap,
		VMin:          &vmin,
		VMax:          &vmax,
		XMin:          ptr(0.0),
		XMax:          ptr(3.0),
		YMin:          ptr(0.0),
		YMax:          ptr(3.0),
		Origin:        core.ImageOriginLower,
		Interpolation: ptr("bilinear"),
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

	if err := core.SavePNG(fig, r, "image_heatmap.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}
	fmt.Println("saved image_heatmap.png")
}

func ptr[T any](v T) *T {
	return &v
}
