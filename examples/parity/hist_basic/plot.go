package hist_basic

import (
	"github.com/cwbudde/matplotlib-go/examples/parity/internal/common"
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	ax.SetTitle("Basic Histogram")

	data := common.NormalData(42, 0, 500, 5.0, 1.5)

	blue := render.Color{R: 0.26, G: 0.53, B: 0.80, A: 0.8}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 0.8
	ax.Hist(data, core.HistOptions{
		Color:     &blue,
		EdgeColor: &black,
		EdgeWidth: &ew,
	})
	ax.AutoScale(0.05)
	_, yMax := ax.YScale.Domain()
	ax.SetYLim(0, yMax)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
