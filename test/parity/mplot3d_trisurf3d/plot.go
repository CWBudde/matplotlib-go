package mplot3d_trisurf3d

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)


const (
	Width  = 720
	Height = 560
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
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

	// Points are laid out as: 1 + angleIdx*nRadii + ring
	triangles := make([][3]int, 0, nAngles+((nRadii-1)*nAngles*2))
	// Center fan: connect origin to innermost ring
	for angleIdx := 0; angleIdx < nAngles; angleIdx++ {
		next := (angleIdx + 1) % nAngles
		triangles = append(triangles, [3]int{0, 1 + angleIdx*nRadii, 1 + next*nRadii})
	}
	// Adjacent rings
	for ring := 0; ring < nRadii-1; ring++ {
		for angleIdx := 0; angleIdx < nAngles; angleIdx++ {
			next := (angleIdx + 1) % nAngles
			a := 1 + angleIdx*nRadii + ring
			b := 1 + angleIdx*nRadii + (ring + 1)
			c := 1 + next*nRadii + (ring + 1)
			d := 1 + next*nRadii + ring
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
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
