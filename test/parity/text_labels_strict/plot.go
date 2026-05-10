package text_labels_strict

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/test/parity/internal/common"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func Render() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	ax.SetTitle("Text Labels")
	ax.SetXLabel("Group")
	ax.SetYLabel("Value")
	return common.RenderFixtureFigure(fig, 640, 360)
}
