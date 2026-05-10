package axes_grid1_showcase

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	axesgrid1showcase "github.com/cwbudde/matplotlib-go/examples/axes_grid1/showcase"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := axesgrid1showcase.AxesGrid1Showcase()
	r, err := agg.New(axesgrid1showcase.Width, axesgrid1showcase.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(axesgrid1showcase.DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}
