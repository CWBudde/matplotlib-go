package boxplot_basic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	boxplotbasic "github.com/cwbudde/matplotlib-go/examples/boxplot/basic"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := boxplotbasic.Build()
	r, err := agg.New(boxplotbasic.Width, boxplotbasic.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(boxplotbasic.DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}
