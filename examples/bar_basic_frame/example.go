// Package bar_basic_frame is a showcase of the bare bar-chart frame: axes
// rectangle only, no ticks, tick labels, or title.
package bar_basic_frame

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	return common.PlotBarBasicScaffold(false, false, false)
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	return common.RenderBarBasicScaffold(false, false, false)
}
