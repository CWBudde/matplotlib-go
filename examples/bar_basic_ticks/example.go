package bar_basic_ticks

import (
	"image"

	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(true, false, false)
}
