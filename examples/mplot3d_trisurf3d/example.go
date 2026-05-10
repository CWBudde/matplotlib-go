package mplot3d_trisurf3d

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
		nRadii  = 8
		nAngles = 36
	)
	radii := make([]float64, nRadii)
	angles := make([]float64, nAngles)
	for i := range nRadii {
		radii[i] = 0.125 + (float64(i)/float64(nRadii-1))*(1.0-0.125)
	}
	for i := range nAngles {
		angles[i] = 2 * math.Pi * float64(i) / float64(nAngles)
	}

	x := make([]float64, 1+nRadii*nAngles)
	y := make([]float64, len(x))
	z := make([]float64, len(x))
	x[0], y[0], z[0] = 0, 0, 0

	index := 1
	for _, angle := range angles {
		for _, radius := range radii {
			x[index] = radius * math.Cos(angle)
			y[index] = radius * math.Sin(angle)
			z[index] = math.Sin(-x[index] * y[index])
			index++
		}
	}

	triangles := make([][3]int, 0, (nRadii-1)*nAngles*2)
	for ring := 0; ring < nRadii-1; ring++ {
		for angleIdx := 0; angleIdx < nAngles; angleIdx++ {
			next := (angleIdx + 1) % nAngles
			a := 1 + ring*nAngles + angleIdx
			b := 1 + (ring+1)*nAngles + angleIdx
			c := 1 + (ring+1)*nAngles + next
			d := 1 + ring*nAngles + next
			triangles = append(triangles, [3]int{a, b, c})
			triangles = append(triangles, [3]int{a, c, d})
		}
	}

	cmap := "Blues"
	vmin := 2 * common.MinInSlice(z)
	ax.Trisurf(core.Triangulation{X: x, Y: y, Triangles: triangles}, z, core.PlotOptions{
		Colormap: &cmap,
		VMin:     &vmin,
	})
	common.DisableMplot3DTickLabels(ax)

	r, err := agg.New(720, 560, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
