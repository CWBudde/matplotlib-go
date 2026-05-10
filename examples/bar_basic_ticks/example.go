package bar_basic_ticks

import (
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"image"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(true, false, false)
}
