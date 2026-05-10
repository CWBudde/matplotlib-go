package geo_hammer_axes

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/test/parity/internal/common"
)

func Render() image.Image {
	return common.RenderGeoProjectionAxes("hammer", "Hammer Projection", -math.Pi, math.Pi)
}
