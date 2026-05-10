package unstructured_showcase

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	unstructuredshowcase "github.com/cwbudde/matplotlib-go/examples/unstructured/showcase"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := unstructuredshowcase.UnstructuredShowcase()
	r, err := agg.New(unstructuredshowcase.Width, unstructuredshowcase.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(unstructuredshowcase.DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}
