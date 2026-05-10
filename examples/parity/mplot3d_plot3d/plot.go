package mplot3d_plot3d

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := core.NewFigure(720, 560)
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		panic(err)
	}

	const n = 100
	x := make([]float64, n)
	y := make([]float64, n)
	z := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(n-1)
		x[i] = t
		y[i] = math.Sin(6 * math.Pi * t)
		z[i] = math.Cos(6 * math.Pi * t)
	}
	ax.Plot3D(x, y, z)

	r, err := agg.New(720, 560, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
