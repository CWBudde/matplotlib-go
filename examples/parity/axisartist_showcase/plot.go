package axisartist_showcase

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	axisartistshowcase "github.com/cwbudde/matplotlib-go/examples/axisartist/showcase"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := axisartistshowcase.AxisArtistShowcase()
	r, err := agg.New(axisartistshowcase.Width, axisartistshowcase.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(axisartistshowcase.DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}
