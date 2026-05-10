package image_heatmap

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})
	ax.SetTitle("Image Heatmap")
	ax.SetXLim(0, 3)
	ax.SetYLim(0, 3)

	data := [][]float64{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
	}

	cmap := "viridis"
	vmin, vmax := 0.0, 8.0
	xmin, xmax := 0.0, 3.0
	ymin, ymax := 0.0, 3.0
	ax.Image(data, core.ImageOptions{
		Colormap: &cmap,
		VMin:     &vmin,
		VMax:     &vmax,
		XMin:     &xmin,
		XMax:     &xmax,
		YMin:     &ymin,
		YMax:     &ymax,
		Origin:   core.ImageOriginLower,
	})

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
