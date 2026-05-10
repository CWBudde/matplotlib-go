package mplot3d_wire3d

import (
	"github.com/cwbudde/matplotlib-go/test/parity/internal/common"
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

	x, y, z := common.Get3DWireframeTestData(0.05)

	rStride := 10
	cStride := 10
	ax.Wireframe(x, y, z, core.PlotOptions{
		RStride: &rStride,
		CStride: &cStride,
	})
	common.DisableMplot3DTickLabels(ax)

	r, err := agg.New(720, 560, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
