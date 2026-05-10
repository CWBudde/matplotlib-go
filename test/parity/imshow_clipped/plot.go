package imshow_clipped

import (
	"github.com/cwbudde/matplotlib-go/test/parity/internal/common"
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func Render() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Clipped Imshow")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	cmap := "viridis"
	nearest := "nearest"
	extent := [4]float64{0, 8, 0, 8}
	ax.ImShow(common.WaveImageData(8, 8), core.ImShowOptions{
		Colormap:      &cmap,
		Extent:        &extent,
		Origin:        core.ImageOriginLower,
		Aspect:        "auto",
		Interpolation: &nearest,
	})
	ax.SetXLim(2, 6)
	ax.SetYLim(1, 7)

	return common.RenderImageFixture(fig, 640, 360)
}
