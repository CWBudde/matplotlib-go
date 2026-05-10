// Package geo_aitoff_axes is a showcase of the Aitoff (equal-area) projection
// with a sinusoidal latitude trace.
package geo_aitoff_axes

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	return common.PlotGeoProjectionAxes("aitoff", "Aitoff Projection", -math.Pi, math.Pi)
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	return common.RenderGeoProjectionAxes("aitoff", "Aitoff Projection", -math.Pi, math.Pi)
}
