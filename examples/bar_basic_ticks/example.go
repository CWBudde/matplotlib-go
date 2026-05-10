// Package bar_basic_ticks adds tick marks (no labels) on top of the bar
// scaffold from bar_basic_frame.
package bar_basic_ticks

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	return common.PlotBarBasicScaffold(true, false, false)
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	return common.RenderBarBasicScaffold(true, false, false)
}
