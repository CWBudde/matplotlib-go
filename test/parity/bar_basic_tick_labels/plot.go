// Package bar_basic_tick_labels adds tick labels (numbers under the marks)
// on top of the bar_basic_ticks scaffold.
package bar_basic_tick_labels

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	return common.PlotBarBasicScaffold(true, true, false)
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	return common.RenderBarBasicScaffold(true, true, false)
}
