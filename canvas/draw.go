package canvas

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

// DrawFigure renders a figure through a renderer.
func DrawFigure(fig *Figure, r render.Renderer) {
	if fig == nil || r == nil {
		return
	}
	core.DrawFigure(fig, r)
}
