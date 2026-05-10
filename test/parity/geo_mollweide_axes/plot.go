package geo_mollweide_axes

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/test/parity/internal/common"
)

func Render() image.Image {
	return common.RenderGeoProjectionAxes("mollweide", "Mollweide Projection", -math.Pi, math.Pi)
}
