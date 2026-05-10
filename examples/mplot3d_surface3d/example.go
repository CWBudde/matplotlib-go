package mplot3d_surface3d

import (
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
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

	const (
		xMin = -5.0
		xMax = 5.0
		step = 0.25
	)
	xCount := int((xMax - xMin) / step)
	x := make([]float64, xCount)
	y := make([]float64, xCount)
	z := make([][]float64, xCount)
	for yi := range xCount {
		y[yi] = xMin + step*float64(yi)
		z[yi] = make([]float64, xCount)
	}
	for xi := range xCount {
		x[xi] = xMin + step*float64(xi)
	}
	for yi := 0; yi < xCount; yi++ {
		for xi := 0; xi < xCount; xi++ {
			r := math.Hypot(x[xi], y[yi])
			z[yi][xi] = math.Sin(r)
		}
	}

	vmin := 2 * common.MinInGrid(z)
	cmap := "Blues"
	common.DisableMplot3DTickLabels(ax)
	ax.Surface(x, y, z, core.PlotOptions{
		VMin:     &vmin,
		Colormap: &cmap,
	})

	r, err := agg.New(720, 560, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
