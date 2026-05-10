// Package bar_basic_title finishes the bar scaffold progression by adding
// the figure title on top of bar_basic_tick_labels.
package bar_basic_title

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	return common.PlotBarBasicScaffold(true, true, true)
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	return common.RenderBarBasicScaffold(true, true, true)
}
