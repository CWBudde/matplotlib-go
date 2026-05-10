package mplot3d_voxels

import (
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"image"

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

	const n = 8
	filled := make([][][]bool, n)
	for i := 0; i < n; i++ {
		filled[i] = make([][]bool, n)
		for j := 0; j < n; j++ {
			filled[i][j] = make([]bool, n)
		}
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			for k := 0; k < n; k++ {
				if (i < 3 && j < 3 && k < 3) || (i >= 5 && j >= 5 && k >= 5) {
					filled[i][j][k] = true
				}
			}
		}
	}

	edgeColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	ax.Voxels(filled, core.VoxelOptions{
		EdgeColor: &edgeColor,
	})
	common.DisableMplot3DTickLabels(ax)

	r, err := agg.New(720, 560, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
