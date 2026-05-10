package bar_basic

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
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Basic Bars")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 10)

	bar := &core.Bar2D{
		X:           []float64{1, 2, 3, 4, 5},
		Heights:     []float64{3, 7, 2, 8, 5},
		Width:       0.6,
		Color:       render.Color{R: 0.2, G: 0.6, B: 0.8, A: 1},
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(bar)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
