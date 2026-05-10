package arrays_showcase

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	arraysshowcase "github.com/cwbudde/matplotlib-go/examples/arrays/showcase"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := arraysshowcase.ArraysShowcase()
	r, err := agg.New(arraysshowcase.Width, arraysshowcase.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(arraysshowcase.DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}
