package dashes

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/examples/lines/dashes"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := dashes.Dashes()
	r, err := agg.New(dashes.Width, dashes.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
