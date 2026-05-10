package bar_basic_ticks

import (
	"github.com/cwbudde/matplotlib-go/test/parity/internal/common"
	"image"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(true, false, false)
}
