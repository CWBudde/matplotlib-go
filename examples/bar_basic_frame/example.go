package bar_basic_frame

import (
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"image"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(false, false, false)
}
