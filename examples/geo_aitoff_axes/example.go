package geo_aitoff_axes

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Render() image.Image {
	return common.RenderGeoProjectionAxes("aitoff", "Aitoff Projection", -math.Pi, math.Pi)
}
