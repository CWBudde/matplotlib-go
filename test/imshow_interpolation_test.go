package test

import (
	"image"
	"testing"

	"matplotlib-go/backends/agg"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

const (
	imshowInterpFigureWidth  = 256
	imshowInterpFigureHeight = 256
)

// renderImshowFiltered renders a small high-frequency pattern with the given
// matplotlib-style interpolation filter, upsampled into a 256x256 figure so
// that filter differences (bilinear vs bicubic vs nearest) are visible.
func renderImshowFiltered(filter string) func() image.Image {
	return func() image.Image {
		const N = 32
		data := make([][]float64, N)
		for j := range data {
			row := make([]float64, N)
			for i := range row {
				// Diagonal-stripe pattern: enough variation that a smoothing
				// filter (bilinear/bicubic) produces visibly different output
				// than nearest-neighbour.
				if (i+j)%2 == 0 {
					row[i] = 1.0
				} else {
					row[i] = 0.0
				}
			}
			data[j] = row
		}

		fig := core.NewFigure(imshowInterpFigureWidth, imshowInterpFigureHeight)
		ax := fig.AddAxes(geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 1, Y: 1},
		})

		cmap := "gray"
		vmin, vmax := 0.0, 1.0
		xmin, xmax := 0.0, float64(N)
		ymin, ymax := 0.0, float64(N)
		ax.Image(data, core.ImageOptions{
			Colormap:      &cmap,
			VMin:          &vmin,
			VMax:          &vmax,
			XMin:          &xmin,
			XMax:          &xmax,
			YMin:          &ymin,
			YMax:          &ymax,
			Origin:        core.ImageOriginLower,
			Interpolation: &filter,
		})

		r, err := agg.New(imshowInterpFigureWidth, imshowInterpFigureHeight, render.Color{R: 1, G: 1, B: 1, A: 1})
		if err != nil {
			panic(err)
		}
		core.DrawFigure(fig, r)
		return r.GetImage()
	}
}

func TestImshowBilinear_Golden(t *testing.T) {
	runGoldenTest(t, "imshow_bilinear", renderImshowFiltered("bilinear"))
}

func TestImshowBicubic_Golden(t *testing.T) {
	runGoldenTest(t, "imshow_bicubic", renderImshowFiltered("bicubic"))
}
