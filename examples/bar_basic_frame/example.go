package bar_basic_frame

import (
	"image"

	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(false, false, false)
}
