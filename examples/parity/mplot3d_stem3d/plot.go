package mplot3d_stem3d

import (
	"github.com/cwbudde/matplotlib-go/examples/parity/internal/common"
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

	const n = 20
	x := make([]float64, n)
	y := make([]float64, n)
	z := make([]float64, n)
	for i := range n {
		t := 2 * math.Pi * float64(i) / float64(n-1)
		x[i] = math.Sin(t)
		y[i] = math.Cos(t)
		z[i] = float64(i) / float64(n-1)
	}
	ax.Stem(x, y, z)
	common.DisableMplot3DTickLabels(ax)

	r, err := agg.New(720, 560, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
