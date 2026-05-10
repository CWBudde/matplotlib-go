// Package geo_hammer_axes is a showcase of the Hammer (equal-area) projection
// with a sinusoidal latitude trace.
package geo_hammer_axes

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	return common.PlotGeoProjectionAxes("hammer", "Hammer Projection", -math.Pi, math.Pi)
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	return common.RenderGeoProjectionAxes("hammer", "Hammer Projection", -math.Pi, math.Pi)
}
