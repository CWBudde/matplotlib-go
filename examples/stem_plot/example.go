package stem_plot

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := core.NewFigure(720, 420)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.16}, Max: geom.Pt{X: 0.94, Y: 0.86}})
	ax.SetTitle("Stem")
	ax.SetXLabel("Sample")
	ax.SetYLabel("Amplitude")
	ax.SetXLim(0.5, 7.5)
	ax.SetYLim(-0.2, 4.2)
	ax.AddYGrid()
	stemColor := render.Color{R: 0.15, G: 0.42, B: 0.73, A: 1}
	baseline := 0.3
	markerSize := 7.0
	ax.Stem(
		[]float64{1, 2, 3, 4, 5, 6, 7},
		[]float64{0.9, 2.2, 1.6, 3.3, 2.4, 3.7, 2.1},
		core.StemOptions{
			Color:         &stemColor,
			Baseline:      &baseline,
			MarkerSize:    &markerSize,
			Label:         "samples",
			BaselineColor: &render.Color{R: 0.32, G: 0.32, B: 0.32, A: 1},
		},
	)

	r, err := agg.New(720, 420, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
